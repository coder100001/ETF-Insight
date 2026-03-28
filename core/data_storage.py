"""
数据存储模块
处理 ETF 数据的存储、更新和查询
"""

import pandas as pd
from datetime import datetime, date
from typing import Optional, List, Dict, Any
import logging
import os

logger = logging.getLogger(__name__)


class DataStorage:
    """数据存储基类"""
    
    def save_historical_data(self, symbol: str, data: pd.DataFrame) -> bool:
        """保存历史数据"""
        raise NotImplementedError
    
    def get_latest_date(self, symbol: str) -> Optional[date]:
        """获取最新数据日期"""
        raise NotImplementedError
    
    def get_historical_data(
        self,
        symbol: str,
        start_date: Optional[date] = None,
        end_date: Optional[date] = None
    ) -> Optional[pd.DataFrame]:
        """获取历史数据"""
        raise NotImplementedError


class DjangoDBStorage(DataStorage):
    """
    Django ORM 数据存储
    用于与 workflow/models.py 集成
    """
    
    def __init__(self):
        self._models = None
        self._batch_size = 1000
    
    def _get_models(self):
        """延迟导入模型"""
        if self._models is None:
            try:
                from workflow.models import ETFData
                self._models = {'ETFData': ETFData}
            except ImportError as e:
                logger.error(f"导入 Django 模型失败: {e}")
                raise
        return self._models
    
    def save_historical_data(self, symbol: str, data: pd.DataFrame) -> bool:
        """
        保存历史数据到数据库
        
        Args:
            symbol: ETF 代码
            data: DataFrame 数据
        
        Returns:
            是否成功
        """
        if data is None or data.empty:
            logger.warning(f"{symbol} 数据为空，跳过保存")
            return False
        
        try:
            models = self._get_models()
            ETFData = models['ETFData']
            
            records_to_create = []
            records_to_update = []
            
            for index, row in data.iterrows():
                record_date = index.date() if hasattr(index, 'date') else index
                
                # 检查记录是否已存在
                existing = ETFData.objects.filter(
                    symbol=symbol,
                    date=record_date
                ).first()
                
                record_data = {
                    'open_price': float(row.get('open', row.get('Open', 0))),
                    'high_price': float(row.get('high', row.get('High', 0))),
                    'low_price': float(row.get('low', row.get('Low', 0))),
                    'close_price': float(row.get('close', row.get('Close', 0))),
                    'volume': int(row.get('volume', row.get('Volume', 0))),
                }
                
                if existing:
                    # 更新现有记录
                    for key, value in record_data.items():
                        setattr(existing, key, value)
                    records_to_update.append(existing)
                else:
                    # 创建新记录
                    records_to_create.append(ETFData(
                        symbol=symbol,
                        date=record_date,
                        **record_data
                    ))
            
            # 批量更新
            if records_to_update:
                from django.db import transaction
                with transaction.atomic():
                    for record in records_to_update:
                        record.save()
                logger.info(f"{symbol} 更新 {len(records_to_update)} 条记录")
            
            # 批量创建
            if records_to_create:
                ETFData.objects.bulk_create(
                    records_to_create,
                    batch_size=self._batch_size,
                    ignore_conflicts=True
                )
                logger.info(f"{symbol} 创建 {len(records_to_create)} 条记录")
            
            return True
            
        except Exception as e:
            logger.error(f"保存 {symbol} 数据失败: {e}")
            return False
    
    def get_latest_date(self, symbol: str) -> Optional[date]:
        """获取最新数据日期"""
        try:
            models = self._get_models()
            ETFData = models['ETFData']
            
            latest = ETFData.objects.filter(symbol=symbol).order_by('-date').first()
            if latest:
                return latest.date
            return None
            
        except Exception as e:
            logger.error(f"获取 {symbol} 最新日期失败: {e}")
            return None
    
    def get_historical_data(
        self,
        symbol: str,
        start_date: Optional[date] = None,
        end_date: Optional[date] = None
    ) -> Optional[pd.DataFrame]:
        """获取历史数据"""
        try:
            models = self._get_models()
            ETFData = models['ETFData']
            
            queryset = ETFData.objects.filter(symbol=symbol)
            
            if start_date:
                queryset = queryset.filter(date__gte=start_date)
            if end_date:
                queryset = queryset.filter(date__lte=end_date)
            
            data = list(queryset.order_by('date').values())
            
            if not data:
                return None
            
            df = pd.DataFrame(data)
            df.set_index('date', inplace=True)
            df.index = pd.to_datetime(df.index)
            
            return df
            
        except Exception as e:
            logger.error(f"获取 {symbol} 历史数据失败: {e}")
            return None


class CSVStorage(DataStorage):
    """CSV 文件存储（用于备份和导出）"""
    
    def __init__(self, data_dir: str = 'etf_data'):
        self.data_dir = data_dir
        os.makedirs(data_dir, exist_ok=True)
    
    def save_historical_data(self, symbol: str, data: pd.DataFrame) -> bool:
        """保存到 CSV"""
        if data is None or data.empty:
            return False
        
        try:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            filename = f"{symbol}_{timestamp}.csv"
            filepath = os.path.join(self.data_dir, filename)
            
            data.to_csv(filepath)
            logger.info(f"数据已保存到: {filepath}")
            return True
            
        except Exception as e:
            logger.error(f"保存 CSV 失败: {e}")
            return False
    
    def get_latest_date(self, symbol: str) -> Optional[date]:
        """从 CSV 获取最新日期"""
        # 简化实现，实际项目中可能需要更复杂的逻辑
        return None
    
    def get_historical_data(
        self,
        symbol: str,
        start_date: Optional[date] = None,
        end_date: Optional[date] = None
    ) -> Optional[pd.DataFrame]:
        """从 CSV 读取历史数据"""
        # 简化实现
        return None
    
    def save_to_excel(
        self,
        data_dict: Dict[str, pd.DataFrame],
        filename: Optional[str] = None
    ) -> str:
        """
        保存多个数据到 Excel
        
        Args:
            data_dict: {sheet_name: DataFrame}
            filename: 文件名
        
        Returns:
            文件路径
        """
        if filename is None:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            filename = f"ETF_Data_{timestamp}.xlsx"
        
        filepath = os.path.join(self.data_dir, filename)
        
        with pd.ExcelWriter(filepath, engine='openpyxl') as writer:
            for sheet_name, df in data_dict.items():
                if df is not None and not df.empty:
                    # 移除时区信息
                    df_copy = df.copy()
                    if isinstance(df_copy.index, pd.DatetimeIndex):
                        df_copy.index = df_copy.index.tz_localize(None)
                    
                    for col in df_copy.columns:
                        if pd.api.types.is_datetime64tz_dtype(df_copy[col]):
                            df_copy[col] = df_copy[col].dt.tz_localize(None)
                    
                    df_copy.to_excel(writer, sheet_name=sheet_name)
        
        logger.info(f"Excel 已保存到: {filepath}")
        return filepath


class HybridStorage(DataStorage):
    """
    混合存储策略
    主存储：Django ORM
    备份存储：CSV
    """
    
    def __init__(self, data_dir: str = 'etf_data'):
        self.db_storage = DjangoDBStorage()
        self.csv_storage = CSVStorage(data_dir)
    
    def save_historical_data(self, symbol: str, data: pd.DataFrame, backup: bool = True) -> bool:
        """
        保存数据到数据库，可选 CSV 备份
        
        Args:
            symbol: ETF 代码
            data: DataFrame 数据
            backup: 是否备份到 CSV
        """
        success = self.db_storage.save_historical_data(symbol, data)
        
        if success and backup:
            self.csv_storage.save_historical_data(symbol, data)
        
        return success
    
    def get_latest_date(self, symbol: str) -> Optional[date]:
        return self.db_storage.get_latest_date(symbol)
    
    def get_historical_data(
        self,
        symbol: str,
        start_date: Optional[date] = None,
        end_date: Optional[date] = None
    ) -> Optional[pd.DataFrame]:
        return self.db_storage.get_historical_data(symbol, start_date, end_date)
    
    def export_to_excel(self, symbols: List[str], filename: Optional[str] = None) -> str:
        """
        导出指定 ETF 数据到 Excel
        
        Args:
            symbols: ETF 代码列表
            filename: 文件名
        
        Returns:
            文件路径
        """
        data_dict = {}
        for symbol in symbols:
            data = self.get_historical_data(symbol)
            if data is not None:
                data_dict[symbol] = data
        
        return self.csv_storage.save_to_excel(data_dict, filename)


# 全局单例
_hybrid_storage = None


def get_storage() -> HybridStorage:
    """获取 HybridStorage 单例"""
    global _hybrid_storage
    if _hybrid_storage is None:
        _hybrid_storage = HybridStorage()
    return _hybrid_storage

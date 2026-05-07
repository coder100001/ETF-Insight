#!/usr/bin/env python3
"""
AKShare数据服务
提供HTTP API接口供Go后端调用
"""

import math

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
import akshare as ak
import pandas as pd
from datetime import datetime, timedelta
from typing import List, Dict, Optional
import uvicorn
import os

app = FastAPI(title="AKShare Data Service", version="2.6.0")

allowed_origins = os.getenv("AKSHARE_CORS_ORIGINS", "http://localhost:5173,http://localhost:8080").split(",")
app.add_middleware(
    CORSMiddleware,
    allow_origins=allowed_origins,
    allow_credentials=True,
    allow_methods=["GET", "POST", "OPTIONS"],
    allow_headers=["Content-Type", "Authorization"],
)


@app.get("/")
def root():
    return {"message": "AKShare Data Service", "status": "running"}


@app.get("/api/etf/list")
def get_etf_list():
    """获取A股ETF列表"""
    try:
        # 获取ETF基金列表
        etf_df = ak.fund_etf_spot_em()

        etf_list = []
        for _, row in etf_df.iterrows():
            def safe_float(val, default=0):
                try:
                    v = float(val)
                    return default if math.isnan(v) else v
                except (ValueError, TypeError):
                    return default

            def safe_int(val, default=0):
                try:
                    fv = float(val)
                    if math.isnan(fv):
                        return default
                    return int(fv)
                except (ValueError, TypeError):
                    return default

            etf_list.append({
                "symbol": str(row.get("代码", "")).strip(),
                "name": str(row.get("名称", "")).strip(),
                "full_name": str(row.get("名称", "")).strip(),
                "exchange": "SSE" if str(row.get("代码", "")).startswith("51") else "SZSE",
                "current_price": safe_float(row.get("最新价")),
                "previous_close": safe_float(row.get("昨收")),
                "price_change": safe_float(row.get("涨跌额")),
                "price_change_pct": safe_float(row.get("涨跌幅")),
                "volume": safe_int(row.get("成交量")),
                "turnover": safe_float(row.get("成交额")),
            })

        return {
            "success": True,
            "data": etf_list,
            "count": len(etf_list)
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


@app.get("/api/etf/quote/{symbol}")
def get_etf_quote(symbol: str):
    """获取单个ETF实时行情"""
    try:
        # 获取ETF实时行情
        etf_df = ak.fund_etf_spot_em()
        etf_row = etf_df[etf_df["代码"] == symbol]

        if etf_row.empty:
            raise HTTPException(status_code=404, detail=f"ETF {symbol} not found")

        row = etf_row.iloc[0]
        quote = {
            "symbol": symbol,
            "name": row.get("名称", ""),
            "current_price": float(row.get("最新价", 0) or 0),
            "previous_close": float(row.get("昨收", 0) or 0),
            "open": float(row.get("今开", 0) or 0),
            "high": float(row.get("最高", 0) or 0),
            "low": float(row.get("最低", 0) or 0),
            "volume": int(row.get("成交量", 0) or 0),
            "turnover": float(row.get("成交额", 0) or 0),
            "price_change": float(row.get("涨跌额", 0) or 0),
            "price_change_pct": float(row.get("涨跌幅", 0) or 0),
            "bid_price": float(row.get("买一", 0) or 0),
            "ask_price": float(row.get("卖一", 0) or 0),
            "bid_volume": int(row.get("买一手", 0) or 0),
            "ask_volume": int(row.get("卖一手", 0) or 0),
            "update_time": datetime.now().isoformat()
        }

        return {
            "success": True,
            "data": quote
        }
    except HTTPException:
        raise
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


@app.get("/api/etf/quotes")
def get_etf_quotes(symbols: List[str]):
    """批量获取ETF行情"""
    try:
        etf_df = ak.fund_etf_spot_em()
        quotes = {}

        for symbol in symbols:
            etf_row = etf_df[etf_df["代码"] == symbol]
            if not etf_row.empty:
                row = etf_row.iloc[0]
                quotes[symbol] = {
                    "symbol": symbol,
                    "name": row.get("名称", ""),
                    "current_price": float(row.get("最新价", 0) or 0),
                    "previous_close": float(row.get("昨收", 0) or 0),
                    "price_change": float(row.get("涨跌额", 0) or 0),
                    "price_change_pct": float(row.get("涨跌幅", 0) or 0),
                    "volume": int(row.get("成交量", 0) or 0),
                    "turnover": float(row.get("成交额", 0) or 0),
                    "update_time": datetime.now().isoformat()
                }

        return {
            "success": True,
            "data": quotes
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


@app.get("/api/etf/historical/{symbol}")
def get_historical_data(
    symbol: str,
    start_date: Optional[str] = None,
    end_date: Optional[str] = None
):
    """获取ETF历史数据"""
    try:
        # 默认获取最近一年数据
        if not end_date:
            end_date = datetime.now().strftime("%Y%m%d")
        if not start_date:
            start = datetime.strptime(end_date, "%Y%m%d") - timedelta(days=365)
            start_date = start.strftime("%Y%m%d")

        # 获取历史数据
        hist_df = ak.fund_etf_hist_em(
            symbol=symbol,
            period="daily",
            start_date=start_date,
            end_date=end_date,
            adjust="qfq"  # 前复权
        )

        data = []
        for _, row in hist_df.iterrows():
            data.append({
                "symbol": symbol,
                "date": row.get("日期", ""),
                "open": float(row.get("开盘", 0) or 0),
                "high": float(row.get("最高", 0) or 0),
                "low": float(row.get("最低", 0) or 0),
                "close": float(row.get("收盘", 0) or 0),
                "volume": int(row.get("成交量", 0) or 0),
                "turnover": float(row.get("成交额", 0) or 0),
                "amount": float(row.get("成交额", 0) or 0)
            })

        return {
            "success": True,
            "data": data,
            "count": len(data)
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


@app.get("/api/etf/dividends/{symbol}")
def get_dividend_history(symbol: str):
    """获取ETF分红历史"""
    try:
        # 获取基金分红数据
        div_df = ak.fund_etf_dividend_em()
        div_rows = div_df[div_df["基金代码"] == symbol]

        dividends = []
        for _, row in div_rows.iterrows():
            dividends.append({
                "symbol": symbol,
                "ex_dividend_date": row.get("除息日", ""),
                "dividend_per_share": float(row.get("每份分红", 0) or 0),
                "dividend_yield": float(row.get("分红率", 0) or 0)
            })

        return {
            "success": True,
            "data": dividends
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


@app.get("/api/etf/fund_holdings/{symbol}")
def get_fund_holdings(symbol: str):
    """获取ETF持仓明细"""
    try:
        # 获取ETF持仓
        holdings_df = ak.fund_etf_hold_em(symbol=symbol)

        holdings = []
        for _, row in holdings_df.head(10).iterrows():
            holdings.append({
                "stock_code": row.get("股票代码", ""),
                "stock_name": row.get("股票名称", ""),
                "weight": float(row.get("持仓占比", 0) or 0),
                "shares": int(row.get("持股数", 0) or 0),
                "market_value": float(row.get("持仓市值", 0) or 0)
            })

        return {
            "success": True,
            "data": holdings
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


@app.get("/api/market/index_components/{index_code}")
def get_index_components(index_code: str):
    """获取指数成分股"""
    try:
        # 获取指数成分股
        components_df = ak.index_stock_cons_csindex(symbol=index_code)

        components = []
        for _, row in components_df.iterrows():
            components.append({
                "stock_code": row.get("成分券代码", ""),
                "stock_name": row.get("成分券名称", ""),
                "weight": float(row.get("权重", 0) or 0)
            })

        return {
            "success": True,
            "data": components
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


@app.get("/api/market/sector_performance")
def get_sector_performance():
    """获取板块表现"""
    try:
        # 获取行业板块涨跌幅
        sector_df = ak.stock_sector_spot_em()

        sectors = []
        for _, row in sector_df.head(20).iterrows():
            sectors.append({
                "sector_name": row.get("板块", ""),
                "change_pct": float(row.get("涨跌幅", 0) or 0),
                "total_volume": float(row.get("总成交量", 0) or 0),
                "total_amount": float(row.get("总成交额", 0) or 0),
                "leading_stock": row.get("领涨股", "")
            })

        return {
            "success": True,
            "data": sectors
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)

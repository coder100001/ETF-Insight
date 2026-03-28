"""
ETF 基础数据层 (Data Layer)
包含 ETF 基础信息、持仓数据、实时/历史行情等核心数据模型
"""

from django.db import models
from django.utils import timezone


class ETFBaseInfo(models.Model):
    """
    ETF 基础信息表
    存储 ETF 的基本资料、发行方信息、费率等
    """
    MARKET_CHOICES = [
        ('US', '美股'),
        ('CN', 'A股'),
        ('HK', '港股'),
        ('JP', '日股'),
        ('EU', '欧股'),
    ]

    ASSET_CLASS_CHOICES = [
        ('EQUITY', '股票型'),
        ('BOND', '债券型'),
        ('COMMODITY', '商品型'),
        ('CURRENCY', '货币型'),
        ('ALTERNATIVE', '另类投资'),
        ('MULTI_ASSET', '多资产'),
    ]

    symbol = models.CharField(max_length=20, unique=True, verbose_name='ETF代码')
    name = models.CharField(max_length=200, verbose_name='ETF名称')
    name_en = models.CharField(max_length=200, null=True, blank=True, verbose_name='英文名称')

    # 基础信息
    market = models.CharField(max_length=10, choices=MARKET_CHOICES, verbose_name='市场')
    asset_class = models.CharField(max_length=20, choices=ASSET_CLASS_CHOICES, default='EQUITY', verbose_name='资产类别')
    category = models.CharField(max_length=50, null=True, blank=True, verbose_name='分类')

    # 发行方信息
    issuer = models.CharField(max_length=100, verbose_name='发行方')
    issuer_website = models.URLField(null=True, blank=True, verbose_name='发行方官网')

    # 跟踪信息
    tracking_index = models.CharField(max_length=200, null=True, blank=True, verbose_name='跟踪指数')
    tracking_index_symbol = models.CharField(max_length=50, null=True, blank=True, verbose_name='指数代码')
    tracking_method = models.CharField(max_length=50, null=True, blank=True, verbose_name='跟踪方式')  # 完全复制/抽样复制/合成复制

    # 时间与规模
    inception_date = models.DateField(null=True, blank=True, verbose_name='成立日期')
    aum = models.DecimalField(max_digits=20, decimal_places=2, null=True, blank=True, verbose_name='资产管理规模(AUM)')
    aum_currency = models.CharField(max_length=10, default='USD', verbose_name='AUM货币')
    shares_outstanding = models.BigIntegerField(null=True, blank=True, verbose_name='流通股数')

    # 费率信息
    expense_ratio = models.DecimalField(max_digits=6, decimal_places=4, null=True, blank=True, verbose_name='管理费率(%)')
    management_fee = models.DecimalField(max_digits=6, decimal_places=4, null=True, blank=True, verbose_name='管理费(%)')
    other_expenses = models.DecimalField(max_digits=6, decimal_places=4, null=True, blank=True, verbose_name='其他费用(%)')

    # 交易信息
    listing_exchange = models.CharField(max_length=50, null=True, blank=True, verbose_name='上市交易所')
    trading_currency = models.CharField(max_length=10, default='USD', verbose_name='交易货币')
    lot_size = models.IntegerField(default=1, verbose_name='每手股数')
    is_leveraged = models.BooleanField(default=False, verbose_name='是否杠杆')
    leverage_ratio = models.DecimalField(max_digits=4, decimal_places=2, null=True, blank=True, verbose_name='杠杆倍数')
    is_inverse = models.BooleanField(default=False, verbose_name='是否反向')

    # 策略信息
    investment_strategy = models.TextField(null=True, blank=True, verbose_name='投资策略')
    investment_objective = models.TextField(null=True, blank=True, verbose_name='投资目标')
    benchmark = models.CharField(max_length=200, null=True, blank=True, verbose_name='业绩基准')

    # 状态
    status = models.IntegerField(choices=[(0, '禁用'), (1, '启用')], default=1, verbose_name='状态')
    sort_order = models.IntegerField(default=0, verbose_name='排序')

    # 元数据
    data_source = models.CharField(max_length=50, null=True, blank=True, verbose_name='数据来源')
    last_updated = models.DateTimeField(auto_now=True, verbose_name='最后更新')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')

    class Meta:
        db_table = 'etf_base_info'
        verbose_name = 'ETF基础信息'
        verbose_name_plural = 'ETF基础信息'
        ordering = ['market', 'sort_order', 'symbol']
        indexes = [
            models.Index(fields=['symbol']),
            models.Index(fields=['market', 'status']),
            models.Index(fields=['issuer']),
            models.Index(fields=['category']),
        ]

    def __str__(self):
        return f"{self.symbol} - {self.name}"

    @property
    def aum_display(self):
        """格式化显示 AUM"""
        if self.aum:
            if self.aum >= 1_000_000_000:
                return f"{self.aum / 1_000_000_000:.2f}B {self.aum_currency}"
            elif self.aum >= 1_000_000:
                return f"{self.aum / 1_000_000:.2f}M {self.aum_currency}"
        return "N/A"


class ETFPrice(models.Model):
    """
    ETF 价格数据表
    存储实时和历史行情数据（K线）
    """
    INTERVAL_CHOICES = [
        ('1m', '1分钟'),
        ('5m', '5分钟'),
        ('15m', '15分钟'),
        ('30m', '30分钟'),
        ('1h', '1小时'),
        ('1d', '日线'),
        ('1w', '周线'),
        ('1M', '月线'),
    ]

    etf = models.ForeignKey(ETFBaseInfo, on_delete=models.CASCADE, related_name='prices', verbose_name='ETF')
    symbol = models.CharField(max_length=20, verbose_name='ETF代码')

    # 时间信息
    trade_date = models.DateField(verbose_name='交易日期')
    trade_time = models.TimeField(null=True, blank=True, verbose_name='交易时间')
    interval = models.CharField(max_length=10, choices=INTERVAL_CHOICES, default='1d', verbose_name='时间周期')

    # 价格数据
    open_price = models.DecimalField(max_digits=12, decimal_places=4, verbose_name='开盘价')
    high_price = models.DecimalField(max_digits=12, decimal_places=4, verbose_name='最高价')
    low_price = models.DecimalField(max_digits=12, decimal_places=4, verbose_name='最低价')
    close_price = models.DecimalField(max_digits=12, decimal_places=4, verbose_name='收盘价')
    pre_close = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='昨收价')

    # 成交量数据
    volume = models.BigIntegerField(verbose_name='成交量')
    turnover = models.DecimalField(max_digits=20, decimal_places=4, null=True, blank=True, verbose_name='成交额')

    # 衍生指标
    change_amount = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='涨跌额')
    change_percent = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='涨跌幅(%)')
    turnover_rate = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='换手率(%)')

    # 实时数据扩展
    bid_price = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='买入价')
    ask_price = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='卖出价')
    bid_volume = models.BigIntegerField(null=True, blank=True, verbose_name='买入量')
    ask_volume = models.BigIntegerField(null=True, blank=True, verbose_name='卖出量')

    # 数据源
    data_source = models.CharField(max_length=50, null=True, blank=True, verbose_name='数据来源')
    is_adjusted = models.BooleanField(default=False, verbose_name='是否复权')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')

    class Meta:
        db_table = 'etf_price'
        verbose_name = 'ETF价格数据'
        verbose_name_plural = 'ETF价格数据'
        unique_together = [['symbol', 'trade_date', 'trade_time', 'interval']]
        ordering = ['-trade_date', '-trade_time', 'symbol']
        indexes = [
            models.Index(fields=['symbol', 'interval', '-trade_date']),
            models.Index(fields=['trade_date', 'symbol']),
            models.Index(fields=['etf', '-trade_date']),
        ]

    def __str__(self):
        return f"{self.symbol} - {self.trade_date} {self.close_price}"

    def save(self, *args, **kwargs):
        # 自动计算涨跌额和涨跌幅
        if self.pre_close and self.close_price:
            self.change_amount = self.close_price - self.pre_close
            if self.pre_close > 0:
                self.change_percent = (self.change_amount / self.pre_close) * 100
        super().save(*args, **kwargs)


class ETFNav(models.Model):
    """
    ETF 净值数据表
    存储 NAV 和市价偏离数据
    """
    etf = models.ForeignKey(ETFBaseInfo, on_delete=models.CASCADE, related_name='navs', verbose_name='ETF')
    symbol = models.CharField(max_length=20, verbose_name='ETF代码')

    # 日期
    nav_date = models.DateField(verbose_name='净值日期')
    nav_time = models.TimeField(null=True, blank=True, verbose_name='净值时间')

    # NAV 数据
    nav = models.DecimalField(max_digits=12, decimal_places=4, verbose_name='单位净值(NAV)')
    nav_change = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='净值涨跌')
    nav_change_percent = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='净值涨跌幅(%)')

    # 市价数据
    market_price = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='市价')

    # 偏离度
    premium_discount = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='溢价率(%)')
    # 溢价率 = (市价 - NAV) / NAV * 100

    # 数据源
    data_source = models.CharField(max_length=50, null=True, blank=True, verbose_name='数据来源')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')

    class Meta:
        db_table = 'etf_nav'
        verbose_name = 'ETF净值数据'
        verbose_name_plural = 'ETF净值数据'
        unique_together = [['symbol', 'nav_date', 'nav_time']]
        ordering = ['-nav_date', '-nav_time', 'symbol']
        indexes = [
            models.Index(fields=['symbol', '-nav_date']),
            models.Index(fields=['premium_discount']),
        ]

    def __str__(self):
        return f"{self.symbol} - NAV: {self.nav}"

    def save(self, *args, **kwargs):
        # 自动计算溢价率
        if self.nav and self.market_price and self.nav > 0:
            self.premium_discount = ((self.market_price - self.nav) / self.nav) * 100
        super().save(*args, **kwargs)


class ETFHolding(models.Model):
    """
    ETF 持仓数据表
    存储成分股/债券等持仓信息
    """
    ASSET_TYPE_CHOICES = [
        ('STOCK', '股票'),
        ('BOND', '债券'),
        ('FUND', '基金'),
        ('COMMODITY', '商品'),
        ('CASH', '现金'),
        ('OTHER', '其他'),
    ]

    etf = models.ForeignKey(ETFBaseInfo, on_delete=models.CASCADE, related_name='holdings', verbose_name='ETF')
    symbol = models.CharField(max_length=20, verbose_name='ETF代码')

    # 持仓标的
    holding_symbol = models.CharField(max_length=20, verbose_name='持仓代码')
    holding_name = models.CharField(max_length=200, verbose_name='持仓名称')
    asset_type = models.CharField(max_length=20, choices=ASSET_TYPE_CHOICES, default='STOCK', verbose_name='资产类型')

    # 持仓数据
    shares = models.DecimalField(max_digits=20, decimal_places=4, null=True, blank=True, verbose_name='持仓数量')
    market_value = models.DecimalField(max_digits=20, decimal_places=4, null=True, blank=True, verbose_name='持仓市值')
    weight = models.DecimalField(max_digits=8, decimal_places=4, verbose_name='权重(%)')
    weight_change = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='权重变化(%)')

    # 价格信息
    price = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='当前价格')
    price_currency = models.CharField(max_length=10, default='USD', verbose_name='价格货币')

    # 报告期
    report_date = models.DateField(verbose_name='报告日期')
    is_estimated = models.BooleanField(default=False, verbose_name='是否估算')

    # 元数据
    data_source = models.CharField(max_length=50, null=True, blank=True, verbose_name='数据来源')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')

    class Meta:
        db_table = 'etf_holding'
        verbose_name = 'ETF持仓数据'
        verbose_name_plural = 'ETF持仓数据'
        unique_together = [['symbol', 'holding_symbol', 'report_date']]
        ordering = ['-report_date', '-weight', 'symbol']
        indexes = [
            models.Index(fields=['symbol', '-report_date']),
            models.Index(fields=['holding_symbol']),
            models.Index(fields=['asset_type']),
            models.Index(fields=['-weight']),
        ]

    def __str__(self):
        return f"{self.symbol} - {self.holding_symbol} ({self.weight}%)"


class ETFHoldingSector(models.Model):
    """
    ETF 行业分布表
    存储持仓的行业分类统计
    """
    etf = models.ForeignKey(ETFBaseInfo, on_delete=models.CASCADE, related_name='sectors', verbose_name='ETF')
    symbol = models.CharField(max_length=20, verbose_name='ETF代码')

    # 行业信息
    sector_name = models.CharField(max_length=100, verbose_name='行业名称')
    sector_code = models.CharField(max_length=20, null=True, blank=True, verbose_name='行业代码')

    # 分布数据
    weight = models.DecimalField(max_digits=8, decimal_places=4, verbose_name='权重(%)')
    weight_change = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='权重变化(%)')
    market_value = models.DecimalField(max_digits=20, decimal_places=4, null=True, blank=True, verbose_name='市值')
    stock_count = models.IntegerField(null=True, blank=True, verbose_name='股票数量')

    # 报告期
    report_date = models.DateField(verbose_name='报告日期')

    # 元数据
    data_source = models.CharField(max_length=50, null=True, blank=True, verbose_name='数据来源')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')

    class Meta:
        db_table = 'etf_holding_sector'
        verbose_name = 'ETF行业分布'
        verbose_name_plural = 'ETF行业分布'
        unique_together = [['symbol', 'sector_name', 'report_date']]
        ordering = ['-report_date', '-weight', 'symbol']
        indexes = [
            models.Index(fields=['symbol', '-report_date']),
            models.Index(fields=['sector_name']),
        ]

    def __str__(self):
        return f"{self.symbol} - {self.sector_name} ({self.weight}%)"


class ETFHoldingRegion(models.Model):
    """
    ETF 地区分布表
    存储持仓的地区分类统计
    """
    etf = models.ForeignKey(ETFBaseInfo, on_delete=models.CASCADE, related_name='regions', verbose_name='ETF')
    symbol = models.CharField(max_length=20, verbose_name='ETF代码')

    # 地区信息
    region_name = models.CharField(max_length=100, verbose_name='地区名称')
    region_code = models.CharField(max_length=20, null=True, blank=True, verbose_name='地区代码')
    country = models.CharField(max_length=100, null=True, blank=True, verbose_name='国家')

    # 分布数据
    weight = models.DecimalField(max_digits=8, decimal_places=4, verbose_name='权重(%)')
    weight_change = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='权重变化(%)')
    market_value = models.DecimalField(max_digits=20, decimal_places=4, null=True, blank=True, verbose_name='市值')
    stock_count = models.IntegerField(null=True, blank=True, verbose_name='股票数量')

    # 报告期
    report_date = models.DateField(verbose_name='报告日期')

    # 元数据
    data_source = models.CharField(max_length=50, null=True, blank=True, verbose_name='数据来源')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')

    class Meta:
        db_table = 'etf_holding_region'
        verbose_name = 'ETF地区分布'
        verbose_name_plural = 'ETF地区分布'
        unique_together = [['symbol', 'region_name', 'report_date']]
        ordering = ['-report_date', '-weight', 'symbol']
        indexes = [
            models.Index(fields=['symbol', '-report_date']),
            models.Index(fields=['region_name']),
            models.Index(fields=['country']),
        ]

    def __str__(self):
        return f"{self.symbol} - {self.region_name} ({self.weight}%)"


class ETFRebalance(models.Model):
    """
    ETF 调仓记录表
    记录季度调仓变化
    """
    CHANGE_TYPE_CHOICES = [
        ('ADD', '新增'),
        ('REMOVE', '移除'),
        ('INCREASE', '增持'),
        ('DECREASE', '减持'),
        ('UNCHANGED', '不变'),
    ]

    etf = models.ForeignKey(ETFBaseInfo, on_delete=models.CASCADE, related_name='rebalances', verbose_name='ETF')
    symbol = models.CharField(max_length=20, verbose_name='ETF代码')

    # 持仓标的
    holding_symbol = models.CharField(max_length=20, verbose_name='持仓代码')
    holding_name = models.CharField(max_length=200, verbose_name='持仓名称')

    # 变化信息
    change_type = models.CharField(max_length=20, choices=CHANGE_TYPE_CHOICES, verbose_name='变化类型')
    old_weight = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='原权重(%)')
    new_weight = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='新权重(%)')
    weight_change = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='权重变化(%)')

    # 报告期
    report_date = models.DateField(verbose_name='报告日期')
    previous_report_date = models.DateField(null=True, blank=True, verbose_name='上期报告日期')

    # 元数据
    data_source = models.CharField(max_length=50, null=True, blank=True, verbose_name='数据来源')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')

    class Meta:
        db_table = 'etf_rebalance'
        verbose_name = 'ETF调仓记录'
        verbose_name_plural = 'ETF调仓记录'
        ordering = ['-report_date', '-weight_change', 'symbol']
        indexes = [
            models.Index(fields=['symbol', '-report_date']),
            models.Index(fields=['change_type']),
            models.Index(fields=['holding_symbol']),
        ]

    def __str__(self):
        return f"{self.symbol} - {self.holding_symbol} {self.get_change_type_display()}"

    def save(self, *args, **kwargs):
        # 自动计算权重变化
        if self.old_weight is not None and self.new_weight is not None:
            self.weight_change = self.new_weight - self.old_weight
        super().save(*args, **kwargs)


class ETFDividend(models.Model):
    """
    ETF 分红数据表
    """
    DIVIDEND_TYPE_CHOICES = [
        ('CASH', '现金分红'),
        ('STOCK', '股票分红'),
        ('SPECIAL', '特别分红'),
    ]

    etf = models.ForeignKey(ETFBaseInfo, on_delete=models.CASCADE, related_name='dividends', verbose_name='ETF')
    symbol = models.CharField(max_length=20, verbose_name='ETF代码')

    # 分红信息
    dividend_type = models.CharField(max_length=20, choices=DIVIDEND_TYPE_CHOICES, default='CASH', verbose_name='分红类型')
    dividend_amount = models.DecimalField(max_digits=12, decimal_places=6, verbose_name='每股分红金额')
    dividend_currency = models.CharField(max_length=10, default='USD', verbose_name='分红货币')

    # 日期
    ex_dividend_date = models.DateField(verbose_name='除息日')
    record_date = models.DateField(null=True, blank=True, verbose_name='股权登记日')
    payment_date = models.DateField(null=True, blank=True, verbose_name='派息日')

    # 派息频率
    frequency = models.CharField(max_length=20, null=True, blank=True, verbose_name='派息频率')  # 月度/季度/年度

    # 计算指标
    dividend_yield = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='股息率(%)')
    annualized_yield = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='年化股息率(%)')

    # 元数据
    data_source = models.CharField(max_length=50, null=True, blank=True, verbose_name='数据来源')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')

    class Meta:
        db_table = 'etf_dividend'
        verbose_name = 'ETF分红数据'
        verbose_name_plural = 'ETF分红数据'
        unique_together = [['symbol', 'ex_dividend_date']]
        ordering = ['-ex_dividend_date', 'symbol']
        indexes = [
            models.Index(fields=['symbol', '-ex_dividend_date']),
            models.Index(fields=['ex_dividend_date']),
        ]

    def __str__(self):
        return f"{self.symbol} - {self.ex_dividend_date} - {self.dividend_amount}"


class ETFIndicator(models.Model):
    """
    ETF 技术指标表
    存储常用的技术分析指标
    """
    etf = models.ForeignKey(ETFBaseInfo, on_delete=models.CASCADE, related_name='indicators', verbose_name='ETF')
    symbol = models.CharField(max_length=20, verbose_name='ETF代码')

    # 日期
    calc_date = models.DateField(verbose_name='计算日期')

    # 移动平均线
    ma5 = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='MA5')
    ma10 = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='MA10')
    ma20 = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='MA20')
    ma60 = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='MA60')
    ma120 = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='MA120')

    # 波动率
    volatility_20d = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='20日波动率(%)')
    volatility_60d = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='60日波动率(%)')

    # RSI
    rsi6 = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='RSI6')
    rsi12 = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='RSI12')
    rsi24 = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='RSI24')

    # MACD
    macd_dif = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='MACD DIF')
    macd_dea = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='MACD DEA')
    macd_histogram = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='MACD柱状图')

    # 布林带
    boll_upper = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='布林上轨')
    boll_middle = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='布林中轨')
    boll_lower = models.DecimalField(max_digits=12, decimal_places=4, null=True, blank=True, verbose_name='布林下轨')

    # 统计指标
    sharpe_ratio = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='夏普比率')
    max_drawdown = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='最大回撤(%)')
    beta = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='Beta')
    alpha = models.DecimalField(max_digits=8, decimal_places=4, null=True, blank=True, verbose_name='Alpha')

    # 计算周期
    calc_period = models.IntegerField(default=252, verbose_name='计算周期(天)')

    # 元数据
    data_source = models.CharField(max_length=50, null=True, blank=True, verbose_name='数据来源')
    created_at = models.DateTimeField(auto_now_add=True, verbose_name='创建时间')

    class Meta:
        db_table = 'etf_indicator'
        verbose_name = 'ETF技术指标'
        verbose_name_plural = 'ETF技术指标'
        unique_together = [['symbol', 'calc_date']]
        ordering = ['-calc_date', 'symbol']
        indexes = [
            models.Index(fields=['symbol', '-calc_date']),
        ]

    def __str__(self):
        return f"{self.symbol} - {self.calc_date}"

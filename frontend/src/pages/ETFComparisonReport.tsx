import { useState, useMemo } from 'react';
import styled from 'styled-components';
import { Table, App, Row, Col } from 'antd';
import { ArrowLeftOutlined, StarFilled, InfoCircleOutlined } from '@ant-design/icons';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, Legend, ResponsiveContainer, ReferenceLine } from 'recharts';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';

const PageHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px 0;
  border-bottom: 1px solid ${theme.colors.border};
  h2 {
    margin: 0;
    font-size: ${theme.fonts.size['2xl']};
    color: ${theme.colors.textPrimary};
    display: flex;
    align-items: center;
    gap: 12px;
  }
`;

const ChartContainer = styled.div`
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.lg};
  padding: 24px;
  margin-bottom: 20px;
  box-shadow: ${theme.shadows.card};
  .chart-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    h3 { margin: 0; font-size: ${theme.fonts.size.lg}; color: ${theme.colors.textPrimary}; }
    .period-tabs { display: flex; gap: 8px; }
    button {
      padding: 6px 16px; border: 1px solid ${theme.colors.border}; border-radius: 20px;
      background: transparent; cursor: pointer; transition: all 0.3s; font-size: ${theme.fonts.size.sm};
      &:hover, &.active { background: ${theme.colors.primary}; color: white; border-color: ${theme.colors.primary}; }
    }
  }
`;

const SectionCard = styled.div`
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.lg};
  padding: 20px;
  margin-bottom: 16px;
  box-shadow: ${theme.shadows.card};
  .section-header {
    display: flex; justify-content: space-between; align-items: center;
    margin-bottom: 16px; padding-bottom: 12px; border-bottom: 1px solid ${theme.colors.border};
    h4 { margin: 0; font-size: ${theme.fonts.size.base}; font-weight: ${theme.fonts.weight.semibold}; color: ${theme.colors.textPrimary}; }
  }
`;

const StyledTable = styled(Table)`
  .ant-table-thead > tr > th { background: #1a1a2e; font-weight: ${theme.fonts.weight.semibold}; color: ${theme.colors.textMuted}; font-size: ${theme.fonts.size.xs}; border-bottom: 1px solid ${theme.colors.border}; padding: 10px 8px; }
  .ant-table-tbody > tr > td { border-bottom: 1px solid ${theme.colors.border}; padding: 10px 8px; font-size: ${theme.fonts.size.sm}; }
  .ant-table-tbody > tr:hover > td { background: rgba(255,255,255,0.03); }
  .positive { color: #f5222d; }
  .negative { color: #52c41a; }
  .neutral { color: ${theme.colors.textMuted}; }
  .star-rating { color: #faad14; letter-spacing: 2px; }
`;

const ETFNameCell = styled.div`
  .symbol { font-weight: ${theme.fonts.weight.semibold}; color: ${theme.colors.textPrimary}; font-size: ${theme.fonts.size.base}; }
  .name { color: ${theme.colors.textMuted}; font-size: ${theme.fonts.size.xs}; margin-top: 2px; }
`;

const ColorDot = styled.span<{ color: string }>`
  display: inline-block; width: 10px; height: 10px; border-radius: 50%;
  background: ${props => props.color}; margin-right: 6px;
`;

interface ETFReportData {
  symbol: string;
  name: string;
  currency: string;
  inceptionDate: string;
  aum: number;
  expenseRatio: number;
  trackingIndex: string;
  priceData: { currentPrice: number; previousClose: number; volume: number; avgVolume10d: number; turnover: number; high52w: number; low52w: number };
  returns: { ytdReturn: number; return1y: number; return3y: number; return5y: number | null; return10y: number | null; return3yAnnualized: number };
  risk: { volatility3y: number; sharpe3y: number; maxDrawdown3y: number; trackingError3y: number; alpha3y: number; beta3y: number; stdDev3y: number; rSquared3y: number; sortino3y: number; riskRanking3y: string };
  dividend: { dividendYieldTTM: number; dividendYieldLFY: number; payoutFrequency: string; lastExDate: string; lastPayDate: string };
  holdings: { numHoldings: number; styleBox: string; topHoldings: Array<{ name: string; weight: number }> };
}

const COLORS = ['#1890ff', '#faad14', '#13c2c2'];

const mockETFData: ETFReportData[] = [
  {
    symbol: 'SPYD', name: 'SPDR S&P 500 High Dividend ETF', currency: 'USD',
    inceptionDate: '2015-10-07', aum: 108.5, expenseRatio: 0.35,
    trackingIndex: 'S&P 500 High Dividend Index',
    priceData: { currentPrice: 45.35, previousClose: 44.82, volume: 2388400, avgVolume10d: 1948500, turnover: 1.08, high52w: 48.041, low52w: 36.224 },
    returns: { ytdReturn: 3.21, return1y: 18.45, return3y: 38.72, return5y: 56.34, return10y: 127.02, return3yAnnualized: 11.26 },
    risk: { volatility3y: 15.23, sharpe3y: 0.455, maxDrawdown3y: -12.182, trackingError3y: 12.27, alpha3y: -3.300, beta3y: 0.800, stdDev3y: 10.785, rSquared3y: 39.110, sortino3y: 0.62, riskRanking3y: '超过83.64%同类' },
    dividend: { dividendYieldTTM: 4.36, dividendYieldLFY: 4.28, payoutFrequency: '季度', lastExDate: '2026/03/23', lastPayDate: '2026/03/25' },
    holdings: { numHoldings: 60, styleBox: '中型价值股', topHoldings: [{ name: 'APA', weight: 1.94 }, { name: 'LYB', weight: 1.88 }, { name: 'DOW', weight: 1.78 }, { name: 'EOG', weight: 1.61 }, { name: 'VZ', weight: 1.54 }, { name: 'PSX', weight: 1.52 }, { name: 'EIX', weight: 1.49 }, { name: 'T', weight: 1.47 }, { name: 'CVX', weight: 1.46 }, { name: 'MO', weight: 1.42 }] },
  },
  {
    symbol: 'SCHD', name: 'Schwab US Dividend Equity ETF', currency: 'USD',
    inceptionDate: '2011-10-20', aum: 62.9, expenseRatio: 0.06,
    trackingIndex: 'Dow Jones U.S. Dividend 100 Index',
    priceData: { currentPrice: 30.51, previousClose: 30.25, volume: 20626200, avgVolume10d: 27552800, turnover: 6.29, high52w: 31.682, low52w: 22.982 },
    returns: { ytdReturn: 2.87, return1y: 17.32, return3y: 42.58, return5y: 68.91, return10y: 218.28, return3yAnnualized: 11.88 },
    risk: { volatility3y: 14.56, sharpe3y: 0.550, maxDrawdown3y: -13.048, trackingError3y: 11.96, alpha3y: -0.880, beta3y: 0.660, stdDev3y: 10.535, rSquared3y: 32.980, sortino3y: 0.75, riskRanking3y: '超过68.13%同类' },
    dividend: { dividendYieldTTM: 3.46, dividendYieldLFY: 3.42, payoutFrequency: '季度', lastExDate: '2026/03/25', lastPayDate: '2026/03/30' },
    holdings: { numHoldings: 100, styleBox: '大型价值股', topHoldings: [{ name: 'CVX', weight: 4.43 }, { name: 'COP', weight: 4.26 }, { name: 'MRK', weight: 4.16 }, { name: 'KO', weight: 4.06 }, { name: 'VZ', weight: 3.99 }, { name: 'TXN', weight: 3.98 }, { name: 'PEP', weight: 3.96 }, { name: 'UNH', weight: 3.93 }, { name: 'AMGN', weight: 3.80 }, { name: 'HD', weight: 3.65 }] },
  },
  {
    symbol: 'JEPQ', name: 'JPMorgan Nasdaq Equity Premium Income ETF', currency: 'USD',
    inceptionDate: '2022-05-03', aum: 33.4, expenseRatio: 0.35,
    trackingIndex: 'Nasdaq-100 Index (with options overlay)',
    priceData: { currentPrice: 55.52, previousClose: 54.85, volume: 6050100, avgVolume10d: 7949900, turnover: 3.34, high52w: 58.564, low52w: 39.516 },
    returns: { ytdReturn: 5.67, return1y: 28.93, return3y: 78.45, return5y: null, return10y: null, return3yAnnualized: 19.60 },
    risk: { volatility3y: 18.92, sharpe3y: 1.291, maxDrawdown3y: -15.760, trackingError3y: 5.28, alpha3y: 3.420, beta3y: 0.780, stdDev3y: 11.748, rSquared3y: 80.790, sortino3y: 1.85, riskRanking3y: '超过12.52%同类' },
    dividend: { dividendYieldTTM: 11.12, dividendYieldLFY: 10.86, payoutFrequency: '月度', lastExDate: '2026/04/01', lastPayDate: '2026/04/06' },
    holdings: { numHoldings: 91, styleBox: '大型成长股', topHoldings: [{ name: 'NVDA', weight: 7.35 }, { name: 'AAPL', weight: 6.38 }, { name: 'GOOG', weight: 5.20 }, { name: 'MSFT', weight: 4.77 }, { name: 'AMZN', weight: 4.04 }, { name: 'META', weight: 3.03 }, { name: 'JPMorgan Principal', weight: 2.95 }, { name: 'TSLA', weight: 2.65 }, { name: 'WMT', weight: 2.46 }, { name: 'AVGO', weight: 2.35 }] },
  },
];

const generatePriceHistory = (basePrice: number, volatility: number, trend: number, days: number) => {
  const data: Array<{ date: string; change: number }> = [];
  let price = basePrice * (1 - trend);
  for (let i = 0; i < days; i++) {
    const change = (Math.random() - 0.48) * volatility + (trend / days);
    price *= (1 + change);
    const date = new Date();
    date.setDate(date.getDate() - (days - i));
    data.push({ date: date.toISOString().split('T')[0], change: parseFloat(((price / basePrice - 1) * 100).toFixed(2)) });
  }
  return data;
};

const PERIODS = [
  { key: '1m', label: '近1月', days: 30 },
  { key: '3m', label: '近3月', days: 90 },
  { key: '6m', label: '近6月', days: 180 },
  { key: '1y', label: '近1年', days: 365 },
  { key: '3y', label: '近3年', days: 1095 },
  { key: '5y', label: '近5年', days: 1825 },
];

const renderStars = (count: number) => (
  <span className="star-rating">{[...Array(5)].map((_, i) => <StarFilled key={i} style={{ opacity: i < count ? 1 : 0.3 }} />)}</span>
);

const formatNumber = (num: number | null, decimals = 2): string => {
  if (num === null || isNaN(num)) return '--';
  if (Math.abs(num) >= 10000) return `${(num / 10000).toFixed(decimals)}万`;
  if (Math.abs(num) >= 100000000) return `${(num / 100000000).toFixed(decimals)}亿`;
  return num.toFixed(decimals);
};

const formatPercent = (value: number | null, showSign = true): string => {
  if (value === null || isNaN(value)) return '--';
  const prefix = showSign && value > 0 ? '+' : '';
  return `${prefix}${value.toFixed(2)}%`;
};

const getChangeClass = (value: number | null): string => {
  if (value === null || value === 0) return 'neutral';
  return value > 0 ? 'positive' : 'negative';
};

const ETFComparisonReport: React.FC = () => {
  App.useApp();
  const [selectedPeriod, setSelectedPeriod] = useState('3m');
  const [selectedETFs] = useState<string[]>(['SPYD', 'SCHD', 'JEPQ']);

  const chartData = useMemo(() => {
    const period = PERIODS.find(p => p.key === selectedPeriod) || PERIODS[2];
    const etfConfigs = mockETFData.filter(e => selectedETFs.includes(e.symbol)).map((etf, idx) => ({
      symbol: etf.symbol, basePrice: etf.priceData.currentPrice,
      volatility: etf.risk.volatility3y / 100 / 20,
      trend: etf.returns.return3y / 100 * (period.days / 1095),
      color: COLORS[idx % COLORS.length],
    }));
    if (etfConfigs.length === 0) return [];
    const dates: Array<Record<string, number | string>> = [];
    for (let i = 0; i < period.days; i++) {
      const date = new Date(); date.setDate(date.getDate() - (period.days - i));
      dates.push({ date: date.toISOString().split('T')[0] });
    }
    etfConfigs.forEach(config => {
      const history = generatePriceHistory(config.basePrice, config.volatility, config.trend, period.days);
      history.forEach((item, idx) => { if (dates[idx]) dates[idx][config.symbol] = item.change; });
    });
    return dates;
  }, [selectedPeriod, selectedETFs]);

  const filteredETFs = useMemo(() => mockETFData.filter(etf => selectedETFs.includes(etf.symbol)), [selectedETFs]);

  const renderNameCell = (record: ETFReportData) => (
    <ETFNameCell>
      <div className="symbol">{record.name}<span style={{ marginLeft: 8, fontSize: '12px', fontWeight: 'normal' }}>{record.symbol}</span></div>
    </ETFNameCell>
  );

  const renderSubNameCell = (record: ETFReportData) => (
    <ETFNameCell><div className="symbol">{record.name}</div><div className="name">{record.symbol}</div></ETFNameCell>
  );

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const marketColumns: any = [
    { title: '股票名称', dataIndex: 'symbol', key: 'name', width: 200, render: (_: unknown, r: ETFReportData) => renderNameCell(r) },
    { title: '币种', dataIndex: ['priceData'], key: 'currency', width: 80, render: (_: unknown, r: ETFReportData) => r.currency },
    { title: '昨收价', dataIndex: ['priceData', 'previousClose'], key: 'prevClose', align: 'center', render: (v: number) => v?.toFixed(3) },
    { title: '成交量', dataIndex: ['priceData', 'volume'], key: 'volume', align: 'right', render: (v: number) => formatNumber(v, 2) },
    { title: '近10日平均成交量', dataIndex: ['priceData', 'avgVolume10d'], key: 'avgVol', align: 'right', render: (v: number) => formatNumber(v, 2) },
    { title: '成交额', dataIndex: ['priceData', 'turnover'], key: 'turnover', align: 'right', render: (v: number) => `${v?.toFixed(2)}亿` },
    { title: '52周最高', dataIndex: ['priceData', 'high52w'], key: 'high52w', align: 'center', render: (v: number) => v?.toFixed(3) },
    { title: '52周最低', dataIndex: ['priceData', 'low52w'], key: 'low52w', align: 'center', render: (v: number) => v?.toFixed(3) },
  ];

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const returnColumns: any = [
    { title: '股票名称', dataIndex: 'symbol', key: 'name', width: 200, render: (_: unknown, r: ETFReportData) => renderSubNameCell(r) },
    { title: '近3年年化收益率', dataIndex: ['returns', 'return3yAnnualized'], key: 'return3yAnn', align: 'center', render: (v: number | null) => <span className={getChangeClass(v)}>{formatPercent(v)}</span> },
    { title: '近10年累计收益率', dataIndex: ['returns', 'return10y'], key: 'return10y', align: 'center', render: (v: number | null) => <span className={getChangeClass(v)}>{formatPercent(v)}</span> },
    { title: '近3年万元收益', dataIndex: ['returns', 'return3y'], key: 'profit3y', align: 'center', render: (v: number | null) => v !== null ? `¥${(10000 * (1 + v / 100)).toFixed(2)}` : '--' },
    { title: '近5年万元收益', dataIndex: ['returns', 'return5y'], key: 'profit5y', align: 'center', render: (v: number | null) => v !== null ? `¥${(10000 * (1 + v / 100)).toFixed(2)}` : '--' },
    { title: '近10年万元收益', dataIndex: ['returns', 'return10y'], key: 'profit10y', align: 'center', render: (v: number | null) => v !== null ? `¥${(10000 * (1 + v / 100)).toFixed(2)}` : '--' },
    { title: '近3年晨星评级', dataIndex: 'symbol', key: 'star3y', align: 'center', render: (s: string) => renderStars(s === 'SPYD' ? 3 : s === 'SCHD' ? 2 : 5) },
    { title: '近3年收益能力同类排名', dataIndex: ['risk', 'riskRanking3y'], key: 'rank3y', align: 'center', render: (v: string) => <span className="positive">{v}</span> },
    { title: '近3年最长连续上涨周数', dataIndex: 'symbol', key: 'upWeeks', align: 'center', render: (s: string) => s === 'SPYD' ? 8 : s === 'SCHD' ? 8 : 13 },
  ];

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const holdingColumns: any = [
    { title: '股票名称', dataIndex: 'symbol', key: 'name', width: 200, render: (_: unknown, r: ETFReportData) => renderSubNameCell(r) },
    { title: '持股数量', dataIndex: ['holdings', 'numHoldings'], key: 'numHoldings', align: 'center' },
    { title: '股票风格箱', dataIndex: ['holdings', 'styleBox'], key: 'styleBox', align: 'center' },
    { title: '固收风格箱', key: 'fixedIncome', align: 'center', render: () => '--' },
  ];

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const topHoldingColumns: any = [
    { title: '', dataIndex: 'rank', key: 'rank', width: 40, align: 'center' },
    { title: '股票名称', dataIndex: 'symbol', key: 'name', width: 200, render: (_: unknown, r: ETFReportData) => renderSubNameCell(r) },
    ...filteredETFs.map((etf, idx) => ({
      title: (<span><ColorDot color={COLORS[idx]} />前10大持仓</span>), key: etf.symbol, align: 'center',
      render: (_: unknown, __: ETFReportData, index: number) => {
        const holding = etf.holdings.topHoldings[index];
        return holding ? `${holding.name} ${holding.weight}%` : '';
      },
    })),
  ];

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const dividendColumns: any = [
    { title: '股票名称', dataIndex: 'symbol', key: 'name', width: 200, render: (_: unknown, r: ETFReportData) => renderSubNameCell(r) },
    { title: '股息率TTM', dataIndex: ['dividend', 'dividendYieldTTM'], key: 'yieldTTM', align: 'center', render: (v: number) => <span className="positive">{v?.toFixed(2)}%</span> },
    { title: '股息率LFY', dataIndex: ['dividend', 'dividendYieldLFY'], key: 'yieldLFY', align: 'center', render: (v: number | null) => v ? `${v.toFixed(2)}%` : '--' },
    { title: '派息频率', dataIndex: ['dividend', 'payoutFrequency'], key: 'frequency', align: 'center' },
    { title: '最近除息日', dataIndex: ['dividend', 'lastExDate'], key: 'exDate', align: 'center' },
    { title: '最近派息日', dataIndex: ['dividend', 'lastPayDate'], key: 'payDate', align: 'center' },
  ];

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const riskColumns: any = [
    { title: '股票名称', dataIndex: 'symbol', key: 'name', width: 200, render: (_: unknown, r: ETFReportData) => renderSubNameCell(r) },
    { title: '近3年风险能力同类排名', dataIndex: ['risk', 'riskRanking3y'], key: 'riskRank', align: 'center', render: (v: string) => <span className="positive">{v}</span> },
    { title: '近3年跟踪误差', dataIndex: ['risk', 'trackingError3y'], key: 'trackingErr', align: 'center', render: (v: number) => `${v?.toFixed(2)}%` },
    { title: '近3年α', dataIndex: ['risk', 'alpha3y'], key: 'alpha', align: 'center', render: (v: number) => <span className={getChangeClass(v)}>{v?.toFixed(3)}</span> },
    { title: '近3年最大回撤', dataIndex: ['risk', 'maxDrawdown3y'], key: 'maxDD', align: 'center', render: (v: number) => <span className="negative">{v?.toFixed(3)}%</span> },
    { title: '近3年最大回撤修复天数', dataIndex: 'symbol', key: 'ddRecovery', align: 'center', render: (s: string) => s === 'SPYD' ? 193 : s === 'SCHD' ? 189 : 86 },
    { title: '近3年年化标准差', dataIndex: ['risk', 'stdDev3y'], key: 'stdDev', align: 'center', render: (v: number) => `${v?.toFixed(3)}%` },
    { title: '近3年β', dataIndex: ['risk', 'beta3y'], key: 'beta', align: 'center', render: (v: number) => v?.toFixed(3) },
    { title: '近3年R²', dataIndex: ['risk', 'rSquared3y'], key: 'rsquared', align: 'center', render: (v: number) => v?.toFixed(3) },
    { title: '近3年夏普比率', dataIndex: ['risk', 'sharpe3y'], key: 'sharpe', align: 'center', render: (v: number) => v?.toFixed(3) },
  ];

  return (
    <Layout>
      <PageHeader>
        <h2><ArrowLeftOutlined style={{ cursor: 'pointer' }} onClick={() => window.history.back()} />ETF对比</h2>
      </PageHeader>

      <ChartContainer>
        <div className="chart-header">
          <h3>价格走势</h3>
          <div className="period-tabs">
            {PERIODS.map(p => (
              <button key={p.key} className={selectedPeriod === p.key ? 'active' : ''} onClick={() => setSelectedPeriod(p.key)}>{p.label}</button>
            ))}
          </div>
        </div>
        <ResponsiveContainer width="100%" height={350}>
          <LineChart data={chartData} margin={{ top: 5, right: 30, left: 0, bottom: 5 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#333" />
            <XAxis dataKey="date" tick={{ fontSize: 11, fill: '#888' }} interval={Math.ceil(chartData.length / 6)} />
            <YAxis tick={{ fontSize: 11, fill: '#888' }} tickFormatter={(v: any) => `${v}%`} domain={['auto', 'auto']} />
            <RechartsTooltip contentStyle={{ backgroundColor: '#1a1a2e', border: '1px solid #333', borderRadius: 8 }} labelStyle={{ color: '#fff' }} formatter={(value: any) => [`${Number(value).toFixed(2)}%`, '']} />
            <Legend wrapperStyle={{ paddingTop: 20 }} />
            <ReferenceLine y={0} stroke="#666" strokeDasharray="3 3" />
            {filteredETFs.map((etf, idx) => (
              <Line key={etf.symbol} type="monotone" dataKey={etf.symbol} name={`${etf.name} (${etf.symbol})`} stroke={COLORS[idx]} strokeWidth={2} dot={false} activeDot={{ r: 4 }} />
            ))}
          </LineChart>
        </ResponsiveContainer>
      </ChartContainer>

      <SectionCard>
        <div className="section-header"><h4>行情指标</h4><InfoCircleOutlined style={{ color: theme.colors.textMuted }} /></div>
        <StyledTable dataSource={filteredETFs} columns={marketColumns} rowKey="symbol" pagination={false} size="middle" />
      </SectionCard>

      <SectionCard>
        <div className="section-header"><h4>收益表现</h4></div>
        <StyledTable dataSource={filteredETFs} columns={returnColumns} rowKey="symbol" pagination={false} size="middle" />
      </SectionCard>

      <Row gutter={[16, 16]}>
        <Col span={12}>
          <SectionCard>
            <div className="section-header"><h4>持仓分析</h4></div>
            <StyledTable dataSource={filteredETFs} columns={holdingColumns} rowKey="symbol" pagination={false} size="small" />
          </SectionCard>
        </Col>
        <Col span={12}>
          <SectionCard>
            <div className="section-header"><h4>分红派息</h4></div>
            <StyledTable dataSource={filteredETFs} columns={dividendColumns} rowKey="symbol" pagination={false} size="small" />
          </SectionCard>
        </Col>
      </Row>

      <SectionCard>
        <div className="section-header"><h4>前10大持仓</h4></div>
        <StyledTable dataSource={[...Array(10)].map((_, i) => ({ rank: i + 1 }))} columns={topHoldingColumns} rowKey="rank" pagination={false} size="small" />
      </SectionCard>

      <SectionCard>
        <div className="section-header"><h4>风险分析</h4></div>
        <StyledTable dataSource={filteredETFs} columns={riskColumns} rowKey="symbol" pagination={false} size="middle" />
      </SectionCard>
    </Layout>
  );
};

export default ETFComparisonReport;

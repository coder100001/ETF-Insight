import { useState } from 'react';
import { Card, Row, Col, Select, Button, DatePicker, InputNumber, Table, Tag, Spin, Alert, Statistic, Tabs } from 'antd';
import { XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, AreaChart, Area, BarChart, Bar } from 'recharts';
import { HistoryOutlined, PlayCircleOutlined, RiseOutlined, FallOutlined } from '@ant-design/icons';
import Layout from '../components/Layout';
import type { BacktestResult, PortfolioAllocation } from '../types';
import dayjs from 'dayjs';

const { Option } = Select;
const { RangePicker } = DatePicker;
const { TabPane } = Tabs;

// 模拟ETF列表
const ETF_LIST = [
  { symbol: 'SCHD', name: 'Schwab US Dividend Equity ETF' },
  { symbol: 'VYM', name: 'Vanguard High Dividend Yield ETF' },
  { symbol: 'JEPI', name: 'JPMorgan Equity Premium Income ETF' },
  { symbol: 'SPY', name: 'SPDR S&P 500 ETF Trust' },
  { symbol: 'QQQ', name: 'Invesco QQQ Trust' },
  { symbol: 'VTI', name: 'Vanguard Total Stock Market ETF' },
  { symbol: 'BND', name: 'Vanguard Total Bond Market ETF' },
];

const BacktestAnalysis = () => {
  const [portfolio, setPortfolio] = useState<PortfolioAllocation[]>([
    { symbol: 'SCHD', weight: 40 },
    { symbol: 'VYM', weight: 30 },
    { symbol: 'QQQ', weight: 30 },
  ]);
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>([
    dayjs().subtract(3, 'year'),
    dayjs(),
  ]);
  const [initialCapital, setInitialCapital] = useState(10000);
  const [rebalanceFreq, setRebalanceFreq] = useState('quarterly');
  const [commissionRate, setCommissionRate] = useState(0.001);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<BacktestResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleAddETF = () => {
    setPortfolio([...portfolio, { symbol: '', weight: 0 }]);
  };

  const handleRemoveETF = (index: number) => {
    setPortfolio(portfolio.filter((_, i) => i !== index));
  };

  const handlePortfolioChange = (index: number, field: keyof PortfolioAllocation, value: string | number) => {
    const newPortfolio = [...portfolio];
    newPortfolio[index] = { ...newPortfolio[index], [field]: value };
    setPortfolio(newPortfolio);
  };

  const handleRunBacktest = async () => {
    const totalWeight = portfolio.reduce((sum, p) => sum + (p.weight || 0), 0);
    if (Math.abs(totalWeight - 100) > 0.1) {
      setError(`权重总和必须为100%，当前为${totalWeight.toFixed(1)}%`);
      return;
    }

    if (!dateRange) {
      setError('请选择回测时间范围');
      return;
    }

    setLoading(true);
    setError(null);

    // 模拟回测结果
    setTimeout(() => {
      const days = dateRange[1].diff(dateRange[0], 'day');
      const years = days / 365;

      const mockResult: BacktestResult = {
        config: {
          start_date: dateRange[0].format('YYYY-MM-DD'),
          end_date: dateRange[1].format('YYYY-MM-DD'),
          initial_capital: initialCapital,
          rebalance_freq: rebalanceFreq as 'quarterly',
          commission_rate: commissionRate,
        },
        performance: {
          total_return: 35 + Math.random() * 40,
          annualized_return: 8 + Math.random() * 10,
          volatility: 12 + Math.random() * 8,
          sharpe_ratio: 0.8 + Math.random() * 0.8,
          sortino_ratio: 1.2 + Math.random() * 1.0,
          max_drawdown: -(8 + Math.random() * 15),
          max_drawdown_period: '2022-01-01 ~ 2022-10-01',
          win_rate: 55 + Math.random() * 15,
          profit_factor: 1.2 + Math.random() * 0.6,
          calmar_ratio: 0.6 + Math.random() * 0.8,
          final_capital: initialCapital * (1.35 + Math.random() * 0.4),
        },
        trades: [
          { date: dateRange[0].format('YYYY-MM-DD'), symbol: 'Initial', action: 'buy', shares: 0, price: 0, amount: initialCapital, commission: 0, reason: 'initial_allocation' },
          { date: dateRange[0].add(3, 'month').format('YYYY-MM-DD'), symbol: 'SCHD', action: 'rebalance', shares: 10, price: 75, amount: 750, commission: 0.75, reason: 'rebalance' },
          { date: dateRange[0].add(6, 'month').format('YYYY-MM-DD'), symbol: 'VYM', action: 'rebalance', shares: 8, price: 95, amount: 760, commission: 0.76, reason: 'rebalance' },
        ],
        daily_returns: Array.from({ length: Math.min(days, 252) }, (_, i) => ({
          date: dateRange[0].add(i, 'day').format('YYYY-MM-DD'),
          portfolio_value: initialCapital * (1 + (i / 252) * (0.35 + Math.random() * 0.4) + Math.sin(i / 30) * 0.05),
          daily_return: (Math.random() - 0.48) * 0.02,
          cumulative_return: (i / 252) * (0.35 + Math.random() * 0.4),
        })),
        drawdowns: [
          { start_date: '2022-01-15', end_date: '2022-10-15', peak_value: 12500, trough_value: 10800, drawdown: -13.6, duration: 273 },
          { start_date: '2023-07-20', end_date: '2023-10-30', peak_value: 14200, trough_value: 13500, drawdown: -4.9, duration: 102 },
        ],
        monthly_returns: {
          '2022-01': -3.5, '2022-02': -2.1, '2022-03': 2.8, '2022-04': -4.2, '2022-05': 1.5, '2022-06': -3.8,
          '2022-07': 4.2, '2022-08': -2.5, '2022-09': -5.1, '2022-10': 3.8, '2022-11': 5.2, '2022-12': -1.8,
        },
        yearly_returns: {
          '2022': -8.5, '2023': 18.2, '2024': 12.5,
        },
        risk_metrics: {
          volatility: 15.2,
          sharpe_ratio: 1.15,
          sortino_ratio: 1.65,
          calmar_ratio: 0.85,
          max_drawdown: -12.5,
          max_drawdown_days: 180,
          beta: 0.92,
          alpha: 1.2,
          var_95: -2.1,
          var_99: -3.2,
          cvar_95: -2.8,
          cvar_99: -4.1,
          treynor_ratio: 8.5,
          information_ratio: 0.35,
          tracking_error: 3.2,
          downside_deviation: 9.8,
          upside_potential: 14.2,
          omega_ratio: 1.45,
        },
        benchmark_compare: {
          benchmark_symbol: 'SPY',
          portfolio_return: 42.5,
          benchmark_return: 38.2,
          excess_return: 4.3,
          tracking_error: 3.2,
          information_ratio: 1.34,
          beta: 0.92,
          alpha: 1.2,
          correlation: 0.95,
        },
      };

      setResult(mockResult);
      setLoading(false);
    }, 1500);
  };

  const totalWeight = portfolio.reduce((sum, p) => sum + (p.weight || 0), 0);

  return (
    <Layout>
      <div style={{ padding: '24px' }}>
        <h1 style={{ marginBottom: '24px' }}>
          <HistoryOutlined style={{ marginRight: '8px' }} />
          回测分析
        </h1>

        {/* 回测配置 */}
        <Card title="回测配置" style={{ marginBottom: '24px' }}>
          <Row gutter={24}>
            <Col span={24}>
              <h4>投资组合配置</h4>
              {portfolio.map((item, index) => (
                <Row gutter={16} key={index} style={{ marginBottom: '8px' }}>
                  <Col span={8}>
                    <Select
                      placeholder="选择ETF"
                      value={item.symbol || undefined}
                      onChange={(value) => handlePortfolioChange(index, 'symbol', value)}
                      style={{ width: '100%' }}
                    >
                      {ETF_LIST.map(etf => (
                        <Option key={etf.symbol} value={etf.symbol}>{etf.symbol} - {etf.name}</Option>
                      ))}
                    </Select>
                  </Col>
                  <Col span={6}>
                    <InputNumber
                      placeholder="权重(%)"
                      min={0}
                      max={100}
                      value={item.weight}
                      onChange={(value) => handlePortfolioChange(index, 'weight', value || 0)}
                      style={{ width: '100%' }}
                      addonAfter="%"
                    />
                  </Col>
                  <Col span={4}>
                    <Button danger onClick={() => handleRemoveETF(index)}>删除</Button>
                  </Col>
                </Row>
              ))}
              <Button type="dashed" onClick={handleAddETF} style={{ marginTop: '8px' }}>+ 添加ETF</Button>
              <div style={{ marginTop: '8px' }}>
                <Tag color={Math.abs(totalWeight - 100) < 0.1 ? 'green' : 'red'}>
                  总权重: {totalWeight.toFixed(1)}%
                </Tag>
              </div>
            </Col>
          </Row>

          <Row gutter={24} style={{ marginTop: '24px' }}>
            <Col span={6}>
              <label>回测时间范围</label>
              <RangePicker
                value={dateRange}
                onChange={(dates) => setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null)}
                style={{ width: '100%' }}
              />
            </Col>
            <Col span={4}>
              <label>初始资金</label>
              <InputNumber
                min={1000}
                value={initialCapital}
                onChange={(value) => setInitialCapital(value || 10000)}
                style={{ width: '100%' }}
                formatter={(value) => `$ ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
              />
            </Col>
            <Col span={4}>
              <label>再平衡频率</label>
              <Select value={rebalanceFreq} onChange={setRebalanceFreq} style={{ width: '100%' }}>
                <Option value="monthly">月度</Option>
                <Option value="quarterly">季度</Option>
                <Option value="yearly">年度</Option>
                <Option value="none">不再平衡</Option>
              </Select>
            </Col>
            <Col span={4}>
              <label>手续费率</label>
              <InputNumber
                min={0}
                max={0.01}
                step={0.0001}
                value={commissionRate}
                onChange={(value) => setCommissionRate(value || 0)}
                style={{ width: '100%' }}
                formatter={(value) => `${(Number(value) * 100).toFixed(2)}%`}
              />
            </Col>
            <Col span={6}>
              <Button
                type="primary"
                size="large"
                onClick={handleRunBacktest}
                loading={loading}
                icon={<PlayCircleOutlined />}
                style={{ marginTop: '22px' }}
                disabled={Math.abs(totalWeight - 100) > 0.1}
              >
                运行回测
              </Button>
            </Col>
          </Row>
          {error && <Alert message={error} type="error" showIcon style={{ marginTop: '16px' }} />}
        </Card>

        {loading ? (
          <div style={{ textAlign: 'center', padding: '48px' }}>
            <Spin size="large" />
            <p style={{ marginTop: '16px' }}>正在运行回测...</p>
          </div>
        ) : result ? (
          <>
            {/* 业绩概览 */}
            <Row gutter={16} style={{ marginBottom: '24px' }}>
              <Col span={4}>
                <Card>
                  <Statistic
                    title="总收益"
                    value={result.performance.total_return}
                    precision={2}
                    suffix="%"
                    valueStyle={{ color: result.performance.total_return > 0 ? '#52c41a' : '#ff4d4f' }}
                    prefix={result.performance.total_return > 0 ? <RiseOutlined /> : <FallOutlined />}
                  />
                </Card>
              </Col>
              <Col span={4}>
                <Card>
                  <Statistic
                    title="年化收益"
                    value={result.performance.annualized_return}
                    precision={2}
                    suffix="%"
                    valueStyle={{ color: '#1890ff' }}
                  />
                </Card>
              </Col>
              <Col span={4}>
                <Card>
                  <Statistic
                    title="夏普比率"
                    value={result.performance.sharpe_ratio}
                    precision={2}
                    valueStyle={{ color: result.performance.sharpe_ratio > 1 ? '#52c41a' : '#faad14' }}
                  />
                </Card>
              </Col>
              <Col span={4}>
                <Card>
                  <Statistic
                    title="最大回撤"
                    value={result.performance.max_drawdown}
                    precision={2}
                    suffix="%"
                    valueStyle={{ color: '#ff4d4f' }}
                  />
                </Card>
              </Col>
              <Col span={4}>
                <Card>
                  <Statistic
                    title="最终资产"
                    value={result.performance.final_capital}
                    precision={0}
                    prefix="$"
                    valueStyle={{ color: '#52c41a' }}
                  />
                </Card>
              </Col>
              <Col span={4}>
                <Card>
                  <Statistic
                    title="超额收益(vs SPY)"
                    value={result.benchmark_compare.excess_return}
                    precision={2}
                    suffix="%"
                    valueStyle={{ color: result.benchmark_compare.excess_return > 0 ? '#52c41a' : '#ff4d4f' }}
                  />
                </Card>
              </Col>
            </Row>

            {/* 图表分析 */}
            <Tabs defaultActiveKey="equity">
              <TabPane tab="权益曲线" key="equity">
                <Card>
                  <ResponsiveContainer width="100%" height={400}>
                    <AreaChart data={result.daily_returns}>
                      <defs>
                        <linearGradient id="colorValue" x1="0" y1="0" x2="0" y2="1">
                          <stop offset="5%" stopColor="#1890ff" stopOpacity={0.8}/>
                          <stop offset="95%" stopColor="#1890ff" stopOpacity={0}/>
                        </linearGradient>
                      </defs>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="date" tickFormatter={(date) => dayjs(date).format('YYYY-MM')} />
                      <YAxis />
                      <Tooltip formatter={(value: number) => `$${Number(value).toFixed(2)}`} />
                      <Area type="monotone" dataKey="portfolio_value" stroke="#1890ff" fillOpacity={1} fill="url(#colorValue)" />
                    </AreaChart>
                  </ResponsiveContainer>
                </Card>
              </TabPane>

              <TabPane tab="回撤分析" key="drawdown">
                <Card>
                  <ResponsiveContainer width="100%" height={400}>
                    <AreaChart data={result.daily_returns.map(d => ({
                      ...d,
                      drawdown: Math.min(0, (d.portfolio_value / (result?.performance.final_capital || 1) - 1) * 100)
                    }))}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="date" tickFormatter={(date) => dayjs(date).format('YYYY-MM')} />
                      <YAxis />
                      <Tooltip />
                      <Area type="monotone" dataKey="drawdown" stroke="#ff4d4f" fill="#ff4d4f" fillOpacity={0.3} />
                    </AreaChart>
                  </ResponsiveContainer>
                </Card>
              </TabPane>

              <TabPane tab="月度收益" key="monthly">
                <Card>
                  <ResponsiveContainer width="100%" height={400}>
                    <BarChart data={Object.entries(result.monthly_returns).map(([month, ret]) => ({ month, return: ret }))}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="month" />
                      <YAxis />
                      <Tooltip />
                      <Bar dataKey="return" fill="#1890ff" />
                    </BarChart>
                  </ResponsiveContainer>
                </Card>
              </TabPane>
            </Tabs>

            {/* 交易记录 */}
            <Card title="交易记录" style={{ marginTop: '24px' }}>
              <Table
                dataSource={result.trades.map((t, i) => ({ ...t, key: i }))}
                columns={[
                  { title: '日期', dataIndex: 'date', key: 'date' },
                  { title: 'ETF', dataIndex: 'symbol', key: 'symbol' },
                  { title: '操作', dataIndex: 'action', key: 'action', render: (v: string) => <Tag>{v}</Tag> },
                  { title: '数量', dataIndex: 'shares', key: 'shares' },
                  { title: '价格', dataIndex: 'price', key: 'price', render: (v: number) => `$${v}` },
                  { title: '金额', dataIndex: 'amount', key: 'amount', render: (v: number) => `$${v.toFixed(2)}` },
                  { title: '手续费', dataIndex: 'commission', key: 'commission', render: (v: number) => `$${v.toFixed(2)}` },
                  { title: '原因', dataIndex: 'reason', key: 'reason' },
                ]}
                pagination={{ pageSize: 5 }}
                size="small"
              />
            </Card>
          </>
        ) : null}
      </div>
    </Layout>
  );
};

export default BacktestAnalysis;

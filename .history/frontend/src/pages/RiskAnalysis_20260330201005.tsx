import { useState } from 'react';
import { Card, Row, Col, Select, Button, Statistic, Table, Tag, Spin, Alert, Slider, Form } from 'antd';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, RadarChart, PolarGrid, PolarAngleAxis, PolarRadiusAxis, Radar } from 'recharts';
import { ExclamationCircleOutlined, SafetyOutlined, LineChartOutlined } from '@ant-design/icons';
import Layout from '../components/Layout';
import { riskMetricsAPI } from '../services/api';
import type { RiskMetrics } from '../types';

const { Option } = Select;

// 模拟ETF列表
const ETF_LIST = [
  { symbol: 'SCHD', name: 'Schwab US Dividend Equity ETF' },
  { symbol: 'VYM', name: 'Vanguard High Dividend Yield ETF' },
  { symbol: 'JEPI', name: 'JPMorgan Equity Premium Income ETF' },
  { symbol: 'SPY', name: 'SPDR S&P 500 ETF Trust' },
  { symbol: 'QQQ', name: 'Invesco QQQ Trust' },
  { symbol: 'VTI', name: 'Vanguard Total Stock Market ETF' },
  { symbol: 'VXUS', name: 'Vanguard Total International Stock ETF' },
  { symbol: 'BND', name: 'Vanguard Total Bond Market ETF' },
];

const RiskAnalysis = () => {
  const [selectedETFs, setSelectedETFs] = useState<string[]>([]);
  const [period, setPeriod] = useState('1y');
  const [riskFreeRate, setRiskFreeRate] = useState(4.5);
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState<Record<string, RiskMetrics>>({});
  const [error, setError] = useState<string | null>(null);

  const handleAnalyze = async () => {
    if (selectedETFs.length === 0) {
      setError('请至少选择一个ETF');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      // 模拟数据 - 实际项目中应该调用API
      // const response = await riskMetricsAPI.calculateBatch(selectedETFs, period, riskFreeRate / 100);
      // if (response.success) {
      //   setResults(response.data || {});
      // }

      // 模拟数据
      const mockResults: Record<string, RiskMetrics> = {};
      selectedETFs.forEach(symbol => {
        mockResults[symbol] = {
          volatility: 15 + Math.random() * 10,
          sharpe_ratio: 0.8 + Math.random() * 1.2,
          sortino_ratio: 1.2 + Math.random() * 1.5,
          calmar_ratio: 0.5 + Math.random() * 1.0,
          max_drawdown: -(5 + Math.random() * 20),
          max_drawdown_days: 30 + Math.floor(Math.random() * 100),
          beta: 0.7 + Math.random() * 0.6,
          alpha: -2 + Math.random() * 4,
          var_95: -(1 + Math.random() * 3),
          var_99: -(2 + Math.random() * 4),
          cvar_95: -(1.5 + Math.random() * 3),
          cvar_99: -(3 + Math.random() * 5),
          treynor_ratio: 5 + Math.random() * 10,
          information_ratio: -0.5 + Math.random(),
          tracking_error: 2 + Math.random() * 8,
          downside_deviation: 8 + Math.random() * 6,
          upside_potential: 12 + Math.random() * 8,
          omega_ratio: 1.2 + Math.random() * 0.8,
        };
      });

      setTimeout(() => {
        setResults(mockResults);
        setLoading(false);
      }, 800);
    } catch (err) {
      setError('分析失败，请重试');
      setLoading(false);
    }
  };

  // 雷达图数据
  const getRadarData = () => {
    const metrics = ['sharpe_ratio', 'sortino_ratio', 'calmar_ratio', 'omega_ratio', 'treynor_ratio'];
    return metrics.map(metric => {
      const data: Record<string, number | string> = { metric: getMetricLabel(metric) };
      Object.entries(results).forEach(([symbol, metrics]) => {
        data[symbol] = Number((metrics[metric as keyof RiskMetrics] as number)?.toFixed(2)) || 0;
      });
      return data;
    });
  };

  const getMetricLabel = (key: string) => {
    const labels: Record<string, string> = {
      sharpe_ratio: '夏普比率',
      sortino_ratio: '索提诺比率',
      calmar_ratio: '卡玛比率',
      omega_ratio: 'Omega比率',
      treynor_ratio: '特雷诺比率',
    };
    return labels[key] || key;
  };

  // 表格列
  const columns = [
    { title: 'ETF', dataIndex: 'symbol', key: 'symbol', fixed: 'left' },
    {
      title: '波动率',
      dataIndex: 'volatility',
      key: 'volatility',
      render: (v: number) => `${v?.toFixed(2)}%`,
      sorter: (a: RiskMetrics, b: RiskMetrics) => a.volatility - b.volatility,
    },
    {
      title: '夏普比率',
      dataIndex: 'sharpe_ratio',
      key: 'sharpe_ratio',
      render: (v: number) => (
        <Tag color={v > 1 ? 'green' : v > 0.5 ? 'orange' : 'red'}>
          {v?.toFixed(2)}
        </Tag>
      ),
      sorter: (a: RiskMetrics, b: RiskMetrics) => a.sharpe_ratio - b.sharpe_ratio,
    },
    {
      title: '索提诺比率',
      dataIndex: 'sortino_ratio',
      key: 'sortino_ratio',
      render: (v: number) => v?.toFixed(2),
      sorter: (a: RiskMetrics, b: RiskMetrics) => a.sortino_ratio - b.sortino_ratio,
    },
    {
      title: '最大回撤',
      dataIndex: 'max_drawdown',
      key: 'max_drawdown',
      render: (v: number) => (
        <Tag color={v > -10 ? 'green' : v > -20 ? 'orange' : 'red'}>
          {v?.toFixed(2)}%
        </Tag>
      ),
      sorter: (a: RiskMetrics, b: RiskMetrics) => a.max_drawdown - b.max_drawdown,
    },
    {
      title: '回撤天数',
      dataIndex: 'max_drawdown_days',
      key: 'max_drawdown_days',
      render: (v: number) => `${v}天`,
    },
    {
      title: 'Beta',
      dataIndex: 'beta',
      key: 'beta',
      render: (v: number) => v?.toFixed(2),
    },
    {
      title: 'Alpha',
      dataIndex: 'alpha',
      key: 'alpha',
      render: (v: number) => (
        <Tag color={v > 0 ? 'green' : 'red'}>{v > 0 ? '+' : ''}{v?.toFixed(2)}%</Tag>
      ),
    },
    {
      title: 'VaR(95%)',
      dataIndex: 'var_95',
      key: 'var_95',
      render: (v: number) => `${v?.toFixed(2)}%`,
    },
    {
      title: '卡玛比率',
      dataIndex: 'calmar_ratio',
      key: 'calmar_ratio',
      render: (v: number) => v?.toFixed(2),
    },
  ];

  const tableData = Object.entries(results).map(([symbol, metrics]) => ({
    symbol,
    ...metrics,
    key: symbol,
  }));

  // 风险等级判断
  const getRiskLevel = (metrics: RiskMetrics) => {
    if (metrics.sharpe_ratio > 1.2 && metrics.max_drawdown > -15) return { level: '低风险', color: 'green' };
    if (metrics.sharpe_ratio > 0.8 && metrics.max_drawdown > -25) return { level: '中风险', color: 'orange' };
    return { level: '高风险', color: 'red' };
  };

  return (
    <Layout>
      <div style={{ padding: '24px' }}>
        <h1 style={{ marginBottom: '24px' }}>
          <SafetyOutlined style={{ marginRight: '8px' }} />
          风险指标分析
        </h1>

        {/* 参数配置 */}
        <Card title="分析参数" style={{ marginBottom: '24px' }}>
          <Row gutter={24} align="middle">
            <Col span={8}>
              <Form.Item label="选择ETF">
                <Select
                  mode="multiple"
                  placeholder="选择要分析的ETF"
                  value={selectedETFs}
                  onChange={setSelectedETFs}
                  style={{ width: '100%' }}
                  maxTagCount={3}
                >
                  {ETF_LIST.map(etf => (
                    <Option key={etf.symbol} value={etf.symbol}>
                      {etf.symbol} - {etf.name}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={4}>
              <Form.Item label="时间周期">
                <Select value={period} onChange={setPeriod} style={{ width: '100%' }}>
                  <Option value="1m">1个月</Option>
                  <Option value="3m">3个月</Option>
                  <Option value="6m">6个月</Option>
                  <Option value="1y">1年</Option>
                  <Option value="3y">3年</Option>
                  <Option value="5y">5年</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item label={`无风险利率: ${riskFreeRate}%`}>
                <Slider
                  min={0}
                  max={10}
                  step={0.1}
                  value={riskFreeRate}
                  onChange={setRiskFreeRate}
                />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Button
                type="primary"
                size="large"
                onClick={handleAnalyze}
                loading={loading}
                icon={<LineChartOutlined />}
                style={{ marginTop: '30px' }}
              >
                开始分析
              </Button>
            </Col>
          </Row>
          {error && <Alert message={error} type="error" showIcon style={{ marginTop: '16px' }} />}
        </Card>

        {loading ? (
          <div style={{ textAlign: 'center', padding: '48px' }}>
            <Spin size="large" />
            <p style={{ marginTop: '16px' }}>正在计算风险指标...</p>
          </div>
        ) : Object.keys(results).length > 0 ? (
          <>
            {/* 风险概览卡片 */}
            <Row gutter={16} style={{ marginBottom: '24px' }}>
              {Object.entries(results).map(([symbol, metrics]) => {
                const risk = getRiskLevel(metrics);
                return (
                  <Col span={6} key={symbol}>
                    <Card size="small" title={symbol}>
                      <Tag color={risk.color}>{risk.level}</Tag>
                      <Row gutter={8} style={{ marginTop: '12px' }}>
                        <Col span={12}>
                          <Statistic
                            title="夏普比率"
                            value={metrics.sharpe_ratio}
                            precision={2}
                            valueStyle={{ color: metrics.sharpe_ratio > 1 ? '#52c41a' : '#faad14' }}
                          />
                        </Col>
                        <Col span={12}>
                          <Statistic
                            title="最大回撤"
                            value={metrics.max_drawdown}
                            precision={2}
                            suffix="%"
                            valueStyle={{ color: '#ff4d4f' }}
                          />
                        </Col>
                      </Row>
                    </Card>
                  </Col>
                );
              })}
            </Row>

            {/* 雷达图对比 */}
            <Card title="风险调整收益对比" style={{ marginBottom: '24px' }}>
              <ResponsiveContainer width="100%" height={400}>
                <RadarChart data={getRadarData()}>
                  <PolarGrid />
                  <PolarAngleAxis dataKey="metric" />
                  <PolarRadiusAxis angle={30} domain={[0, 'auto']} />
                  {Object.keys(results).map((symbol, index) => (
                    <Radar
                      key={symbol}
                      name={symbol}
                      dataKey={symbol}
                      stroke={['#1890ff', '#52c41a', '#faad14', '#ff4d4f', '#722ed1'][index % 5]}
                      fill={['#1890ff', '#52c41a', '#faad14', '#ff4d4f', '#722ed1'][index % 5]}
                      fillOpacity={0.1}
                    />
                  ))}
                  <Legend />
                  <Tooltip />
                </RadarChart>
              </ResponsiveContainer>
            </Card>

            {/* 详细数据表格 */}
            <Card title="详细风险指标">
              <Table
                columns={columns}
                dataSource={tableData}
                scroll={{ x: 1500 }}
                pagination={false}
                size="small"
              />
            </Card>

            {/* 风险说明 */}
            <Card title="指标说明" style={{ marginTop: '24px' }}>
              <Row gutter={24}>
                <Col span={8}>
                  <h4><ExclamationCircleOutlined /> 夏普比率 (Sharpe Ratio)</h4>
                  <p>衡量每单位总风险所获得的超额收益。&gt;1优秀，&gt;0.5良好。</p>
                </Col>
                <Col span={8}>
                  <h4><ExclamationCircleOutlined /> 索提诺比率 (Sortino Ratio)</h4>
                  <p>只考虑下行风险的收益调整指标。比夏普比率更关注实际损失风险。</p>
                </Col>
                <Col span={8}>
                  <h4><ExclamationCircleOutlined /> 最大回撤 (Max Drawdown)</h4>
                  <p>从峰值到谷底的最大跌幅。反映最坏情况下的损失幅度。</p>
                </Col>
              </Row>
            </Card>
          </>
        ) : null}
      </div>
    </Layout>
  );
};

export default RiskAnalysis;

import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Typography,
  Select,
  InputNumber,
  Button,
  Table,
  Tag,
  Statistic,
  Divider,
  Collapse,
  message,
  Space,
} from 'antd';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  BarChart,
  Bar,
} from 'recharts';
import {
  ArrowUpOutlined,
  ArrowDownOutlined,
  DashboardOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { Panel } = Collapse;

interface ScenarioAssumptions {
  scenario: string;
  annual_return: number;
  volatility: number;
  sharpe_ratio: number;
  max_drawdown: number;
  description: string;
  risk_free_rate: number;
  market_condition: string;
  dividend_yield: number;
}

interface PortfolioProjection {
  year: number;
  start_date: string;
  end_date: string;
  start_value: number;
  end_value: number;
  annual_return: number;
  cumulative_return: number;
  volatility: number;
  max_drawdown: number;
  sharpe_ratio: number;
  dividend_income: number;
  total_value: number;
}

interface ScenarioResult {
  assumptions: ScenarioAssumptions;
  projections: PortfolioProjection[];
  final_value: number;
  total_return: number;
  avg_annual_return: number;
  best_year: number;
  worst_year: number;
}

interface ComparisonMetrics {
  best_scenario: string;
  worst_scenario: string;
  value_difference: number;
  return_spread: number;
  risk_adjusted_winner: string;
}

interface ScenarioAnalysisResult {
  portfolio: Record<string, number>;
  total_investment: number;
  time_horizon_years: number;
  scenarios: Record<string, ScenarioResult>;
  comparison_metrics: ComparisonMetrics;
  methodology: string;
  assumptions: string;
  limitations: string;
}

const PortfolioAnalysis: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<ScenarioAnalysisResult | null>(null);
  const [totalInvestment, setTotalInvestment] = useState(100000);
  const [timeHorizon, setTimeHorizon] = useState(10);
  const [schdWeight, setSchdWeight] = useState(70);
  const [jepqWeight, setJepqWeight] = useState(30);
  const [activeScenario, setActiveScenario] = useState<string>('all');

  const handleAnalyze = async () => {
    if (schdWeight + jepqWeight !== 100) {
      message.error('权重总和必须为 100%');
      return;
    }

    setLoading(true);
    try {
      const response = await fetch('/api/portfolio/scenarios', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          portfolio: {
            SCHD: schdWeight / 100,
            JEPQ: jepqWeight / 100,
          },
          total_investment: totalInvestment,
          time_horizon_years: timeHorizon,
        }),
      });

      const data = await response.json();
      if (data.success) {
        setResult(data.data);
        message.success('分析完成');
      } else {
        message.error(data.error || '分析失败');
      }
    } catch (error) {
      message.error('请求失败');
    } finally {
      setLoading(false);
    }
  };

  const getScenarioColor = (scenario: string): string => {
    switch (scenario) {
      case 'optimistic':
        return 'green';
      case 'pessimistic':
        return 'red';
      default:
        return 'blue';
    }
  };

  const getScenarioName = (scenario: string): string => {
    switch (scenario) {
      case 'optimistic':
        return '乐观';
      case 'pessimistic':
        return '悲观';
      default:
        return '中性';
    }
  };

  const formatCurrency = (value: number): string => {
    return new Intl.NumberFormat('zh-CN', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
    }).format(value);
  };

  const formatPercent = (value: number): string => {
    return `${value.toFixed(2)}%`;
  };

  const comparisonColumns = [
    {
      title: '指标',
      dataIndex: 'metric',
      key: 'metric',
    },
    {
      title: '悲观',
      dataIndex: 'pessimistic',
      key: 'pessimistic',
      render: (value: number) => (
        <Text type="danger">{formatPercent(value)}</Text>
      ),
    },
    {
      title: '中性',
      dataIndex: 'neutral',
      key: 'neutral',
      render: (value: number) => <Text>{formatPercent(value)}</Text>,
    },
    {
      title: '乐观',
      dataIndex: 'optimistic',
      key: 'optimistic',
      render: (value: number) => (
        <Text type="success">{formatPercent(value)}</Text>
      ),
    },
  ];

  const projectionColumns = [
    {
      title: '年份',
      dataIndex: 'year',
      key: 'year',
      width: 80,
    },
    {
      title: '期初价值',
      dataIndex: 'start_value',
      key: 'start_value',
      render: (value: number) => formatCurrency(value),
    },
    {
      title: '期末价值',
      dataIndex: 'end_value',
      key: 'end_value',
      render: (value: number) => formatCurrency(value),
    },
    {
      title: '年度收益',
      dataIndex: 'annual_return',
      key: 'annual_return',
      render: (value: number) => (
        <Tag color={value >= 0 ? 'green' : 'red'}>
          {value >= 0 ? '+' : ''}
          {formatPercent(value)}
        </Tag>
      ),
    },
    {
      title: '累计收益',
      dataIndex: 'cumulative_return',
      key: 'cumulative_return',
      render: (value: number) => (
        <Tag color={value >= 0 ? 'green' : 'red'}>
          {value >= 0 ? '+' : ''}
          {formatPercent(value)}
        </Tag>
      ),
    },
    {
      title: '股息收入',
      dataIndex: 'dividend_income',
      key: 'dividend_income',
      render: (value: number) => formatCurrency(value),
    },
  ];

  const generateComparisonData = () => {
    if (!result) return [];

    const scenarios = ['pessimistic', 'neutral', 'optimistic'];
    return scenarios.map((scenario) => {
      const data: Record<string, unknown> = { metric: scenario };
      const scenarioResult = result.scenarios[scenario];
      if (scenarioResult) {
        data.avg_annual_return = scenarioResult.avg_annual_return;
        data.volatility = scenarioResult.assumptions.volatility * 100;
        data.sharpe_ratio = scenarioResult.assumptions.sharpe_ratio;
        data.max_drawdown = scenarioResult.assumptions.max_drawdown * 100;
        data.total_return = scenarioResult.total_return;
      }
      return data;
    });
  };

  const generateProjectionChartData = () => {
    if (!result) return [];

    const scenarios = ['pessimistic', 'neutral', 'optimistic'];
    const years = result.time_horizon_years;

    return Array.from({ length: years }, (_, i) => {
      const data: Record<string, unknown> = { year: i + 1 };
      scenarios.forEach((scenario) => {
        const scenarioResult = result.scenarios[scenario];
        if (scenarioResult && scenarioResult.projections[i]) {
          data[scenario] = scenarioResult.projections[i].end_value;
        }
      });
      return data;
    });
  };

  return (
    <div style={{ padding: '24px' }}>
      <Title level={2}>
        <DashboardOutlined /> 投资组合情景分析
      </Title>
      <Paragraph type="secondary">
        基于历史表现和蒙特卡洛模拟，分析不同市场情景下的投资组合表现
      </Paragraph>

      <Card title="投资参数设置" style={{ marginBottom: 24 }}>
        <Row gutter={[16, 16]}>
          <Col span={6}>
            <Text>总投资金额 (USD)</Text>
            <InputNumber
              style={{ width: '100%', marginTop: 8 }}
              value={totalInvestment}
              onChange={(value) => setTotalInvestment(value || 100000)}
              min={1000}
              step={10000}
              formatter={(value) =>
                `$ ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')
              }
              parser={(value) =>
                value ? parseFloat(value.replace(/\$\s?|(,*)/g, '')) : 0
              }
            />
          </Col>
          <Col span={6}>
            <Text>投资年限</Text>
            <InputNumber
              style={{ width: '100%', marginTop: 8 }}
              value={timeHorizon}
              onChange={(value) => setTimeHorizon(value || 10)}
              min={1}
              max={30}
              suffix="年"
            />
          </Col>
          <Col span={6}>
            <Text>SCHD 权重</Text>
            <InputNumber
              style={{ width: '100%', marginTop: 8 }}
              value={schdWeight}
              onChange={(value) => {
                setSchdWeight(value || 70);
                setJepqWeight(100 - (value || 70));
              }}
              min={0}
              max={100}
              suffix="%"
            />
          </Col>
          <Col span={6}>
            <Text>JEPQ 权重</Text>
            <InputNumber
              style={{ width: '100%', marginTop: 8 }}
              value={jepqWeight}
              onChange={(value) => {
                setJepqWeight(value || 30);
                setSchdWeight(100 - (value || 30));
              }}
              min={0}
              max={100}
              suffix="%"
            />
          </Col>
        </Row>
        <Divider />
        <Button
          type="primary"
          size="large"
          loading={loading}
          onClick={handleAnalyze}
          block
        >
          开始分析
        </Button>
      </Card>

      {result && (
        <>
          <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
            <Col span={8}>
              <Card>
                <Statistic
                  title="最终价值 (乐观)"
                  value={result.scenarios.optimistic?.final_value || 0}
                  precision={2}
                  prefix={<ArrowUpOutlined />}
                  suffix="USD"
                  valueStyle={{ color: '#3f8600' }}
                />
              </Card>
            </Col>
            <Col span={8}>
              <Card>
                <Statistic
                  title="最终价值 (中性)"
                  value={result.scenarios.neutral?.final_value || 0}
                  precision={2}
                  suffix="USD"
                />
              </Card>
            </Col>
            <Col span={8}>
              <Card>
                <Statistic
                  title="最终价值 (悲观)"
                  value={result.scenarios.pessimistic?.final_value || 0}
                  precision={2}
                  prefix={<ArrowDownOutlined />}
                  suffix="USD"
                  valueStyle={{ color: '#cf1322' }}
                />
              </Card>
            </Col>
          </Row>

          <Card title="情景对比" style={{ marginBottom: 24 }}>
            <Table
              dataSource={generateComparisonData()}
              columns={comparisonColumns}
              pagination={false}
              rowKey="metric"
            />
          </Card>

          <Card title="价值增长趋势" style={{ marginBottom: 24 }}>
            <ResponsiveContainer width="100%" height={400}>
              <LineChart data={generateProjectionChartData()}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="year" label={{ value: '年份', position: 'insideBottom', offset: -5 }} />
                <YAxis
                  label={{ value: '价值 (USD)', angle: -90, position: 'insideLeft' }}
                  tickFormatter={(value) => `$${(value / 1000).toFixed(0)}k`}
                />
                <Tooltip
                  formatter={(value: number) => formatCurrency(value)}
                  labelFormatter={(label) => `第 ${label} 年`}
                />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="optimistic"
                  stroke="#52c41a"
                  name="乐观"
                  strokeWidth={2}
                />
                <Line
                  type="monotone"
                  dataKey="neutral"
                  stroke="#1890ff"
                  name="中性"
                  strokeWidth={2}
                />
                <Line
                  type="monotone"
                  dataKey="pessimistic"
                  stroke="#ff4d4f"
                  name="悲观"
                  strokeWidth={2}
                />
              </LineChart>
            </ResponsiveContainer>
          </Card>

          <Card title="年度收益分布" style={{ marginBottom: 24 }}>
            <ResponsiveContainer width="100%" height={400}>
              <BarChart data={generateProjectionChartData()}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="year" label={{ value: '年份', position: 'insideBottom', offset: -5 }} />
                <YAxis
                  label={{ value: '价值 (USD)', angle: -90, position: 'insideLeft' }}
                  tickFormatter={(value) => `$${(value / 1000).toFixed(0)}k`}
                />
                <Tooltip
                  formatter={(value: number) => formatCurrency(value)}
                  labelFormatter={(label) => `第 ${label} 年`}
                />
                <Legend />
                <Bar dataKey="optimistic" fill="#52c41a" name="乐观" />
                <Bar dataKey="neutral" fill="#1890ff" name="中性" />
                <Bar dataKey="pessimistic" fill="#ff4d4f" name="悲观" />
              </BarChart>
            </ResponsiveContainer>
          </Card>

          <Collapse style={{ marginBottom: 24 }}>
            <Panel header="各情景详细预测" key="1">
              {['pessimistic', 'neutral', 'optimistic'].map((scenario) => (
                <div key={scenario} style={{ marginBottom: 24 }}>
                  <Title level={4}>
                    <Tag color={getScenarioColor(scenario)}>
                      {getScenarioName(scenario)}
                    </Tag>
                  </Title>
                  <Row gutter={16} style={{ marginBottom: 16 }}>
                    <Col span={6}>
                      <Statistic
                        title="平均年化收益"
                        value={result.scenarios[scenario]?.avg_annual_return || 0}
                        precision={2}
                        suffix="%"
                      />
                    </Col>
                    <Col span={6}>
                      <Statistic
                        title="波动率"
                        value={result.scenarios[scenario]?.assumptions.volatility || 0}
                        precision={2}
                        suffix="%"
                      />
                    </Col>
                    <Col span={6}>
                      <Statistic
                        title="夏普比率"
                        value={result.scenarios[scenario]?.assumptions.sharpe_ratio || 0}
                        precision={2}
                      />
                    </Col>
                    <Col span={6}>
                      <Statistic
                        title="最大回撤"
                        value={result.scenarios[scenario]?.assumptions.max_drawdown || 0}
                        precision={2}
                        suffix="%"
                      />
                    </Col>
                  </Row>
                  <Table
                    dataSource={result.scenarios[scenario]?.projections || []}
                    columns={projectionColumns}
                    pagination={false}
                    rowKey="year"
                    size="small"
                  />
                </div>
              ))}
            </Panel>
            <Panel header="方法论与假设" key="2">
              <Title level={4}>方法论</Title>
              <Paragraph>{result.methodology}</Paragraph>
              <Title level={4}>假设条件</Title>
              <Paragraph style={{ whiteSpace: 'pre-line' }}>
                {result.assumptions}
              </Paragraph>
              <Title level={4}>局限性</Title>
              <Paragraph style={{ whiteSpace: 'pre-line' }}>
                {result.limitations}
              </Paragraph>
            </Panel>
          </Collapse>
        </>
      )}
    </div>
  );
};

export default PortfolioAnalysis;

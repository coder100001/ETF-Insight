import React, { useState } from 'react';
import { Card, Row, Col, Button, Select, InputNumber, message, Table, Tabs, Spin, Alert, Slider, Form } from 'antd';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip as RechartsTooltip, Legend, ScatterChart, Scatter, XAxis, YAxis, CartesianGrid, ZAxis } from 'recharts';
import { optimizationAPI } from '../services/api';
import Layout from '../components/Layout';
import styled from 'styled-components';

const { TabPane } = Tabs;
const { Option } = Select;

const Container = styled.div`
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
`;

const Title = styled.h1`
  font-size: 28px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24px;
`;

const StyledCard = styled(Card)`
  margin-bottom: 24px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
`;

const MetricCard = styled.div`
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 20px;
  border-radius: 12px;
  text-align: center;
  margin-bottom: 16px;
`;

const MetricValue = styled.div`
  font-size: 32px;
  font-weight: 700;
  margin-bottom: 8px;
`;

const MetricLabel = styled.div`
  font-size: 14px;
  opacity: 0.9;
`;

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82CA9D', '#FFC658', '#FF6B6B'];

// 预设ETF数据
const PRESET_ETFS = [
  { symbol: 'VTI', name: 'Vanguard Total Stock Market', return: 0.12, volatility: 0.16 },
  { symbol: 'VOO', name: 'Vanguard S&P 500', return: 0.11, volatility: 0.15 },
  { symbol: 'QQQ', name: 'Invesco QQQ Trust', return: 0.14, volatility: 0.20 },
  { symbol: 'IWM', name: 'iShares Russell 2000', return: 0.09, volatility: 0.19 },
  { symbol: 'EFA', name: 'iShares MSCI EAFE', return: 0.08, volatility: 0.17 },
  { symbol: 'EEM', name: 'iShares Emerging Markets', return: 0.10, volatility: 0.22 },
  { symbol: 'AGG', name: 'iShares Core U.S. Aggregate Bond', return: 0.04, volatility: 0.05 },
  { symbol: 'TLT', name: 'iShares 20+ Year Treasury Bond', return: 0.03, volatility: 0.14 },
  { symbol: 'GLD', name: 'SPDR Gold Shares', return: 0.06, volatility: 0.15 },
  { symbol: 'VNQ', name: 'Vanguard Real Estate', return: 0.08, volatility: 0.18 },
];

interface OptimizationResult {
  weights: Record<string, number>;
  expected_return: number;
  volatility: number;
  sharpe_ratio: number;
  sortino_ratio: number;
  diversification_ratio: number;
  risk_contribution: Record<string, number>;
  herfindahl_index: number;
}

interface EfficientFrontierPoint {
  target_return: number;
  min_volatility: number;
  optimal_weights: Record<string, number>;
  sharpe_ratio: number;
}

interface RiskParityResult {
  weights: Record<string, number>;
  risk_contributions: Record<string, number>;
  expected_return: number;
  volatility: number;
  leverage: number;
  target_volatility: number;
  diversification_ratio: number;
}

interface BlackLittermanResult {
  prior_returns: Record<string, number>;
  posterior_returns: Record<string, number>;
  implied_returns: Record<string, number>;
  optimal_weights: Record<string, number>;
  expected_return: number;
  volatility: number;
  sharpe_ratio: number;
  confidence: number;
}

const PortfolioOptimization: React.FC = () => {
  const [selectedETFs, setSelectedETFs] = useState<string[]>(['VTI', 'VOO', 'AGG']);
  const [objective, setObjective] = useState<'min_volatility' | 'max_sharpe' | 'target_return'>('max_sharpe');
  const [targetReturn, setTargetReturn] = useState<number>(0.10);
  const [riskFreeRate, setRiskFreeRate] = useState<number>(0.045);
  const [minWeights, setMinWeights] = useState<Record<string, number>>({});
  const [maxWeights, setMaxWeights] = useState<Record<string, number>>({});
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<OptimizationResult | null>(null);
  const [frontier, setFrontier] = useState<EfficientFrontierPoint[]>([]);
  const [activeTab, setActiveTab] = useState('1');

  // 风险平价状态
  const [rpMethod, setRpMethod] = useState<'parity' | 'inverse_vol' | 'budget'>('parity');
  const [rpResult, setRpResult] = useState<RiskParityResult | null>(null);
  const [targetVol, setTargetVol] = useState<number>(0.10);
  const [useLeverage] = useState<boolean>(false);

  // Black-Litterman状态
  const [blResult, setBlResult] = useState<BlackLittermanResult | null>(null);
  const [absoluteViews, setAbsoluteViews] = useState<Record<string, number>>({});
  const [relativeViews] = useState<Array<{asset1: string; asset2: string; expected_diff: number; confidence: number}>>([]);
  const [tau, setTau] = useState<number>(0.025);
  const [riskAversion, setRiskAversion] = useState<number>(2.5);

  // 生成协方差矩阵（简化模型）
  const generateCovarianceMatrix = (symbols: string[]) => {
    const matrix: Record<string, Record<string, number>> = {};

    symbols.forEach(s1 => {
      matrix[s1] = {};
      const etf1 = PRESET_ETFS.find(e => e.symbol === s1);

      symbols.forEach(s2 => {
        const etf2 = PRESET_ETFS.find(e => e.symbol === s2);

        if (s1 === s2) {
          // 对角线：方差 = 波动率^2
          matrix[s1][s2] = (etf1?.volatility || 0.15) ** 2;
        } else {
          // 非对角线：协方差 = 相关系数 * 波动率1 * 波动率2
          // 假设股票-股票相关性0.8，股票-债券0.2，债券-债券0.9
          let correlation = 0.5;
          const isBond1 = s1.includes('AGG') || s1.includes('TLT');
          const isBond2 = s2.includes('AGG') || s2.includes('TLT');

          if (isBond1 && isBond2) correlation = 0.9;
          else if (isBond1 || isBond2) correlation = 0.2;
          else correlation = 0.8;

          matrix[s1][s2] = correlation * (etf1?.volatility || 0.15) * (etf2?.volatility || 0.15);
        }
      });
    });

    return matrix;
  };

  // 生成收益率映射
  const generateReturns = (symbols: string[]) => {
    const returns: Record<string, number> = {};
    symbols.forEach(symbol => {
      const etf = PRESET_ETFS.find(e => e.symbol === symbol);
      returns[symbol] = etf?.return || 0.10;
    });
    return returns;
  };

  const handleOptimize = async () => {
    if (selectedETFs.length < 2) {
      message.warning('请至少选择2个ETF');
      return;
    }

    setLoading(true);
    try {
      const returns = generateReturns(selectedETFs);
      const covMatrix = generateCovarianceMatrix(selectedETFs);

      const response = await optimizationAPI.mptOptimize({
        symbols: selectedETFs,
        returns,
        cov_matrix: covMatrix,
        objective,
        target_return: targetReturn,
        risk_free_rate: riskFreeRate,
        constraints: {
          min_weights: minWeights,
          max_weights: maxWeights,
          allow_short: false,
        },
      });

      if (response.success && response.data) {
        setResult(response.data);
        message.success('优化完成');
        setActiveTab('2');
      } else {
        message.error(response.message || '优化失败');
      }
    } catch (error) {
      message.error('优化请求失败');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const handleCalculateFrontier = async () => {
    if (selectedETFs.length < 2) {
      message.warning('请至少选择2个ETF');
      return;
    }

    setLoading(true);
    try {
      const returns = generateReturns(selectedETFs);
      const covMatrix = generateCovarianceMatrix(selectedETFs);

      const response = await optimizationAPI.efficientFrontier({
        symbols: selectedETFs,
        returns,
        cov_matrix: covMatrix,
        num_points: 20,
        constraints: {
          min_weights: minWeights,
          max_weights: maxWeights,
        },
      });

      if (response.success && response.data) {
        setFrontier(response.data);
        message.success('有效前沿计算完成');
        setActiveTab('3');
      } else {
        message.error(response.message || '计算失败');
      }
    } catch (error) {
      message.error('计算请求失败');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  // 风险平价优化
  const handleRiskParityOptimize = async () => {
    if (selectedETFs.length < 2) {
      message.warning('请至少选择2个ETF');
      return;
    }

    setLoading(true);
    try {
      const returns = generateReturns(selectedETFs);
      const covMatrix = generateCovarianceMatrix(selectedETFs);

      const response = await optimizationAPI.riskParityOptimize({
        symbols: selectedETFs,
        returns,
        cov_matrix: covMatrix,
        method: rpMethod,
        constraints: {
          min_weights: minWeights,
          max_weights: maxWeights,
          target_volatility: targetVol,
          use_leverage: useLeverage,
          max_leverage: 2.0,
        },
      });

      if (response.success && response.data) {
        setRpResult(response.data);
        message.success('风险平价优化完成');
        setActiveTab('4');
      } else {
        message.error(response.message || '优化失败');
      }
    } catch (error) {
      message.error('优化请求失败');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  // Black-Litterman优化
  const handleBlackLittermanOptimize = async () => {
    if (selectedETFs.length < 2) {
      message.warning('请至少选择2个ETF');
      return;
    }

    setLoading(true);
    try {
      // 构建市场权重（等权重）
      const marketWeights: Record<string, number> = {};
      selectedETFs.forEach(symbol => {
        marketWeights[symbol] = 1.0 / selectedETFs.length;
      });

      const covMatrix = generateCovarianceMatrix(selectedETFs);

      const response = await optimizationAPI.blackLittermanOptimize({
        market_weights: marketWeights,
        cov_matrix: covMatrix,
        absolute_views: absoluteViews,
        relative_views: relativeViews,
        risk_aversion: riskAversion,
        tau: tau,
        risk_free_rate: riskFreeRate,
        constraints: {
          min_weights: minWeights,
          max_weights: maxWeights,
        },
      });

      if (response.success && response.data) {
        setBlResult(response.data);
        message.success('Black-Litterman优化完成');
        setActiveTab('5');
      } else {
        message.error(response.message || '优化失败');
      }
    } catch (error) {
      message.error('优化请求失败');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const weightData = result ? Object.entries(result.weights)
    .filter(([, weight]) => weight > 0.001)
    .map(([symbol, weight]) => ({
      name: symbol,
      value: weight * 100,
    })) : [];

  const riskContributionData = result ? Object.entries(result.risk_contribution)
    .map(([symbol, contribution]) => ({
      symbol,
      contribution: contribution * 100,
    }))
    .sort((a, b) => b.contribution - a.contribution) : [];

  const frontierData = frontier.map(point => ({
    volatility: point.min_volatility * 100,
    return: point.target_return * 100,
    sharpe: point.sharpe_ratio,
  }));

  const weightColumns = [
    {
      title: 'ETF代码',
      dataIndex: 'symbol',
      key: 'symbol',
    },
    {
      title: '权重',
      dataIndex: 'weight',
      key: 'weight',
      render: (value: number) => `${(value * 100).toFixed(2)}%`,
    },
    {
      title: '风险贡献',
      dataIndex: 'riskContribution',
      key: 'riskContribution',
      render: (value: number) => `${(value * 100).toFixed(2)}%`,
    },
  ];

  const weightTableData = result ? Object.entries(result.weights)
    .map(([symbol, weight]) => ({
      symbol,
      weight,
      riskContribution: result.risk_contribution[symbol] || 0,
    }))
    .sort((a, b) => b.weight - a.weight) : [];

  return (
    <Layout>
      <Container>
        <Title>投资组合优化 (MPT / 风险平价 / Black-Litterman)</Title>

        <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab="参数配置" key="1">
          <StyledCard title="选择ETF">
            <Select
              mode="multiple"
              style={{ width: '100%' }}
              placeholder="选择ETF构建组合"
              value={selectedETFs}
              onChange={setSelectedETFs}
            >
              {PRESET_ETFS.map(etf => (
                <Option key={etf.symbol} value={etf.symbol}>
                  {etf.symbol} - {etf.name} (预期收益: {(etf.return * 100).toFixed(1)}%, 波动率: {(etf.volatility * 100).toFixed(1)}%)
                </Option>
              ))}
            </Select>
          </StyledCard>

          <StyledCard title="优化目标">
            <Form layout="vertical">
              <Form.Item label="优化目标">
                <Select value={objective} onChange={setObjective} style={{ width: 200 }}>
                  <Option value="max_sharpe">最大化夏普比率</Option>
                  <Option value="min_volatility">最小化波动率</Option>
                  <Option value="target_return">目标收益率</Option>
                </Select>
              </Form.Item>

              {objective === 'target_return' && (
                <Form.Item label="目标收益率 (%)">
                  <Slider
                    min={0}
                    max={20}
                    step={0.5}
                    value={targetReturn * 100}
                    onChange={(value) => setTargetReturn(value / 100)}
                    marks={{ 0: '0%', 10: '10%', 20: '20%' }}
                  />
                </Form.Item>
              )}

              <Form.Item label="无风险利率 (%)">
                <InputNumber
                  min={0}
                  max={10}
                  step={0.1}
                  value={riskFreeRate * 100}
                  onChange={(value) => setRiskFreeRate((value || 0) / 100)}
                  formatter={(value) => `${value}%`}
                  parser={(value) => parseFloat(value?.replace('%', '') || '0')}
                />
              </Form.Item>
            </Form>
          </StyledCard>

          <StyledCard title="权重约束">
            <Alert
              message="权重约束设置"
              description="可以为每个ETF设置最小和最大权重限制。如果不设置，默认为0-100%。"
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />
            <Row gutter={[16, 16]}>
              {selectedETFs.map(symbol => (
                <Col span={12} key={symbol}>
                  <Card size="small" title={symbol}>
                    <div style={{ marginBottom: 8 }}>
                      <span>最小权重: </span>
                      <InputNumber
                        min={0}
                        max={100}
                        value={(minWeights[symbol] || 0) * 100}
                        onChange={(value) => setMinWeights({ ...minWeights, [symbol]: (value || 0) / 100 })}
                        formatter={(value) => `${value}%`}
                        parser={(value) => parseFloat(value?.replace('%', '') || '0')}
                        style={{ width: 80 }}
                      />
                    </div>
                    <div>
                      <span>最大权重: </span>
                      <InputNumber
                        min={0}
                        max={100}
                        value={(maxWeights[symbol] || 1) * 100}
                        onChange={(value) => setMaxWeights({ ...maxWeights, [symbol]: (value || 0) / 100 })}
                        formatter={(value) => `${value}%`}
                        parser={(value) => parseFloat(value?.replace('%', '') || '0')}
                        style={{ width: 80 }}
                      />
                    </div>
                  </Card>
                </Col>
              ))}
            </Row>
          </StyledCard>

          <Button
            type="primary"
            size="large"
            onClick={handleOptimize}
            loading={loading}
            style={{ marginRight: 16 }}
          >
            执行优化
          </Button>
          <Button
            size="large"
            onClick={handleCalculateFrontier}
            loading={loading}
          >
            计算有效前沿
          </Button>
        </TabPane>

        <TabPane tab="优化结果" key="2">
          {result ? (
            <>
              <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
                <Col span={6}>
                  <MetricCard>
                    <MetricValue>{(result.expected_return * 100).toFixed(2)}%</MetricValue>
                    <MetricLabel>预期年化收益率</MetricLabel>
                  </MetricCard>
                </Col>
                <Col span={6}>
                  <MetricCard style={{ background: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)' }}>
                    <MetricValue>{(result.volatility * 100).toFixed(2)}%</MetricValue>
                    <MetricLabel>年化波动率</MetricLabel>
                  </MetricCard>
                </Col>
                <Col span={6}>
                  <MetricCard style={{ background: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)' }}>
                    <MetricValue>{result.sharpe_ratio.toFixed(2)}</MetricValue>
                    <MetricLabel>夏普比率</MetricLabel>
                  </MetricCard>
                </Col>
                <Col span={6}>
                  <MetricCard style={{ background: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)' }}>
                    <MetricValue>{result.sortino_ratio.toFixed(2)}</MetricValue>
                    <MetricLabel>索提诺比率</MetricLabel>
                  </MetricCard>
                </Col>
              </Row>

              <Row gutter={[24, 24]}>
                <Col span={12}>
                  <StyledCard title="资产配置">
                    <ResponsiveContainer width="100%" height={300}>
                      <PieChart>
                        <Pie
                          data={weightData}
                          cx="50%"
                          cy="50%"
                          labelLine={false}
                          label={({ name, value }) => `${name}: ${value.toFixed(1)}%`}
                          outerRadius={100}
                          fill="#8884d8"
                          dataKey="value"
                        >
                          {weightData.map((_, index) => (
                            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                          ))}
                        </Pie>
                        <RechartsTooltip />
                        <Legend />
                      </PieChart>
                    </ResponsiveContainer>
                  </StyledCard>
                </Col>
                <Col span={12}>
                  <StyledCard title="风险贡献">
                    <ResponsiveContainer width="100%" height={300}>
                      <PieChart>
                        <Pie
                          data={riskContributionData}
                          cx="50%"
                          cy="50%"
                          innerRadius={60}
                          outerRadius={100}
                          paddingAngle={5}
                          dataKey="contribution"
                          label={({ name, value }) => `${name}: ${Number(value).toFixed(1)}%`}
                        >
                          {riskContributionData.map((_, index) => (
                            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                          ))}
                        </Pie>
                        <RechartsTooltip />
                      </PieChart>
                    </ResponsiveContainer>
                  </StyledCard>
                </Col>
              </Row>

              <StyledCard title="详细数据">
                <Table
                  dataSource={weightTableData}
                  columns={weightColumns}
                  rowKey="symbol"
                  pagination={false}
                />
              </StyledCard>

              <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
                <Col span={12}>
                  <StyledCard title="组合指标">
                    <p><strong>分散化比率:</strong> {result.diversification_ratio.toFixed(2)}</p>
                    <p><strong>赫芬达尔指数:</strong> {result.herfindahl_index.toFixed(4)}</p>
                    <p style={{ color: '#666', fontSize: 12 }}>
                      分散化比率 &gt; 1 表示有分散化效益；赫芬达尔指数越接近0表示分散度越高
                    </p>
                  </StyledCard>
                </Col>
              </Row>
            </>
          ) : (
            <Alert message="请先执行优化" type="info" showIcon />
          )}
        </TabPane>

        <TabPane tab="有效前沿" key="3">
          {frontier.length > 0 ? (
            <StyledCard title="有效前沿曲线">
              <ResponsiveContainer width="100%" height={500}>
                <ScatterChart margin={{ top: 20, right: 20, bottom: 20, left: 20 }}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis
                    type="number"
                    dataKey="volatility"
                    name="波动率"
                    unit="%"
                    domain={['dataMin - 1', 'dataMax + 1']}
                    label={{ value: '年化波动率 (%)', position: 'insideBottom', offset: -10 }}
                  />
                  <YAxis
                    type="number"
                    dataKey="return"
                    name="收益率"
                    unit="%"
                    domain={['dataMin - 1', 'dataMax + 1']}
                    label={{ value: '预期年化收益率 (%)', angle: -90, position: 'insideLeft' }}
                  />
                  <ZAxis type="number" dataKey="sharpe" range={[50, 400]} name="夏普比率" />
                  <RechartsTooltip
                    cursor={{ strokeDasharray: '3 3' }}
                  />
                  <Legend />
                  <Scatter
                    name="有效前沿"
                    data={frontierData}
                    fill="#8884d8"
                    shape="circle"
                  />
                </ScatterChart>
              </ResponsiveContainer>
              <p style={{ textAlign: 'center', color: '#666', marginTop: 16 }}>
                圆点大小代表夏普比率，越大表示风险调整后收益越好
              </p>
            </StyledCard>
          ) : (
            <Alert message="请先计算有效前沿" type="info" showIcon />
          )}
        </TabPane>

        <TabPane tab="风险平价" key="4">
          <Row gutter={[24, 24]}>
            <Col span={8}>
              <StyledCard title="风险平价配置">
                <Form layout="vertical">
                  <Form.Item label="优化方法">
                    <Select value={rpMethod} onChange={setRpMethod} style={{ width: '100%' }}>
                      <Option value="parity">风险平价 (等风险贡献)</Option>
                      <Option value="inverse_vol">逆波动率加权</Option>
                      <Option value="budget">风险预算</Option>
                    </Select>
                  </Form.Item>
                  <Form.Item label="目标波动率 (%)" help="设置目标波动率以调整杠杆">
                    <Slider
                      min={5}
                      max={30}
                      step={1}
                      value={targetVol * 100}
                      onChange={(value) => setTargetVol(value / 100)}
                      marks={{ 5: '5%', 15: '15%', 30: '30%' }}
                    />
                  </Form.Item>
                  <Form.Item>
                    <Button type="primary" onClick={handleRiskParityOptimize} loading={loading} block>
                      执行风险平价优化
                    </Button>
                  </Form.Item>
                </Form>
              </StyledCard>
            </Col>
            <Col span={16}>
              {rpResult ? (
                <>
                  <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
                    <Col span={8}>
                      <MetricCard>
                        <MetricValue>{(rpResult.expected_return * 100).toFixed(2)}%</MetricValue>
                        <MetricLabel>预期收益</MetricLabel>
                      </MetricCard>
                    </Col>
                    <Col span={8}>
                      <MetricCard style={{ background: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)' }}>
                        <MetricValue>{(rpResult.volatility * 100).toFixed(2)}%</MetricValue>
                        <MetricLabel>波动率</MetricLabel>
                      </MetricCard>
                    </Col>
                    <Col span={8}>
                      <MetricCard style={{ background: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)' }}>
                        <MetricValue>{rpResult.diversification_ratio.toFixed(2)}</MetricValue>
                        <MetricLabel>分散化比率</MetricLabel>
                      </MetricCard>
                    </Col>
                  </Row>
                  <StyledCard title="风险平价权重配置">
                    <Table
                      dataSource={Object.entries(rpResult.weights).map(([symbol, weight]) => ({
                        symbol,
                        weight,
                        riskContribution: rpResult.risk_contributions[symbol] || 0,
                      })).sort((a, b) => b.weight - a.weight)}
                      columns={[
                        { title: 'ETF', dataIndex: 'symbol', key: 'symbol' },
                        { title: '权重', dataIndex: 'weight', key: 'weight', render: (v: number) => `${(v * 100).toFixed(2)}%` },
                        { title: '风险贡献', dataIndex: 'riskContribution', key: 'riskContribution', render: (v: number) => `${(v * 100).toFixed(2)}%` },
                      ]}
                      rowKey="symbol"
                      pagination={false}
                    />
                  </StyledCard>
                </>
              ) : (
                <Alert message="请先执行风险平价优化" type="info" showIcon />
              )}
            </Col>
          </Row>
        </TabPane>

        <TabPane tab="Black-Litterman" key="5">
          <Row gutter={[24, 24]}>
            <Col span={8}>
              <StyledCard title="Black-Litterman配置">
                <Form layout="vertical">
                  <Form.Item label="Tau (缩放参数)" help="通常取0.025-0.05">
                    <InputNumber
                      min={0.001}
                      max={0.1}
                      step={0.001}
                      value={tau}
                      onChange={(v) => setTau(v || 0.025)}
                      style={{ width: '100%' }}
                    />
                  </Form.Item>
                  <Form.Item label="风险厌恶系数" help="通常取2-4">
                    <Slider
                      min={1}
                      max={5}
                      step={0.5}
                      value={riskAversion}
                      onChange={setRiskAversion}
                    />
                  </Form.Item>
                  <Form.Item label="绝对观点 (资产→预期收益)">
                    <Select
                      mode="multiple"
                      placeholder="选择要设置观点的资产"
                      style={{ width: '100%', marginBottom: 8 }}
                      onChange={(symbols) => {
                        const newViews: Record<string, number> = {};
                        symbols.forEach((s: string) => {
                          newViews[s] = absoluteViews[s] || 0.1;
                        });
                        setAbsoluteViews(newViews);
                      }}
                    >
                      {selectedETFs.map(s => <Option key={s} value={s}>{s}</Option>)}
                    </Select>
                    {Object.entries(absoluteViews).map(([symbol, view]) => (
                      <div key={symbol} style={{ marginBottom: 8 }}>
                        <span>{symbol}: </span>
                        <InputNumber
                          min={-0.5}
                          max={0.5}
                          step={0.01}
                          value={view}
                          onChange={(v) => setAbsoluteViews({ ...absoluteViews, [symbol]: v || 0 })}
                          formatter={(v) => `${(v * 100).toFixed(0)}%`}
                          parser={(v) => parseFloat(v?.replace('%', '') || '0') / 100}
                          style={{ width: 100 }}
                        />
                      </div>
                    ))}
                  </Form.Item>
                  <Form.Item>
                    <Button type="primary" onClick={handleBlackLittermanOptimize} loading={loading} block>
                      执行Black-Litterman优化
                    </Button>
                  </Form.Item>
                </Form>
              </StyledCard>
            </Col>
            <Col span={16}>
              {blResult ? (
                <>
                  <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
                    <Col span={8}>
                      <MetricCard>
                        <MetricValue>{(blResult.expected_return * 100).toFixed(2)}%</MetricValue>
                        <MetricLabel>预期收益</MetricLabel>
                      </MetricCard>
                    </Col>
                    <Col span={8}>
                      <MetricCard style={{ background: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)' }}>
                        <MetricValue>{(blResult.volatility * 100).toFixed(2)}%</MetricValue>
                        <MetricLabel>波动率</MetricLabel>
                      </MetricCard>
                    </Col>
                    <Col span={8}>
                      <MetricCard style={{ background: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)' }}>
                        <MetricValue>{blResult.sharpe_ratio.toFixed(2)}</MetricValue>
                        <MetricLabel>夏普比率</MetricLabel>
                      </MetricCard>
                    </Col>
                  </Row>
                  <StyledCard title="收益估计对比">
                    <Table
                      dataSource={selectedETFs.map(symbol => ({
                        symbol,
                        prior: blResult.prior_returns[symbol] || 0,
                        posterior: blResult.posterior_returns[symbol] || 0,
                      }))}
                      columns={[
                        { title: 'ETF', dataIndex: 'symbol', key: 'symbol' },
                        { title: '先验收益', dataIndex: 'prior', key: 'prior', render: (v: number) => `${(v * 100).toFixed(2)}%` },
                        { title: '后验收益', dataIndex: 'posterior', key: 'posterior', render: (v: number) => `${(v * 100).toFixed(2)}%` },
                      ]}
                      rowKey="symbol"
                      pagination={false}
                    />
                  </StyledCard>
                  <StyledCard title="最优权重" style={{ marginTop: 16 }}>
                    <Table
                      dataSource={Object.entries(blResult.optimal_weights).map(([symbol, weight]) => ({
                        symbol,
                        weight,
                      })).sort((a, b) => b.weight - a.weight)}
                      columns={[
                        { title: 'ETF', dataIndex: 'symbol', key: 'symbol' },
                        { title: '权重', dataIndex: 'weight', key: 'weight', render: (v: number) => `${(v * 100).toFixed(2)}%` },
                      ]}
                      rowKey="symbol"
                      pagination={false}
                    />
                  </StyledCard>
                </>
              ) : (
                <Alert message="请先执行Black-Litterman优化" type="info" showIcon />
              )}
            </Col>
          </Row>
        </TabPane>
      </Tabs>

        {loading && (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <Spin size="large" />
            <p style={{ marginTop: 16 }}>正在计算最优组合...</p>
          </div>
        )}
      </Container>
    </Layout>
  );
};

export default PortfolioOptimization;

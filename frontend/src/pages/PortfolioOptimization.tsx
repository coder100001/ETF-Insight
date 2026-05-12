import React, { useState } from 'react';
import { Card, Row, Col, Button, Select, InputNumber, message, Table, Tabs, Spin, Alert, Slider, Form } from 'antd';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip as RechartsTooltip, Legend, ScatterChart, Scatter, XAxis, YAxis, CartesianGrid, ZAxis } from 'recharts';
import Layout from '../components/Layout';
import styled from 'styled-components';
import { useOptimization } from '../hooks/useOptimization';
import { useETFData } from '../hooks/useETFData';
import type {
  OptimizationResult,
  EfficientFrontierPoint,
  RiskParityResult,
  BlackLittermanResult
} from '../types';

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

// 权重约束组件
const WeightConstraints: React.FC<{
  selectedETFs: string[];
  minWeights: Record<string, number>;
  maxWeights: Record<string, number>;
  onMinWeightsChange: (weights: Record<string, number>) => void;
  onMaxWeightsChange: (weights: Record<string, number>) => void;
}> = ({ selectedETFs, minWeights, maxWeights, onMinWeightsChange, onMaxWeightsChange }) => (
  <StyledCard title="权重约束">
    <Alert
      title="权重约束设置"
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
                onChange={(value) => onMinWeightsChange({ ...minWeights, [symbol]: (value || 0) / 100 })}
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
                onChange={(value) => onMaxWeightsChange({ ...maxWeights, [symbol]: (value || 0) / 100 })}
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
);

// 优化结果展示组件
const OptimizationResultDisplay: React.FC<{ result: OptimizationResult }> = ({ result }) => {
  const weightData = Object.entries(result.weights)
    .filter(([, weight]) => weight > 0.001)
    .map(([symbol, weight]) => ({
      name: symbol,
      value: weight * 100,
    }));

  const riskContributionData = Object.entries(result.risk_contribution)
    .map(([symbol, contribution]) => ({
      symbol,
      contribution: contribution * 100,
    }))
    .sort((a, b) => b.contribution - a.contribution);

  const weightTableData = Object.entries(result.weights)
    .map(([symbol, weight]) => ({
      symbol,
      weight,
      riskContribution: result.risk_contribution[symbol] || 0,
    }))
    .sort((a, b) => b.weight - a.weight);

  return (
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
          columns={[
            { title: 'ETF代码', dataIndex: 'symbol', key: 'symbol' },
            { title: '权重', dataIndex: 'weight', key: 'weight', render: (value: number) => `${(value * 100).toFixed(2)}%` },
            { title: '风险贡献', dataIndex: 'riskContribution', key: 'riskContribution', render: (value: number) => `${(value * 100).toFixed(2)}%` },
          ]}
          rowKey="symbol"
          pagination={false}
        />
      </StyledCard>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col span={12}>
          <StyledCard title="组合指标">
            <p><strong>分散化比率:</strong> {result.diversification_ratio.toFixed(2)}</p>
            <p style={{ color: '#666', fontSize: 12 }}>
              分散化比率 &gt; 1 表示有分散化效益
            </p>
          </StyledCard>
        </Col>
      </Row>
    </>
  );
};

// 有效前沿展示组件
const EfficientFrontierDisplay: React.FC<{ frontier: EfficientFrontierPoint[] }> = ({ frontier }) => {
  const frontierData = frontier.map(point => ({
    volatility: point.min_volatility * 100,
    return: point.target_return * 100,
    sharpe: point.sharpe_ratio,
  }));

  const volMin = Math.min(...frontierData.map(d => d.volatility));
  const volMax = Math.max(...frontierData.map(d => d.volatility));
  const retMin = Math.min(...frontierData.map(d => d.return));
  const retMax = Math.max(...frontierData.map(d => d.return));

  return (
    <StyledCard title="有效前沿曲线">
      <ResponsiveContainer width="100%" height={500}>
        <ScatterChart margin={{ top: 20, right: 20, bottom: 50, left: 60 }}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis
            type="number"
            dataKey="volatility"
            name="波动率"
            unit="%"
            domain={[Math.max(0, volMin - 0.1), volMax + 0.1]}
            tickFormatter={(v) => v.toFixed(2)}
            label={{ value: '年化波动率 (%)', position: 'insideBottom', offset: -10 }}
          />
          <YAxis
            type="number"
            dataKey="return"
            name="收益率"
            unit="%"
            domain={[Math.max(0, retMin - 1), retMax + 1]}
            tickFormatter={(v) => v.toFixed(1)}
            label={{ value: '预期年化收益率 (%)', angle: -90, position: 'insideLeft' }}
          />
          <ZAxis type="number" dataKey="sharpe" range={[50, 400]} name="夏普比率" />
          <RechartsTooltip cursor={{ strokeDasharray: '3 3' }} formatter={(value, name) => {
            const numValue = Array.isArray(value) ? value[0] : value;
            const nameStr = String(name);
            if (nameStr === '波动率') return [`${Number(numValue).toFixed(2)}%`, nameStr];
            if (nameStr === '收益率') return [`${Number(numValue).toFixed(2)}%`, nameStr];
            return [Number(numValue).toFixed(2), nameStr];
          }} />
          <Legend />
          <Scatter name="有效前沿" data={frontierData} fill="#8884d8" shape="circle" />
        </ScatterChart>
      </ResponsiveContainer>
      <p style={{ textAlign: 'center', color: '#666', marginTop: 16 }}>
        圆点大小代表夏普比率，越大表示风险调整后收益越好
      </p>
    </StyledCard>
  );
};

// 风险平价结果展示组件
const RiskParityDisplay: React.FC<{ result: RiskParityResult }> = ({ result }) => (
  <>
    <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
      <Col span={12}>
        <MetricCard style={{ background: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)' }}>
          <MetricValue>{(result.volatility * 100).toFixed(2)}%</MetricValue>
          <MetricLabel>波动率</MetricLabel>
        </MetricCard>
      </Col>
      <Col span={12}>
        <MetricCard style={{ background: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)' }}>
          <MetricValue>{result.diversification_ratio.toFixed(2)}</MetricValue>
          <MetricLabel>分散化比率</MetricLabel>
        </MetricCard>
      </Col>
    </Row>
    <StyledCard title="风险平价权重配置">
      <Table
        dataSource={Object.entries(result.weights).map(([symbol, weight]) => ({
          symbol,
          weight,
          riskContribution: result.risk_contributions[symbol] || 0,
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
);

// Black-Litterman结果展示组件
const BlackLittermanDisplay: React.FC<{ result: BlackLittermanResult; selectedETFs: string[] }> = ({ result, selectedETFs }) => (
  <>
    <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
      <Col span={8}>
        <MetricCard>
          <MetricValue>{(result.expected_return * 100).toFixed(2)}%</MetricValue>
          <MetricLabel>预期收益</MetricLabel>
        </MetricCard>
      </Col>
      <Col span={12}>
        <MetricCard style={{ background: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)' }}>
          <MetricValue>{result.sharpe_ratio.toFixed(2)}</MetricValue>
          <MetricLabel>夏普比率</MetricLabel>
        </MetricCard>
      </Col>
    </Row>
    <StyledCard title="后验收益估计">
      <Table
        dataSource={selectedETFs.map(symbol => ({
          symbol,
          posterior: result.posterior_returns[symbol] || 0,
        }))}
        columns={[
          { title: 'ETF', dataIndex: 'symbol', key: 'symbol' },
          { title: '后验收益', dataIndex: 'posterior', key: 'posterior', render: (v: number) => `${(v * 100).toFixed(2)}%` },
        ]}
        rowKey="symbol"
        pagination={false}
      />
    </StyledCard>
    <StyledCard title="最优权重" style={{ marginTop: 16 }}>
      <Table
        dataSource={Object.entries(result.optimal_weights).map(([symbol, weight]) => ({
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
);

const PortfolioOptimization: React.FC = () => {
  const [selectedETFs, setSelectedETFs] = useState<string[]>(['VTI', 'VOO', 'AGG']);
  const [objective, setObjective] = useState<'min_volatility' | 'max_sharpe' | 'target_return'>('max_sharpe');
  const [targetReturn, setTargetReturn] = useState<number>(0.10);
  const [riskFreeRate, setRiskFreeRate] = useState<number>(0.045);
  const [minWeights, setMinWeights] = useState<Record<string, number>>({});
  const [maxWeights, setMaxWeights] = useState<Record<string, number>>({});
  const [activeTab, setActiveTab] = useState('1');

  // 风险平价状态
  const [rpMethod, setRpMethod] = useState<'parity' | 'inverse_vol' | 'budget'>('parity');
  const [targetVol, setTargetVol] = useState<number>(0.10);
  const [riskBudget, setRiskBudget] = useState<Record<string, number>>({});

  // Black-Litterman状态
  const [absoluteViews, setAbsoluteViews] = useState<Record<string, number>>({});
  const [tau, setTau] = useState<number>(0.025);
  const [riskAversion, setRiskAversion] = useState<number>(2.5);

  // 使用自定义Hooks
  const { availableETFs, etfStatistics, statsLoading } = useETFData();
  const {
    result,
    frontier,
    rpResult,
    blResult,
    loading,
    optimize,
    calculateFrontier,
    calculateRiskParity,
    calculateBlackLitterman,
  } = useOptimization();

  const handleOptimize = async () => {
    await optimize({
      symbols: selectedETFs,
      objective,
      targetReturn: objective === 'target_return' ? targetReturn : undefined,
    });
    setActiveTab('2');
  };

  const handleCalculateFrontier = async () => {
    await calculateFrontier(selectedETFs, 20);
    setActiveTab('3');
  };

  const handleRiskParityOptimize = async () => {
    if (rpMethod === 'budget') {
      const hasRiskBudget = selectedETFs.some(etf => riskBudget[etf] && riskBudget[etf] > 0);
      if (!hasRiskBudget) {
        message.warning('风险预算方法需要为至少一个ETF设置风险预算');
        return;
      }
    }
    await calculateRiskParity(selectedETFs);
    setActiveTab('4');
  };

  const handleBlackLittermanOptimize = async () => {
    const views = Object.entries(absoluteViews).map(([symbol, ret]) => ({
      symbol,
      return: ret,
      confidence: 0.5,
    }));
    await calculateBlackLitterman(selectedETFs, views);
    setActiveTab('5');
  };

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
                loading={statsLoading}
              >
                {availableETFs.map(etf => {
                  const stats = etfStatistics[etf.symbol];
                  const displayReturn = stats ? (stats.annualized * 100).toFixed(1) : '-';
                  const displayVol = stats ? (stats.volatility * 100).toFixed(1) : '-';
                  return (
                    <Option key={etf.symbol} value={etf.symbol}>
                      {etf.symbol} - {etf.name} (预期收益: {displayReturn}%, 波动率: {displayVol}%)
                    </Option>
                  );
                })}
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

            <WeightConstraints
              selectedETFs={selectedETFs}
              minWeights={minWeights}
              maxWeights={maxWeights}
              onMinWeightsChange={setMinWeights}
              onMaxWeightsChange={setMaxWeights}
            />

            <Button type="primary" size="large" onClick={handleOptimize} loading={loading} style={{ marginRight: 16 }}>
              执行优化
            </Button>
            <Button size="large" onClick={handleCalculateFrontier} loading={loading}>
              计算有效前沿
            </Button>
          </TabPane>

          <TabPane tab="优化结果" key="2">
            {result ? <OptimizationResultDisplay result={result} /> : <Alert title="请先执行优化" type="info" showIcon />}
          </TabPane>

          <TabPane tab="有效前沿" key="3">
            {frontier.length > 0 ? <EfficientFrontierDisplay frontier={frontier} /> : <Alert title="请先计算有效前沿" type="info" showIcon />}
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
                    {rpMethod === 'budget' && (
                      <Form.Item label="风险预算配置" help="为每个ETF设置风险预算比例（总和应为100%）">
                        {selectedETFs.map(symbol => (
                          <div key={symbol} style={{ marginBottom: 8 }}>
                            <span style={{ display: 'inline-block', width: 80 }}>{symbol}:</span>
                            <InputNumber
                              min={0}
                              max={100}
                              step={5}
                              value={(riskBudget[symbol] || 0) * 100}
                              onChange={(value) => setRiskBudget(prev => ({ ...prev, [symbol]: (value || 0) / 100 }))}
                              formatter={(value) => `${value}%`}
                              parser={(value) => parseFloat(value?.replace('%', '') || '0')}
                              style={{ width: 100 }}
                            />
                          </div>
                        ))}
                      </Form.Item>
                    )}
                    <Form.Item label="目标波动率 (%)">
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
                {rpResult ? <RiskParityDisplay result={rpResult} /> : <Alert title="请先执行风险平价优化" type="info" showIcon />}
              </Col>
            </Row>
          </TabPane>

          <TabPane tab="Black-Litterman" key="5">
            <Row gutter={[24, 24]}>
              <Col span={8}>
                <StyledCard title="Black-Litterman配置">
                  <Form layout="vertical">
                    <Form.Item label="Tau (缩放参数)">
                      <InputNumber min={0.001} max={0.1} step={0.001} value={tau} onChange={(v) => setTau(v || 0.025)} style={{ width: '100%' }} />
                    </Form.Item>
                    <Form.Item label="风险厌恶系数">
                      <Slider min={1} max={5} step={0.5} value={riskAversion} onChange={setRiskAversion} />
                    </Form.Item>
                    <Form.Item label="绝对观点">
                      <Select
                        mode="multiple"
                        placeholder="选择要设置观点的资产"
                        style={{ width: '100%', marginBottom: 8 }}
                        onChange={(symbols) => {
                          const newViews: Record<string, number> = {};
                          symbols.forEach((s: string) => { newViews[s] = absoluteViews[s] || 0.1; });
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
                            formatter={(v) => `${((v ?? 0) * 100).toFixed(0)}%`}
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
                {blResult ? <BlackLittermanDisplay result={blResult} selectedETFs={selectedETFs} /> : <Alert title="请先执行Black-Litterman优化" type="info" showIcon />}
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

import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Button, Select, InputNumber, Table, Tabs, Spin, Alert, Slider, Form } from 'antd';
import ReactEChart from '../components/ReactEChart';
import type {
  EChartsOption,
  DefaultLabelFormatterCallbackParams,
  TooltipComponentFormatterCallbackParams,
} from 'echarts';
import Layout from '../components/Layout';
import styled from 'styled-components';
import { useOptimization } from '../hooks/useOptimization';
import { useETFData } from '../hooks/useETFData';
import { useFinancialConfig } from '../hooks/useFinancialConfig';
import type {
  OptimizationResult,
  EfficientFrontierPoint,
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

  // 资产配置饼图（对应原 Recharts PieChart，outerRadius=100）
  const weightPieOption: EChartsOption = {
    tooltip: {
      trigger: 'item',
      formatter: (params: TooltipComponentFormatterCallbackParams) => {
        const p = Array.isArray(params) ? params[0] : params;
        return `${p.name}: ${Number(p.value).toFixed(1)}%`;
      },
    },
    legend: {},
    series: [
      {
        name: '资产配置',
        type: 'pie',
        radius: '60%',
        center: ['50%', '50%'],
        label: {
          formatter: (params: DefaultLabelFormatterCallbackParams) => `${params.name}: ${Number(params.value).toFixed(1)}%`,
        },
        labelLine: { show: false },
        data: weightData.map((d, idx) => ({
          name: d.name,
          value: d.value,
          itemStyle: { color: COLORS[idx % COLORS.length] },
        })),
      },
    ],
  };

  // 风险贡献环形图（对应原 Recharts PieChart，innerRadius=60, outerRadius=100, paddingAngle=5）
  const riskPieOption: EChartsOption = {
    tooltip: {
      trigger: 'item',
      formatter: (params: TooltipComponentFormatterCallbackParams) => {
        const p = Array.isArray(params) ? params[0] : params;
        return `${p.name}: ${Number(p.value).toFixed(1)}%`;
      },
    },
    series: [
      {
        name: '风险贡献',
        type: 'pie',
        radius: ['40%', '65%'],
        center: ['50%', '50%'],
        padAngle: 5,
        label: {
          formatter: (params: DefaultLabelFormatterCallbackParams) => `${params.name}: ${Number(params.value).toFixed(1)}%`,
        },
        data: riskContributionData.map((d, idx) => ({
          name: d.symbol,
          value: d.contribution,
          itemStyle: { color: COLORS[idx % COLORS.length] },
        })),
      },
    ],
  };

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
            <ReactEChart option={weightPieOption} height={300} />
          </StyledCard>
        </Col>
        <Col span={12}>
          <StyledCard title="风险贡献">
            <ReactEChart option={riskPieOption} height={300} />
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
  const sharpeMin = Math.min(...frontierData.map(d => d.sharpe));
  const sharpeMax = Math.max(...frontierData.map(d => d.sharpe));

  // 有效前沿散点图（对应原 Recharts ScatterChart，ZAxis range=[50,400] 映射为 symbolSize）
  const frontierOption: EChartsOption = {
    tooltip: {
      trigger: 'item',
      formatter: (params: TooltipComponentFormatterCallbackParams) => {
        const p = Array.isArray(params) ? params[0] : params;
        const d = p.data as number[];
        return `波动率: ${Number(d[0]).toFixed(2)}%<br/>收益率: ${Number(d[1]).toFixed(2)}%<br/>夏普比率: ${Number(d[2]).toFixed(2)}`;
      },
    },
    legend: {},
    grid: {
      left: 70,
      right: 30,
      top: 40,
      bottom: 60,
    },
    xAxis: {
      type: 'value',
      name: '年化波动率 (%)',
      nameLocation: 'middle',
      nameGap: 30,
      min: Math.max(0, volMin - 0.1),
      max: volMax + 0.1,
      axisLabel: { formatter: (v: number) => v.toFixed(2) },
      splitLine: { lineStyle: { type: 'dashed' } },
    },
    yAxis: {
      type: 'value',
      name: '预期年化收益率 (%)',
      nameLocation: 'middle',
      nameGap: 45,
      min: Math.max(0, retMin - 1),
      max: retMax + 1,
      axisLabel: { formatter: (v: number) => v.toFixed(1) },
      splitLine: { lineStyle: { type: 'dashed' } },
    },
    series: [
      {
        name: '有效前沿',
        type: 'scatter',
        data: frontierData.map(d => [d.volatility, d.return, d.sharpe]),
        symbolSize: (data: number[]) => {
          if (sharpeMax === sharpeMin) return 50;
          return 50 + ((Number(data[2]) - sharpeMin) / (sharpeMax - sharpeMin)) * 350;
        },
        itemStyle: { color: '#8884d8' },
      },
    ],
  };

  return (
    <StyledCard title="有效前沿曲线">
      <ReactEChart option={frontierOption} height={500} />
      <p style={{ textAlign: 'center', color: '#666', marginTop: 16 }}>
        圆点大小代表夏普比率，越大表示风险调整后收益越好
      </p>
    </StyledCard>
  );
};

const PortfolioOptimization: React.FC = () => {
  // 使用自定义Hooks (must be before useState that references finConfig)
  const { config: finConfig } = useFinancialConfig();
  const { availableETFs, etfStatistics, statsLoading, refreshStatistics } = useETFData();

  // 按需获取统计数据（不再由 store 自动触发）
  useEffect(() => {
    if (availableETFs.length > 0) {
      refreshStatistics(availableETFs.map(e => e.symbol));
    }
  }, [availableETFs, refreshStatistics]);

  const [selectedETFs, setSelectedETFs] = useState<string[]>(['VTI', 'VOO', 'AGG']);
  const [objective, setObjective] = useState<'min_volatility' | 'max_sharpe' | 'target_return'>('max_sharpe');
  const [targetReturn, setTargetReturn] = useState<number>(0.10);
  const [riskFreeRate, setRiskFreeRate] = useState<number>(finConfig.risk_free_rate);
  const [minWeights, setMinWeights] = useState<Record<string, number>>({});
  const [maxWeights, setMaxWeights] = useState<Record<string, number>>({});
  const [activeTab, setActiveTab] = useState('1');
  const { result, frontier, loading, optimize, calculateFrontier } = useOptimization();

  const handleOptimize = async () => {
    await optimize({
      symbols: selectedETFs,
      objective,
      targetReturn: objective === 'target_return' ? targetReturn : undefined,
      riskFreeRate,
    });
    setActiveTab('2');
  };

  const handleCalculateFrontier = async () => {
    await calculateFrontier(selectedETFs, 20);
    setActiveTab('3');
  };

  return (
    <Layout>
      <Container>
        <Title>投资组合优化 (MPT)</Title>

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

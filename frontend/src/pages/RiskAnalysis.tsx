import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Select, Button, Spin, Alert, Statistic, Table, Tag, message } from 'antd';
import { WarningOutlined, SafetyOutlined, BarChartOutlined, RadarChartOutlined } from '@ant-design/icons';
import ReactEChart from '../components/ReactEChart';
import type { EChartsOption } from 'echarts';
import Layout from '../components/Layout';
import styled from 'styled-components';
import { theme } from '../styles/theme';
import { portfolioAPI } from '../services/api';

const { Option } = Select;

const PageContainer = styled.div`
  padding: ${theme.spacing.lg};
`;

const PageHeader = styled.div`
  margin-bottom: ${theme.spacing.lg};
`;

const PageTitle = styled.h1`
  font-size: ${theme.fonts.size['2xl']};
  font-weight: ${theme.fonts.weight.bold};
  color: ${theme.colors.textPrimary};
  margin-bottom: ${theme.spacing.sm};
`;

const RiskCard = styled(Card)<{ $riskLevel?: 'low' | 'medium' | 'high' }>`
  margin-bottom: ${theme.spacing.lg};
  border-left: 4px solid ${props => {
    switch (props.$riskLevel) {
      case 'low': return theme.colors.up;
      case 'medium': return theme.colors.warning || '#faad14';
      case 'high': return theme.colors.down;
      default: return theme.colors.primary;
    }
  }};
`;

const StatGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: ${theme.spacing.md};
  margin-bottom: ${theme.spacing.lg};
`;

// 预定义投资组合配置
const portfolioConfigs: Record<string, Record<string, number>> = {
  conservative: { BND: 0.80, VTI: 0.20 },
  balanced: { VTI: 0.60, BND: 0.40 },
  aggressive: { QQQ: 0.80, BND: 0.20 },
  income: { SCHD: 0.50, JEPQ: 0.50 },
  dividend_growth: { SCHD: 0.40, VYM: 0.30, VTI: 0.30 },
  tech_focus: { QQQ: 0.70, VTI: 0.30 },
};

// API返回的风险数据类型
interface PortfolioRisk {
  symbol: string;
  weight: number;
  componentVar: number;
  marginalVar: number;
}

interface RiskAnalysisResult {
  portfolio: Record<string, number>;
  period: string;
  confidence: number;
  risk_level: string;
  var_95: number;
  var_99: number;
  cvar_95: number;
  volatility: number;
  sharpe_ratio: number;
  sortino_ratio: number;
  max_drawdown: number;
  calmar_ratio: number;
  beta: number;
  alpha: number;
  portfolio_risks: PortfolioRisk[];
  data_points: number;
}

const RiskAnalysis: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [selectedPortfolio, setSelectedPortfolio] = useState<string>('conservative');
  const [riskData, setRiskData] = useState<RiskAnalysisResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  // 获取风险数据
  const fetchRiskData = async () => {
    setLoading(true);
    setError(null);
    try {
      const portfolio = portfolioConfigs[selectedPortfolio];
      if (!portfolio) {
        setError('Invalid portfolio selection');
        return;
      }

      const response = await portfolioAPI.analyzeRisk(portfolio);

      if (response.success && response.data) {
        const apiData = response.data;
        const riskResult: RiskAnalysisResult = {
          portfolio: apiData.portfolio,
          period: apiData.period,
          confidence: apiData.confidence,
          risk_level: apiData.risk_level,
          var_95: apiData.var_95,
          var_99: apiData.var_99,
          cvar_95: apiData.cvar_95,
          volatility: apiData.volatility,
          sharpe_ratio: apiData.sharpe_ratio,
          sortino_ratio: apiData.sortino_ratio,
          max_drawdown: apiData.max_drawdown,
          calmar_ratio: apiData.calmar_ratio,
          beta: apiData.beta,
          alpha: apiData.alpha,
          portfolio_risks: apiData.portfolio_risks,
          data_points: apiData.data_points,
        };
        setRiskData(riskResult);
      } else {
        setError(response.message || 'Failed to fetch risk data');
        message.error(response.message || '获取风险数据失败');
      }
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMsg);
      message.error('获取风险数据失败: ' + errorMsg);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRiskData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedPortfolio]);

  // 获取风险等级
  const getRiskLevel = (volatility: number): 'low' | 'medium' | 'high' => {
    if (volatility < 10) return 'low';
    if (volatility < 15) return 'medium';
    return 'high';
  };

  // 表格列定义
  const columns = [
    {
      title: '资产',
      dataIndex: 'symbol',
      key: 'symbol',
    },
    {
      title: '权重',
      dataIndex: 'weight',
      key: 'weight',
      render: (value: number) => `${(value * 100).toFixed(1)}%`,
    },
    {
      title: '成分 VaR',
      dataIndex: 'componentVar',
      key: 'componentVar',
      render: (value: number) => `${value.toFixed(2)}%`,
    },
    {
      title: '边际 VaR',
      dataIndex: 'marginalVar',
      key: 'marginalVar',
      render: (value: number) => `${value.toFixed(2)}%`,
    },
    {
      title: '风险贡献',
      key: 'contribution',
      render: (_: unknown, record: PortfolioRisk) => {
        const contribution = riskData ? (record.componentVar / riskData.var_95) * 100 : 0;
        return (
          <Tag color={contribution > 40 ? 'red' : contribution > 25 ? 'orange' : 'green'}>
            {contribution.toFixed(1)}%
          </Tag>
        );
      },
    },
  ];

  // 风险雷达图配置
  const radarData = riskData ? [
    { subject: 'VaR', value: Math.min(Math.abs(riskData.var_95) * 10, 100) },
    { subject: 'CVaR', value: Math.min(Math.abs(riskData.cvar_95) * 10, 100) },
    { subject: '波动率', value: Math.min(riskData.volatility * 5, 100) },
    { subject: 'Beta', value: Math.min(Math.abs(riskData.beta) * 50, 100) },
    { subject: 'Alpha', value: Math.min(Math.max(riskData.alpha * 5 + 50, 0), 100) },
    { subject: '夏普', value: Math.min(Math.max(riskData.sharpe_ratio * 30, 0), 100) },
  ] : null;

  const radarOption: EChartsOption | null = radarData ? {
    tooltip: {},
    radar: {
      indicator: radarData.map(d => ({ name: d.subject, max: 100 })),
      radius: '65%',
    },
    series: [{
      type: 'radar',
      data: [{ value: radarData.map(d => d.value), name: '风险指标' }],
      areaStyle: { opacity: 0.6 },
      lineStyle: { color: '#1890ff' },
      itemStyle: { color: '#1890ff' },
    }],
  } : null;

  // 风险指标对比柱状图配置
  const barData = riskData ? [
    { name: 'VaR(95%)', value: Math.abs(riskData.var_95), fill: '#ff4d4f' },
    { name: 'VaR(99%)', value: Math.abs(riskData.var_99), fill: '#cf1322' },
    { name: 'CVaR(95%)', value: Math.abs(riskData.cvar_95), fill: '#a8071a' },
    { name: '波动率', value: riskData.volatility, fill: '#fa8c16' },
    { name: '最大回撤', value: Math.abs(riskData.max_drawdown), fill: '#722ed1' },
  ] : null;

  const barOption: EChartsOption | null = barData ? {
    tooltip: {
      trigger: 'axis',
      valueFormatter: (value: unknown) => `${Number(value).toFixed(2)}%`,
    },
    legend: {},
    xAxis: { type: 'category', data: barData.map(d => d.name) },
    yAxis: { type: 'value' },
    series: [{
      type: 'bar',
      data: barData.map(d => ({ value: d.value, itemStyle: { color: d.fill } })),
      name: '百分比',
    }],
  } : null;

  return (
    <Layout>
      <PageContainer>
        <PageHeader>
          <PageTitle>
            <WarningOutlined style={{ marginRight: theme.spacing.sm }} />
            风险分析
          </PageTitle>
        </PageHeader>

        <Card style={{ marginBottom: theme.spacing.lg }}>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={12} md={8}>
            <div style={{ marginBottom: theme.spacing.xs }}>选择投资组合</div>
            <Select
              style={{ width: '100%' }}
              value={selectedPortfolio}
              onChange={setSelectedPortfolio}
            >
              <Option value="conservative">保守型组合</Option>
              <Option value="balanced">平衡型组合</Option>
              <Option value="aggressive">激进型组合</Option>
              <Option value="income">收入型组合</Option>
              <Option value="dividend_growth">股息增长型</Option>
              <Option value="tech_focus">科技聚焦型</Option>
            </Select>
          </Col>
          <Col xs={24} sm={12} md={8}>
            <Button
              type="primary"
              onClick={fetchRiskData}
              loading={loading}
              style={{ marginTop: 24 }}
              block
            >
              更新分析
            </Button>
          </Col>
        </Row>
      </Card>

      {error && (
        <Alert
          title="数据获取失败"
          description={error}
          type="error"
          style={{ marginBottom: theme.spacing.lg }}
          closable
          onClose={() => setError(null)}
        />
      )}

      {loading ? (
        <div style={{ textAlign: 'center', padding: theme.spacing.xl }}>
          <Spin size="large" />
          <p>正在计算风险指标...</p>
        </div>
      ) : riskData ? (
        <>
          {/* 风险等级提示 */}
          <RiskCard $riskLevel={getRiskLevel(riskData.volatility)}>
            <Row align="middle" gutter={[16, 16]}>
              <Col>
                <SafetyOutlined style={{ fontSize: 48, color: theme.colors.primary }} />
              </Col>
              <Col flex="auto">
                <h3 style={{ marginBottom: theme.spacing.xs }}>
                  风险等级: {riskData.risk_level === 'low' ? '低风险' : riskData.risk_level === 'medium' ? '中等风险' : '高风险'}
                </h3>
                <p style={{ color: theme.colors.textSecondary, margin: 0 }}>
                  当前投资组合的年化波动率为 {riskData.volatility.toFixed(2)}%，
                  最大回撤为 {riskData.max_drawdown.toFixed(2)}%。
                  {riskData.risk_level === 'low'
                    ? '风险水平较低，适合保守型投资者。'
                    : riskData.risk_level === 'medium'
                    ? '风险水平适中，适合平衡型投资者。'
                    : '风险水平较高，适合激进型投资者。'}
                </p>
              </Col>
            </Row>
          </RiskCard>

          {/* VaR / CVaR 指标 */}
          <StatGrid>
            <Card>
              <Statistic
                title="VaR (95%)"
                value={riskData.var_95}
                precision={2}
                suffix="%"
                styles={{ content: { color: '#cf1322' } }}
                prefix={<WarningOutlined />}
              />
              <div style={{ marginTop: theme.spacing.sm, fontSize: 12, color: theme.colors.textSecondary }}>
                单日最大损失概率 5%
              </div>
            </Card>
            <Card>
              <Statistic
                title="VaR (99%)"
                value={riskData.var_99}
                precision={2}
                suffix="%"
                styles={{ content: { color: '#cf1322' } }}
              />
              <div style={{ marginTop: theme.spacing.sm, fontSize: 12, color: theme.colors.textSecondary }}>
                单日最大损失概率 1%
              </div>
            </Card>
            <Card>
              <Statistic
                title="CVaR (95%)"
                value={riskData.cvar_95}
                precision={2}
                suffix="%"
                styles={{ content: { color: '#cf1322' } }}
              />
              <div style={{ marginTop: theme.spacing.sm, fontSize: 12, color: theme.colors.textSecondary }}>
                超过 VaR 的平均损失
              </div>
            </Card>
            <Card>
              <Statistic
                title="波动率"
                value={riskData.volatility}
                precision={2}
                suffix="%"
              />
              <div style={{ marginTop: theme.spacing.sm, fontSize: 12, color: theme.colors.textSecondary }}>
                年化标准差
              </div>
            </Card>
          </StatGrid>

          {/* 风险调整收益指标 */}
          <Row gutter={[16, 16]} style={{ marginBottom: theme.spacing.lg }}>
            <Col xs={24} md={12}>
              <Card title="风险调整收益指标">
                <Row gutter={[16, 16]}>
                  <Col span={12}>
                    <Statistic
                      title="夏普比率"
                      value={riskData.sharpe_ratio}
                      precision={2}
                      styles={{
                        content: {
                          color: riskData.sharpe_ratio > 1 ? '#3f8600' : riskData.sharpe_ratio > 0.5 ? '#faad14' : '#cf1322'
                        }
                      }}
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic
                      title="索提诺比率"
                      value={riskData.sortino_ratio}
                      precision={2}
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic
                      title="卡尔玛比率"
                      value={riskData.calmar_ratio}
                      precision={2}
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic
                      title="最大回撤"
                      value={riskData.max_drawdown}
                      precision={2}
                      suffix="%"
                      styles={{ content: { color: '#cf1322' } }}
                    />
                  </Col>
                </Row>
              </Card>
            </Col>
            <Col xs={24} md={12}>
              <Card title="市场风险指标">
                <Row gutter={[16, 16]}>
                  <Col span={12}>
                    <Statistic
                      title="Beta"
                      value={riskData.beta}
                      precision={2}
                      styles={{
                        content: {
                          color: riskData.beta > 1 ? '#cf1322' : riskData.beta < 0.8 ? '#3f8600' : '#666'
                        }
                      }}
                    />
                    <div style={{ fontSize: 12, color: theme.colors.textSecondary }}>
                      {riskData.beta > 1 ? '高于市场波动' : riskData.beta < 0.8 ? '低于市场波动' : '与市场同步'}
                    </div>
                  </Col>
                  <Col span={12}>
                    <Statistic
                      title="Alpha"
                      value={riskData.alpha}
                      precision={2}
                      suffix="%"
                      styles={{
                        content: {
                          color: riskData.alpha > 0 ? '#3f8600' : riskData.alpha < 0 ? '#cf1322' : '#666'
                        }
                      }}
                    />
                    <div style={{ fontSize: 12, color: theme.colors.textSecondary }}>
                      {riskData.alpha > 0 ? '超额收益' : riskData.alpha < 0 ? '跑输市场' : '与市场持平'}
                    </div>
                  </Col>
                </Row>
                <Alert
                  style={{ marginTop: theme.spacing.md }}
                  title="风险解释"
                  description="Beta 衡量相对于市场的波动性，Alpha 表示超额收益能力。"
                  type="info"
                  showIcon
                />
              </Card>
            </Col>
          </Row>

          {/* 组合风险分解 */}
          <Card
            title={
              <span>
                <BarChartOutlined style={{ marginRight: 8 }} />
                组合风险分解
              </span>
            }
          >
            <Table
              dataSource={riskData.portfolio_risks}
              columns={columns}
              rowKey="symbol"
              pagination={false}
            />
            <Alert
              style={{ marginTop: theme.spacing.md }}
              title="成分 VaR 说明"
              description="成分 VaR 表示每个资产对组合总风险的贡献度。边际 VaR 表示增加该资产配置对组合风险的影响程度。"
              type="info"
              showIcon
            />
          </Card>

          {/* 风险雷达图 */}
          <Row gutter={[16, 16]} style={{ marginTop: theme.spacing.lg }}>
            <Col xs={24} lg={12}>
              <Card
                title={
                  <span>
                    <RadarChartOutlined style={{ marginRight: 8 }} />
                    风险雷达图
                  </span>
                }
              >
                {radarOption && <ReactEChart option={radarOption} height={300} />}
                <div style={{ textAlign: 'center', fontSize: 12, color: theme.colors.textSecondary }}>
                  各指标已归一化到 0-100 范围，便于对比
                </div>
              </Card>
            </Col>
            <Col xs={24} lg={12}>
              <Card
                title={
                  <span>
                    <BarChartOutlined style={{ marginRight: 8 }} />
                    风险指标对比
                  </span>
                }
              >
                {barOption && <ReactEChart option={barOption} height={300} />}
              </Card>
            </Col>
          </Row>
        </>
      ) : (
        <Alert title="请选择投资组合以查看风险分析" type="info" />
      )}
      </PageContainer>
    </Layout>
  );
};

export default RiskAnalysis;

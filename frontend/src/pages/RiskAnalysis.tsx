import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Select, Button, Spin, Alert, Statistic, Table, Tag } from 'antd';
import { WarningOutlined, SafetyOutlined, BarChartOutlined } from '@ant-design/icons';
import styled from 'styled-components';
import { theme } from '../styles/theme';

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

// 模拟数据类型
interface RiskMetrics {
  volatility: number;
  sharpeRatio: number;
  sortinoRatio: number;
  maxDrawdown: number;
  calmarRatio: number;
  beta: number;
  alpha: number;
  var95: number;
  var99: number;
  cvar95: number;
}

interface PortfolioRisk {
  symbol: string;
  weight: number;
  componentVar: number;
  marginalVar: number;
}

const RiskAnalysis: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [selectedPortfolio, setSelectedPortfolio] = useState<string>('conservative');
  const [riskMetrics, setRiskMetrics] = useState<RiskMetrics | null>(null);
  const [portfolioRisks, setPortfolioRisks] = useState<PortfolioRisk[]>([]);

  // 模拟获取风险数据
  const fetchRiskData = async () => {
    setLoading(true);
    try {
      // 模拟不同投资组合的风险数据
      const mockData: Record<string, RiskMetrics> = {
        conservative: {
          volatility: 8.5,
          sharpeRatio: 1.2,
          sortinoRatio: 1.8,
          maxDrawdown: 12.3,
          calmarRatio: 0.85,
          beta: 0.65,
          alpha: 1.5,
          var95: 2.1,
          var99: 3.5,
          cvar95: 2.8,
        },
        balanced: {
          volatility: 12.8,
          sharpeRatio: 0.95,
          sortinoRatio: 1.4,
          maxDrawdown: 18.5,
          calmarRatio: 0.72,
          beta: 0.85,
          alpha: 0.8,
          var95: 3.2,
          var99: 5.1,
          cvar95: 4.2,
        },
        aggressive: {
          volatility: 18.2,
          sharpeRatio: 0.75,
          sortinoRatio: 1.1,
          maxDrawdown: 28.3,
          calmarRatio: 0.58,
          beta: 1.15,
          alpha: -0.5,
          var95: 4.8,
          var99: 7.2,
          cvar95: 6.1,
        },
      };

      setRiskMetrics(mockData[selectedPortfolio]);

      // 模拟组合风险分解
      const mockPortfolioRisks: PortfolioRisk[] = [
        { symbol: 'SPY', weight: 0.4, componentVar: 1.25, marginalVar: 3.12 },
        { symbol: 'BND', weight: 0.3, componentVar: 0.35, marginalVar: 1.17 },
        { symbol: 'QQQ', weight: 0.2, componentVar: 1.15, marginalVar: 5.75 },
        { symbol: 'VTI', weight: 0.1, componentVar: 0.45, marginalVar: 4.50 },
      ];
      setPortfolioRisks(mockPortfolioRisks);
    } catch (error) {
      console.error('Failed to fetch risk data:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRiskData();
  }, [selectedPortfolio]);

  // 获取风险等级
  const getRiskLevel = (metrics: RiskMetrics): 'low' | 'medium' | 'high' => {
    if (metrics.volatility < 10) return 'low';
    if (metrics.volatility < 15) return 'medium';
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
      render: (_: any, record: PortfolioRisk) => {
        const contribution = (record.componentVar / (riskMetrics?.var95 || 1)) * 100;
        return (
          <Tag color={contribution > 40 ? 'red' : contribution > 25 ? 'orange' : 'green'}>
            {contribution.toFixed(1)}%
          </Tag>
        );
      },
    },
  ];

  return (
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

      {loading ? (
        <div style={{ textAlign: 'center', padding: theme.spacing.xl }}>
          <Spin size="large" />
          <p>正在计算风险指标...</p>
        </div>
      ) : riskMetrics ? (
        <>
          {/* 风险等级提示 */}
          <RiskCard $riskLevel={getRiskLevel(riskMetrics)}>
            <Row align="middle" gutter={[16, 16]}>
              <Col>
                <SafetyOutlined style={{ fontSize: 48, color: theme.colors.primary }} />
              </Col>
              <Col flex="auto">
                <h3 style={{ marginBottom: theme.spacing.xs }}>
                  风险等级: {getRiskLevel(riskMetrics) === 'low' ? '低风险' : getRiskLevel(riskMetrics) === 'medium' ? '中等风险' : '高风险'}
                </h3>
                <p style={{ color: theme.colors.textSecondary, margin: 0 }}>
                  当前投资组合的年化波动率为 {riskMetrics.volatility}%，
                  最大回撤为 {riskMetrics.maxDrawdown}%。
                  {getRiskLevel(riskMetrics) === 'low'
                    ? '风险水平较低，适合保守型投资者。'
                    : getRiskLevel(riskMetrics) === 'medium'
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
                value={riskMetrics.var95}
                precision={2}
                suffix="%"
                valueStyle={{ color: '#cf1322' }}
                prefix={<WarningOutlined />}
              />
              <div style={{ marginTop: theme.spacing.sm, fontSize: 12, color: theme.colors.textSecondary }}>
                单日最大损失概率 5%
              </div>
            </Card>
            <Card>
              <Statistic
                title="VaR (99%)"
                value={riskMetrics.var99}
                precision={2}
                suffix="%"
                valueStyle={{ color: '#cf1322' }}
              />
              <div style={{ marginTop: theme.spacing.sm, fontSize: 12, color: theme.colors.textSecondary }}>
                单日最大损失概率 1%
              </div>
            </Card>
            <Card>
              <Statistic
                title="CVaR (95%)"
                value={riskMetrics.cvar95}
                precision={2}
                suffix="%"
                valueStyle={{ color: '#cf1322' }}
              />
              <div style={{ marginTop: theme.spacing.sm, fontSize: 12, color: theme.colors.textSecondary }}>
                超过 VaR 的平均损失
              </div>
            </Card>
            <Card>
              <Statistic
                title="波动率"
                value={riskMetrics.volatility}
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
                      value={riskMetrics.sharpeRatio}
                      precision={2}
                      valueStyle={{
                        color: riskMetrics.sharpeRatio > 1 ? '#3f8600' : riskMetrics.sharpeRatio > 0.5 ? '#faad14' : '#cf1322'
                      }}
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic
                      title="索提诺比率"
                      value={riskMetrics.sortinoRatio}
                      precision={2}
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic
                      title="卡尔玛比率"
                      value={riskMetrics.calmarRatio}
                      precision={2}
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic
                      title="最大回撤"
                      value={riskMetrics.maxDrawdown}
                      precision={2}
                      suffix="%"
                      valueStyle={{ color: '#cf1322' }}
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
                      value={riskMetrics.beta}
                      precision={2}
                      valueStyle={{
                        color: riskMetrics.beta > 1 ? '#cf1322' : riskMetrics.beta < 0.8 ? '#3f8600' : '#666'
                      }}
                    />
                    <div style={{ fontSize: 12, color: theme.colors.textSecondary }}>
                      {riskMetrics.beta > 1 ? '高于市场波动' : riskMetrics.beta < 0.8 ? '低于市场波动' : '与市场同步'}
                    </div>
                  </Col>
                  <Col span={12}>
                    <Statistic
                      title="Alpha"
                      value={riskMetrics.alpha}
                      precision={2}
                      suffix="%"
                      valueStyle={{
                        color: riskMetrics.alpha > 0 ? '#3f8600' : riskMetrics.alpha < 0 ? '#cf1322' : '#666'
                      }}
                    />
                    <div style={{ fontSize: 12, color: theme.colors.textSecondary }}>
                      {riskMetrics.alpha > 0 ? '超额收益' : riskMetrics.alpha < 0 ? '跑输市场' : '与市场持平'}
                    </div>
                  </Col>
                </Row>
                <Alert
                  style={{ marginTop: theme.spacing.md }}
                  message="风险解释"
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
              dataSource={portfolioRisks}
              columns={columns}
              rowKey="symbol"
              pagination={false}
            />
            <Alert
              style={{ marginTop: theme.spacing.md }}
              message="成分 VaR 说明"
              description="成分 VaR 表示每个资产对组合总风险的贡献度。边际 VaR 表示增加该资产配置对组合风险的影响程度。"
              type="info"
              showIcon
            />
          </Card>
        </>
      ) : (
        <Alert message="请选择投资组合以查看风险分析" type="info" />
      )}
    </PageContainer>
  );
};

export default RiskAnalysis;

import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Button, Select, Switch, Table, Tabs, Spin, Alert, Tag, message } from 'antd';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, RadarChart, PolarGrid, PolarAngleAxis, PolarRadiusAxis, Radar } from 'recharts';
import { factorAPI, etfAPI } from '../services/api';
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
  font-size: 28px;
  font-weight: 700;
  margin-bottom: 8px;
`;

const MetricLabel = styled.div`
  font-size: 14px;
  opacity: 0.9;
`;



interface ETFInfo {
  symbol: string;
  name: string;
}

interface FactorExposure {
  market: number;
  size: number;
  value: number;
  profitability?: number;
  investment?: number;
  alpha: number;
  r2: number;
  adj_r2: number;
}

interface FactorAttribution {
  exposures: FactorExposure;
  contributions: Record<string, number>;
  total_return: number;
  explained_return: number;
  unexplained_return: number;
  annualized_alpha: number;
  t_statistics: Record<string, number>;
  p_values: Record<string, number>;
}

interface FactorStats {
  name: string;
  annualized: number;
  volatility: number;
  sharpe: number;
  max_drawdown: number;
}

const FactorAnalysis: React.FC = () => {
  const [selectedETFs, setSelectedETFs] = useState<string[]>([]);
  const [useFiveFactor, setUseFiveFactor] = useState<boolean>(false);
  const [loading, setLoading] = useState(false);
  const [factorStats, setFactorStats] = useState<FactorStats[]>([]);
  const [attribution, setAttribution] = useState<FactorAttribution | null>(null);
  const [multiAttribution, setMultiAttribution] = useState<Record<string, FactorAttribution>>({});
  const [activeTab, setActiveTab] = useState('1');
  const [error, setError] = useState<string | null>(null);
  const [availableETFs, setAvailableETFs] = useState<ETFInfo[]>([]);
  const [etfLoading, setEtfLoading] = useState(false);

  // 获取可用ETF列表
  useEffect(() => {
    const fetchETFs = async () => {
      setEtfLoading(true);
      try {
        const response = await etfAPI.getList();
        if (response.success && response.data) {
          setAvailableETFs(response.data);
          // 设置默认值：取列表前3个（如果不足3个则取全部）
          // 使用函数式更新避免依赖 selectedETFs
          setSelectedETFs(prev => {
            const data = response.data;
            if (data && data.length > 0 && prev.length === 0) {
              const defaultCount = Math.min(3, data.length);
              return data.slice(0, defaultCount).map(etf => etf.symbol);
            }
            return prev;
          });
        }
      } catch {
        message.error('获取ETF列表失败');
      } finally {
        setEtfLoading(false);
      }
    };
    fetchETFs();
  }, []);

  // 加载因子统计信息
  useEffect(() => {
    loadFactorStats();
  }, [useFiveFactor]);

  const loadFactorStats = async () => {
    try {
      const response = await factorAPI.getStatistics();
      if (response.success && response.data) {
        // 将API返回的数据转换为组件期望的格式
        const stats: FactorStats[] = response.data.factors.map(f => ({
          name: f.name,
          annualized: f.annualized_return,
          volatility: f.volatility,
          sharpe: f.annualized_return / f.volatility,
          max_drawdown: 0,
        }));
        setFactorStats(stats);
      } else if (response.error) {
        console.error('加载因子统计失败:', response.error);
      }
    } catch (error) {
      console.error('加载因子统计失败:', error);
    }
  };

  // 分析单个ETF
  const analyzeSingleETF = async (symbol: string) => {
    setLoading(true);
    setError(null);
    try {
      const factors = useFiveFactor
        ? ['market', 'size', 'value', 'profitability', 'investment']
        : ['market', 'size', 'value'];

      const response = await factorAPI.analyzeExposure(symbol, factors);

      if (response.success && response.data) {
        // 将API返回的数据转换为组件期望的格式
        const exposures = response.data.exposures;
        const attribution: FactorAttribution = {
          exposures: {
            market: exposures.market || 0,
            size: exposures.size || 0,
            value: exposures.value || 0,
            profitability: exposures.profitability || 0,
            investment: exposures.investment || 0,
            alpha: 0,
            r2: response.data.r_squared,
            adj_r2: response.data.r_squared,
          },
          contributions: {
            market: exposures.market || 0,
            smb: exposures.size || 0,
            hml: exposures.value || 0,
            rmw: exposures.profitability || 0,
            cma: exposures.investment || 0,
          },
          total_return: 0,
          explained_return: 0,
          unexplained_return: 0,
          annualized_alpha: 0,
          t_statistics: {},
          p_values: {},
        };
        setAttribution(attribution);
        setActiveTab('2');
        message.success('因子分析完成');
      } else {
        setError(response.error || '分析失败');
        message.error(response.error || '分析失败');
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : '分析失败';
      setError(errorMsg);
      message.error(errorMsg);
      console.error('因子分析失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 分析投资组合
  const analyzePortfolio = async () => {
    if (selectedETFs.length === 0) {
      message.warning('请至少选择一个ETF');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const weights: Record<string, number> = {};
      selectedETFs.forEach(symbol => {
        weights[symbol] = 1.0 / selectedETFs.length;
      });

      const response = await factorAPI.analyzePortfolio(weights);

      if (response.success && response.data) {
        // 将API返回的数据转换为组件期望的格式
        const portfolioExposures = response.data.portfolio_exposures;
        const factorContributions = response.data.factor_contributions;
        const attribution: FactorAttribution = {
          exposures: {
            market: portfolioExposures.market || 0,
            size: portfolioExposures.size || 0,
            value: portfolioExposures.value || 0,
            profitability: portfolioExposures.profitability || 0,
            investment: portfolioExposures.investment || 0,
            alpha: 0,
            r2: 0,
            adj_r2: 0,
          },
          contributions: {
            market: factorContributions.market || 0,
            smb: factorContributions.size || 0,
            hml: factorContributions.value || 0,
            rmw: factorContributions.profitability || 0,
            cma: factorContributions.investment || 0,
          },
          total_return: 0,
          explained_return: 0,
          unexplained_return: 0,
          annualized_alpha: 0,
          t_statistics: {},
          p_values: {},
        };
        setAttribution(attribution);
        setActiveTab('2');
        message.success('组合因子分析完成');
      } else {
        setError(response.error || '分析失败');
        message.error(response.error || '分析失败');
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : '分析失败';
      setError(errorMsg);
      message.error(errorMsg);
      console.error('组合因子分析失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 批量分析多个ETF
  const analyzeMultipleETFs = async () => {
    if (selectedETFs.length === 0) {
      message.warning('请至少选择一个ETF');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const response = await factorAPI.analyzeMultipleAssets(selectedETFs);

      if (response.success && response.data) {
        // 将API返回的数据转换为组件期望的格式
        const multiAttr: Record<string, FactorAttribution> = {};
        Object.entries(response.data).forEach(([symbol, exposures]) => {
          multiAttr[symbol] = {
            exposures: {
              market: exposures.market || 0,
              size: exposures.size || 0,
              value: exposures.value || 0,
              profitability: exposures.profitability || 0,
              investment: exposures.investment || 0,
              alpha: 0,
              r2: 0,
              adj_r2: 0,
            },
            contributions: {
              market: exposures.market || 0,
              smb: exposures.size || 0,
              hml: exposures.value || 0,
              rmw: exposures.profitability || 0,
              cma: exposures.investment || 0,
            },
            total_return: 0,
            explained_return: 0,
            unexplained_return: 0,
            annualized_alpha: 0,
            t_statistics: {},
            p_values: {},
          };
        });
        setMultiAttribution(multiAttr);
        setActiveTab('3');
        message.success('批量因子分析完成');
      } else {
        setError(response.error || '分析失败');
        message.error(response.error || '分析失败');
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : '分析失败';
      setError(errorMsg);
      message.error(errorMsg);
      console.error('批量因子分析失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 准备雷达图数据
  const radarData = attribution ? [
    { factor: '市场', exposure: attribution.exposures.market, fullMark: 2 },
    { factor: '市值', exposure: attribution.exposures.size, fullMark: 2 },
    { factor: '价值', exposure: attribution.exposures.value, fullMark: 2 },
    ...(useFiveFactor ? [
      { factor: '盈利', exposure: attribution.exposures.profitability || 0, fullMark: 2 },
      { factor: '投资', exposure: attribution.exposures.investment || 0, fullMark: 2 },
    ] : []),
  ] : [];

  // 准备贡献柱状图数据
  const contributionData = attribution ? [
    { name: '市场因子', contribution: attribution.contributions.market * 100 },
    { name: '市值因子', contribution: attribution.contributions.smb * 100 },
    { name: '价值因子', contribution: attribution.contributions.hml * 100 },
    ...(useFiveFactor ? [
      { name: '盈利因子', contribution: (attribution.contributions.rmw || 0) * 100 },
      { name: '投资因子', contribution: (attribution.contributions.cma || 0) * 100 },
    ] : []),
    { name: 'Alpha', contribution: attribution.contributions.alpha * 100 },
  ] : [];

  // 多资产对比表格数据
  const comparisonColumns = [
    { title: 'ETF', dataIndex: 'symbol', key: 'symbol' },
    { title: '市场暴露', dataIndex: 'market', key: 'market', render: (v: number) => v?.toFixed(3) },
    { title: '市值暴露', dataIndex: 'size', key: 'size', render: (v: number) => v?.toFixed(3) },
    { title: '价值暴露', dataIndex: 'value', key: 'value', render: (v: number) => v?.toFixed(3) },
    ...(useFiveFactor ? [
      { title: '盈利暴露', dataIndex: 'profitability', key: 'profitability', render: (v: number) => {
        if (v === undefined || v === null || isNaN(v) || !isFinite(v)) return '0.000';
        return v.toFixed(3);
      }},
      { title: '投资暴露', dataIndex: 'investment', key: 'investment', render: (v: number) => {
        if (v === undefined || v === null || isNaN(v) || !isFinite(v)) return '0.000';
        return v.toFixed(3);
      }},
    ] : []),
    { title: 'Alpha', dataIndex: 'alpha', key: 'alpha', render: (v: number) => `${(v * 100).toFixed(2)}%` },
    { title: 'R²', dataIndex: 'r2', key: 'r2', render: (v: number) => v?.toFixed(3) },
  ];

  const comparisonData = Object.entries(multiAttribution).map(([symbol, attr]) => ({
    symbol,
    market: attr.exposures.market,
    size: attr.exposures.size,
    value: attr.exposures.value,
    profitability: attr.exposures.profitability ?? 0,
    investment: attr.exposures.investment ?? 0,
    alpha: attr.annualized_alpha,
    r2: attr.exposures.r2,
  }));

  return (
    <Layout>
      <Container>
        <Title>Fama-French 因子归因分析</Title>

        {error && (
          <Alert
            message="错误"
            description={error}
            type="error"
            closable
            onClose={() => setError(null)}
            style={{ marginBottom: 16 }}
          />
        )}

        <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab="配置" key="1">
          <StyledCard title="分析配置">
            <Row gutter={[24, 24]}>
              <Col span={12}>
                <h4>选择ETF</h4>
                <Select
                  mode="multiple"
                  style={{ width: '100%', marginBottom: 16 }}
                  placeholder="选择要分析的ETF"
                  value={selectedETFs}
                  onChange={setSelectedETFs}
                  loading={etfLoading}
                >
                  {availableETFs.map(etf => (
                    <Option key={etf.symbol} value={etf.symbol}>
                      {etf.symbol} - {etf.name}
                    </Option>
                  ))}
                </Select>

                <div style={{ marginBottom: 16 }}>
                  <span style={{ marginRight: 8 }}>使用五因子模型:</span>
                  <Switch checked={useFiveFactor} onChange={setUseFiveFactor} />
                  {useFiveFactor && <Tag color="blue" style={{ marginLeft: 8 }}>RMW + CMA</Tag>}
                </div>

                <Button type="primary" onClick={analyzePortfolio} loading={loading} style={{ marginRight: 8 }}>
                  分析组合
                </Button>
                <Button onClick={analyzeMultipleETFs} loading={loading}>
                  批量对比
                </Button>
              </Col>
              <Col span={12}>
                <h4>快速分析单个ETF</h4>
                <Select
                  style={{ width: '100%', marginBottom: 16 }}
                  placeholder="选择ETF"
                  onChange={analyzeSingleETF}
                  loading={etfLoading}
                >
                  {availableETFs.map(etf => (
                    <Option key={etf.symbol} value={etf.symbol}>
                      {etf.symbol} - {etf.name}
                    </Option>
                  ))}
                </Select>
              </Col>
            </Row>
          </StyledCard>

          <StyledCard title="因子统计信息">
            <Table
              dataSource={factorStats}
              columns={[
                { title: '因子', dataIndex: 'name', key: 'name' },
                { title: '年化收益', dataIndex: 'annualized', key: 'annualized', render: (v: number) => `${(v * 100).toFixed(2)}%` },
                { title: '年化波动率', dataIndex: 'volatility', key: 'volatility', render: (v: number) => `${(v * 100).toFixed(2)}%` },
                { title: '夏普比率', dataIndex: 'sharpe', key: 'sharpe', render: (v: number) => v?.toFixed(2) },
                { title: '最大回撤', dataIndex: 'max_drawdown', key: 'max_drawdown', render: (v: number) => `${(v * 100).toFixed(2)}%` },
              ]}
              rowKey="name"
              pagination={false}
              size="small"
            />
          </StyledCard>
        </TabPane>

        <TabPane tab="因子暴露" key="2">
          {attribution ? (
            <>
              <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
                <Col span={6}>
                  <MetricCard>
                    <MetricValue>{(attribution.total_return * 100).toFixed(2)}%</MetricValue>
                    <MetricLabel>总收益</MetricLabel>
                  </MetricCard>
                </Col>
                <Col span={6}>
                  <MetricCard style={{ background: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)' }}>
                    <MetricValue>{(attribution.explained_return * 100).toFixed(2)}%</MetricValue>
                    <MetricLabel>因子解释收益</MetricLabel>
                  </MetricCard>
                </Col>
                <Col span={6}>
                  <MetricCard style={{ background: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)' }}>
                    <MetricValue>{(attribution.annualized_alpha * 100).toFixed(2)}%</MetricValue>
                    <MetricLabel>年化Alpha</MetricLabel>
                  </MetricCard>
                </Col>
                <Col span={6}>
                  <MetricCard style={{ background: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)' }}>
                    <MetricValue>{(attribution.exposures.r2 * 100).toFixed(1)}%</MetricValue>
                    <MetricLabel>解释度 R²</MetricLabel>
                  </MetricCard>
                </Col>
              </Row>

              <Row gutter={[24, 24]}>
                <Col span={12}>
                  <StyledCard title="因子暴露雷达图">
                    <ResponsiveContainer width="100%" height={400}>
                      <RadarChart data={radarData}>
                        <PolarGrid />
                        <PolarAngleAxis dataKey="factor" />
                        <PolarRadiusAxis angle={90} domain={[-1, 2]} />
                        <Radar
                          name="因子暴露"
                          dataKey="exposure"
                          stroke="#8884d8"
                          fill="#8884d8"
                          fillOpacity={0.6}
                        />
                        <Legend />
                        <Tooltip />
                      </RadarChart>
                    </ResponsiveContainer>
                  </StyledCard>
                </Col>
                <Col span={12}>
                  <StyledCard title="收益贡献分解">
                    <ResponsiveContainer width="100%" height={400}>
                      <BarChart data={contributionData} layout="vertical">
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis type="number" unit="%" />
                        <YAxis dataKey="name" type="category" width={80} />
                        <Tooltip />
                        <Bar dataKey="contribution" fill="#8884d8" />
                      </BarChart>
                    </ResponsiveContainer>
                  </StyledCard>
                </Col>
              </Row>

              <StyledCard title="因子暴露详情">
                <Table
                  dataSource={[
                    { factor: '市场因子 (Market)', exposure: attribution.exposures.market, t_stat: attribution.t_statistics.market, p_value: attribution.p_values.market, description: '市场贝塔，衡量系统性风险' },
                    { factor: '市值因子 (SMB)', exposure: attribution.exposures.size, t_stat: attribution.t_statistics.smb, p_value: attribution.p_values.smb, description: '小市值溢价' },
                    { factor: '价值因子 (HML)', exposure: attribution.exposures.value, t_stat: attribution.t_statistics.hml, p_value: attribution.p_values.hml, description: '高账面市值比溢价' },
                    ...(useFiveFactor ? [
                      { factor: '盈利因子 (RMW)', exposure: attribution.exposures.profitability || 0, t_stat: attribution.t_statistics.rmw, p_value: attribution.p_values.rmw, description: '高盈利能力溢价' },
                      { factor: '投资因子 (CMA)', exposure: attribution.exposures.investment || 0, t_stat: attribution.t_statistics.cma, p_value: attribution.p_values.cma, description: '保守投资溢价' },
                    ] : []),
                    { factor: 'Alpha', exposure: attribution.exposures.alpha, t_stat: attribution.t_statistics.alpha, p_value: attribution.p_values.alpha, description: '超额收益' },
                  ]}
                  columns={[
                    { title: '因子', dataIndex: 'factor', key: 'factor' },
                    { title: '暴露', dataIndex: 'exposure', key: 'exposure', render: (v: number) => {
                      if (v === undefined || v === null || isNaN(v) || !isFinite(v)) return '0.0000';
                      return v.toFixed(4);
                    }},
                    { title: 'T统计量', dataIndex: 't_stat', key: 't_stat', render: (v: number) => {
                      if (v === undefined || v === null || isNaN(v) || !isFinite(v)) return '0.00';
                      return v.toFixed(2);
                    }},
                    { title: 'P值', dataIndex: 'p_value', key: 'p_value', render: (v: number) => {
                      if (v === undefined || v === null || isNaN(v) || !isFinite(v)) return '0.0000';
                      return v.toFixed(4);
                    }},
                    { title: '说明', dataIndex: 'description', key: 'description' },
                  ]}
                  rowKey="factor"
                  pagination={false}
                />
              </StyledCard>
            </>
          ) : (
            <Alert title="请先执行因子分析" type="info" showIcon />
          )}
        </TabPane>

        <TabPane tab="多资产对比" key="3">
          {Object.keys(multiAttribution).length > 0 ? (
            <StyledCard title="多资产因子暴露对比">
              <Table
                dataSource={comparisonData}
                columns={comparisonColumns}
                rowKey="symbol"
                pagination={false}
              />
            </StyledCard>
          ) : (
            <Alert title="请先执行批量分析" type="info" showIcon />
          )}
        </TabPane>
      </Tabs>

        {loading && (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <Spin size="large" />
            <p style={{ marginTop: 16 }}>正在进行因子分析...</p>
          </div>
        )}
      </Container>
    </Layout>
  );
};

export default FactorAnalysis;

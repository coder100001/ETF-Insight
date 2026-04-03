import { useState } from 'react';
import { Card, Row, Col, Select, Button, Table, Tag, Progress, Spin, Alert, Statistic, List, Tabs } from 'antd';
import { PieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { InteractionOutlined, WarningOutlined, ClusterOutlined } from '@ant-design/icons';
import Layout from '../components/Layout';
import type { OverlapResult, PortfolioOverlap, CommonHolding } from '../types';

const { Option } = Select;
const { TabPane } = Tabs;

// 模拟ETF列表
const ETF_LIST = [
  { symbol: 'SCHD', name: 'Schwab US Dividend Equity ETF' },
  { symbol: 'VYM', name: 'Vanguard High Dividend Yield ETF' },
  { symbol: 'JEPI', name: 'JPMorgan Equity Premium Income ETF' },
  { symbol: 'HDV', name: 'iShares Core High Dividend ETF' },
  { symbol: 'SPYD', name: 'SPDR Portfolio S&P 500 High Dividend ETF' },
  { symbol: 'DVY', name: 'iShares Select Dividend ETF' },
];

const COLORS = ['#1890ff', '#52c41a', '#faad14', '#ff4d4f', '#722ed1', '#13c2c2'];

const OverlapAnalysis = () => {
  const [selectedETFs, setSelectedETFs] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [pairwiseResult, setPairwiseResult] = useState<OverlapResult | null>(null);
  const [portfolioResult, setPortfolioResult] = useState<PortfolioOverlap | null>(null);
  const [activeTab, setActiveTab] = useState('pairwise');
  const [error, setError] = useState<string | null>(null);

  const handleAnalyze = async () => {
    if (selectedETFs.length < 2) {
      setError('请至少选择2只ETF');
      return;
    }

    setLoading(true);
    setError(null);

    // 模拟数据
    setTimeout(() => {
      if (selectedETFs.length === 2) {
        // 两两对比
        const mockResult: OverlapResult = {
          etf1_symbol: selectedETFs[0],
          etf2_symbol: selectedETFs[1],
          overlap_score: 25 + Math.random() * 50,
          common_holdings: [
            { symbol: 'AAPL', name: 'Apple Inc.', etf1_weight: 4.5, etf2_weight: 3.8, total_weight: 8.3, sector: 'Technology' },
            { symbol: 'MSFT', name: 'Microsoft Corp.', etf1_weight: 3.2, etf2_weight: 4.1, total_weight: 7.3, sector: 'Technology' },
            { symbol: 'JNJ', name: 'Johnson & Johnson', etf1_weight: 2.8, etf2_weight: 2.5, total_weight: 5.3, sector: 'Healthcare' },
            { symbol: 'JPM', name: 'JPMorgan Chase', etf1_weight: 2.1, etf2_weight: 2.3, total_weight: 4.4, sector: 'Financials' },
            { symbol: 'XOM', name: 'Exxon Mobil', etf1_weight: 1.8, etf2_weight: 1.9, total_weight: 3.7, sector: 'Energy' },
          ],
          etf1_only: [
            { symbol: 'CVX', name: 'Chevron Corp.', weight: 2.5, sector: 'Energy' },
            { symbol: 'PG', name: 'Procter & Gamble', weight: 2.2, sector: 'Consumer Staples' },
          ],
          etf2_only: [
            { symbol: 'UNH', name: 'UnitedHealth Group', weight: 2.8, sector: 'Healthcare' },
            { symbol: 'HD', name: 'Home Depot', weight: 2.1, sector: 'Consumer Discretionary' },
          ],
          sector_overlap: {
            'Technology': 35.5,
            'Healthcare': 18.2,
            'Financials': 15.8,
            'Energy': 12.3,
            'Consumer Staples': 10.5,
            'Others': 7.7,
          },
          country_overlap: {
            'USA': 95.5,
            'International': 4.5,
          },
        };
        setPairwiseResult(mockResult);
        setPortfolioResult(null);
      } else {
        // 投资组合分析
        const mockPortfolio: PortfolioOverlap = {
          etf_count: selectedETFs.length,
          pairwise_results: [],
          overall_score: 30 + Math.random() * 40,
          concentration: {
            hhi: 1200 + Math.random() * 1000,
            top5_concentration: 25 + Math.random() * 15,
            top10_concentration: 40 + Math.random() * 20,
            risk_level: 'Medium',
          },
          diversification: {
            effective_n: 15 + Math.random() * 20,
            sector_count: 8 + Math.floor(Math.random() * 4),
            country_count: 3 + Math.floor(Math.random() * 5),
            stock_count: 80 + Math.floor(Math.random() * 100),
          },
        };
        setPortfolioResult(mockPortfolio);
        setPairwiseResult(null);
      }
      setLoading(false);
    }, 800);
  };

  const getRiskLevelColor = (level: string) => {
    switch (level) {
      case 'Low': return 'green';
      case 'Medium': return 'orange';
      case 'High': return 'red';
      default: return 'default';
    }
  };

  const getRiskLevelText = (level: string) => {
    switch (level) {
      case 'Low': return '低风险';
      case 'Medium': return '中风险';
      case 'High': return '高风险';
      default: return level;
    }
  };

  const columns = [
    { title: '股票代码', dataIndex: 'symbol', key: 'symbol' },
    { title: '公司名称', dataIndex: 'name', key: 'name' },
    {
      title: `${pairwiseResult?.etf1_symbol || 'ETF1'} 权重`,
      dataIndex: 'etf1_weight',
      key: 'etf1_weight',
      render: (v: number) => `${v?.toFixed(2)}%`,
    },
    {
      title: `${pairwiseResult?.etf2_symbol || 'ETF2'} 权重`,
      dataIndex: 'etf2_weight',
      key: 'etf2_weight',
      render: (v: number) => `${v?.toFixed(2)}%`,
    },
    {
      title: '合计权重',
      dataIndex: 'total_weight',
      key: 'total_weight',
      render: (v: number) => <Tag color="blue">{v?.toFixed(2)}%</Tag>,
      sorter: (a: CommonHolding, b: CommonHolding) => a.total_weight - b.total_weight,
    },
    { title: '行业', dataIndex: 'sector', key: 'sector' },
  ];

  return (
    <Layout>
      <div style={{ padding: '24px' }}>
        <h1 style={{ marginBottom: '24px' }}>
          <InteractionOutlined style={{ marginRight: '8px' }} />
          持仓重叠分析
        </h1>

        {/* 参数配置 */}
        <Card title="分析参数" style={{ marginBottom: '24px' }}>
          <Row gutter={24} align="middle">
            <Col span={12}>
              <label style={{ display: 'block', marginBottom: '8px' }}>选择ETF（2只对比，3只以上组合分析）</label>
              <Select
                mode="multiple"
                placeholder="选择要分析的ETF"
                value={selectedETFs}
                onChange={setSelectedETFs}
                style={{ width: '100%' }}
                maxTagCount={5}
              >
                {ETF_LIST.map(etf => (
                  <Option key={etf.symbol} value={etf.symbol}>
                    {etf.symbol} - {etf.name}
                  </Option>
                ))}
              </Select>
            </Col>
            <Col span={6}>
              <Tabs activeKey={activeTab} onChange={setActiveTab}>
                <TabPane tab="两两对比" key="pairwise" />
                <TabPane tab="组合分析" key="portfolio" />
              </Tabs>
            </Col>
            <Col span={6}>
              <Button
                type="primary"
                size="large"
                onClick={handleAnalyze}
                loading={loading}
                icon={<ClusterOutlined />}
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
            <p style={{ marginTop: '16px' }}>正在计算持仓重叠度...</p>
          </div>
        ) : pairwiseResult ? (
          <>
            {/* 重叠度概览 */}
            <Row gutter={16} style={{ marginBottom: '24px' }}>
              <Col span={8}>
                <Card>
                  <Statistic
                    title="重叠度评分"
                    value={pairwiseResult.overlap_score}
                    precision={1}
                    suffix="%"
                    valueStyle={{
                      color: pairwiseResult.overlap_score > 50 ? '#ff4d4f' : pairwiseResult.overlap_score > 25 ? '#faad14' : '#52c41a'
                    }}
                  />
                  <Progress
                    percent={pairwiseResult.overlap_score}
                    status={pairwiseResult.overlap_score > 50 ? 'exception' : 'active'}
                    strokeColor={pairwiseResult.overlap_score > 50 ? '#ff4d4f' : pairwiseResult.overlap_score > 25 ? '#faad14' : '#52c41a'}
                  />
                </Card>
              </Col>
              <Col span={8}>
                <Card>
                  <Statistic
                    title="共同持仓数量"
                    value={pairwiseResult.common_holdings.length}
                    suffix="只"
                  />
                  <p style={{ marginTop: '8px', color: '#666' }}>
                    占 {pairwiseResult.etf1_symbol} 持仓的 {((pairwiseResult.common_holdings.length / 50) * 100).toFixed(1)}%
                  </p>
                </Card>
              </Col>
              <Col span={8}>
                <Card>
                  <Statistic
                    title="风险等级"
                    value={pairwiseResult.overlap_score > 50 ? '高' : pairwiseResult.overlap_score > 25 ? '中' : '低'}
                    valueStyle={{
                      color: pairwiseResult.overlap_score > 50 ? '#ff4d4f' : pairwiseResult.overlap_score > 25 ? '#faad14' : '#52c41a'
                    }}
                  />
                  <p style={{ marginTop: '8px', color: '#666' }}>
                    {pairwiseResult.overlap_score > 50
                      ? '高度重叠，建议分散投资'
                      : pairwiseResult.overlap_score > 25
                      ? '适度重叠，可接受范围'
                      : '分散良好，风险较低'}
                  </p>
                </Card>
              </Col>
            </Row>

            {/* 共同持仓表格 */}
            <Card title="共同持仓明细" style={{ marginBottom: '24px' }}>
              <Table
                columns={columns}
                dataSource={pairwiseResult.common_holdings}
                pagination={{ pageSize: 10 }}
                size="small"
              />
            </Card>

            {/* 行业分布对比 */}
            <Row gutter={16}>
              <Col span={12}>
                <Card title="行业重叠分布">
                  <ResponsiveContainer width="100%" height={300}>
                    <PieChart>
                      <Pie
                        data={Object.entries(pairwiseResult.sector_overlap).map(([name, value]) => ({ name, value }))}
                        cx="50%"
                        cy="50%"
                        innerRadius={60}
                        outerRadius={100}
                        paddingAngle={5}
                        dataKey="value"
                        label
                      >
                        {Object.entries(pairwiseResult.sector_overlap).map((_, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip />
                      <Legend />
                    </PieChart>
                  </ResponsiveContainer>
                </Card>
              </Col>
              <Col span={12}>
                <Card title="分析建议">
                  <List>
                    <List.Item>
                      <List.Item.Meta
                        avatar={<WarningOutlined style={{ color: '#faad14', fontSize: '24px' }} />}
                        title="集中度风险"
                        description={`前5大共同持仓占总重叠度的 ${pairwiseResult.common_holdings.slice(0, 5).reduce((sum, h) => sum + h.total_weight, 0).toFixed(1)}%`}
                      />
                    </List.Item>
                    <List.Item>
                      <List.Item.Meta
                        avatar={<ClusterOutlined style={{ color: '#52c41a', fontSize: '24px' }} />}
                        title="行业分散度"
                        description={`科技行业重叠度最高(${pairwiseResult.sector_overlap['Technology']?.toFixed(1)}%)，建议关注行业集中度风险`}
                      />
                    </List.Item>
                  </List>
                </Card>
              </Col>
            </Row>
          </>
        ) : portfolioResult ? (
          <>
            {/* 组合分析概览 */}
            <Row gutter={16} style={{ marginBottom: '24px' }}>
              <Col span={6}>
                <Card>
                  <Statistic
                    title="整体重叠度"
                    value={portfolioResult.overall_score}
                    precision={1}
                    suffix="%"
                    valueStyle={{
                      color: portfolioResult.overall_score > 50 ? '#ff4d4f' : portfolioResult.overall_score > 30 ? '#faad14' : '#52c41a'
                    }}
                  />
                </Card>
              </Col>
              <Col span={6}>
                <Card>
                  <Statistic
                    title="HHI集中度"
                    value={portfolioResult.concentration.hhi}
                    precision={0}
                    valueStyle={{
                      color: portfolioResult.concentration.hhi > 2500 ? '#ff4d4f' : portfolioResult.concentration.hhi > 1500 ? '#faad14' : '#52c41a'
                    }}
                  />
                  <Tag color={getRiskLevelColor(portfolioResult.concentration.risk_level)}>
                    {getRiskLevelText(portfolioResult.concentration.risk_level)}
                  </Tag>
                </Card>
              </Col>
              <Col span={6}>
                <Card>
                  <Statistic
                    title="有效分散数量"
                    value={portfolioResult.diversification.effective_n.toFixed(1)}
                  />
                  <p style={{ marginTop: '8px', color: '#666' }}>
                    覆盖 {portfolioResult.diversification.stock_count} 只股票
                  </p>
                </Card>
              </Col>
              <Col span={6}>
                <Card>
                  <Statistic
                    title="行业覆盖"
                    value={portfolioResult.diversification.sector_count}
                    suffix="个"
                  />
                  <p style={{ marginTop: '8px', color: '#666' }}>
                    覆盖 {portfolioResult.diversification.country_count} 个国家/地区
                  </p>
                </Card>
              </Col>
            </Row>

            {/* 集中度分析 */}
            <Card title="持仓集中度分析" style={{ marginBottom: '24px' }}>
              <Row gutter={16}>
                <Col span={12}>
                  <Statistic
                    title="前5大持仓集中度"
                    value={portfolioResult.concentration.top5_concentration}
                    precision={1}
                    suffix="%"
                  />
                  <Progress
                    percent={portfolioResult.concentration.top5_concentration}
                    status={portfolioResult.concentration.top5_concentration > 40 ? 'exception' : 'active'}
                  />
                </Col>
                <Col span={12}>
                  <Statistic
                    title="前10大持仓集中度"
                    value={portfolioResult.concentration.top10_concentration}
                    precision={1}
                    suffix="%"
                  />
                  <Progress
                    percent={portfolioResult.concentration.top10_concentration}
                    status={portfolioResult.concentration.top10_concentration > 60 ? 'exception' : 'active'}
                  />
                </Col>
              </Row>
            </Card>

            {/* 分散度评估 */}
            <Card title="分散度评估">
              <Alert
                message="分析结论"
                description={
                  portfolioResult.overall_score > 50
                    ? '投资组合持仓重叠度较高，存在明显的集中度风险。建议增加不同策略或行业的ETF，降低整体相关性。'
                    : portfolioResult.overall_score > 30
                    ? '投资组合分散度适中，部分持仓存在重叠。可适当调整权重或引入新的ETF以进一步优化分散效果。'
                    : '投资组合分散良好，持仓重叠度较低，风险分散效果较好。'
                }
                type={portfolioResult.overall_score > 50 ? 'error' : portfolioResult.overall_score > 30 ? 'warning' : 'success'}
                showIcon
              />
            </Card>
          </>
        ) : null}
      </div>
    </Layout>
  );
};

export default OverlapAnalysis;

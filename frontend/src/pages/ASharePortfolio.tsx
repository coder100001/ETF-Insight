import { useState, useEffect, useMemo } from 'react';
import styled from 'styled-components';
import {
  Row, Col, Card, Table, InputNumber, Statistic, Tag, Space,
  Tabs, Button, Alert, Divider
} from 'antd';
import {
  PieChartOutlined, BarChartOutlined, WalletOutlined,
  PercentageOutlined, MoneyCollectOutlined, CalendarOutlined,
  InfoCircleOutlined, EditOutlined, ReloadOutlined
} from '@ant-design/icons';
import {
  PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis,
  CartesianGrid, Tooltip as RechartsTooltip, Legend, ResponsiveContainer
} from 'recharts';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import { aSharePortfolioAPI } from '../services/api';
import type { AShareDividendCalculation, AShareHoldingDetail } from '../types';
import { App } from 'antd';

const { TabPane } = Tabs;

const PageHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;

  h2 {
    margin: 0;
    font-size: ${theme.fonts.size['2xl']};
    color: ${theme.colors.textPrimary};
    display: flex;
    align-items: center;
    gap: 10px;
  }
`;

const SummaryCard = styled(Card)`
  margin-bottom: 20px;
  box-shadow: ${theme.shadows.card};
  
  .ant-card-body {
    padding: 20px;
  }
`;

const SummaryGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  
  @media (max-width: ${theme.breakpoints.xl}) {
    grid-template-columns: repeat(2, 1fr);
  }
  
  @media (max-width: ${theme.breakpoints.md}) {
    grid-template-columns: 1fr;
  }
`;

const SummaryItem = styled.div<{ $color?: string }>`
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.md};
  padding: 16px;
  border-left: 4px solid ${props => props.$color || theme.colors.primary};
  
  .label {
    font-size: ${theme.fonts.size.sm};
    color: ${theme.colors.textSecondary};
    margin-bottom: 8px;
  }
  
  .value {
    font-size: ${theme.fonts.size['2xl']};
    font-weight: ${theme.fonts.weight.bold};
    color: ${props => props.$color || theme.colors.textPrimary};
  }
  
  .unit {
    font-size: ${theme.fonts.size.sm};
    color: ${theme.colors.textSecondary};
    margin-left: 4px;
  }
`;

const ChartCard = styled(Card)`
  margin-bottom: 20px;
  box-shadow: ${theme.shadows.card};
  
  .ant-card-head {
    background: ${theme.colors.background};
    border-bottom: 1px solid ${theme.colors.border};
  }
`;

const StyledTable = styled(Table)`
  .ant-table-thead > tr > th {
    background: ${theme.colors.background};
    font-weight: ${theme.fonts.weight.semibold};
  }
  
  .ant-table-tbody > tr:hover > td {
    background: ${theme.colors.background};
  }
`;

const EditableCell = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  
  .ant-input-number {
    width: 120px;
  }
`;

const FrequencyTag = styled(Tag)`
  &.monthly {
    background: #e6f7ff;
    color: #1890ff;
    border-color: #91d5ff;
  }
  &.quarterly {
    background: #f6ffed;
    color: #52c41a;
    border-color: #b7eb8f;
  }
  &.yearly {
    background: #fff7e6;
    color: #fa8c16;
    border-color: #ffd591;
  }
`;

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82CA9D', '#FFC658', '#FF6B6B'];

// 格式化金额
const formatMoney = (value: number) => {
  return new Intl.NumberFormat('zh-CN', {
    style: 'currency',
    currency: 'CNY',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(value);
};

// 格式化百分比
const formatPercent = (value: number) => {
  return `${value.toFixed(2)}%`;
};

// 默认投资金额（万元）
const DEFAULT_INVESTMENTS: Record<string, number> = {
  '515080': 125000,  // 中证红利ETF
  '515180': 100000,  // 红利ETF
  '515300': 150000,  // 中证红利低波动
  '510720': 80000,   // 红利国企ETF
  '520900': 100000,  // 红利低波ETF
  '159545': 75000,   // 红利ETF易方达
  '520550': 50000,   // 红利质量ETF
  '513820': 50000,   // 港股红利ETF
};

export default function ASharePortfolioPage() {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [portfolioData, setPortfolioData] = useState<AShareDividendCalculation | null>(null);
  const [investments, setInvestments] = useState<Record<string, number>>(DEFAULT_INVESTMENTS);
  const [editing, setEditing] = useState(false);

  // 加载默认组合数据
  const loadPortfolio = async () => {
    setLoading(true);
    try {
      const response = await aSharePortfolioAPI.getDefaultPortfolio();
      if (response.success && response.data) {
        setPortfolioData(response.data);
        // 同步投资金额
        const newInvestments: Record<string, number> = {};
        response.data.holdings.forEach(h => {
          newInvestments[h.symbol] = h.investment;
        });
        setInvestments(newInvestments);
      }
    } catch {
      message.error('加载组合数据失败');
    } finally {
      setLoading(false);
    }
  };

  // 分析组合
  const analyzePortfolio = async (newInvestments: Record<string, number>) => {
    try {
      const response = await aSharePortfolioAPI.analyzePortfolio(newInvestments);
      if (response.success && response.data) {
        setPortfolioData(response.data);
      }
    } catch {
      message.error('分析组合失败');
    }
  };

  useEffect(() => {
    loadPortfolio();
  }, []);

  // 更新投资金额
  const handleInvestmentChange = (symbol: string, value: number | null) => {
    if (value === null) return;
    
    const newInvestments = { ...investments, [symbol]: value };
    setInvestments(newInvestments);
    analyzePortfolio(newInvestments);
  };

  // 重置为默认值
  const handleReset = () => {
    setInvestments(DEFAULT_INVESTMENTS);
    analyzePortfolio(DEFAULT_INVESTMENTS);
    message.success('已重置为默认配置');
  };

  // 表格列定义
  const columns: import('antd').TableColumnsType<AShareHoldingDetail> = [
    {
      title: 'ETF代码',
      dataIndex: 'symbol',
      key: 'symbol',
      width: 100,
      render: (text) => <strong>{text as string}</strong>,
    },
    {
      title: 'ETF名称',
      dataIndex: 'name',
      key: 'name',
      width: 180,
    },
    {
      title: '投资金额',
      dataIndex: 'investment',
      key: 'investment',
      width: 150,
      render: (value, record) => {
        const r = record as AShareHoldingDetail;
        return editing ? (
          <EditableCell>
            <InputNumber
              value={investments[r.symbol] || (value as number)}
              onChange={(v) => handleInvestmentChange(r.symbol, v)}
              formatter={value => `¥ ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
              parser={value => value!.replace(/¥\s?|(,*)/g, '') as unknown as number}
              step={1000}
              min={0}
            />
          </EditableCell>
        ) : (
          formatMoney(value as number)
        );
      },
    },
    {
      title: '占比',
      dataIndex: 'weight',
      key: 'weight',
      width: 100,
      render: (value) => formatPercent(value as number),
    },
    {
      title: '股息率',
      dataIndex: 'dividend_yield',
      key: 'dividend_yield',
      width: 100,
      render: (value) => (
        <span style={{ color: theme.colors.success }}>{formatPercent(value as number)}</span>
      ),
    },
    {
      title: '分红频率',
      dataIndex: 'dividend_frequency',
      key: 'dividend_frequency',
      width: 100,
      render: (freq) => {
        const f = freq as string;
        const className = f === '月分' ? 'monthly' : f === '季分' ? 'quarterly' : 'yearly';
        return <FrequencyTag className={className}>{f}</FrequencyTag>;
      },
    },
    {
      title: '预期年分红',
      dataIndex: 'expected_dividend',
      key: 'expected_dividend',
      width: 150,
      render: (value) => formatMoney(value as number),
    },
    {
      title: '分红贡献',
      dataIndex: 'dividend_contribution',
      key: 'dividend_contribution',
      width: 100,
      render: (value) => formatPercent(value as number),
    },
  ];

  // 饼图数据
  const pieData = useMemo(() => {
    if (!portfolioData) return [];
    return portfolioData.holdings.map(h => ({
      name: h.symbol,
      value: h.investment,
      fullName: h.name,
    }));
  }, [portfolioData]);

  // 柱状图数据
  const barData = useMemo(() => {
    if (!portfolioData) return [];
    return portfolioData.holdings.map(h => ({
      name: h.symbol,
      预期分红: h.expected_dividend,
      投资金额: h.investment / 10, // 缩小比例以便显示
    }));
  }, [portfolioData]);

  if (!portfolioData) {
    return (
      <Layout>
        <Card loading={true} />
      </Layout>
    );
  }

  return (
    <Layout>
      <PageHeader>
        <h2>
          <WalletOutlined />
          A股红利ETF组合分析
        </h2>
        <Space>
          <Button
            type={editing ? 'primary' : 'default'}
            icon={<EditOutlined />}
            onClick={() => setEditing(!editing)}
          >
            {editing ? '完成编辑' : '调整金额'}
          </Button>
          <Button icon={<ReloadOutlined />} onClick={handleReset}>
            重置默认
          </Button>
        </Space>
      </PageHeader>

      <Alert
        message="A股红利ETF组合说明"
        description="本组合精选8只A股市场优质红利ETF，涵盖中证红利、红利低波、红利国企等多个策略。按每10万元投资计算，预期年化分红约5000元（股息率约5%）。"
        type="info"
        showIcon
        icon={<InfoCircleOutlined />}
        style={{ marginBottom: 20 }}
      />

      {/* 组合概览 */}
      <SummaryCard>
        <SummaryGrid>
          <SummaryItem $color="#1890ff">
            <div className="label">
              <WalletOutlined /> 总投资金额
            </div>
            <div className="value">
              {formatMoney(portfolioData.total_investment)}
            </div>
          </SummaryItem>
          
          <SummaryItem $color="#52c41a">
            <div className="label">
              <MoneyCollectOutlined /> 预期年分红
            </div>
            <div className="value">
              {formatMoney(portfolioData.expected_annual_dividend)}
            </div>
          </SummaryItem>
          
          <SummaryItem $color="#fa8c16">
            <div className="label">
              <PercentageOutlined /> 平均股息率
            </div>
            <div className="value">
              {formatPercent(portfolioData.average_dividend_yield)}
            </div>
          </SummaryItem>
          
          <SummaryItem $color="#722ed1">
            <div className="label">
              <CalendarOutlined /> 月均分红
            </div>
            <div className="value">
              {formatMoney(portfolioData.monthly_dividend)}
            </div>
          </SummaryItem>
        </SummaryGrid>
      </SummaryCard>

      {/* 图表区域 */}
      <Row gutter={20}>
        <Col span={12}>
          <ChartCard
            title={
              <Space>
                <PieChartOutlined />
                投资占比分布
              </Space>
            }
          >
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => `${name} ${(percent * 100).toFixed(1)}%`}
                  outerRadius={100}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {pieData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <RechartsTooltip
                  formatter={(value) => formatMoney(Number(value))}
                />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          </ChartCard>
        </Col>
        
        <Col span={12}>
          <ChartCard
            title={
              <Space>
                <BarChartOutlined />
                各ETF预期分红贡献
              </Space>
            }
          >
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={barData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <RechartsTooltip
                  formatter={(value, name) => {
                    if (name === '投资金额') {
                      return [formatMoney(Number(value) * 10), name as string];
                    }
                    return [formatMoney(Number(value)), name as string];
                  }}
                />
                <Legend />
                <Bar dataKey="预期分红" fill="#52c41a" />
                <Bar dataKey="投资金额" fill="#1890ff" />
              </BarChart>
            </ResponsiveContainer>
          </ChartCard>
        </Col>
      </Row>

      {/* ETF配置表格 */}
      <Card
        title={
          <Space>
            <BarChartOutlined />
            ETF配置明细
          </Space>
        }
        style={{ boxShadow: theme.shadows.card }}
      >
        <StyledTable
          dataSource={portfolioData.holdings}
          columns={columns}
          rowKey="symbol"
          pagination={false}
          loading={loading}
          summary={() => (
            <Table.Summary fixed>
              <Table.Summary.Row>
                <Table.Summary.Cell index={0} colSpan={2}>
                  <strong>合计</strong>
                </Table.Summary.Cell>
                <Table.Summary.Cell index={2}>
                  <strong>{formatMoney(portfolioData.total_investment)}</strong>
                </Table.Summary.Cell>
                <Table.Summary.Cell index={3}>
                  <strong>100%</strong>
                </Table.Summary.Cell>
                <Table.Summary.Cell index={4}>
                  <strong>{formatPercent(portfolioData.average_dividend_yield)}</strong>
                </Table.Summary.Cell>
                <Table.Summary.Cell index={5}>-</Table.Summary.Cell>
                <Table.Summary.Cell index={6}>
                  <strong style={{ color: theme.colors.success }}>
                    {formatMoney(portfolioData.expected_annual_dividend)}
                  </strong>
                </Table.Summary.Cell>
                <Table.Summary.Cell index={7}>
                  <strong>100%</strong>
                </Table.Summary.Cell>
              </Table.Summary.Row>
            </Table.Summary>
          )}
        />
      </Card>

      <Divider />

      {/* 分红时间维度展示 */}
      <Card
        title={
          <Space>
            <CalendarOutlined />
            分红收益时间维度
          </Space>
        }
        style={{ boxShadow: theme.shadows.card }}
      >
        <Tabs defaultActiveKey="quarterly">
          <TabPane tab="按年" key="yearly">
            <Row gutter={16}>
              <Col span={8}>
                <Statistic
                  title="年度分红总额"
                  value={portfolioData.expected_annual_dividend}
                  precision={0}
                  formatter={(value) => formatMoney(Number(value))}
                  valueStyle={{ color: '#52c41a' }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="相当于每月"
                  value={portfolioData.monthly_dividend}
                  precision={0}
                  formatter={(value) => formatMoney(Number(value))}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="收益率"
                  value={portfolioData.average_dividend_yield}
                  precision={2}
                  suffix="%"
                  valueStyle={{ color: '#1890ff' }}
                />
              </Col>
            </Row>
          </TabPane>
          <TabPane tab="按季" key="quarterly">
            <Row gutter={16}>
              <Col span={8}>
                <Statistic
                  title="季度分红总额"
                  value={portfolioData.quarterly_dividend}
                  precision={0}
                  formatter={(value) => formatMoney(Number(value))}
                  valueStyle={{ color: '#52c41a' }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="全年4个季度"
                  value={portfolioData.expected_annual_dividend}
                  precision={0}
                  formatter={(value) => formatMoney(Number(value))}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="季度收益率"
                  value={portfolioData.average_dividend_yield / 4}
                  precision={2}
                  suffix="%"
                  valueStyle={{ color: '#1890ff' }}
                />
              </Col>
            </Row>
          </TabPane>
          <TabPane tab="按月" key="monthly">
            <Row gutter={16}>
              <Col span={8}>
                <Statistic
                  title="月度分红总额"
                  value={portfolioData.monthly_dividend}
                  precision={0}
                  formatter={(value) => formatMoney(Number(value))}
                  valueStyle={{ color: '#52c41a' }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="全年12个月"
                  value={portfolioData.expected_annual_dividend}
                  precision={0}
                  formatter={(value) => formatMoney(Number(value))}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="月度收益率"
                  value={portfolioData.average_dividend_yield / 12}
                  precision={2}
                  suffix="%"
                  valueStyle={{ color: '#1890ff' }}
                />
              </Col>
            </Row>
          </TabPane>
        </Tabs>
      </Card>
    </Layout>
  );
}

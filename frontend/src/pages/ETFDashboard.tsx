import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import styled from 'styled-components';
import { Card, Table, Badge, Button, App, Row, Col } from 'antd';
import { BarChartOutlined, WalletOutlined } from '@ant-design/icons';
import { FaBalanceScale } from 'react-icons/fa';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import { etfAPI } from '../services/api';
import type { ETFData } from '../types';
import HoldingPieChart from '../components/HoldingPieChart';
import SectorBarChart from '../components/SectorBarChart';

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

const ActionButtons = styled.div`
  display: flex;
  gap: 10px;
`;

const ETFGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-bottom: 20px;

  @media (max-width: ${theme.breakpoints.xl}) {
    grid-template-columns: repeat(2, 1fr);
  }

  @media (max-width: ${theme.breakpoints.md}) {
    grid-template-columns: 1fr;
  }
`;

const ETFCard = styled(Card)`
  box-shadow: ${theme.shadows.card};
  transition: all ${theme.transitions.normal};

  &:hover {
    box-shadow: ${theme.shadows.hover};
  }

  .ant-card-head {
    background: ${theme.colors.background};
    border-bottom: 1px solid ${theme.colors.border};
  }

  .ant-card-body {
    padding: 16px;
  }
`;

const CardHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
`;

const ETFTitle = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;

  strong {
    font-size: ${theme.fonts.size.xl};
  }

  small {
    color: ${theme.colors.textMuted};
  }
`;

const PriceDisplay = styled.div`
  text-align: center;
  margin: 16px 0;

  h3 {
    font-size: ${theme.fonts.size['3xl']};
    margin: 0;
    color: ${theme.colors.textPrimary};
  }
`;

const InfoTable = styled.table`
  width: 100%;
  font-size: ${theme.fonts.size.sm};
  margin-bottom: 12px;

  td {
    padding: 6px 0;
    border-bottom: 1px solid ${theme.colors.border};

    &:first-child {
      color: ${theme.colors.textSecondary};
    }

    &:last-child {
      text-align: right;
    }
  }

  tr:last-child td {
    border-bottom: none;
  }
`;

const StrategyText = styled.p`
  font-size: ${theme.fonts.size.sm};
  color: ${theme.colors.textMuted};
  margin: 0 0 12px 0;
`;

const StyledTable = styled(Table)`
  .ant-table-thead > tr > th {
    background: ${theme.colors.background};
    font-weight: ${theme.fonts.weight.semibold};
  }

  .ant-table-tbody > tr:hover > td {
    background: #f8f9fa;
  }
`;

interface ETFApiItem {
  symbol: string;
  name: string;
  current_price?: number;
  previous_close?: number;
  change?: number;
  change_percent?: number;
  open_price?: number;
  high_price?: number;
  low_price?: number;
  volume?: number;
  dividend_yield?: number;
  volatility?: number;
  total_return?: number;
  max_drawdown?: number;
  sharpe_ratio?: number;
  expense_ratio?: number;
  focus?: string;
  strategy?: string;
}

const ETFDashboard: React.FC = () => {
  const { message } = App.useApp();
  const [etfs, setEtfs] = useState<ETFData[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetchETFData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const fetchETFData = async () => {
    setLoading(true);
    try {
      const response = await etfAPI.getList();
      if (response.success && response.data) {
        // 转换后端数据为前端格式
        const formattedData: ETFData[] = response.data.map((item: ETFApiItem) => ({
          symbol: item.symbol,
          name: item.name,
          current_price: item.current_price || 0,
          previous_close: item.previous_close || 0,
          change: item.change || 0,
          change_percent: item.change_percent || 0,
          open_price: item.open_price || 0,
          high_price: item.high_price || 0,
          low_price: item.low_price || 0,
          volume: item.volume || 0,
          dividend_yield: item.dividend_yield || 0,
          volatility: item.volatility || 0,
          total_return: item.total_return || 0,
          max_drawdown: item.max_drawdown || 0,
          sharpe_ratio: item.sharpe_ratio || 0,
          expense_ratio: item.expense_ratio || 0,
          info: {
            focus: item.focus || '',
            strategy: item.strategy || '',
          },
        }));
        setEtfs(formattedData);
      } else {
        message.error('获取ETF数据失败');
      }
    } catch (error) {
      message.error('获取ETF数据失败: ' + (error as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const comparisonColumns = [
    { title: '指标', dataIndex: 'indicator', key: 'indicator', fixed: 'left' as const },
    ...etfs.map(etf => ({
      title: etf.symbol,
      dataIndex: etf.symbol,
      key: etf.symbol,
      align: 'center' as const,
    })),
  ];

  const comparisonData = [
    {
      indicator: '当前价格',
      ...etfs.reduce((acc, etf) => ({
        ...acc,
        [etf.symbol]: `$${etf.current_price.toFixed(2)}`,
      }), {}),
    },
    {
      indicator: '今日涨跌',
      ...etfs.reduce((acc, etf) => ({
        ...acc,
        [etf.symbol]: (
          <span style={{ color: etf.change_percent >= 0 ? theme.colors.success : theme.colors.danger }}>
            {etf.change_percent >= 0 ? '+' : ''}{etf.change_percent.toFixed(2)}%
          </span>
        ),
      }), {}),
    },
    {
      indicator: '股息率',
      ...etfs.reduce((acc, etf) => ({
        ...acc,
        [etf.symbol]: etf.dividend_yield ? `${etf.dividend_yield.toFixed(2)}%` : '-',
      }), {}),
    },
    {
      indicator: '年度收益',
      ...etfs.reduce((acc, etf) => ({
        ...acc,
        [etf.symbol]: (
          <span style={{ color: etf.total_return >= 0 ? theme.colors.success : theme.colors.danger }}>
            {etf.total_return >= 0 ? '+' : ''}{etf.total_return.toFixed(2)}%
          </span>
        ),
      }), {}),
    },
    {
      indicator: '年化波动率',
      ...etfs.reduce((acc, etf) => ({
        ...acc,
        [etf.symbol]: `${etf.volatility.toFixed(2)}%`,
      }), {}),
    },
    {
      indicator: '夏普比率',
      ...etfs.reduce((acc, etf) => ({
        ...acc,
        [etf.symbol]: etf.sharpe_ratio.toFixed(2),
      }), {}),
    },
    {
      indicator: '最大回撤',
      ...etfs.reduce((acc, etf) => ({
        ...acc,
        [etf.symbol]: <span style={{ color: theme.colors.danger }}>{etf.max_drawdown.toFixed(2)}%</span>,
      }), {}),
    },
    {
      indicator: '策略类型',
      ...etfs.reduce((acc, etf) => ({
        ...acc,
        [etf.symbol]: <small>{etf.info.strategy}</small>,
      }), {}),
    },
  ];

  return (
    <Layout>
      <PageHeader>
        <h2>
          <BarChartOutlined />
          ETF数据分析
        </h2>
        <ActionButtons>
          <Link to="/etf-comparison">
            <Button type="primary" icon={<FaBalanceScale />}>
              对比分析
            </Button>
          </Link>
          <Link to="/portfolio-analysis">
            <Button icon={<WalletOutlined />} style={{ background: theme.colors.success, borderColor: theme.colors.success, color: '#fff' }}>
              组合分析
            </Button>
          </Link>
        </ActionButtons>
      </PageHeader>

      {/* ETF卡片列表 */}
      <ETFGrid>
        {loading && <Card loading style={{ gridColumn: '1 / -1' }} />}
        {!loading && etfs.map(etf => (
          <ETFCard
            key={etf.symbol}
            title={
              <CardHeader>
                <ETFTitle>
                  <strong>{etf.symbol}</strong>
                  <small>{etf.info.focus}</small>
                </ETFTitle>
                <Badge
                  count={`${etf.change_percent >= 0 ? '+' : ''}${etf.change_percent.toFixed(2)}%`}
                  style={{
                    backgroundColor: etf.change_percent >= 0 ? theme.colors.success : theme.colors.danger,
                  }}
                />
              </CardHeader>
            }
          >
            <PriceDisplay>
              <h3>${etf.current_price.toFixed(2)}</h3>
            </PriceDisplay>

            <InfoTable>
              <tbody>
                <tr>
                  <td>名称</td>
                  <td><small>{etf.name.length > 30 ? `${etf.name.slice(0, 30)}...` : etf.name}</small></td>
                </tr>
                <tr>
                  <td>股息率</td>
                  <td>{etf.dividend_yield ? `${etf.dividend_yield.toFixed(2)}%` : '-'}</td>
                </tr>
                <tr>
                  <td>年化波动率</td>
                  <td>{etf.volatility.toFixed(2)}%</td>
                </tr>
                <tr>
                  <td>年度收益</td>
                  <td style={{ color: etf.total_return >= 0 ? theme.colors.success : theme.colors.danger }}>
                    {etf.total_return.toFixed(2)}%
                  </td>
                </tr>
                <tr>
                  <td>最大回撤</td>
                  <td style={{ color: theme.colors.danger }}>{etf.max_drawdown.toFixed(2)}%</td>
                </tr>
                <tr>
                  <td>夏普比率</td>
                  <td>{etf.sharpe_ratio.toFixed(2)}</td>
                </tr>
                <tr>
                  <td>费率</td>
                  <td>{etf.expense_ratio}%</td>
                </tr>
              </tbody>
            </InfoTable>

            <StrategyText>{etf.info.strategy}</StrategyText>

            <Link to={`/etf-detail/${etf.symbol}`}>
              <Button block icon={<BarChartOutlined />}>
                查看详情
              </Button>
            </Link>
          </ETFCard>
        ))}
      </ETFGrid>

      {/* 数据可视化区域 */}
      {etfs.length > 0 && (
        <>
          <Row gutter={16} style={{ marginBottom: 20 }}>
            <Col xs={24} lg={12}>
              <HoldingPieChart
                data={etfs.map(etf => ({
                  symbol: etf.symbol,
                  name: etf.name,
                  weight: 100 / etfs.length, // 平均权重
                  value: etf.current_price * 100, // 示例值
                }))}
                title="ETF 持仓分布"
              />
            </Col>
            <Col xs={24} lg={12}>
              <SectorBarChart
                data={[
                  { name: '科技', weight: 35, value: 35000 },
                  { name: '金融', weight: 25, value: 25000 },
                  { name: '医疗', weight: 15, value: 15000 },
                  { name: '消费', weight: 15, value: 15000 },
                  { name: '能源', weight: 10, value: 10000 },
                ]}
                title="行业分布"
              />
            </Col>
          </Row>
        </>
      )}

      {/* 快速对比表格 */}
      <Card
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <FaBalanceScale />
            <span>快速对比</span>
          </div>
        }
        style={{ boxShadow: theme.shadows.card }}
      >
        <StyledTable
          dataSource={comparisonData}
          columns={comparisonColumns}
          rowKey="indicator"
          pagination={false}
          scroll={{ x: 'max-content' }}
          bordered
          size="middle"
        />
      </Card>
    </Layout>
  );
};

export default ETFDashboard;

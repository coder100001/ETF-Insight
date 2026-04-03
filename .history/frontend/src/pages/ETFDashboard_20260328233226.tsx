import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import styled from 'styled-components';
import { Card, Table, Badge, Button, App } from 'antd';
import { BarChartOutlined, WalletOutlined } from '@ant-design/icons';
import { FaBalanceScale } from 'react-icons/fa';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import { etfAPI } from '../services/api';
import type { ETFData } from '../types';

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

// 模拟ETF数据
const mockETFData: ETFData[] = [
  {
    symbol: 'SCHD',
    name: 'Schwab US Dividend Equity ETF',
    current_price: 30.44,
    previous_close: 31.67,
    change: -1.23,
    change_percent: -3.88,
    open_price: 30.35,
    high_price: 30.59,
    low_price: 30.20,
    volume: 8500000,
    dividend_yield: 3.45,
    volatility: 15.2,
    total_return: 12.5,
    max_drawdown: -8.3,
    sharpe_ratio: 1.2,
    expense_ratio: 0.06,
    info: {
      focus: '美股高股息',
      strategy: '质量因子筛选',
    },
  },
  {
    symbol: 'SPYD',
    name: 'SPDR S&P 500 High Dividend ETF',
    current_price: 47.85,
    previous_close: 48.14,
    change: -0.29,
    change_percent: -0.60,
    open_price: 47.71,
    high_price: 48.09,
    low_price: 47.47,
    volume: 6200000,
    dividend_yield: 4.12,
    volatility: 16.8,
    total_return: 8.3,
    max_drawdown: -10.5,
    sharpe_ratio: 0.9,
    expense_ratio: 0.07,
    info: {
      focus: 'S&P高股息',
      strategy: '股息率加权',
    },
  },
  {
    symbol: 'JEPQ',
    name: 'JPMorgan Nasdaq Equity Premium Income ETF',
    current_price: 57.20,
    previous_close: 57.51,
    change: -0.31,
    change_percent: -0.54,
    open_price: 57.03,
    high_price: 57.49,
    low_price: 56.74,
    volume: 4800000,
    dividend_yield: 11.2,
    volatility: 18.5,
    total_return: 15.8,
    max_drawdown: -12.1,
    sharpe_ratio: 1.1,
    expense_ratio: 0.35,
    info: {
      focus: '纳斯达克备兑',
      strategy: '期权收益增强',
    },
  },
  {
    symbol: 'JEPI',
    name: 'JPMorgan Equity Premium Income ETF',
    current_price: 58.90,
    previous_close: 59.31,
    change: -0.41,
    change_percent: -0.69,
    open_price: 58.72,
    high_price: 59.19,
    low_price: 58.43,
    volume: 9200000,
    dividend_yield: 9.8,
    volatility: 14.2,
    total_return: 11.2,
    max_drawdown: -7.5,
    sharpe_ratio: 1.3,
    expense_ratio: 0.35,
    info: {
      focus: '美股备兑',
      strategy: '期权收益增强',
    },
  },
  {
    symbol: 'VYM',
    name: 'Vanguard High Dividend Yield ETF',
    current_price: 154.50,
    previous_close: 155.37,
    change: -0.87,
    change_percent: -0.56,
    open_price: 154.04,
    high_price: 155.27,
    low_price: 153.26,
    volume: 3800000,
    dividend_yield: 2.95,
    volatility: 13.8,
    total_return: 9.6,
    max_drawdown: -9.2,
    sharpe_ratio: 1.0,
    expense_ratio: 0.06,
    info: {
      focus: '美股高股息',
      strategy: '股息率筛选',
    },
  },
];

const ETFDashboard: React.FC = () => {
  const { message } = App.useApp();
  const [etfs, setEtfs] = useState<ETFData[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetchETFData();
  }, []);

  const fetchETFData = async () => {
    setLoading(true);
    try {
      const response = await etfAPI.getList();
      if (response.success && response.data) {
        // 转换后端数据为前端格式
        const formattedData: ETFData[] = response.data.map((item: any) => ({
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
        {etfs.map(etf => (
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

import { useState } from 'react';
import styled from 'styled-components';
import { Card, Table, Button, Select, message } from 'antd';
import { BarChartOutlined } from '@ant-design/icons';
import { FaBalanceScale } from 'react-icons/fa';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
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

const FilterSection = styled.div`
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
  padding: 16px;
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.md};
  box-shadow: ${theme.shadows.card};
`;

const StyledTable = styled(Table)`
  .ant-table-thead > tr > th {
    background: ${theme.colors.background};
    font-weight: ${theme.fonts.weight.semibold};
  }

  .ant-table-tbody > tr:hover > td {
    background: #f8f9fa;
  }
` as typeof Table;

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

const ETFComparison: React.FC = () => {
  const [etfs] = useState<ETFData[]>(mockETFData);
  const [selectedETFs, setSelectedETFs] = useState<string[]>(['SCHD', 'SPYD', 'JEPQ']);

  const handleCompare = () => {
    if (selectedETFs.length < 2) {
      message.warning('请至少选择2个ETF进行对比');
      return;
    }
    message.success(`已选择 ${selectedETFs.length} 个ETF进行对比`);
  };

  const filteredETFs = etfs.filter(etf => selectedETFs.includes(etf.symbol));

  const columns = [
    {
      title: 'ETF',
      dataIndex: 'symbol',
      key: 'symbol',
      render: (text: string, record: ETFData) => (
        <div>
          <strong>{text}</strong>
          <br />
          <small style={{ color: theme.colors.textMuted }}>{record.name}</small>
        </div>
      ),
    },
    {
      title: '当前价格',
      dataIndex: 'current_price',
      key: 'current_price',
      align: 'center' as const,
      render: (value: number) => `$${value.toFixed(2)}`,
    },
    {
      title: '今日涨跌',
      dataIndex: 'change_percent',
      key: 'change_percent',
      align: 'center' as const,
      render: (value: number) => (
        <span style={{ color: value >= 0 ? theme.colors.success : theme.colors.danger }}>
          {value >= 0 ? '+' : ''}{value.toFixed(2)}%
        </span>
      ),
    },
    {
      title: '股息率',
      dataIndex: 'dividend_yield',
      key: 'dividend_yield',
      align: 'center' as const,
      render: (value?: number) => value ? `${value.toFixed(2)}%` : '-',
    },
    {
      title: '年化波动率',
      dataIndex: 'volatility',
      key: 'volatility',
      align: 'center' as const,
      render: (value: number) => `${value.toFixed(2)}%`,
    },
    {
      title: '夏普比率',
      dataIndex: 'sharpe_ratio',
      key: 'sharpe_ratio',
      align: 'center' as const,
      render: (value: number) => value.toFixed(2),
    },
    {
      title: '最大回撤',
      dataIndex: 'max_drawdown',
      key: 'max_drawdown',
      align: 'center' as const,
      render: (value: number) => (
        <span style={{ color: theme.colors.danger }}>{value.toFixed(2)}%</span>
      ),
    },
    {
      title: '费率',
      dataIndex: 'expense_ratio',
      key: 'expense_ratio',
      align: 'center' as const,
      render: (value: number) => `${value}%`,
    },
    {
      title: '策略',
      dataIndex: ['info', 'strategy'],
      key: 'strategy',
      align: 'center' as const,
    },
  ];

  return (
    <Layout>
      <PageHeader>
        <h2>
          <FaBalanceScale />
          ETF对比分析
        </h2>
      </PageHeader>

      <FilterSection>
        <Select
          mode="multiple"
          placeholder="选择要对比的ETF"
          value={selectedETFs}
          onChange={setSelectedETFs}
          style={{ minWidth: 300 }}
          options={etfs.map(etf => ({
            label: `${etf.symbol} - ${etf.name}`,
            value: etf.symbol,
          }))}
        />
        <Button type="primary" icon={<BarChartOutlined />} onClick={handleCompare}>
          开始对比
        </Button>
      </FilterSection>

      <Card style={{ boxShadow: theme.shadows.card }}>
        <StyledTable
          dataSource={filteredETFs}
          columns={columns as any}
          rowKey="symbol"
          pagination={false}
          bordered
        />
      </Card>
    </Layout>
  );
};

export default ETFComparison;

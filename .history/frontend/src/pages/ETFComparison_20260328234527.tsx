import { useState, useEffect } from 'react';
import styled from 'styled-components';
import { Card, Table, Button, Select, App } from 'antd';
import { BarChartOutlined } from '@ant-design/icons';
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

const ETFComparison: React.FC = () => {
  const { message } = App.useApp();
  const [etfs, setEtfs] = useState<ETFData[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedETFs, setSelectedETFs] = useState<string[]>(['SCHD', 'SPYD', 'JEPQ']);

  useEffect(() => {
    fetchETFData();
  }, []);

  const fetchETFData = async () => {
    setLoading(true);
    try {
      const response = await etfAPI.getList();
      if (response.success && response.data) {
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

  const handleCompare = () => {
    if (selectedETFs.length < 2) {
      message.warning('请至少选择2个ETF进行对比');
      return;
    }
    message.success(`已选择 ${selectedETFs.length} 个ETF进行对比`);
  };

  const filteredETFs = etfs.filter(etf => selectedETFs.includes(etf.symbol));

  const columns: import('antd').TableProps<ETFData>['columns'] = [
    {
      title: 'ETF',
      dataIndex: 'symbol',
      key: 'symbol',
      render: (text, record) => (
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
      render: (value) => `$${(value as number).toFixed(2)}`,
    },
    {
      title: '今日涨跌',
      dataIndex: 'change_percent',
      key: 'change_percent',
      align: 'center' as const,
      render: (value) => (
        <span style={{ color: (value as number) >= 0 ? theme.colors.success : theme.colors.danger }}>
          {(value as number) >= 0 ? '+' : ''}{(value as number).toFixed(2)}%
        </span>
      ),
    },
    {
      title: '股息率',
      dataIndex: 'dividend_yield',
      key: 'dividend_yield',
      align: 'center' as const,
      render: (value) => value ? `${(value as number).toFixed(2)}%` : '-',
    },
    {
      title: '年化波动率',
      dataIndex: 'volatility',
      key: 'volatility',
      align: 'center' as const,
      render: (value) => `${(value as number).toFixed(2)}%`,
    },
    {
      title: '夏普比率',
      dataIndex: 'sharpe_ratio',
      key: 'sharpe_ratio',
      align: 'center' as const,
      render: (value) => (value as number).toFixed(2),
    },
    {
      title: '最大回撤',
      dataIndex: 'max_drawdown',
      key: 'max_drawdown',
      align: 'center' as const,
      render: (value) => (
        <span style={{ color: theme.colors.danger }}>{(value as number).toFixed(2)}%</span>
      ),
    },
    {
      title: '费率',
      dataIndex: 'expense_ratio',
      key: 'expense_ratio',
      align: 'center' as const,
      render: (value) => `${value}%`,
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
          columns={columns}
          rowKey="symbol"
          pagination={false}
          bordered
          loading={loading}
        />
      </Card>
    </Layout>
  );
};

export default ETFComparison;

import { useState, useEffect } from 'react';
import styled from 'styled-components';
import { Card, Table, Button, Badge, Space, App } from 'antd';
import { SwapOutlined, ReloadOutlined, EditOutlined, HistoryOutlined } from '@ant-design/icons';
import axios from 'axios';

import Layout from '../components/Layout';
import { theme } from '../styles/theme';

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

const StatsRow = styled.div`
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 20px;

  @media (max-width: ${theme.breakpoints.lg}) {
    grid-template-columns: repeat(2, 1fr);
  }

  @media (max-width: ${theme.breakpoints.sm}) {
    grid-template-columns: 1fr;
  }
`;

const StatCard = styled.div<{ $borderColor?: string }>`
  background: ${theme.colors.surface};
  padding: 20px;
  border-radius: ${theme.borderRadius.md};
  box-shadow: ${theme.shadows.card};
  border-left: 4px solid ${props => props.$borderColor || theme.colors.primary};

  h3 {
    font-size: ${theme.fonts.size['2xl']};
    margin: 0 0 5px 0;
    color: ${props => props.$borderColor || theme.colors.primary};
  }

  p {
    color: ${theme.colors.textSecondary};
    margin: 0;
    font-size: ${theme.fonts.size.base};
  }
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

interface ExchangeRate {
  id: number;
  from_currency: string;
  to_currency: string;
  rate: number;
  previous_rate: number;
  change_percent: number;
  updated_at: string;
  source: string;
}

const ExchangeRatePage: React.FC = () => {
  const { message } = App.useApp();
  const [rates, setRates] = useState<ExchangeRate[]>([]);
  const [loading, setLoading] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<string>('');

  useEffect(() => {
    fetchRates();
    // 每5分钟自动刷新一次
    const interval = setInterval(fetchRates, 5 * 60 * 1000);
    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const fetchRates = async () => {
    setLoading(true);
    try {
      const response = await axios.get('/api/exchange-rates');
      if (response.data.success && response.data.data) {
        setRates(response.data.data);
        setLastUpdated(new Date().toLocaleString('zh-CN'));
      } else {
        message.error('获取汇率数据失败');
      }
    } catch (error) {
      message.error('获取汇率数据失败: ' + (error as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    fetchRates();
    message.success('汇率数据已更新');
  };

  const columns: import('antd').TableProps<ExchangeRate>['columns'] = [
    {
      title: '货币对',
      dataIndex: 'from_currency',
      key: 'pair',
      render: (_, record) => (
        <strong>{record.from_currency}/{record.to_currency}</strong>
      ),
    },
    {
      title: '当前汇率',
      dataIndex: 'rate',
      key: 'rate',
      align: 'center' as const,
      render: (rate) => <strong>{(rate as number).toFixed(4)}</strong>,
    },
    {
      title: '涨跌',
      dataIndex: 'change_percent',
      key: 'change_percent',
      align: 'center' as const,
      render: (value) => (
        <Badge
          count={`${(value as number) >= 0 ? '+' : ''}${(value as number).toFixed(2)}%`}
          style={{
            backgroundColor: (value as number) >= 0 ? theme.colors.success : theme.colors.danger,
          }}
        />
      ),
    },
    {
      title: '数据源',
      dataIndex: 'source',
      key: 'source',
      align: 'center' as const,
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      align: 'center' as const,
    },
    {
      title: '操作',
      key: 'action',
      align: 'center' as const,
      render: () => (
        <Space>
          <Button size="small" icon={<EditOutlined />}>编辑</Button>
          <Button size="small" icon={<HistoryOutlined />}>历史</Button>
        </Space>
      ),
    },
  ];

  return (
    <Layout>
      <PageHeader>
        <h2>
          <SwapOutlined />
          外汇管理
        </h2>
        <Button type="primary" icon={<ReloadOutlined />} onClick={handleRefresh}>
          刷新汇率
        </Button>
      </PageHeader>

      <StatsRow>
        <StatCard $borderColor={theme.colors.primary}>
          <h3>USD/CNY</h3>
          <p>7.2345 (+0.26%)</p>
        </StatCard>
        <StatCard $borderColor={theme.colors.success}>
          <h3>EUR/CNY</h3>
          <p>7.8234 (-0.28%)</p>
        </StatCard>
        <StatCard $borderColor={theme.colors.warning}>
          <h3>GBP/CNY</h3>
          <p>9.1234 (+0.20%)</p>
        </StatCard>
        <StatCard $borderColor={theme.colors.info}>
          <h3>JPY/CNY</h3>
          <p>0.0478 (-0.62%)</p>
        </StatCard>
      </StatsRow>

      <Card 
        style={{ boxShadow: theme.shadows.card }}
        title={
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>汇率列表</span>
            {lastUpdated && (
              <span style={{ fontSize: '12px', color: theme.colors.textSecondary }}>
                最后更新: {lastUpdated}
              </span>
            )}
          </div>
        }
      >
        <StyledTable
          dataSource={rates}
          columns={columns}
          rowKey="id"
          pagination={false}
          loading={loading}
        />
      </Card>
    </Layout>
  );
};

export default ExchangeRatePage;

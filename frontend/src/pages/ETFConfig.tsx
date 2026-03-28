import { useState } from 'react';
import styled from 'styled-components';
import { Card, Table, Button, Space, Switch, Tag, type TableColumnsType } from 'antd';
import { SettingOutlined, EditOutlined, ReloadOutlined } from '@ant-design/icons';
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

const StyledTable = styled(Table)`
  .ant-table-thead > tr > th {
    background: ${theme.colors.background};
    font-weight: ${theme.fonts.weight.semibold};
  }

  .ant-table-tbody > tr:hover > td {
    background: #f8f9fa;
  }
` as typeof Table;

interface ETFConfig {
  id: number;
  symbol: string;
  name: string;
  is_active: boolean;
  auto_update: boolean;
  update_frequency: string;
  last_updated: string;
  data_source: string;
}

const mockConfigs: ETFConfig[] = [
  {
    id: 1,
    symbol: 'SCHD',
    name: 'Schwab US Dividend Equity ETF',
    is_active: true,
    auto_update: true,
    update_frequency: '每日',
    last_updated: '2024-03-28 09:00:00',
    data_source: 'Yahoo Finance',
  },
  {
    id: 2,
    symbol: 'SPYD',
    name: 'SPDR S&P 500 High Dividend ETF',
    is_active: true,
    auto_update: true,
    update_frequency: '每日',
    last_updated: '2024-03-28 09:00:00',
    data_source: 'Yahoo Finance',
  },
  {
    id: 3,
    symbol: 'JEPQ',
    name: 'JPMorgan Nasdaq Equity Premium Income ETF',
    is_active: true,
    auto_update: true,
    update_frequency: '每日',
    last_updated: '2024-03-28 09:00:00',
    data_source: 'Yahoo Finance',
  },
  {
    id: 4,
    symbol: 'JEPI',
    name: 'JPMorgan Equity Premium Income ETF',
    is_active: true,
    auto_update: true,
    update_frequency: '每日',
    last_updated: '2024-03-28 09:00:00',
    data_source: 'Yahoo Finance',
  },
  {
    id: 5,
    symbol: 'VYM',
    name: 'Vanguard High Dividend Yield ETF',
    is_active: true,
    auto_update: true,
    update_frequency: '每日',
    last_updated: '2024-03-28 09:00:00',
    data_source: 'Yahoo Finance',
  },
];

const ETFConfigPage: React.FC = () => {
  const [configs, setConfigs] = useState<ETFConfig[]>(mockConfigs);

  const handleToggleActive = (id: number, checked: boolean) => {
    setConfigs(prev =>
      prev.map(config =>
        config.id === id ? { ...config, is_active: checked } : config
      )
    );
  };

  const handleToggleAutoUpdate = (id: number, checked: boolean) => {
    setConfigs(prev =>
      prev.map(config =>
        config.id === id ? { ...config, auto_update: checked } : config
      )
    );
  };

  const columns: TableColumnsType<ETFConfig> = [
    {
      title: '代码',
      dataIndex: 'symbol',
      key: 'symbol',
      render: (text: string) => <strong>{text}</strong>,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      align: 'center',
      render: (is_active: boolean, record: ETFConfig) => (
        <Switch
          checked={is_active}
          onChange={(checked) => handleToggleActive(record.id, checked)}
        />
      ),
    },
    {
      title: '自动更新',
      dataIndex: 'auto_update',
      key: 'auto_update',
      align: 'center',
      render: (auto_update: boolean, record: ETFConfig) => (
        <Switch
          checked={auto_update}
          onChange={(checked) => handleToggleAutoUpdate(record.id, checked)}
        />
      ),
    },
    {
      title: '更新频率',
      dataIndex: 'update_frequency',
      key: 'update_frequency',
      align: 'center',
      render: (freq: string) => <Tag>{freq}</Tag>,
    },
    {
      title: '数据源',
      dataIndex: 'data_source',
      key: 'data_source',
      align: 'center',
    },
    {
      title: '最后更新',
      dataIndex: 'last_updated',
      key: 'last_updated',
      align: 'center',
    },
    {
      title: '操作',
      key: 'action',
      align: 'center',
      render: () => (
        <Space>
          <Button size="small" icon={<ReloadOutlined />}>更新</Button>
          <Button size="small" icon={<EditOutlined />}>编辑</Button>
        </Space>
      ),
    },
  ];

  return (
    <Layout>
      <PageHeader>
        <h2>
          <SettingOutlined />
          ETF配置
        </h2>
        <Button type="primary">添加ETF</Button>
      </PageHeader>

      <Card style={{ boxShadow: theme.shadows.card }}>
        <StyledTable
          dataSource={configs}
          columns={columns}
          rowKey="id"
          pagination={false}
        />
      </Card>
    </Layout>
  );
};

export default ETFConfigPage;

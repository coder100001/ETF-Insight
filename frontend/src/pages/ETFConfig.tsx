import { useState, useEffect } from 'react';
import styled from 'styled-components';
import { Card, Table, Button, Space, Switch, Tag, type TableColumnsType, message } from 'antd';
import { SettingOutlined, EditOutlined, ReloadOutlined } from '@ant-design/icons';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import { etfConfigAPI } from '../services/api';
import type { ETFConfig } from '../types';

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

const ETFConfigPage: React.FC = () => {
  const [configs, setConfigs] = useState<ETFConfig[]>([]);
  const [loading, setLoading] = useState(false);

  // 加载ETF配置列表
  const loadConfigs = async () => {
    setLoading(true);
    try {
      const response = await etfConfigAPI.getConfigs();
      if (response.success && response.data) {
        const formattedConfigs = response.data.map(config => ({
          ...config,
          is_active: config.status === 1,
          auto_update: config.auto_update ?? true,
          update_frequency: config.update_frequency ?? '每日',
          last_updated: config.last_updated ?? config.updated_at ?? '-',
          data_source: config.data_source ?? 'Finage',
        }));
        setConfigs(formattedConfigs);
      }
    } catch (error) {
      console.error('Failed to load ETF configs:', error);
      message.error('加载ETF配置失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadConfigs();
  }, []);

  // 切换状态
  const handleToggleActive = async (id: number, checked: boolean) => {
    // 先更新本地状态
    setConfigs(prev =>
      prev.map(config =>
        config.id === id ? { ...config, is_active: checked, status: checked ? 1 : 0 } : config
      )
    );

    try {
      const response = await etfConfigAPI.toggleStatus(id, checked ? 1 : 0);
      if (response.success) {
        message.success(`已${checked ? '启用' : '禁用'}ETF`);
      } else {
        // 失败时恢复状态
        setConfigs(prev =>
          prev.map(config =>
            config.id === id ? { ...config, is_active: !checked, status: !checked ? 1 : 0 } : config
          )
        );
        message.error('更新状态失败');
      }
    } catch (error) {
      console.error('Failed to toggle status:', error);
      // 失败时恢复状态
      setConfigs(prev =>
        prev.map(config =>
          config.id === id ? { ...config, is_active: !checked, status: !checked ? 1 : 0 } : config
        )
      );
      message.error('更新状态失败');
    }
  };

  // 切换自动更新
  const handleToggleAutoUpdate = async (id: number, checked: boolean) => {
    // 先更新本地状态
    setConfigs(prev =>
      prev.map(config =>
        config.id === id ? { ...config, auto_update: checked } : config
      )
    );

    try {
      const response = await etfConfigAPI.toggleAutoUpdate(id, checked);
      if (response.success) {
        message.success(`已${checked ? '开启' : '关闭'}自动更新`);
      } else {
        // 失败时恢复状态
        setConfigs(prev =>
          prev.map(config =>
            config.id === id ? { ...config, auto_update: !checked } : config
          )
        );
        message.error('更新自动更新设置失败');
      }
    } catch (error) {
      console.error('Failed to toggle auto update:', error);
      // 失败时恢复状态
      setConfigs(prev =>
        prev.map(config =>
          config.id === id ? { ...config, auto_update: !checked } : config
        )
      );
      message.error('更新自动更新设置失败');
    }
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
          loading={loading}
        />
      </Card>
    </Layout>
  );
};

export default ETFConfigPage;

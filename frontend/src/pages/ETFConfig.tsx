import { useState, useEffect } from 'react';
import styled from 'styled-components';
import { Card, Table, Button, Space, Switch, Tag, Modal, Form, Input, InputNumber, Select, type TableColumnsType, message } from 'antd';
import { SettingOutlined, EditOutlined, ReloadOutlined, PlusOutlined } from '@ant-design/icons';
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
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingConfig, setEditingConfig] = useState<ETFConfig | null>(null);
  const [form] = Form.useForm();

  // 加载ETF配置列表
  const loadConfigs = async () => {
    setLoading(true);
    try {
      const response = await etfConfigAPI.getAll();
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

  const handleEdit = (record: ETFConfig) => {
    setEditingConfig(record);
    form.setFieldsValue({
      symbol: record.symbol,
      name: record.name,
      description: record.description,
      strategy: record.strategy,
      focus: record.focus,
      expense_ratio: record.expense_ratio,
      currency: record.currency,
      exchange: record.exchange,
      category: record.category,
      provider: record.provider,
    });
    setIsModalVisible(true);
  };

  const handleAdd = () => {
    setEditingConfig(null);
    form.resetFields();
    setIsModalVisible(true);
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();

      if (editingConfig) {
        const result = await etfConfigAPI.update(String(editingConfig.id), {
          name: values.name,
          description: values.description,
          strategy: values.strategy,
          focus: values.focus,
          expense_ratio: values.expense_ratio,
          currency: values.currency,
          exchange: values.exchange,
          category: values.category,
          provider: values.provider,
        });
        if (result.success) {
          message.success('ETF配置已更新');
          setIsModalVisible(false);
          form.resetFields();
          loadConfigs();
        } else {
          message.error(result.error || '更新ETF配置失败');
        }
      } else {
        const result = await etfConfigAPI.create({
          symbol: values.symbol,
          name: values.name,
          description: values.description,
          strategy: values.strategy,
          focus: values.focus,
          expense_ratio: values.expense_ratio,
          currency: values.currency || 'USD',
          exchange: values.exchange,
          category: values.category,
          provider: values.provider,
          status: 1,
        });
        if (result.success) {
          message.success('ETF配置已添加');
          setIsModalVisible(false);
          form.resetFields();
          loadConfigs();
        } else {
          message.error(result.error || '添加ETF配置失败');
        }
      }
    } catch (error) {
      console.error('Validation failed:', error);
    }
  };

  // 切换状态
  const handleToggleActive = async (id: number, checked: boolean) => {
    // 先更新本地状态
    setConfigs(prev =>
      prev.map(config =>
        config.id === id ? { ...config, is_active: checked, status: checked ? 1 : 0 } : config
      )
    );

    try {
      const response = await etfConfigAPI.toggleStatus(String(id));
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
      const response = await etfConfigAPI.toggleAutoUpdate(String(id));
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
      render: (_text: unknown, record: ETFConfig) => (
        <Space>
          <Button size="small" icon={<ReloadOutlined />}>更新</Button>
          <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>编辑</Button>
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
        <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>添加ETF</Button>
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

      <Modal
        title={editingConfig ? '编辑ETF配置' : '添加ETF'}
        open={isModalVisible}
        onOk={handleModalOk}
        onCancel={() => setIsModalVisible(false)}
        width={600}
        okText={editingConfig ? '保存' : '添加'}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="symbol"
            label="ETF代码"
            rules={[{ required: true, message: '请输入ETF代码' }]}
          >
            <Input placeholder="例如：SPY" disabled={!!editingConfig} />
          </Form.Item>
          <Form.Item
            name="name"
            label="ETF名称"
            rules={[{ required: true, message: '请输入ETF名称' }]}
          >
            <Input placeholder="例如：SPDR S&P 500 ETF" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea placeholder="ETF描述信息" rows={2} />
          </Form.Item>
          <Form.Item name="focus" label="投资方向">
            <Input placeholder="例如：大盘股" />
          </Form.Item>
          <Form.Item name="strategy" label="投资策略">
            <Input placeholder="例如：被动指数跟踪" />
          </Form.Item>
          <Form.Item name="category" label="分类">
            <Select
              placeholder="选择分类"
              options={[
                { value: 'equity', label: '股票型' },
                { value: 'bond', label: '债券型' },
                { value: 'commodity', label: '商品型' },
                { value: 'reit', label: 'REITs' },
                { value: 'sector', label: '行业型' },
                { value: 'international', label: '国际型' },
                { value: 'mixed', label: '混合型' },
              ]}
            />
          </Form.Item>
          <Form.Item name="provider" label="基金公司">
            <Input placeholder="例如：Vanguard" />
          </Form.Item>
          <Form.Item name="exchange" label="交易所">
            <Input placeholder="例如：NYSE" />
          </Form.Item>
          <Form.Item name="currency" label="币种">
            <Select
              placeholder="选择币种"
              options={[
                { value: 'USD', label: 'USD' },
                { value: 'CNY', label: 'CNY' },
                { value: 'HKD', label: 'HKD' },
                { value: 'EUR', label: 'EUR' },
              ]}
            />
          </Form.Item>
          <Form.Item name="expense_ratio" label="费率 (%)">
            <InputNumber
              style={{ width: '100%' }}
              min={0}
              max={5}
              step={0.01}
              placeholder="例如：0.03"
            />
          </Form.Item>
        </Form>
      </Modal>
    </Layout>
  );
};

export default ETFConfigPage;

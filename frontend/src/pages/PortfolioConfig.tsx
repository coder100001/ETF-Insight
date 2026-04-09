import { useState, useEffect } from 'react';
import styled from 'styled-components';
import { Card, Table, Button, Badge, Space, Modal, Form, Input, InputNumber, Select, message, Popconfirm } from 'antd';
import { PieChartOutlined, PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import { etfAPI } from '../services/api';

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

interface PortfolioConfig {
  id: number;
  name: string;
  description: string;
  etfs: string[];
  allocation: Record<string, number>;
  total_investment: number;
  created_at: string;
  updated_at: string;
  is_default: boolean;
}

const mockConfigs: PortfolioConfig[] = [
  {
    id: 1,
    name: '稳健型组合',
    description: '适合保守投资者，注重稳定收益',
    etfs: ['SCHD', 'VYM', 'JEPI'],
    allocation: { SCHD: 40, VYM: 40, JEPI: 20 },
    total_investment: 100000,
    created_at: '2024-03-01',
    updated_at: '2024-03-28',
    is_default: true,
  },
  {
    id: 2,
    name: '进取型组合',
    description: '适合激进投资者，追求高收益',
    etfs: ['JEPQ', 'SPYD', 'JEPI'],
    allocation: { JEPQ: 40, SPYD: 30, JEPI: 30 },
    total_investment: 50000,
    created_at: '2024-03-15',
    updated_at: '2024-03-28',
    is_default: false,
  },
  {
    id: 3,
    name: '平衡型组合',
    description: '平衡风险与收益',
    etfs: ['SCHD', 'SPYD', 'JEPQ', 'JEPI', 'VYM'],
    allocation: { SCHD: 30, SPYD: 20, JEPQ: 20, JEPI: 15, VYM: 15 },
    total_investment: 200000,
    created_at: '2024-03-20',
    updated_at: '2024-03-28',
    is_default: false,
  },
];

interface ETFOption {
  value: string;
  label: string;
}

const PortfolioConfigPage: React.FC = () => {
  const [configs, setConfigs] = useState<PortfolioConfig[]>([]);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingConfig, setEditingConfig] = useState<PortfolioConfig | null>(null);
  const [form] = Form.useForm();
  const [etfOptions, setEtfOptions] = useState<ETFOption[]>([]);
  const [loading, setLoading] = useState(false);

  // 加载 ETF 选项
  const loadEtfOptions = async () => {
    setLoading(true);
    try {
      const response = await etfAPI.getList();
      if (response.success && response.data) {
        const options = response.data.map((etf: any) => ({
          value: etf.symbol,
          label: `${etf.symbol} - ${etf.name}`,
        }));
        setEtfOptions(options);
      }
    } catch (error) {
      console.error('Failed to load ETF options:', error);
    } finally {
      setLoading(false);
    }
  };

  // 加载组合配置 - 从API获取
  const loadConfigs = async () => {
    try {
      const response = await fetch('/api/portfolio-configs/');
      if (response.ok) {
        const result = await response.json();
        if (result.success && result.data) {
          setConfigs(result.data.map((c: any) => ({
            id: c.id,
            name: c.name,
            description: c.description || '',
            etfs: c.allocation ? Object.keys(JSON.parse(c.allocation)) : [],
            allocation: c.allocation ? JSON.parse(c.allocation) : {},
            total_investment: c.total_investment || 0,
            created_at: c.created_at?.split('T')[0] || '',
            updated_at: c.updated_at?.split('T')[0] || '',
            is_default: c.is_default || false,
          })));
        }
      }
    } catch (error) {
      console.error('Failed to load portfolio configs:', error);
    }
  };

  const columns = [
    {
      title: '配置名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: PortfolioConfig) => (
        <div>
          <strong>{text}</strong>
          {record.is_default && (
            <Badge
              count="默认"
              style={{ backgroundColor: theme.colors.primary, marginLeft: 8 }}
            />
          )}
          <br />
          <small style={{ color: theme.colors.textMuted }}>{record.description}</small>
        </div>
      ),
    },
    {
      title: '包含ETF',
      dataIndex: 'etfs',
      key: 'etfs',
      render: (etfs: string[]) => (
        <Space>
          {etfs.map(etf => (
            <Badge key={etf} count={etf} style={{ backgroundColor: theme.colors.info }} />
          ))}
        </Space>
      ),
    },
    {
      title: '投资金额',
      dataIndex: 'total_investment',
      key: 'total_investment',
      align: 'right' as const,
      render: (value: number) => `$${value.toLocaleString()}`,
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
      render: (_: any, record: PortfolioConfig) => (
        <Space>
          <Button 
            size="small" 
            icon={<EditOutlined />} 
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个配置吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const handleCreate = () => {
    setEditingConfig(null);
    form.resetFields();
    setIsModalVisible(true);
  };

  const handleEdit = (config: PortfolioConfig) => {
    setEditingConfig(config);
    form.setFieldsValue({
      name: config.name,
      description: config.description,
      etfs: config.etfs,
      allocation: config.allocation,
      total_investment: config.total_investment,
    });
    setIsModalVisible(true);
  };

  const handleDelete = (id: number) => {
    setConfigs(prev => prev.filter(config => config.id !== id));
    message.success('配置已删除');
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      
      // 计算 ETF 权重
      const allocation: Record<string, number> = {};
      if (values.etfs && values.etfs.length > 0) {
        const weight = 100 / values.etfs.length;
        values.etfs.forEach(etf => {
          allocation[etf] = weight;
        });
      }

      const configData = {
        ...values,
        allocation,
        etfs: values.etfs || [],
      };

      if (editingConfig) {
        // 编辑模式
        setConfigs(prev => prev.map(config => 
          config.id === editingConfig.id 
            ? { ...config, ...configData, updated_at: new Date().toISOString().split('T')[0] }
            : config
        ));
        message.success('配置已更新');
      } else {
        // 新建模式
        const newConfig: PortfolioConfig = {
          id: Date.now(),
          ...configData,
          created_at: new Date().toISOString().split('T')[0],
          updated_at: new Date().toISOString().split('T')[0],
          is_default: false,
        };
        setConfigs(prev => [...prev, newConfig]);
        message.success('配置已创建');
      }

      setIsModalVisible(false);
      form.resetFields();
    } catch (error) {
      console.error('Validation failed:', error);
    }
  };
  
  useEffect(() => {
    loadEtfOptions();
    loadConfigs();
  }, []);

  return (
    <Layout>
      <PageHeader>
        <h2>
          <PieChartOutlined />
          组合配置
        </h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          新建配置
        </Button>
      </PageHeader>

      <Card style={{ boxShadow: theme.shadows.card }}>
        <StyledTable
          dataSource={configs}
          columns={columns as any}
          rowKey="id"
          pagination={false}
        />
      </Card>

      <Modal
        title={editingConfig ? '编辑组合配置' : '新建组合配置'}
        open={isModalVisible}
        onOk={handleModalOk}
        onCancel={() => setIsModalVisible(false)}
        width={700}
        okText={editingConfig ? '保存' : '创建'}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="name"
            label="配置名称"
            rules={[{ required: true, message: '请输入配置名称' }]}
          >
            <Input placeholder="例如：稳健型组合" />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
          >
            <Input.TextArea placeholder="描述这个组合的特点" rows={3} />
          </Form.Item>
          <Form.Item
            name="etfs"
            label="选择ETF"
            rules={[{ required: true, message: '请至少选择一个ETF' }]}
          >
            <Select
              mode="multiple"
              placeholder="选择组合中的ETF"
              options={etfOptions}
              loading={loading}
              style={{ width: '100%' }}
            />
          </Form.Item>
          <Form.Item
            name="total_investment"
            label="投资金额"
            rules={[{ required: true, message: '请输入投资金额' }]}
          >
            <InputNumber
              style={{ width: '100%' }}
              prefix="$"
              min={1000}
              step={1000}
              formatter={value => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
              parser={(value: string | undefined) => Number(value?.replace(/\$\s?|(,*)/g, '') || 0)}
            />
          </Form.Item>
        </Form>
      </Modal>
    </Layout>
  );
};

export default PortfolioConfigPage;

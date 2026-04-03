import { useState } from 'react';
import styled from 'styled-components';
import { Card, Table, Button, Badge, Space, Modal, Form, Input, InputNumber } from 'antd';
import { PieChartOutlined, PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
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

interface PortfolioConfig {
  id: number;
  name: string;
  description: string;
  etfs: string[];
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
    total_investment: 200000,
    created_at: '2024-03-20',
    updated_at: '2024-03-28',
    is_default: false,
  },
];

const PortfolioConfigPage: React.FC = () => {
  const [configs] = useState<PortfolioConfig[]>(mockConfigs);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [form] = Form.useForm();

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
      render: () => (
        <Space>
          <Button size="small" icon={<EditOutlined />}>编辑</Button>
          <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
        </Space>
      ),
    },
  ];

  const handleCreate = () => {
    setIsModalVisible(true);
  };

  const handleModalOk = () => {
    form.validateFields().then(values => {
      console.log('Form values:', values);
      setIsModalVisible(false);
      form.resetFields();
    });
  };

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
        title="新建组合配置"
        open={isModalVisible}
        onOk={handleModalOk}
        onCancel={() => setIsModalVisible(false)}
        width={600}
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

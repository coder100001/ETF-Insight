import { useState } from 'react';
import styled from 'styled-components';
import { Card, Table, DatePicker, Select, Button, Tag } from 'antd';
import { FileTextOutlined } from '@ant-design/icons';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';

const { RangePicker } = DatePicker;

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
  flex-wrap: wrap;
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

interface OperationLog {
  id: number;
  user: string;
  action: string;
  target: string;
  details: string;
  ip_address: string;
  created_at: string;
}

const mockLogs: OperationLog[] = [
  {
    id: 1,
    user: 'admin',
    action: '更新ETF数据',
    target: 'SCHD',
    details: '手动触发ETF数据更新',
    ip_address: '192.168.1.100',
    created_at: '2024-03-28 10:30:00',
  },
  {
    id: 2,
    user: 'admin',
    action: '修改配置',
    target: '投资组合配置',
    details: '修改了稳健型组合的配比',
    ip_address: '192.168.1.100',
    created_at: '2024-03-28 09:15:00',
  },
  {
    id: 3,
    user: 'user1',
    action: '查看报表',
    target: 'ETF对比分析',
    details: '导出了ETF对比报表',
    ip_address: '192.168.1.101',
    created_at: '2024-03-27 16:45:00',
  },
  {
    id: 4,
    user: 'admin',
    action: '更新汇率',
    target: 'USD/CNY',
    details: '手动更新汇率数据',
    ip_address: '192.168.1.100',
    created_at: '2024-03-27 14:20:00',
  },
  {
    id: 5,
    user: 'user2',
    action: '登录系统',
    target: '用户认证',
    details: '成功登录系统',
    ip_address: '192.168.1.102',
    created_at: '2024-03-27 09:00:00',
  },
];

const OperationLogs: React.FC = () => {
  const [logs] = useState<OperationLog[]>(mockLogs);

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '用户',
      dataIndex: 'user',
      key: 'user',
      render: (user: string) => <Tag color="blue">{user}</Tag>,
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
    },
    {
      title: '对象',
      dataIndex: 'target',
      key: 'target',
    },
    {
      title: '详情',
      dataIndex: 'details',
      key: 'details',
    },
    {
      title: 'IP地址',
      dataIndex: 'ip_address',
      key: 'ip_address',
      align: 'center' as const,
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      align: 'center' as const,
    },
  ];

  return (
    <Layout>
      <PageHeader>
        <h2>
          <FileTextOutlined />
          操作记录
        </h2>
      </PageHeader>

      <FilterSection>
        <RangePicker placeholder={['开始日期', '结束日期']} />
        <Select placeholder="选择用户" style={{ width: 150 }} allowClear>
          <Select.Option value="admin">admin</Select.Option>
          <Select.Option value="user1">user1</Select.Option>
          <Select.Option value="user2">user2</Select.Option>
        </Select>
        <Select placeholder="操作类型" style={{ width: 150 }} allowClear>
          <Select.Option value="update">更新数据</Select.Option>
          <Select.Option value="create">创建</Select.Option>
          <Select.Option value="delete">删除</Select.Option>
          <Select.Option value="view">查看</Select.Option>
        </Select>
        <Button type="primary">筛选</Button>
        <Button>导出</Button>
      </FilterSection>

      <Card style={{ boxShadow: theme.shadows.card }}>
        <StyledTable
          dataSource={logs}
          columns={columns}
          rowKey="id"
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>
    </Layout>
  );
};

export default OperationLogs;

import { useState } from 'react';
import styled from 'styled-components';
import { Card, Table, Button, Badge, Tag, Space } from 'antd';
import { ProjectOutlined, PlayCircleOutlined, PauseCircleOutlined, EditOutlined } from '@ant-design/icons';
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

interface Workflow {
  id: number;
  name: string;
  description: string;
  status: 'active' | 'paused' | 'error';
  last_run: string;
  next_run: string;
  schedule: string;
}

const mockWorkflows: Workflow[] = [
  {
    id: 1,
    name: 'ETF数据更新',
    description: '每日更新ETF价格、股息等数据',
    status: 'active',
    last_run: '2024-03-28 09:00:00',
    next_run: '2024-03-29 09:00:00',
    schedule: '每天 09:00',
  },
  {
    id: 2,
    name: '汇率数据获取',
    description: '获取最新汇率数据',
    status: 'active',
    last_run: '2024-03-28 08:00:00',
    next_run: '2024-03-29 08:00:00',
    schedule: '每天 08:00',
  },
  {
    id: 3,
    name: '投资组合分析',
    description: '分析投资组合表现',
    status: 'paused',
    last_run: '2024-03-27 18:00:00',
    next_run: '-',
    schedule: '每周一 18:00',
  },
  {
    id: 4,
    name: '数据备份',
    description: '备份数据库和配置文件',
    status: 'active',
    last_run: '2024-03-28 00:00:00',
    next_run: '2024-03-29 00:00:00',
    schedule: '每天 00:00',
  },
];

const WorkflowList: React.FC = () => {
  const [workflows] = useState<Workflow[]>(mockWorkflows);

  const columns = [
    {
      title: '工作流名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Workflow) => (
        <div>
          <strong>{text}</strong>
          <br />
          <small style={{ color: theme.colors.textMuted }}>{record.description}</small>
        </div>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      align: 'center' as const,
      render: (status: string) => {
        const statusMap: { [key: string]: { text: string; color: 'success' | 'warning' | 'error' | 'default' | 'processing' } } = {
          active: { text: '运行中', color: 'success' },
          paused: { text: '已暂停', color: 'warning' },
          error: { text: '错误', color: 'error' },
        };
        const { text, color } = statusMap[status] || statusMap.active;
        return <Badge status={color} text={text} />;
      },
    },
    {
      title: '调度',
      dataIndex: 'schedule',
      key: 'schedule',
      align: 'center' as const,
      render: (schedule: string) => <Tag>{schedule}</Tag>,
    },
    {
      title: '上次执行',
      dataIndex: 'last_run',
      key: 'last_run',
      align: 'center' as const,
    },
    {
      title: '下次执行',
      dataIndex: 'next_run',
      key: 'next_run',
      align: 'center' as const,
    },
    {
      title: '操作',
      key: 'action',
      align: 'center' as const,
      render: (_: unknown, record: Workflow) => (
        <Space>
          {record.status === 'active' ? (
            <Button size="small" icon={<PauseCircleOutlined />}>暂停</Button>
          ) : (
            <Button size="small" type="primary" icon={<PlayCircleOutlined />}>启动</Button>
          )}
          <Button size="small" icon={<EditOutlined />}>编辑</Button>
        </Space>
      ),
    },
  ];

  return (
    <Layout>
      <PageHeader>
        <h2>
          <ProjectOutlined />
          工作流管理
        </h2>
        <Button type="primary">创建工作流</Button>
      </PageHeader>

      <Card style={{ boxShadow: theme.shadows.card }}>
        <StyledTable
          dataSource={workflows}
          columns={columns}
          rowKey="id"
          pagination={false}
        />
      </Card>
    </Layout>
  );
};

export default WorkflowList;

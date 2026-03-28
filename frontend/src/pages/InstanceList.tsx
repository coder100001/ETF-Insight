import { useState } from 'react';
import styled from 'styled-components';
import { Card, Table, Button, Badge, Tag, Space, Select } from 'antd';
import { HistoryOutlined, ReloadOutlined, EyeOutlined } from '@ant-design/icons';
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

interface Instance {
  id: number;
  workflow_name: string;
  status: 'success' | 'failed' | 'running';
  start_time: string;
  end_time: string;
  duration: string;
  trigger: string;
}

const mockInstances: Instance[] = [
  {
    id: 1,
    workflow_name: 'ETF数据更新',
    status: 'success',
    start_time: '2024-03-28 09:00:00',
    end_time: '2024-03-28 09:02:35',
    duration: '2分35秒',
    trigger: '定时触发',
  },
  {
    id: 2,
    workflow_name: '汇率数据获取',
    status: 'success',
    start_time: '2024-03-28 08:00:00',
    end_time: '2024-03-28 08:00:45',
    duration: '45秒',
    trigger: '定时触发',
  },
  {
    id: 3,
    workflow_name: '投资组合分析',
    status: 'failed',
    start_time: '2024-03-27 18:00:00',
    end_time: '2024-03-27 18:05:12',
    duration: '5分12秒',
    trigger: '手动触发',
  },
  {
    id: 4,
    workflow_name: '数据备份',
    status: 'running',
    start_time: '2024-03-28 00:00:00',
    end_time: '-',
    duration: '-',
    trigger: '定时触发',
  },
  {
    id: 5,
    workflow_name: 'ETF数据更新',
    status: 'success',
    start_time: '2024-03-27 09:00:00',
    end_time: '2024-03-27 09:01:58',
    duration: '1分58秒',
    trigger: '定时触发',
  },
];

const InstanceList: React.FC = () => {
  const [instances] = useState<Instance[]>(mockInstances);

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '工作流',
      dataIndex: 'workflow_name',
      key: 'workflow_name',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      align: 'center' as const,
      render: (status: string) => {
        const statusMap: { [key: string]: { text: string; color: 'success' | 'error' | 'processing' | 'default' | 'warning' } } = {
          success: { text: '成功', color: 'success' },
          failed: { text: '失败', color: 'error' },
          running: { text: '运行中', color: 'processing' },
        };
        const { text, color } = statusMap[status] || statusMap.success;
        return <Badge status={color} text={text} />;
      },
    },
    {
      title: '触发方式',
      dataIndex: 'trigger',
      key: 'trigger',
      align: 'center' as const,
      render: (trigger: string) => <Tag>{trigger}</Tag>,
    },
    {
      title: '开始时间',
      dataIndex: 'start_time',
      key: 'start_time',
      align: 'center' as const,
    },
    {
      title: '结束时间',
      dataIndex: 'end_time',
      key: 'end_time',
      align: 'center' as const,
    },
    {
      title: '耗时',
      dataIndex: 'duration',
      key: 'duration',
      align: 'center' as const,
    },
    {
      title: '操作',
      key: 'action',
      align: 'center' as const,
      render: () => (
        <Space>
          <Button size="small" icon={<EyeOutlined />}>查看</Button>
          <Button size="small" icon={<ReloadOutlined />}>重试</Button>
        </Space>
      ),
    },
  ];

  return (
    <Layout>
      <PageHeader>
        <h2>
          <HistoryOutlined />
          执行记录
        </h2>
      </PageHeader>

      <FilterSection>
        <Select placeholder="选择工作流" style={{ width: 200 }} allowClear>
          <Select.Option value="etf">ETF数据更新</Select.Option>
          <Select.Option value="exchange">汇率数据获取</Select.Option>
          <Select.Option value="portfolio">投资组合分析</Select.Option>
          <Select.Option value="backup">数据备份</Select.Option>
        </Select>
        <Select placeholder="选择状态" style={{ width: 150 }} allowClear>
          <Select.Option value="success">成功</Select.Option>
          <Select.Option value="failed">失败</Select.Option>
          <Select.Option value="running">运行中</Select.Option>
        </Select>
        <Button type="primary">筛选</Button>
      </FilterSection>

      <Card style={{ boxShadow: theme.shadows.card }}>
        <StyledTable
          dataSource={instances}
          columns={columns as any}
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

export default InstanceList;

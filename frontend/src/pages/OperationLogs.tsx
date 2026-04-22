import { useState, useEffect, useCallback } from 'react';
import styled from 'styled-components';
import {
  Card,
  Table,
  DatePicker,
  Select,
  Button,
  Tag,
  Input,
  Space,
  message,
  Modal,
  Spin,
  Pagination,
} from 'antd';
import {
  FileTextOutlined,
  SearchOutlined,
  ExportOutlined,
  EyeOutlined,
  FilterOutlined,
} from '@ant-design/icons';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import LogStatusTag from '../components/LogStatusTag';
import {
  operationLogsAPI,
  type LogFilterParams,
  type UnifiedLog,
  type LogTypesResponse,
  type ActionTypesResponse,
  type UsersResponse,
} from '../services/operationLogsService';
import debounce from 'lodash/debounce';

const { RangePicker } = DatePicker;
const { Option } = Select;

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

const FilterCard = styled(Card)`
  margin-bottom: 20px;
  .ant-card-body {
    padding: 16px;
  }
`;

const FilterSection = styled.div`
  display: flex;
  flex-direction: column;
  gap: 16px;
`;

const FilterRow = styled.div`
  display: flex;
  gap: 16px;
  align-items: flex-start;
  flex-wrap: wrap;
`;

const FilterItem = styled.div`
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 200px;

  label {
    font-size: ${theme.fonts.size.sm};
    color: ${theme.colors.textSecondary};
    font-weight: ${theme.fonts.weight.medium};
  }
`;

const ActionsRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8px;
`;

const StyledTable = styled(Table)`
  .ant-table-thead > tr > th {
    background: ${theme.colors.background};
    font-weight: ${theme.fonts.weight.semibold};
    color: ${theme.colors.textPrimary};
  }

  .ant-table-tbody > tr:hover > td {
    background: #f8f9fa;
  }

  .ant-table-row {
    cursor: pointer;
  }
` as typeof Table;

const DetailModal = styled(Modal)`
  .ant-modal-body {
    max-height: 60vh;
    overflow-y: auto;
  }
`;

const DetailSection = styled.div`
  margin-bottom: 16px;

  h3 {
    font-size: ${theme.fonts.size.md};
    color: ${theme.colors.textPrimary};
    margin-bottom: 8px;
    font-weight: ${theme.fonts.weight.semibold};
  }

  p {
    margin: 0;
    font-size: ${theme.fonts.size.sm};
    color: ${theme.colors.textSecondary};
    line-height: 1.5;
  }

  pre {
    background: #f5f5f5;
    padding: 12px;
    border-radius: ${theme.borderRadius.sm};
    overflow-x: auto;
    font-size: ${theme.fonts.size.sm};
    margin: 0;
    white-space: pre-wrap;
    word-wrap: break-word;
  }
`;

const OperationLogs: React.FC = () => {
  // 状态管理
  const [loading, setLoading] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [logs, setLogs] = useState<UnifiedLog[]>([]);
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 20,
    total: 0,
  });
  const [filterParams, setFilterParams] = useState<LogFilterParams>({
    page: 1,
    pageSize: 20,
  });
  const [selectedLog, setSelectedLog] = useState<UnifiedLog | null>(null);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [logTypes, setLogTypes] = useState<LogTypesResponse['types']>([]);
  const [actionTypes, setActionTypes] = useState<ActionTypesResponse['types']>([]);
  const [users, setUsers] = useState<UsersResponse['users']>([]);

  // 加载元数据（日志类型、操作类型、用户列表）
  useEffect(() => {
    const loadMetadata = async () => {
      try {
        const [logTypesRes, actionTypesRes, usersRes] = await Promise.all([
          operationLogsAPI.getLogTypes(),
          operationLogsAPI.getActionTypes(),
          operationLogsAPI.getUsers(),
        ]);
        setLogTypes(logTypesRes.types);
        setActionTypes(actionTypesRes.types);
        setUsers(usersRes.users);
      } catch (error) {
        console.error('加载元数据失败:', error);
      }
    };
    loadMetadata();
  }, []);

  // 加载日志数据
  const loadLogs = useCallback(async () => {
    setLoading(true);
    try {
      const response = await operationLogsAPI.getLogs(filterParams);
      setLogs(response.data);
      setPagination({
        page: filterParams.page,
        pageSize: filterParams.pageSize,
        total: response.total,
      });
    } catch (error) {
      console.error('加载日志失败:', error);
      message.error('加载日志失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  }, [filterParams]);

  // 初始化加载和筛选条件变化时重新加载
  useEffect(() => {
    loadLogs();
  }, [loadLogs]);

  // 防抖搜索
  const debouncedSearch = useCallback(
    debounce((newParams: LogFilterParams) => {
      setFilterParams(newParams);
    }, 500),
    []
  );

  // 筛选条件变更处理
  const handleFilterChange = (key: keyof LogFilterParams, value: any) => {
    const newParams = { ...filterParams, [key]: value, page: 1 }; // 回到第一页
    debouncedSearch(newParams);
  };

  // 分页处理
  const handlePageChange = (page: number, pageSize?: number) => {
    const newParams = {
      ...filterParams,
      page,
      pageSize: pageSize || filterParams.pageSize,
    };
    setFilterParams(newParams);
  };

  // 导出日志
  const handleExport = async () => {
    setExporting(true);
    try {
      await operationLogsAPI.exportLogs(filterParams);
      message.success('导出成功，文件正在下载');
    } catch (error) {
      console.error('导出失败:', error);
      message.error('导出失败，请稍后重试');
    } finally {
      setExporting(false);
    }
  };

  // 查看详情
  const handleViewDetail = async (log: UnifiedLog) => {
    try {
      const detail = await operationLogsAPI.getLogDetail(log.id);
      setSelectedLog(detail);
      setDetailModalVisible(true);
    } catch (error) {
      console.error('获取详情失败:', error);
      message.error('获取日志详情失败');
    }
  };

  // 表格列定义
  const columns = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      align: 'center' as const,
      render: (timestamp: string) => {
        const date = new Date(timestamp);
        return date.toLocaleString('zh-CN');
      },
      sorter: true,
    },
    {
      title: '用户',
      dataIndex: 'user',
      key: 'user',
      width: 120,
      render: (user: string) => (
        <Tag color="blue" style={{ margin: 0 }}>
          {user || '系统'}
        </Tag>
      ),
    },
    {
      title: '日志类型',
      dataIndex: 'log_type',
      key: 'log_type',
      width: 100,
      render: (type: string) => (
        <Tag color={type === 'audit' ? 'purple' : 'cyan'} style={{ margin: 0 }}>
          {type === 'audit' ? 'API日志' : '操作日志'}
        </Tag>
      ),
      filters: logTypes.map((type) => ({ text: type.label, value: type.value })),
    },
    {
      title: '操作类型',
      dataIndex: 'action_type',
      key: 'action_type',
      width: 120,
      filters: actionTypes.map((type) => ({ text: type.label, value: type.value })),
    },
    {
      title: '模块',
      dataIndex: 'module',
      key: 'module',
      width: 150,
    },
    {
      title: '详情',
      dataIndex: 'details',
      key: 'details',
      ellipsis: true,
      render: (details: string) => (
        <div style={{ maxWidth: 300 }}>
          {details && details.length > 50 ? `${details.substring(0, 50)}...` : details}
        </div>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      align: 'center' as const,
      render: (status: 'success' | 'failure') => (
        <LogStatusTag status={status} />
      ),
      filters: [
        { text: '成功', value: 'success' },
        { text: '失败', value: 'failure' },
      ],
    },
    {
      title: 'IP地址',
      dataIndex: 'ip',
      key: 'ip',
      width: 120,
      align: 'center' as const,
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      align: 'center' as const,
      render: (_: any, record: UnifiedLog) => (
        <Button
          type="text"
          size="small"
          icon={<EyeOutlined />}
          onClick={() => handleViewDetail(record)}
        />
      ),
    },
  ];

  return (
    <Layout>
      <PageHeader>
        <h2>
          <FileTextOutlined />
          操作日志
        </h2>
      </PageHeader>

      <FilterCard>
        <FilterSection>
          <FilterRow>
            <FilterItem>
              <label>时间范围</label>
              <RangePicker
                showTime
                placeholder={['开始时间', '结束时间']}
                onChange={(dates) => {
                  if (dates) {
                    handleFilterChange('startTime', dates[0]?.toISOString());
                    handleFilterChange('endTime', dates[1]?.toISOString());
                  } else {
                    handleFilterChange('startTime', undefined);
                    handleFilterChange('endTime', undefined);
                  }
                }}
                style={{ width: '100%' }}
              />
            </FilterItem>

            <FilterItem>
              <label>操作用户</label>
              <Select
                placeholder="选择用户"
                allowClear
                showSearch
                optionFilterProp="children"
                onChange={(value) => handleFilterChange('user', value)}
                style={{ width: '100%' }}
              >
                {users.map((user) => (
                  <Option key={user} value={user}>
                    {user}
                  </Option>
                ))}
              </Select>
            </FilterItem>

            <FilterItem>
              <label>操作类型</label>
              <Select
                placeholder="选择操作类型"
                allowClear
                onChange={(value) => handleFilterChange('actionType', value)}
                style={{ width: '100%' }}
              >
                {actionTypes.map((type) => (
                  <Option key={type.value} value={type.value}>
                    {type.label}
                  </Option>
                ))}
              </Select>
            </FilterItem>

            <FilterItem>
              <label>日志类型</label>
              <Select
                placeholder="选择日志类型"
                allowClear
                onChange={(value) => handleFilterChange('logType', value)}
                style={{ width: '100%' }}
              >
                {logTypes.map((type) => (
                  <Option key={type.value} value={type.value}>
                    {type.label}
                  </Option>
                ))}
              </Select>
            </FilterItem>

            <FilterItem>
              <label>操作状态</label>
              <Select
                placeholder="选择状态"
                allowClear
                onChange={(value) => handleFilterChange('status', value)}
                style={{ width: '100%' }}
              >
                <Option value="success">成功</Option>
                <Option value="failure">失败</Option>
              </Select>
            </FilterItem>
          </FilterRow>

          <FilterRow>
            <FilterItem>
              <label>搜索详情</label>
              <Input
                placeholder="输入关键字搜索详情"
                allowClear
                prefix={<SearchOutlined />}
                onChange={(e) => handleFilterChange('search', e.target.value)}
                style={{ width: '100%' }}
              />
            </FilterItem>
          </FilterRow>

          <ActionsRow>
            <Space>
              <Button
                type="primary"
                icon={<FilterOutlined />}
                onClick={() => loadLogs()}
                loading={loading}
              >
                筛选
              </Button>
              <Button onClick={() => {
                setFilterParams({ page: 1, pageSize: 20 });
                message.success('筛选条件已重置');
              }}>
                重置
              </Button>
            </Space>
            <Button
              type="default"
              icon={<ExportOutlined />}
              onClick={handleExport}
              loading={exporting}
              disabled={logs.length === 0}
            >
              导出Excel
            </Button>
          </ActionsRow>
        </FilterSection>
      </FilterCard>

      <Card style={{ boxShadow: theme.shadows.card }}>
        <Spin spinning={loading}>
          <StyledTable
            dataSource={logs}
            columns={columns}
            rowKey="id"
            pagination={false}
            size="middle"
            scroll={{ x: 1200 }}
            onChange={(pagination, filters, sorter) => {
              // 处理表格排序和筛选变化
              console.log('表格变化:', { pagination, filters, sorter });
            }}
            onRow={(record) => ({
              onClick: () => handleViewDetail(record),
            })}
          />
          <div style={{ marginTop: 16, textAlign: 'right' }}>
            <Pagination
              current={pagination.page}
              pageSize={pagination.pageSize}
              total={pagination.total}
              showSizeChanger
              showQuickJumper
              showTotal={(total) => `共 ${total} 条记录`}
              onChange={handlePageChange}
              onShowSizeChange={handlePageChange}
            />
          </div>
        </Spin>
      </Card>

      {/* 详情弹窗 */}
      <DetailModal
        title="日志详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>,
        ]}
        width={800}
      >
        {selectedLog && (
          <>
            <DetailSection>
              <h3>基本信息</h3>
              <p>
                <strong>日志ID:</strong> {selectedLog.id}
              </p>
              <p>
                <strong>日志类型:</strong>{' '}
                <Tag color={selectedLog.log_type === 'audit' ? 'purple' : 'cyan'}>
                  {selectedLog.log_type === 'audit' ? 'API日志' : '操作日志'}
                </Tag>
              </p>
              <p>
                <strong>操作时间:</strong>{' '}
                {new Date(selectedLog.timestamp).toLocaleString('zh-CN')}
              </p>
              <p>
                <strong>操作用户:</strong>{' '}
                <Tag color="blue">{selectedLog.user || '系统'}</Tag>
              </p>
              <p>
                <strong>操作状态:</strong>{' '}
                <LogStatusTag status={selectedLog.status} />
              </p>
            </DetailSection>

            <DetailSection>
              <h3>操作详情</h3>
              <p>
                <strong>操作模块:</strong> {selectedLog.module}
              </p>
              <p>
                <strong>操作类型:</strong> {selectedLog.action_type}
              </p>
              <p>
                <strong>详情描述:</strong>
              </p>
              <pre>{selectedLog.details}</pre>
            </DetailSection>

            <DetailSection>
              <h3>网络信息</h3>
              <p>
                <strong>IP地址:</strong> {selectedLog.ip || 'N/A'}
              </p>
              {selectedLog.status_code && (
                <p>
                  <strong>状态码:</strong> {selectedLog.status_code}
                </p>
              )}
              {selectedLog.duration && (
                <p>
                  <strong>耗时:</strong> {selectedLog.duration}ms
                </p>
              )}
            </DetailSection>

            {selectedLog.error_message && (
              <DetailSection>
                <h3>错误信息</h3>
                <pre style={{ color: '#e74c3c' }}>{selectedLog.error_message}</pre>
              </DetailSection>
            )}
          </>
        )}
      </DetailModal>
    </Layout>
  );
};

export default OperationLogs;

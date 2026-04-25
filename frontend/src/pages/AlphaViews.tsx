import React, { useState, useEffect } from 'react';
import {
  Card, Table, Button, Tag, Modal, Form, Input, InputNumber, Select, message, Row, Col, Tabs
} from 'antd';
import { EyeOutlined, CloseCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import styled from 'styled-components';
import Layout from '../components/Layout';
import { alphaViewAPI } from '../services/api';
import type { AlphaView, ViewStatus, ViewMethod, ViewType } from '../types';

const { TabPane } = Tabs;
const { Option } = Select;

const Container = styled.div`
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
`;

const Title = styled.h1`
  font-size: 28px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 24px;
`;

const MetricCard = styled.div`
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 20px;
  border-radius: 12px;
  text-align: center;
  margin-bottom: 16px;
`;

const MetricValue = styled.div`
  font-size: 28px;
  font-weight: 700;
  margin-bottom: 8px;
`;

const MetricLabel = styled.div`
  font-size: 14px;
  opacity: 0.9;
`;

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82CA9D'];

const getStatusColor = (status: ViewStatus): string => {
  switch (status) {
    case 'active': return 'green';
    case 'expired': return 'red';
    case 'validated': return 'blue';
  }
};

const getStatusLabel = (status: ViewStatus): string => {
  switch (status) {
    case 'active': return '活跃';
    case 'expired': return '过期';
    case 'validated': return '已验证';
  }
};

const getMethodLabel = (method: ViewMethod): string => {
  switch (method) {
    case 'factor_timing': return '因子择时';
    case 'momentum': return '动量';
    case 'mean_reversion': return '均值回归';
  }
};

const getTypeLabel = (type: ViewType): string => {
  switch (type) {
    case 'absolute': return '绝对';
    case 'relative': return '相对';
  }
};

const AlphaViews: React.FC = () => {
  const [views, setViews] = useState<AlphaView[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedView, setSelectedView] = useState<AlphaView | null>(null);
  const [activeTab, setActiveTab] = useState('all');
  const [form] = Form.useForm();

  const fetchViews = async () => {
    setLoading(true);
    try {
      const result = await alphaViewAPI.getActive();
      if (result.success && result.data) {
        setViews(result.data);
      } else {
        message.error(result.error || '获取观点失败');
      }
    } catch (error) {
      message.error('获取观点失败: ' + (error instanceof Error ? error.message : '未知错误'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchViews();
  }, []);

  const handleCreate = async (values: Record<string, unknown>) => {
    try {
      const request = {
        portfolio_id: 1,
        asset_symbol: values.asset_symbol as string,
        view_return: (values.view_return as number) / 100,
        confidence: values.confidence as number,
        view_type: values.view_type as ViewType,
        view_method: values.view_method as ViewMethod,
        valid_until: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
        source_factor: values.source_factor as string || 'manual',
      };
      const result = await alphaViewAPI.create(request);
      if (result.success) {
        message.success('观点创建成功');
        setModalVisible(false);
        form.resetFields();
        fetchViews();
      } else {
        message.error(result.error || '创建失败');
      }
    } catch (error) {
      message.error('创建失败: ' + (error instanceof Error ? error.message : '未知错误'));
    }
  };

  const handleDeactivate = async (id: number) => {
    try {
      const result = await alphaViewAPI.deactivate(id);
      if (result.success) {
        message.success('观点已停用');
        fetchViews();
      } else {
        message.error(result.error || '停用失败');
      }
    } catch (error) {
      message.error('停用失败: ' + (error instanceof Error ? error.message : '未知错误'));
    }
  };

  const filteredViews = views.filter(v => {
    if (activeTab === 'all') return true;
    return v.status === activeTab;
  });

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 60 },
    { title: '资产', dataIndex: 'asset_symbol', key: 'asset_symbol' },
    { title: '类型', dataIndex: 'view_type', key: 'type', render: (v: ViewType) => getTypeLabel(v) },
    { title: '方法', dataIndex: 'view_method', key: 'method', render: (v: ViewMethod) => getMethodLabel(v) },
    { title: '观点收益', dataIndex: 'view_return', key: 'return', render: (v: number) => `${(v * 100).toFixed(2)}%` },
    { title: '置信度', dataIndex: 'confidence', key: 'confidence', render: (v: number) => `${v}%` },
    { title: '状态', dataIndex: 'status', key: 'status', render: (v: ViewStatus) => (
      <Tag color={getStatusColor(v)}>{getStatusLabel(v)}</Tag>
    )},
    { title: '来源因子', dataIndex: 'source_factor', key: 'source_factor' },
    { title: '生成时间', dataIndex: 'generated_at', key: 'generated_at', render: (v: string) => new Date(v).toLocaleDateString() },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: AlphaView) => (
        <>
          <Button type="link" icon={<EyeOutlined />} onClick={() => { setSelectedView(record); setDetailModalVisible(true); }}>
            详情
          </Button>
          {record.status === 'active' && (
            <Button type="link" danger icon={<CloseCircleOutlined />} onClick={() => handleDeactivate(record.id)}>
              停用
            </Button>
          )}
        </>
      ),
    },
  ];

  const assetDistribution = views.reduce((acc, v) => {
    acc[v.asset_symbol] = (acc[v.asset_symbol] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  const pieData = Object.entries(assetDistribution).map(([name, value]) => ({ name, value }));

  const methodDistribution = views.reduce((acc, v) => {
    const label = getMethodLabel(v.view_method);
    acc[label] = (acc[label] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  const barData = Object.entries(methodDistribution).map(([name, value]) => ({ name, value }));

  return (
    <Layout>
      <Container>
        <Title>Alpha观点管理</Title>

        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col span={6}>
            <MetricCard>
              <MetricValue>{views.length}</MetricValue>
              <MetricLabel>总观点数</MetricLabel>
            </MetricCard>
          </Col>
          <Col span={6}>
            <MetricCard style={{ background: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)' }}>
              <MetricValue>{views.filter(v => v.status === 'active').length}</MetricValue>
              <MetricLabel>活跃观点</MetricLabel>
            </MetricCard>
          </Col>
          <Col span={6}>
            <MetricCard style={{ background: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)' }}>
              <MetricValue>{views.filter(v => v.status === 'validated').length}</MetricValue>
              <MetricLabel>已验证</MetricLabel>
            </MetricCard>
          </Col>
          <Col span={6}>
            <MetricCard style={{ background: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)' }}>
              <MetricValue>{views.filter(v => v.status === 'expired').length}</MetricValue>
              <MetricLabel>已过期</MetricLabel>
            </MetricCard>
          </Col>
        </Row>

        <Card style={{ marginBottom: 24 }}>
          <Row gutter={16}>
            <Col span={12}>
              <ResponsiveContainer width="100%" height={250}>
                <PieChart>
                  <Pie
                    data={pieData}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    label={({ name, value }) => `${name}: ${value}`}
                    outerRadius={80}
                    fill="#8884d8"
                    dataKey="value"
                  >
                    {pieData.map((_, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
              <div style={{ textAlign: 'center', marginTop: 8 }}>资产分布</div>
            </Col>
            <Col span={12}>
              <ResponsiveContainer width="100%" height={250}>
                <BarChart data={barData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="name" />
                  <YAxis />
                  <Tooltip />
                  <Bar dataKey="value" fill="#8884d8" />
                </BarChart>
              </ResponsiveContainer>
              <div style={{ textAlign: 'center', marginTop: 8 }}>方法分布</div>
            </Col>
          </Row>
        </Card>

        <Card
          title="观点列表"
          extra={
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalVisible(true)}>
              创建新观点
            </Button>
          }
        >
          <Tabs activeKey={activeTab} onChange={setActiveTab}>
            <TabPane tab="全部" key="all" />
            <TabPane tab="活跃" key="active" />
            <TabPane tab="已验证" key="validated" />
            <TabPane tab="已过期" key="expired" />
          </Tabs>
          <Table
            dataSource={filteredViews}
            columns={columns}
            rowKey="id"
            loading={loading}
            pagination={{ pageSize: 10 }}
          />
        </Card>

        <Modal
          title="创建新观点"
          open={modalVisible}
          onCancel={() => setModalVisible(false)}
          footer={null}
        >
          <Form form={form} onFinish={handleCreate} layout="vertical">
            <Form.Item name="asset_symbol" label="资产代码" rules={[{ required: true }]}>
              <Input placeholder="如: SPY" />
            </Form.Item>
            <Form.Item name="view_type" label="观点类型" rules={[{ required: true }]} initialValue="absolute">
              <Select>
                <Option value="absolute">绝对</Option>
                <Option value="relative">相对</Option>
              </Select>
            </Form.Item>
            <Form.Item name="view_method" label="生成方法" rules={[{ required: true }]} initialValue="factor_timing">
              <Select>
                <Option value="factor_timing">因子择时</Option>
                <Option value="momentum">动量</Option>
                <Option value="mean_reversion">均值回归</Option>
              </Select>
            </Form.Item>
            <Form.Item name="view_return" label="观点收益 (%)" rules={[{ required: true }]}>
              <InputNumber style={{ width: '100%' }} placeholder="如: 2.5" />
            </Form.Item>
            <Form.Item name="confidence" label="置信度 (%)" rules={[{ required: true }]} initialValue={75}>
              <InputNumber min={0} max={100} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="source_factor" label="来源因子">
              <Input placeholder="如: Mkt-RF" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" block>
                创建
              </Button>
            </Form.Item>
          </Form>
        </Modal>

        <Modal
          title="观点详情"
          open={detailModalVisible}
          onCancel={() => setDetailModalVisible(false)}
          footer={null}
        >
          {selectedView && (
            <div>
              <p><strong>ID:</strong> {selectedView.id}</p>
              <p><strong>资产:</strong> {selectedView.asset_symbol}</p>
              <p><strong>类型:</strong> {getTypeLabel(selectedView.view_type)}</p>
              <p><strong>方法:</strong> {getMethodLabel(selectedView.view_method)}</p>
              <p><strong>观点收益:</strong> {(selectedView.view_return * 100).toFixed(2)}%</p>
              <p><strong>置信度:</strong> {selectedView.confidence}%</p>
              <p><strong>状态:</strong> <Tag color={getStatusColor(selectedView.status)}>{getStatusLabel(selectedView.status)}</Tag></p>
              <p><strong>来源因子:</strong> {selectedView.source_factor}</p>
              <p><strong>生成时间:</strong> {new Date(selectedView.generated_at).toLocaleString()}</p>
              <p><strong>有效期至:</strong> {new Date(selectedView.valid_until).toLocaleString()}</p>
              {selectedView.performance && (
                <>
                  <p><strong>实际收益:</strong> {(selectedView.performance.actual_return * 100).toFixed(2)}%</p>
                  <p><strong>预测误差:</strong> {(selectedView.performance.prediction_error * 100).toFixed(2)}%</p>
                  <p><strong>是否正确:</strong> {selectedView.performance.is_correct ? '是' : '否'}</p>
                  <p><strong>滚动胜率:</strong> {(selectedView.performance.rolling_win_rate * 100).toFixed(1)}%</p>
                </>
              )}
            </div>
          )}
        </Modal>
      </Container>
    </Layout>
  );
};

export default AlphaViews;

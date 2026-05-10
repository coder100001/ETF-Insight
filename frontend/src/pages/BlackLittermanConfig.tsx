import React, { useState } from 'react';
import {
  Card, Button, Row, Col, Select, InputNumber, message, Table, Tag, Form, Alert, Tabs, Statistic
} from 'antd';
import { CalculatorOutlined, SaveOutlined } from '@ant-design/icons';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip as RechartsTooltip, Legend, BarChart, Bar, XAxis, YAxis, CartesianGrid } from 'recharts';
import styled from 'styled-components';
import Layout from '../components/Layout';
import { blackLittermanAPI, alphaViewAPI } from '../services/api';
import type { BlackLittermanConfig, BLPosteriorReturn, AlphaView, PriorType, OmegaMethod } from '../types';

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

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82CA9D', '#FFC658', '#FF6B6B'];

const PRIOR_OPTIONS = [
  { value: 'equal_weight', label: '等权' },
  { value: 'min_variance', label: '最小方差' },
  { value: 'market_cap', label: '市值加权' },
];

const OMEGA_OPTIONS = [
  { value: 'Idzorek', label: 'Idzorek' },
  { value: 'HeLitterman', label: 'He & Litterman' },
];

const BlackLittermanConfigPage: React.FC = () => {
  const [configs, setConfigs] = useState<BlackLittermanConfig[]>([]);
  const [selectedConfig, setSelectedConfig] = useState<BlackLittermanConfig | null>(null);
  const [availableViews, setAvailableViews] = useState<AlphaView[]>([]);
  const [selectedViews, setSelectedViews] = useState<number[]>([]);
  const [posteriorResult, setPosteriorResult] = useState<BLPosteriorReturn | null>(null);
  const [loading, setLoading] = useState(false);
  const [calculateLoading, setCalculateLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('config');
  const [form] = Form.useForm();

  const fetchAvailableViews = async () => {
    try {
      const result = await alphaViewAPI.getActive();
      if (result.success && result.data) {
        setAvailableViews(result.data);
      }
    } catch (error) {
      console.error('获取观点失败:', error);
    }
  };

  const handleCreateConfig = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      const request = {
        portfolio_id: 1,
        risk_aversion: values.risk_aversion as number,
        prior_type: values.prior_type as PriorType,
        prior_weights: values.prior_weights as Record<string, number> || {},
        omega_method: values.omega_method as OmegaMethod,
      };
      const result = await blackLittermanAPI.createConfig(request);
      if (result.success && result.data) {
        setConfigs([...configs, result.data]);
        setSelectedConfig(result.data);
        message.success('配置创建成功');
        form.resetFields();
        fetchAvailableViews();
      } else {
        message.error(result.error || '创建失败');
      }
    } catch (error) {
      message.error('创建失败: ' + (error instanceof Error ? error.message : '未知错误'));
    } finally {
      setLoading(false);
    }
  };

  const handleCalculate = async () => {
    if (!selectedConfig) {
      message.warning('请先选择配置');
      return;
    }
    if (selectedViews.length === 0) {
      message.warning('请至少选择一个观点');
      return;
    }

    setCalculateLoading(true);
    try {
      const result = await blackLittermanAPI.calculate(selectedConfig.id, selectedViews);
      if (result.success && result.data) {
        setPosteriorResult(result.data);
        message.success('后验收益计算成功');
        setActiveTab('result');
      } else {
        message.error(result.error || '计算失败');
      }
    } catch (error) {
      message.error('计算失败: ' + (error instanceof Error ? error.message : '未知错误'));
    } finally {
      setCalculateLoading(false);
    }
  };

  const viewColumns = [
    { title: 'ID', dataIndex: 'id', key: 'id' },
    { title: '资产', dataIndex: 'asset_symbol', key: 'asset_symbol' },
    { title: '观点收益', dataIndex: 'view_return', key: 'return', render: (v: number) => `${(v * 100).toFixed(2)}%` },
    { title: '置信度', dataIndex: 'confidence', key: 'confidence', render: (v: number) => `${v}%` },
    { title: '来源', dataIndex: 'source_factor', key: 'source' },
  ];

  const posteriorWeights = posteriorResult?.posterior_weights
    ? JSON.parse(posteriorResult.posterior_weights)
    : {};

  const posteriorReturns = posteriorResult?.posterior_returns
    ? JSON.parse(posteriorResult.posterior_returns)
    : {};

  const weightData = Object.entries(posteriorWeights).map(([name, value]) => ({
    name,
    value: Number(value) * 100,
  }));

  const returnData = Object.entries(posteriorReturns).map(([name, value]) => ({
    name,
    value: Number(value) * 100,
  }));

  return (
    <Layout>
      <Container>
        <Title>Black-Litterman模型配置</Title>

        <Alert
          message="Black-Litterman模型说明"
          description="BL模型将市场均衡收益与投资者观点相结合，生成后验收益分布。选择先验类型、Omega方法和Alpha观点进行计算。"
          type="info"
          showIcon
          style={{ marginBottom: 24 }}
        />

        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane tab="配置管理" key="config">
            <Row gutter={24}>
              <Col span={12}>
                <Card title="创建新配置">
                  <Form form={form} onFinish={handleCreateConfig} layout="vertical">
                    <Form.Item name="risk_aversion" label="风险厌恶系数" rules={[{ required: true }]} initialValue={2.5}>
                      <InputNumber min={0.1} max={10} step={0.1} style={{ width: '100%' }} />
                    </Form.Item>
                    <Form.Item name="prior_type" label="先验类型" rules={[{ required: true }]} initialValue="market_cap">
                      <Select>
                        {PRIOR_OPTIONS.map(opt => (
                          <Option key={opt.value} value={opt.value}>{opt.label}</Option>
                        ))}
                      </Select>
                    </Form.Item>
                    <Form.Item name="omega_method" label="Omega方法" rules={[{ required: true }]} initialValue="Idzorek">
                      <Select>
                        {OMEGA_OPTIONS.map(opt => (
                          <Option key={opt.value} value={opt.value}>{opt.label}</Option>
                        ))}
                      </Select>
                    </Form.Item>
                    <Form.Item>
                      <Button type="primary" htmlType="submit" loading={loading} icon={<SaveOutlined />} block>
                        保存配置
                      </Button>
                    </Form.Item>
                  </Form>
                </Card>
              </Col>
              <Col span={12}>
                <Card title="现有配置">
                  <Table
                    dataSource={configs}
                    columns={[
                      { title: 'ID', dataIndex: 'id', key: 'id' },
                      { title: '风险厌恶', dataIndex: 'risk_aversion', key: 'risk_aversion' },
                      { title: '先验', dataIndex: 'prior_type', key: 'prior_type' },
                      { title: '状态', dataIndex: 'is_active', key: 'active', render: (v: boolean) => (
                        <Tag color={v ? 'green' : 'red'}>{v ? '活跃' : '停用'}</Tag>
                      )},
                    ]}
                    rowKey="id"
                    pagination={false}
                    onRow={(record) => ({
                      onClick: () => {
                        setSelectedConfig(record);
                        fetchAvailableViews();
                        message.info(`已选择配置 #${record.id}`);
                      },
                    })}
                  />
                </Card>
              </Col>
            </Row>

            {selectedConfig && (
              <Card title="选择Alpha观点" style={{ marginTop: 24 }}>
                <Alert
                  message={`当前配置: 风险厌恶=${selectedConfig.risk_aversion}, 先验=${selectedConfig.prior_type}`}
                  type="info"
                  style={{ marginBottom: 16 }}
                />
                <Table
                  dataSource={availableViews}
                  columns={viewColumns}
                  rowKey="id"
                  rowSelection={{
                    type: 'checkbox',
                    onChange: (selectedRowKeys) => {
                      setSelectedViews(selectedRowKeys as number[]);
                    },
                  }}
                  pagination={{ pageSize: 5 }}
                />
                <Button
                  type="primary"
                  onClick={handleCalculate}
                  loading={calculateLoading}
                  icon={<CalculatorOutlined />}
                  style={{ marginTop: 16 }}
                  block
                >
                  计算后验收益
                </Button>
              </Card>
            )}
          </TabPane>

          <TabPane tab="计算结果" key="result">
            {posteriorResult ? (
              <>
                <Row gutter={16} style={{ marginBottom: 24 }}>
                  <Col span={8}>
                    <Card>
                      <Statistic title="观点数量" value={posteriorResult.num_views} />
                    </Card>
                  </Col>
                  <Col span={8}>
                    <Card>
                      <Statistic title="观点影响度" value={posteriorResult.view_impact} precision={4} />
                    </Card>
                  </Col>
                  <Col span={8}>
                    <Card>
                      <Statistic title="计算日期" value={new Date(posteriorResult.calculation_date).toLocaleDateString()} />
                    </Card>
                  </Col>
                </Row>

                <Row gutter={24}>
                  <Col span={12}>
                    <Card title="后验权重配置">
                      <ResponsiveContainer width="100%" height={300}>
                        <PieChart>
                          <Pie
                            data={weightData}
                            cx="50%"
                            cy="50%"
                            labelLine={false}
                            label={({ name, value }) => `${name}: ${Number(value).toFixed(1)}%`}
                            outerRadius={100}
                            fill="#8884d8"
                            dataKey="value"
                          >
                            {weightData.map((_, index) => (
                              <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                            ))}
                          </Pie>
                          <RechartsTooltip />
                          <Legend />
                        </PieChart>
                      </ResponsiveContainer>
                    </Card>
                  </Col>
                  <Col span={12}>
                    <Card title="后验收益对比">
                      <ResponsiveContainer width="100%" height={300}>
                        <BarChart data={returnData}>
                          <CartesianGrid strokeDasharray="3 3" />
                          <XAxis dataKey="name" />
                          <YAxis />
                          <RechartsTooltip formatter={(value: unknown) => `${Number(value).toFixed(2)}%`} />
                          <Bar dataKey="value" fill="#8884d8" />
                        </BarChart>
                      </ResponsiveContainer>
                    </Card>
                  </Col>
                </Row>

                <Card title="权重明细" style={{ marginTop: 24 }}>
                  <Table
                    dataSource={weightData}
                    columns={[
                      { title: '资产', dataIndex: 'name', key: 'name' },
                      { title: '权重 (%)', dataIndex: 'value', key: 'value', render: (v: number) => `${v.toFixed(2)}%` },
                    ]}
                    rowKey="name"
                    pagination={false}
                  />
                </Card>
              </>
            ) : (
              <Alert title="请先进行计算" type="warning" />
            )}
          </TabPane>
        </Tabs>
      </Container>
    </Layout>
  );
};

export default BlackLittermanConfigPage;

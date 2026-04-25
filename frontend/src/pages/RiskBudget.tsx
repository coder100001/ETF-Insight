import React, { useState } from 'react';
import {
  Card, Button, Row, Col, Select, InputNumber, message, Table, Tag, Form, Alert, Tabs, Statistic
} from 'antd';
import { CalculatorOutlined, SaveOutlined, ExperimentOutlined } from '@ant-design/icons';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, ResponsiveContainer, Legend, PieChart, Pie, Cell } from 'recharts';
import styled from 'styled-components';
import Layout from '../components/Layout';
import { riskBudgetAPI } from '../services/api';
import type { RiskBudgetConfig, RiskContribution, MonteCarloSimulation, RiskMethod } from '../types';

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

const METHOD_OPTIONS = [
  { value: 'historical', label: '历史模拟法' },
  { value: 'parametric', label: '参数法' },
  { value: 'monte_carlo', label: '蒙特卡洛模拟' },
];

const RiskBudget: React.FC = () => {
  const [configs, setConfigs] = useState<RiskBudgetConfig[]>([]);
  const [selectedConfig, setSelectedConfig] = useState<RiskBudgetConfig | null>(null);
  const [cvarResult, setCvarResult] = useState<{ cvar: number; var: number; risk_contributions: RiskContribution[] } | null>(null);
  const [mcResult, setMcResult] = useState<MonteCarloSimulation | null>(null);
  const [loading, setLoading] = useState(false);
  const [calculateLoading, setCalculateLoading] = useState(false);
  const [mcLoading, setMcLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('config');
  const [weightsInput, setWeightsInput] = useState<string>('0.4, 0.3, 0.2, 0.1');
  const [form] = Form.useForm();

  const handleCreateConfig = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      const request = {
        portfolio_id: 1,
        cvar_limit: values.cvar_limit as number,
        confidence_level: values.confidence_level as number,
        time_horizon: values.time_horizon as number,
        method: values.method as RiskMethod,
        risk_budgets: values.risk_budgets as Record<string, number> || {},
      };
      const result = await riskBudgetAPI.createConfig(request);
      if (result.success && result.data) {
        setConfigs([...configs, result.data]);
        setSelectedConfig(result.data);
        message.success('配置创建成功');
        form.resetFields();
      } else {
        message.error(result.error || '创建失败');
      }
    } catch (error) {
      message.error('创建失败: ' + (error instanceof Error ? error.message : '未知错误'));
    } finally {
      setLoading(false);
    }
  };

  const handleCalculateCVaR = async () => {
    if (!selectedConfig) {
      message.warning('请先选择配置');
      return;
    }

    setCalculateLoading(true);
    try {
      const weights = weightsInput.split(',').map(w => parseFloat(w.trim())).filter(w => !isNaN(w));
      if (weights.length === 0) {
        message.error('请输入有效的权重');
        setCalculateLoading(false);
        return;
      }
      const sum = weights.reduce((a, b) => a + b, 0);
      if (Math.abs(sum - 1.0) > 0.01) {
        message.warning(`权重总和为 ${sum.toFixed(2)}，建议总和为 1.0`);
      }
      const result = await riskBudgetAPI.calculateCVaR(selectedConfig.id, weights);
      if (result.success && result.data) {
        setCvarResult(result.data);
        message.success('CVaR计算成功');
        setActiveTab('cvar');
      } else {
        message.error(result.error || '计算失败');
      }
    } catch (error) {
      message.error('计算失败: ' + (error instanceof Error ? error.message : '未知错误'));
    } finally {
      setCalculateLoading(false);
    }
  };

  const handleRunMonteCarlo = async () => {
    if (!selectedConfig) {
      message.warning('请先选择配置');
      return;
    }

    setMcLoading(true);
    try {
      const result = await riskBudgetAPI.runMonteCarlo(selectedConfig.id, 1000, 252);
      if (result.success && result.data) {
        setMcResult(result.data);
        message.success('蒙特卡洛模拟完成');
        setActiveTab('montecarlo');
      } else {
        message.error(result.error || '模拟失败');
      }
    } catch (error) {
      message.error('模拟失败: ' + (error instanceof Error ? error.message : '未知错误'));
    } finally {
      setMcLoading(false);
    }
  };

  const riskColumns = [
    { title: '资产', dataIndex: 'asset_symbol', key: 'asset' },
    { title: '权重', dataIndex: 'weight', key: 'weight', render: (v: number) => `${(v * 100).toFixed(2)}%` },
    { title: '边际风险', dataIndex: 'marginal_risk', key: 'marginal', render: (v: number) => v?.toFixed(4) },
    { title: '风险贡献', dataIndex: 'risk_contribution', key: 'contribution', render: (v: number) => v?.toFixed(4) },
    { title: '百分比贡献', dataIndex: 'percentage_contribution', key: 'pct', render: (v: number) => `${(v * 100).toFixed(2)}%` },
  ];

  const riskData = cvarResult?.risk_contributions || [];
  const pieData = riskData.map(r => ({
    name: r.asset_symbol,
    value: r.percentage_contribution * 100,
  }));

  const barData = riskData.map(r => ({
    name: r.asset_symbol,
    marginal: r.marginal_risk,
    contribution: r.risk_contribution,
  }));

  return (
    <Layout>
      <Container>
        <Title>风险预算管理</Title>

        <Alert
          message="风险预算管理说明"
          description="通过CVaR约束和风险贡献分解，实现风险预算配置和组合优化。支持历史模拟法、参数法和蒙特卡洛模拟。"
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
                    <Form.Item name="cvar_limit" label="CVaR限制 (%)" rules={[{ required: true }]} initialValue={5}>
                      <InputNumber min={0.1} max={50} step={0.1} style={{ width: '100%' }} />
                    </Form.Item>
                    <Form.Item name="confidence_level" label="置信水平 (%)" rules={[{ required: true }]} initialValue={95}>
                      <InputNumber min={80} max={99.9} step={0.1} style={{ width: '100%' }} />
                    </Form.Item>
                    <Form.Item name="time_horizon" label="时间范围 (天)" rules={[{ required: true }]} initialValue={252}>
                      <InputNumber min={1} max={2520} step={1} style={{ width: '100%' }} />
                    </Form.Item>
                    <Form.Item name="method" label="计算方法" rules={[{ required: true }]} initialValue="historical">
                      <Select>
                        {METHOD_OPTIONS.map(opt => (
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
                      { title: 'CVaR限制', dataIndex: 'cvar_limit', key: 'cvar_limit' },
                      { title: '置信水平', dataIndex: 'confidence_level', key: 'confidence_level', render: (v: number) => `${v}%` },
                      { title: '方法', dataIndex: 'method', key: 'method', render: (v: string) => METHOD_OPTIONS.find(o => o.value === v)?.label || v },
                      { title: '状态', dataIndex: 'is_active', key: 'active', render: (v: boolean) => (
                        <Tag color={v ? 'green' : 'red'}>{v ? '活跃' : '停用'}</Tag>
                      )},
                    ]}
                    rowKey="id"
                    pagination={false}
                    onRow={(record) => ({
                      onClick: () => {
                        setSelectedConfig(record);
                        message.info(`已选择配置 #${record.id}`);
                      },
                    })}
                  />
                </Card>
              </Col>
            </Row>

            {selectedConfig && (
              <Card title="计算操作" style={{ marginTop: 24 }}>
                <Alert
                  message={`当前配置: CVaR限制=${selectedConfig.cvar_limit}%, 置信水平=${selectedConfig.confidence_level}%, 方法=${METHOD_OPTIONS.find(o => o.value === selectedConfig.method)?.label}`}
                  type="info"
                  style={{ marginBottom: 16 }}
                />
                <Row gutter={16} style={{ marginBottom: 16 }}>
                  <Col span={24}>
                    <Form.Item label="资产权重 (用逗号分隔，如: 0.4, 0.3, 0.2, 0.1)">
                      <input
                        type="text"
                        value={weightsInput}
                        onChange={(e) => setWeightsInput(e.target.value)}
                        style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #d9d9d9' }}
                        placeholder="输入权重，如: 0.4, 0.3, 0.2, 0.1"
                      />
                    </Form.Item>
                  </Col>
                </Row>
                <Row gutter={16}>
                  <Col span={12}>
                    <Button
                      type="primary"
                      onClick={handleCalculateCVaR}
                      loading={calculateLoading}
                      icon={<CalculatorOutlined />}
                      block
                    >
                      计算CVaR
                    </Button>
                  </Col>
                  <Col span={12}>
                    <Button
                      type="primary"
                      onClick={handleRunMonteCarlo}
                      loading={mcLoading}
                      icon={<ExperimentOutlined />}
                      block
                    >
                      运行蒙特卡洛模拟
                    </Button>
                  </Col>
                </Row>
              </Card>
            )}
          </TabPane>

          <TabPane tab="CVaR结果" key="cvar">
            {cvarResult ? (
              <>
                <Row gutter={16} style={{ marginBottom: 24 }}>
                  <Col span={12}>
                    <Card>
                      <Statistic title="VaR" value={cvarResult.var * 100} precision={2} suffix="%" />
                    </Card>
                  </Col>
                  <Col span={12}>
                    <Card>
                      <Statistic title="CVaR" value={cvarResult.cvar * 100} precision={2} suffix="%" />
                    </Card>
                  </Col>
                </Row>

                <Row gutter={24}>
                  <Col span={12}>
                    <Card title="风险贡献分布">
                      <ResponsiveContainer width="100%" height={300}>
                        <PieChart>
                          <Pie
                            data={pieData}
                            cx="50%"
                            cy="50%"
                            labelLine={false}
                            label={({ name, value }) => `${name}: ${Number(value).toFixed(1)}%`}
                            outerRadius={100}
                            fill="#8884d8"
                            dataKey="value"
                          >
                            {pieData.map((_, index) => (
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
                    <Card title="边际风险与风险贡献">
                      <ResponsiveContainer width="100%" height={300}>
                        <BarChart data={barData}>
                          <CartesianGrid strokeDasharray="3 3" />
                          <XAxis dataKey="name" />
                          <YAxis />
                          <RechartsTooltip />
                          <Legend />
                          <Bar dataKey="marginal" fill="#8884d8" name="边际风险" />
                          <Bar dataKey="contribution" fill="#82ca9d" name="风险贡献" />
                        </BarChart>
                      </ResponsiveContainer>
                    </Card>
                  </Col>
                </Row>

                <Card title="风险贡献明细" style={{ marginTop: 24 }}>
                  <Table
                    dataSource={riskData}
                    columns={riskColumns}
                    rowKey="id"
                    pagination={false}
                  />
                </Card>
              </>
            ) : (
              <Alert message="请先计算CVaR" type="warning" />
            )}
          </TabPane>

          <TabPane tab="蒙特卡洛结果" key="montecarlo">
            {mcResult ? (
              <>
                <Row gutter={16} style={{ marginBottom: 24 }}>
                  <Col span={6}>
                    <Card>
                      <Statistic title="模拟次数" value={mcResult.num_simulations} />
                    </Card>
                  </Col>
                  <Col span={6}>
                    <Card>
                      <Statistic title="时间步数" value={mcResult.time_steps} />
                    </Card>
                  </Col>
                  <Col span={6}>
                    <Card>
                      <Statistic title="平均收益" value={mcResult.mean_return * 100} precision={2} suffix="%" />
                    </Card>
                  </Col>
                  <Col span={6}>
                    <Card>
                      <Statistic title="标准差" value={mcResult.std_return * 100} precision={2} suffix="%" />
                    </Card>
                  </Col>
                </Row>

                <Row gutter={16} style={{ marginBottom: 24 }}>
                  <Col span={12}>
                    <Card>
                      <Statistic title="5%分位数" value={mcResult.percentile_5 * 100} precision={2} suffix="%" valueStyle={{ color: '#cf1322' }} />
                    </Card>
                  </Col>
                  <Col span={12}>
                    <Card>
                      <Statistic title="95%分位数" value={mcResult.percentile_95 * 100} precision={2} suffix="%" valueStyle={{ color: '#3f8600' }} />
                    </Card>
                  </Col>
                </Row>

                <Card title="模拟统计">
                  <p><strong>模拟日期:</strong> {new Date(mcResult.simulation_date).toLocaleString()}</p>
                  <p><strong>模拟数据:</strong> {mcResult.simulation_data ? '已生成' : '无'}</p>
                </Card>
              </>
            ) : (
              <Alert message="请先运行蒙特卡洛模拟" type="warning" />
            )}
          </TabPane>
        </Tabs>
      </Container>
    </Layout>
  );
};

export default RiskBudget;

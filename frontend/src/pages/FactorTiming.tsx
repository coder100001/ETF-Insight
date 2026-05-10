import React, { useState } from 'react';
import { Card, Select, Button, Row, Col, Statistic, Tag, message, InputNumber, Alert, Table, Tabs } from 'antd';
import { ArrowUpOutlined, ArrowDownOutlined, MinusOutlined, RobotOutlined } from '@ant-design/icons';
import { XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, AreaChart, Area } from 'recharts';
import styled from 'styled-components';
import Layout from '../components/Layout';
import { factorTimingAPI } from '../services/api';
import type { FactorTimingSignal, SignalStrength, AlphaView } from '../types';

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

const SignalCard = styled(Card)`
  .ant-statistic-title {
    font-size: 14px;
    color: #8c8c8c;
  }
  .ant-statistic-content {
    font-size: 24px;
    font-weight: 600;
  }
`;

const FACTOR_OPTIONS = [
  { value: 'Mkt-RF', label: '市场因子 (Mkt-RF)' },
  { value: 'SMB', label: '规模因子 (SMB)' },
  { value: 'HML', label: '价值因子 (HML)' },
  { value: 'RMW', label: '盈利因子 (RMW)' },
  { value: 'CMA', label: '投资因子 (CMA)' },
];

const ASSET_OPTIONS = [
  { value: 'SPY', label: 'SPY (标普500)' },
  { value: 'QQQ', label: 'QQQ (纳斯达克100)' },
  { value: 'IWM', label: 'IWM (罗素2000)' },
  { value: 'EEM', label: 'EEM (新兴市场)' },
  { value: 'VTI', label: 'VTI (全美股票)' },
  { value: 'VEA', label: 'VEA (发达市场)' },
];

const getSignalIcon = (strength: SignalStrength) => {
  switch (strength) {
    case 'strong_positive':
      return <ArrowUpOutlined style={{ color: '#52c41a', fontSize: 32 }} />;
    case 'weak_positive':
      return <ArrowUpOutlined style={{ color: '#73d13d', fontSize: 28 }} />;
    case 'neutral':
      return <MinusOutlined style={{ color: '#8c8c8c', fontSize: 24 }} />;
    case 'weak_negative':
      return <ArrowDownOutlined style={{ color: '#ff7875', fontSize: 28 }} />;
    case 'strong_negative':
      return <ArrowDownOutlined style={{ color: '#ff4d4f', fontSize: 32 }} />;
  }
};

const getSignalColor = (strength: SignalStrength): string => {
  switch (strength) {
    case 'strong_positive': return 'green';
    case 'weak_positive': return 'lime';
    case 'neutral': return 'default';
    case 'weak_negative': return 'orange';
    case 'strong_negative': return 'red';
  }
};

const getSignalLabel = (strength: SignalStrength): string => {
  switch (strength) {
    case 'strong_positive': return '强正向';
    case 'weak_positive': return '弱正向';
    case 'neutral': return '中性';
    case 'weak_negative': return '弱负向';
    case 'strong_negative': return '强负向';
  }
};

const FactorTiming: React.FC = () => {
  const [selectedFactor, setSelectedFactor] = useState('Mkt-RF');
  const [lookbackDays, setLookbackDays] = useState(60);
  const [signal, setSignal] = useState<FactorTimingSignal | null>(null);
  const [signalHistory, setSignalHistory] = useState<FactorTimingSignal[]>([]);
  const [loading, setLoading] = useState(false);
  const [targetAsset, setTargetAsset] = useState('SPY');
  const [generatedView, setGeneratedView] = useState<AlphaView | null>(null);
  const [viewLoading, setViewLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('1');

  const calculateSignal = async () => {
    setLoading(true);
    try {
      const result = await factorTimingAPI.calculateSignal(selectedFactor, lookbackDays);
      if (result.success && result.data) {
        setSignal(result.data);
        message.success('信号计算成功');

        const historyResult = await factorTimingAPI.getSignalHistory(selectedFactor);
        if (historyResult.success && historyResult.data) {
          setSignalHistory(historyResult.data);
        }
      } else {
        message.error(result.error || '信号计算失败');
      }
    } catch (error) {
      message.error('信号计算失败: ' + (error instanceof Error ? error.message : '未知错误'));
    } finally {
      setLoading(false);
    }
  };

  const generateView = async () => {
    if (!signal) {
      message.warning('请先计算信号');
      return;
    }

    setViewLoading(true);
    try {
      const result = await factorTimingAPI.generateView(signal, targetAsset);
      if (result.success && result.data) {
        setGeneratedView(result.data);
        message.success(`Alpha观点已生成: ${targetAsset} 预期收益 ${(result.data.view_return * 100).toFixed(2)}%`);
      } else {
        message.error(result.error || '观点生成失败');
      }
    } catch (error) {
      message.error('观点生成失败: ' + (error instanceof Error ? error.message : '未知错误'));
    } finally {
      setViewLoading(false);
    }
  };

  const chartData = signalHistory.length > 0
    ? signalHistory.map(s => ({
        date: s.signal_date,
        zScore: s.z_score,
        percentile: s.percentile,
        signal: s.signal_score,
      }))
    : signal
      ? [{ date: signal.signal_date, zScore: signal.z_score, percentile: signal.percentile, signal: signal.signal_score }]
      : [];

  const historyColumns = [
    { title: '日期', dataIndex: 'signal_date', key: 'date' },
    { title: '因子', dataIndex: 'factor_name', key: 'factor' },
    { title: 'Z-Score', dataIndex: 'z_score', key: 'zscore', render: (v: number) => v?.toFixed(2) },
    { title: '百分位数', dataIndex: 'percentile', key: 'percentile', render: (v: number) => `${v?.toFixed(1)}%` },
    { title: '信号强度', dataIndex: 'signal_strength', key: 'strength', render: (v: SignalStrength) => (
      <Tag color={getSignalColor(v)}>{getSignalLabel(v)}</Tag>
    )},
    { title: '预期收益', dataIndex: 'expected_return', key: 'return', render: (v: number) => `${(v * 100).toFixed(2)}%` },
  ];

  return (
    <Layout>
      <Container>
        <Title>因子择时信号分析</Title>

        <Card style={{ marginBottom: 24 }}>
          <Row gutter={16} align="middle">
            <Col>
              <span>选择因子: </span>
              <Select value={selectedFactor} onChange={setSelectedFactor} style={{ width: 180 }}>
                {FACTOR_OPTIONS.map(opt => (
                  <Option key={opt.value} value={opt.value}>{opt.label}</Option>
                ))}
              </Select>
            </Col>
            <Col>
              <span>回看天数: </span>
              <InputNumber value={lookbackDays} onChange={(v) => setLookbackDays(v || 60)} min={30} max={252} />
            </Col>
            <Col>
              <Button type="primary" onClick={calculateSignal} loading={loading}>
                计算信号
              </Button>
            </Col>
          </Row>
        </Card>

        {signal && (
          <>
            <Row gutter={16} style={{ marginBottom: 24 }}>
              <Col span={8}>
                <SignalCard>
                  <Statistic
                    title="MA斜率 (60日)"
                    value={signal.ma_slope_60}
                    precision={4}
                    styles={{ content: { color: signal.ma_slope_60 > 0 ? '#3f8600' : '#cf1322' } }}
                  />
                  <Tag color={signal.ma_slope_60 > 0 ? 'green' : 'red'}>
                    {signal.ma_slope_60 > 0 ? '上升趋势' : '下降趋势'}
                  </Tag>
                </SignalCard>
              </Col>
              <Col span={8}>
                <SignalCard>
                  <Statistic
                    title="Z-Score"
                    value={signal.z_score}
                    precision={2}
                    styles={{ content: { color: Math.abs(signal.z_score) > 1.5 ? '#1890ff' : '#8c8c8c' } }}
                  />
                  <Tag color={Math.abs(signal.z_score) > 2 ? 'blue' : 'default'}>
                    {Math.abs(signal.z_score) > 2 ? '极端值' : '正常范围'}
                  </Tag>
                </SignalCard>
              </Col>
              <Col span={8}>
                <SignalCard>
                  <Statistic
                    title="百分位数"
                    value={signal.percentile}
                    suffix="%"
                    precision={1}
                  />
                  <Tag color={signal.percentile > 80 ? 'red' : signal.percentile < 20 ? 'green' : 'default'}>
                    {signal.percentile > 80 ? '历史高位' : signal.percentile < 20 ? '历史低位' : '中等水平'}
                  </Tag>
                </SignalCard>
              </Col>
            </Row>

            <Card title="信号强度" style={{ marginBottom: 24 }}>
              <Row gutter={16} align="middle">
                <Col span={4}>
                  {getSignalIcon(signal.signal_strength)}
                </Col>
                <Col span={6}>
                  <Statistic title="预期收益" value={signal.expected_return * 100} precision={2} suffix="%" />
                </Col>
                <Col span={6}>
                  <Statistic title="置信度" value={signal.confidence} precision={0} suffix="%" />
                </Col>
                <Col span={8}>
                  <Tag color={getSignalColor(signal.signal_strength)} style={{ fontSize: 16, padding: '4px 12px' }}>
                    {getSignalLabel(signal.signal_strength)}
                  </Tag>
                </Col>
              </Row>
            </Card>

            <Tabs activeKey={activeTab} onChange={setActiveTab} style={{ marginBottom: 24 }}>
              <TabPane tab="生成Alpha观点" key="1">
                <Card>
                  <Row gutter={16} align="middle">
                    <Col>
                      <span>目标资产: </span>
                      <Select value={targetAsset} onChange={setTargetAsset} style={{ width: 180 }}>
                        {ASSET_OPTIONS.map(opt => (
                          <Option key={opt.value} value={opt.value}>{opt.label}</Option>
                        ))}
                      </Select>
                    </Col>
                    <Col>
                      <Button type="primary" onClick={generateView} loading={viewLoading} icon={<RobotOutlined />}>
                        生成Alpha观点
                      </Button>
                    </Col>
                  </Row>
                  {generatedView && (
                    <Alert
                      message="Alpha观点已生成"
                      description={
                        <div>
                          <p><strong>资产:</strong> {generatedView.asset_symbol}</p>
                          <p><strong>观点收益:</strong> {(generatedView.view_return * 100).toFixed(2)}%</p>
                          <p><strong>置信度:</strong> {generatedView.confidence}%</p>
                          <p><strong>类型:</strong> {generatedView.view_type === 'absolute' ? '绝对' : '相对'}</p>
                          <p><strong>来源因子:</strong> {generatedView.source_factor}</p>
                        </div>
                      }
                      type="success"
                      style={{ marginTop: 16 }}
                    />
                  )}
                </Card>
              </TabPane>
              <TabPane tab="信号历史" key="2">
                <Card>
                  <Table
                    dataSource={signalHistory}
                    columns={historyColumns}
                    rowKey="id"
                    pagination={{ pageSize: 10 }}
                    locale={{ emptyText: '暂无历史数据' }}
                  />
                </Card>
              </TabPane>
              <TabPane tab="趋势图表" key="3">
                <Card>
                  <ResponsiveContainer width="100%" height={300}>
                    <AreaChart data={chartData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="date" />
                      <YAxis />
                      <Tooltip />
                      <Area type="monotone" dataKey="zScore" stroke="#8884d8" fill="#8884d8" fillOpacity={0.3} name="Z-Score" />
                      <Area type="monotone" dataKey="percentile" stroke="#82ca9d" fill="#82ca9d" fillOpacity={0.3} name="百分位数" />
                    </AreaChart>
                  </ResponsiveContainer>
                </Card>
              </TabPane>
            </Tabs>
          </>
        )}
      </Container>
    </Layout>
  );
};

export default FactorTiming;

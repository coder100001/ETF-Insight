import React, { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Tabs,
  Form,
  InputNumber,
  Select,
  Button,
  Row,
  Col,
  Statistic,
  Table,
  message,
  Spin,
  Input,
  Alert,
} from 'antd';
import { ReloadOutlined } from '@ant-design/icons';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import Layout from '../components/Layout';
import { quantlibAPI } from '../services/api';
import type {
  OptionResult,
  BondResult,
  YieldCurveResult,
  VaRResult,
} from '../types/quantlib';

const { TabPane } = Tabs;
const { Option } = Select;
const { TextArea } = Input;

interface ReferenceDataCache<T> {
  data: T;
  timestamp: number;
  ttl: number;
}

function useReferenceData<T>(
  fetchFn: () => Promise<unknown>,
  cacheKey: string,
  ttlMinutes: number = 5
): { data: T[] | null; loading: boolean; refresh: () => void } {
  const [data, setData] = useState<T[] | null>(null);
  const [loading, setLoading] = useState(false);

  const loadData = useCallback(async (forceRefresh: boolean = false) => {
    if (!forceRefresh) {
      try {
        const cached = localStorage.getItem(cacheKey);
        if (cached) {
          const parsed: ReferenceDataCache<T[]> = JSON.parse(cached);
          const now = Date.now();
          if (now - parsed.timestamp < parsed.ttl) {
            setData(parsed.data);
            return;
          }
        }
      } catch {
        console.warn(`Failed to parse cache for ${cacheKey}`);
      }
    }

    setLoading(true);
    try {
      const response = await fetchFn();
      if (response && Array.isArray(response)) {
        const result = response as T[];
        setData(result);

        const cacheEntry: ReferenceDataCache<T[]> = {
          data: result,
          timestamp: Date.now(),
          ttl: ttlMinutes * 60 * 1000,
        };
        localStorage.setItem(cacheKey, JSON.stringify(cacheEntry));
      }
    } catch (error) {
      console.error(`Failed to load reference data for ${cacheKey}:`, error);
    } finally {
      setLoading(false);
    }
  }, [fetchFn, cacheKey, ttlMinutes]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  return { data, loading, refresh: () => loadData(true) };
}

const QuantLibAnalysis: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [optionResult, setOptionResult] = useState<OptionResult | null>(null);
  const [bondResult, setBondResult] = useState<BondResult | null>(null);
  const [yieldCurveData, setYieldCurveData] = useState<YieldCurveResult | null>(null);
  const [varResult, setVarResult] = useState<VaRResult | null>(null);
  const [americanOptionResult, setAmericanOptionResult] = useState<OptionResult | null>(null);
  const [greeksResult, setGreeksResult] = useState<OptionResult | null>(null);
  const [activeGreeksTab, setActiveGreeksTab] = useState(false);

  const [optionForm] = Form.useForm();
  const [bondForm] = Form.useForm();
  const [yieldCurveForm] = Form.useForm();
  const [varForm] = Form.useForm();
  const [americanOptionForm] = Form.useForm();

  const { data: currencies, loading: loadingCurrencies, refresh: refreshCurrencies } =
    useReferenceData<string>(
      async () => {
        const response = await quantlibAPI.getReferenceData('currencies');
        return response.success ? response.data : null;
      },
      'quantlib_currencies',
      5
    );

  const handleOptionPrice = async (values: {
    spot: number;
    strike: number;
    rate: number;
    volatility: number;
    time_to_expiry: number;
    option_type: 'call' | 'put';
  }) => {
    setLoading(true);
    try {
      const response = await quantlibAPI.priceEuropeanOption(values);
      if (response.success && response.data) {
        setOptionResult(response.data);
        message.success('期权定价完成');
      } else {
        message.error(response.error || '期权定价失败');
      }
    } catch {
      message.error('请求失败');
    } finally {
      setLoading(false);
    }
  };

  const handleBondPrice = async (values: {
    face_value: number;
    coupon_rate: number;
    frequency: number;
    maturity: string;
    yield_to_maturity: number;
  }) => {
    setLoading(true);
    try {
      const response = await quantlibAPI.priceBond(values);
      if (response.success && response.data) {
        setBondResult(response.data);
        message.success('债券定价完成');
      } else {
        message.error(response.error || '债券定价失败');
      }
    } catch {
      message.error('请求失败');
    } finally {
      setLoading(false);
    }
  };

  const handleYieldCurve = async (values: {
    currency: string;
    tenors: string;
    rates: string;
  }) => {
    setLoading(true);
    try {
      const tenorList = values.tenors.split(',').map((t) => t.trim());
      const rateList = values.rates.split(',').map((r) => parseFloat(r.trim()));

      if (tenorList.length !== rateList.length) {
        message.error('期限和利率数量必须一致');
        setLoading(false);
        return;
      }

      const response = await quantlibAPI.buildYieldCurve({
        currency: values.currency,
        tenors: tenorList,
        rates: rateList,
      });

      if (response.success && response.data) {
        setYieldCurveData(response.data);
        message.success('收益率曲线构建完成');
      } else {
        message.error(response.error || '收益率曲线构建失败');
      }
    } catch {
      message.error('请求失败');
    } finally {
      setLoading(false);
    }
  };

  const handleVaRCalculate = async (values: {
    portfolio_value: number;
    returns: string;
    confidence: number;
    holding_period: number;
    method: 'historical' | 'parametric' | 'monte_carlo';
  }) => {
    setLoading(true);
    try {
      const returnList = values.returns.split(',').map((r) => parseFloat(r.trim()));

      const response = await quantlibAPI.calculateVaR({
        portfolio_value: values.portfolio_value,
        returns: returnList,
        confidence: values.confidence,
        holding_period: values.holding_period,
        method: values.method,
      });

      if (response.success && response.data) {
        setVarResult(response.data);
        message.success('VaR 计算完成');
      } else {
        message.error(response.error || 'VaR 计算失败');
      }
    } catch {
      message.error('请求失败');
    } finally {
      setLoading(false);
    }
  };

  const handleAmericanOptionPrice = async (values: {
    spot: number;
    strike: number;
    rate: number;
    volatility: number;
    time_to_expiry: number;
    option_type: 'call' | 'put';
    steps: number;
  }) => {
    setLoading(true);
    try {
      const response = await quantlibAPI.priceAmericanOption(values);
      if (response.success && response.data) {
        setAmericanOptionResult(response.data);
        message.success('美式期权定价完成');
      } else {
        message.error(response.error || '美式期权定价失败');
      }
    } catch {
      message.error('请求失败');
    } finally {
      setLoading(false);
    }
  };

  const handleGreeksCalculate = async (values: {
    spot: number;
    strike: number;
    rate: number;
    volatility: number;
    time_to_expiry: number;
    option_type: 'call' | 'put';
  }) => {
    setLoading(true);
    try {
      const response = await quantlibAPI.calculateGreeks(values);
      if (response.success && response.data) {
        setGreeksResult(response.data);
        message.success('Greeks 计算完成');
      } else {
        message.error(response.error || 'Greeks 计算失败');
      }
    } catch {
      message.error('请求失败');
    } finally {
      setLoading(false);
    }
  };

  const optionColumns = [
    { title: '指标', dataIndex: 'label', key: 'label' },
    { title: '值', dataIndex: 'value', key: 'value' },
  ];

  const getOptionTableData = () => {
    if (!optionResult) return [];
    return [
      { key: 'price', label: '期权价格', value: parseFloat(optionResult.price).toFixed(4) },
      { key: 'delta', label: 'Delta', value: parseFloat(optionResult.delta).toFixed(4) },
      { key: 'gamma', label: 'Gamma', value: parseFloat(optionResult.gamma).toFixed(4) },
      { key: 'theta', label: 'Theta', value: parseFloat(optionResult.theta).toFixed(4) },
      { key: 'vega', label: 'Vega', value: parseFloat(optionResult.vega).toFixed(4) },
      { key: 'rho', label: 'Rho', value: parseFloat(optionResult.rho).toFixed(4) },
    ];
  };

  const getYieldCurveChartData = () => {
    if (!yieldCurveData) return [];
    return yieldCurveData.tenors.map((tenor, index) => ({
      tenor,
      spot: parseFloat(yieldCurveData.rates[index]),
      zero: parseFloat(yieldCurveData.zero_rates[index]),
      forward: parseFloat(yieldCurveData.forward_rates[index]),
    }));
  };

  return (
    <Layout>
      <Spin spinning={loading}>
        <div style={{ padding: 24 }}>
          <h2>QuantLib 量化分析</h2>
          <p style={{ color: '#888', marginBottom: 24 }}>
            基于 QuantLib 的期权定价、债券分析、收益率曲线和风险度量
          </p>

        <Tabs defaultActiveKey="option">
          <TabPane tab="期权定价" key="option">
            <Row gutter={24}>
              <Col span={10}>
                <Card title="欧式期权参数">
                  <Form
                    form={optionForm}
                    layout="vertical"
                    onFinish={handleOptionPrice}
                    initialValues={{
                      spot: 100,
                      strike: 100,
                      rate: 0.05,
                      volatility: 0.2,
                      time_to_expiry: 1.0,
                      option_type: 'call',
                    }}
                  >
                    <Form.Item label="标的价格" name="spot" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={0} step={1} />
                    </Form.Item>
                    <Form.Item label="行权价" name="strike" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={0} step={1} />
                    </Form.Item>
                    <Form.Item label="无风险利率" name="rate" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={0} max={1} step={0.01} />
                    </Form.Item>
                    <Form.Item label="波动率" name="volatility" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={0} max={5} step={0.01} />
                    </Form.Item>
                    <Form.Item label="到期时间(年)" name="time_to_expiry" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={0.01} step={0.1} />
                    </Form.Item>
                    <Form.Item label="期权类型" name="option_type" rules={[{ required: true }]}>
                      <Select>
                        <Option value="call">看涨 (Call)</Option>
                        <Option value="put">看跌 (Put)</Option>
                      </Select>
                    </Form.Item>
                    <Form.Item>
                      <Button type="primary" htmlType="submit" block>
                        计算期权价格
                      </Button>
                    </Form.Item>
                  </Form>
                </Card>
              </Col>
              <Col span={14}>
                <Card title="定价结果与希腊字母" extra={
                  <Button size="small" onClick={() => setActiveGreeksTab(!activeGreeksTab)}>
                    {activeGreeksTab ? '返回定价结果' : 'Greeks 敏感性分析'}
                  </Button>
                }>
                  {activeGreeksTab ? (
                    greeksResult ? (
                      <div>
                        <Alert
                          message="Greeks 敏感性分析"
                          description="展示各希腊字母对期权价格的影响"
                          type="info"
                          showIcon
                          style={{ marginBottom: 16 }}
                        />
                        <Table
                          dataSource={[
                            { key: 'delta', label: 'Delta (Δ)', value: parseFloat(greeksResult.delta).toFixed(6), desc: '价格对标的价格敏感度' },
                            { key: 'gamma', label: 'Gamma (Γ)', value: parseFloat(greeksResult.gamma).toFixed(6), desc: 'Delta 对标的价格变化率' },
                            { key: 'theta', label: 'Theta (Θ)', value: parseFloat(greeksResult.theta).toFixed(6), desc: '时间衰减率' },
                            { key: 'vega', label: 'Vega (ν)', value: parseFloat(greeksResult.vega).toFixed(6), desc: '对波动率的敏感度' },
                            { key: 'rho', label: 'Rho (ρ)', value: parseFloat(greeksResult.rho).toFixed(6), desc: '对利率的敏感度' },
                          ]}
                          columns={[
                            { title: '指标', dataIndex: 'label', key: 'label' },
                            { title: '数值', dataIndex: 'value', key: 'value' },
                            { title: '解释', dataIndex: 'desc', key: 'desc' },
                          ]}
                          pagination={false}
                          size="small"
                        />
                        <Button
                          type="link"
                          onClick={() => {
                            const formValues = optionForm.getFieldsValue();
                            handleGreeksCalculate(formValues);
                          }}
                          style={{ marginTop: 12 }}
                        >
                          使用当前参数重新计算 Greeks
                        </Button>
                      </div>
                    ) : (
                      <div style={{ textAlign: 'center', padding: 40, color: '#aaa' }}>
                        请先计算期权价格，然后查看 Greeks 分析
                      </div>
                    )
                  ) : (
                    optionResult ? (
                      <Table
                        dataSource={getOptionTableData()}
                        columns={optionColumns}
                        pagination={false}
                        size="small"
                      />
                    ) : (
                      <div style={{ textAlign: 'center', padding: 40, color: '#aaa' }}>
                        请填写参数并点击计算
                      </div>
                    )
                  )}
                </Card>
              </Col>
            </Row>
          </TabPane>

          <TabPane tab="美式期权" key="americanOption">
            <Row gutter={24}>
              <Col span={10}>
                <Card title="美式期权参数（二叉树法）">
                  <Form
                    form={americanOptionForm}
                    layout="vertical"
                    onFinish={handleAmericanOptionPrice}
                    initialValues={{
                      spot: 100,
                      strike: 100,
                      rate: 0.05,
                      volatility: 0.25,
                      time_to_expiry: 1.0,
                      option_type: 'call',
                      steps: 200,
                    }}
                  >
                    <Form.Item label="标的价格" name="spot" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={0} step={1} />
                    </Form.Item>
                    <Form.Item label="行权价" name="strike" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={0} step={1} />
                    </Form.Item>
                    <Form.Item label="无风险利率" name="rate" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={0} max={1} step={0.01} />
                    </Form.Item>
                    <Form.Item label="波动率" name="volatility" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={0} max={5} step={0.01} />
                    </Form.Item>
                    <Form.Item label="到期时间(年)" name="time_to_expiry" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={0.01} step={0.1} />
                    </Form.Item>
                    <Form.Item label="期权类型" name="option_type" rules={[{ required: true }]}>
                      <Select>
                        <Option value="call">看涨 (Call)</Option>
                        <Option value="put">看跌 (Put)</Option>
                      </Select>
                    </Form.Item>
                    <Form.Item label="二叉树步数" name="steps" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={10} max={1000} step={10} />
                    </Form.Item>
                    <Form.Item>
                      <Button type="primary" htmlType="submit" block>
                        计算美式期权价格
                      </Button>
                    </Form.Item>
                  </Form>
                </Card>
              </Col>
              <Col span={14}>
                <Card title="美式期权定价结果">
                  {americanOptionResult ? (
                    <Table
                      dataSource={[
                        { key: 'price', label: '期权价格', value: parseFloat(americanOptionResult.price).toFixed(4) },
                        { key: 'delta', label: 'Delta', value: parseFloat(americanOptionResult.delta).toFixed(4) },
                        { key: 'gamma', label: 'Gamma', value: parseFloat(americanOptionResult.gamma).toFixed(4) },
                        { key: 'theta', label: 'Theta', value: parseFloat(americanOptionResult.theta).toFixed(4) },
                        { key: 'vega', label: 'Vega', value: parseFloat(americanOptionResult.vega).toFixed(4) },
                        { key: 'rho', label: 'Rho', value: parseFloat(americanOptionResult.rho).toFixed(4) },
                      ]}
                      columns={optionColumns}
                      pagination={false}
                      size="small"
                    />
                  ) : (
                    <div style={{ textAlign: 'center', padding: 40, color: '#aaa' }}>
                      请填写参数并点击计算
                    </div>
                  )}
                </Card>
              </Col>
            </Row>
          </TabPane>

          <TabPane tab="债券定价" key="bond">
            <Row gutter={24}>
              <Col span={10}>
                <Card title="债券参数">
                  <Form
                    form={bondForm}
                    layout="vertical"
                    onFinish={handleBondPrice}
                    initialValues={{
                      face_value: 1000,
                      coupon_rate: 0.05,
                      frequency: 2,
                      maturity: '2030-01-01',
                      yield_to_maturity: 0.04,
                    }}
                  >
                    <Form.Item label="面值" name="face_value" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={0} step={100} />
                    </Form.Item>
                    <Form.Item label="票面利率" name="coupon_rate" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={0} max={1} step={0.005} />
                    </Form.Item>
                    <Form.Item label="付息频率(次/年)" name="frequency" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={1} max={12} />
                    </Form.Item>
                    <Form.Item label="到期日" name="maturity" rules={[{ required: true }]}>
                      <Input placeholder="YYYY-MM-DD" />
                    </Form.Item>
                    <Form.Item label="到期收益率" name="yield_to_maturity" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={0} max={1} step={0.005} />
                    </Form.Item>
                    <Form.Item>
                      <Button type="primary" htmlType="submit" block>
                        计算债券价格
                      </Button>
                    </Form.Item>
                  </Form>
                </Card>
              </Col>
              <Col span={14}>
                <Card title="债券分析结果">
                  {bondResult ? (
                    <Row gutter={[16, 16]}>
                      <Col span={12}>
                        <Statistic title="净价 (Clean Price)" value={parseFloat(bondResult.clean_price)} precision={4} />
                      </Col>
                      <Col span={12}>
                        <Statistic title="全价 (Dirty Price)" value={parseFloat(bondResult.dirty_price)} precision={4} />
                      </Col>
                      <Col span={12}>
                        <Statistic title="久期 (Duration)" value={parseFloat(bondResult.duration)} precision={4} suffix="年" />
                      </Col>
                      <Col span={12}>
                        <Statistic title="修正久期" value={parseFloat(bondResult.modified_duration)} precision={4} />
                      </Col>
                      <Col span={12}>
                        <Statistic title="凸性 (Convexity)" value={parseFloat(bondResult.convexity)} precision={4} />
                      </Col>
                      <Col span={12}>
                        <Statistic title="应计利息" value={parseFloat(bondResult.accrued_interest)} precision={4} />
                      </Col>
                    </Row>
                  ) : (
                    <div style={{ textAlign: 'center', padding: 40, color: '#aaa' }}>
                      请填写参数并点击计算
                    </div>
                  )}
                </Card>
              </Col>
            </Row>
          </TabPane>

          <TabPane tab="收益率曲线" key="yieldCurve">
            <Row gutter={24}>
              <Col span={10}>
                <Card title="收益率曲线参数">
                  <Form
                    form={yieldCurveForm}
                    layout="vertical"
                    onFinish={handleYieldCurve}
                    initialValues={{
                      currency: 'USD',
                      tenors: '1M,3M,6M,1Y,2Y,5Y,10Y,30Y',
                      rates: '0.043,0.044,0.045,0.046,0.047,0.048,0.049,0.05',
                    }}
                  >
                    <Form.Item label="货币" name="currency" rules={[{ required: true }]}>
                      <Input.Group compact>
                        <Select
                          style={{ width: 'calc(100% - 32px)' }}
                          loading={loadingCurrencies}
                          showSearch
                          optionFilterProp="children"
                          notFoundContent={loadingCurrencies ? <Spin size="small" /> : '暂无数据'}
                        >
                          {(currencies || ['USD', 'EUR', 'CNY', 'GBP', 'JPY']).map((currency) => (
                            <Option key={currency} value={currency}>{currency}</Option>
                          ))}
                        </Select>
                        <Button
                          icon={<ReloadOutlined />}
                          onClick={(e) => {
                            e.preventDefault();
                            refreshCurrencies();
                          }}
                          title="刷新货币列表"
                        />
                      </Input.Group>
                    </Form.Item>
                    <Form.Item label="期限 (逗号分隔)" name="tenors" rules={[{ required: true }]}>
                      <TextArea rows={2} placeholder="1M,3M,6M,1Y,2Y,5Y,10Y,30Y" />
                    </Form.Item>
                    <Form.Item label="利率 (逗号分隔)" name="rates" rules={[{ required: true }]}>
                      <TextArea rows={2} placeholder="0.043,0.044,0.045,0.046,0.047,0.048,0.049,0.05" />
                    </Form.Item>
                    <Form.Item>
                      <Button type="primary" htmlType="submit" block>
                        构建收益率曲线
                      </Button>
                    </Form.Item>
                  </Form>
                </Card>
              </Col>
              <Col span={14}>
                <Card title="收益率曲线">
                  {yieldCurveData ? (
                    <ResponsiveContainer width="100%" height={400}>
                      <LineChart data={getYieldCurveChartData()}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="tenor" />
                        <YAxis tickFormatter={(v: unknown) => `${(Number(v) * 100).toFixed(1)}%`} />
                        <Tooltip
                          formatter={(value: unknown) => `${(Number(value) * 100).toFixed(2)}%`}
                        />
                        <Legend />
                        <Line
                          type="monotone"
                          dataKey="spot"
                          stroke="#1890ff"
                          name="即期利率"
                          strokeWidth={2}
                        />
                        <Line
                          type="monotone"
                          dataKey="zero"
                          stroke="#52c41a"
                          name="零息利率"
                          strokeWidth={2}
                        />
                        <Line
                          type="monotone"
                          dataKey="forward"
                          stroke="#fa8c16"
                          name="远期利率"
                          strokeWidth={2}
                        />
                      </LineChart>
                    </ResponsiveContainer>
                  ) : (
                    <div style={{ textAlign: 'center', padding: 40, color: '#aaa' }}>
                      请填写参数并点击构建
                    </div>
                  )}
                </Card>
              </Col>
            </Row>
          </TabPane>

          <TabPane tab="VaR 计算" key="var">
            <Row gutter={24}>
              <Col span={10}>
                <Card title="VaR 参数">
                  <Form
                    form={varForm}
                    layout="vertical"
                    onFinish={handleVaRCalculate}
                    initialValues={{
                      portfolio_value: 1000000,
                      returns: '-0.02,0.01,-0.015,0.008,-0.005,0.012,-0.008,0.003,-0.01,0.006',
                      confidence: 0.95,
                      holding_period: 1,
                      method: 'historical',
                    }}
                  >
                    <Form.Item label="组合价值" name="portfolio_value" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={0} step={10000} />
                    </Form.Item>
                    <Form.Item label="历史收益率 (逗号分隔)" name="returns" rules={[{ required: true }]}>
                      <TextArea rows={3} placeholder="-0.02,0.01,-0.015,..." />
                    </Form.Item>
                    <Form.Item label="置信水平" name="confidence" rules={[{ required: true }]}>
                      <Select>
                        <Option value={0.9}>90%</Option>
                        <Option value={0.95}>95%</Option>
                        <Option value={0.99}>99%</Option>
                      </Select>
                    </Form.Item>
                    <Form.Item label="持有期(天)" name="holding_period" rules={[{ required: true }]}>
                      <InputNumber style={{ width: '100%' }} min={1} max={30} />
                    </Form.Item>
                    <Form.Item label="计算方法" name="method" rules={[{ required: true }]}>
                      <Select>
                        <Option value="historical">历史模拟法</Option>
                        <Option value="parametric">参数法</Option>
                        <Option value="monte_carlo">蒙特卡洛法</Option>
                      </Select>
                    </Form.Item>
                    <Form.Item>
                      <Button type="primary" htmlType="submit" block>
                        计算 VaR
                      </Button>
                    </Form.Item>
                  </Form>
                </Card>
              </Col>
              <Col span={14}>
                <Card title="VaR 分析结果">
                  {varResult ? (
                    <Row gutter={[16, 16]}>
                      <Col span={12}>
                        <Statistic
                          title="VaR (风险价值)"
                          value={parseFloat(varResult.var)}
                          precision={2}
                          prefix="$"
                          styles={{ content: { color: '#cf1322' } }}
                        />
                      </Col>
                      <Col span={12}>
                        <Statistic
                          title="CVaR (条件风险价值)"
                          value={parseFloat(varResult.cvar)}
                          precision={2}
                          prefix="$"
                          styles={{ content: { color: '#a8071a' } }}
                        />
                      </Col>
                      <Col span={12}>
                        <Statistic
                          title="置信水平"
                          value={parseFloat(varResult.confidence) * 100}
                          precision={0}
                          suffix="%"
                        />
                      </Col>
                      <Col span={12}>
                        <Statistic
                          title="持有期"
                          value={varResult.holding_period}
                          suffix="天"
                        />
                      </Col>
                      <Col span={24}>
                        <Statistic
                          title="计算方法"
                          value={
                            varResult.method === 'historical'
                              ? '历史模拟法'
                              : varResult.method === 'parametric'
                              ? '参数法'
                              : '蒙特卡洛法'
                          }
                        />
                      </Col>
                    </Row>
                  ) : (
                    <div style={{ textAlign: 'center', padding: 40, color: '#aaa' }}>
                      请填写参数并点击计算
                    </div>
                  )}
                </Card>
              </Col>
            </Row>
          </TabPane>
        </Tabs>
        </div>
      </Spin>
    </Layout>
  );
};

export default QuantLibAnalysis;

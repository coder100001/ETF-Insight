import React, { useState, useEffect } from 'react';
import {
  Card, Row, Col, InputNumber, Button, Table, Typography,
  Slider, Select, Statistic, message,
} from 'antd';
import { DollarOutlined, RiseOutlined, CalendarOutlined } from '@ant-design/icons';
import Layout from '../components/Layout';
import ReactEChart from '../components/ReactEChart';
import type { EChartsOption } from 'echarts';
import { request } from '../services/api';

const { Title, Text, Paragraph } = Typography;

interface DCAInput {
  etfSymbol: string;
  monthlyInvestment: number;
  years: number;
  expectedReturn: number; // 年化预期收益率 (%)
  dividendYield: number;  // 年化股息率 (%)
}

interface DCAYearData {
  year: number;
  investedTotal: number;
  portfolioValue: number;
  annualReturn: number;
  dividendIncome: number;
  totalGain: number;
  gainPercent: number;
}

interface DCAProjection {
  inputs: DCAInput;
  yearlyData: DCAYearData[];
  totalInvested: number;
  finalValue: number;
  totalGain: number;
  totalGainPercent: number;
  dividendContribution: number;
  annualizedReturn: number;
}

interface ETFInfo {
  symbol: string;
  name: string;
  current_price: number;
}

// DCA 计算核心逻辑
const calculateDCA = (input: DCAInput): DCAProjection => {
  const { monthlyInvestment, years, expectedReturn, dividendYield } = input;
  const monthlyRate = expectedReturn / 100 / 12;
  const monthlyDividendRate = dividendYield / 100 / 12;

  const yearlyData: DCAYearData[] = [];
  let totalInvested = 0;
  let portfolioValue = 0;
  let cumulativeDividend = 0;

  for (let year = 1; year <= years; year++) {
    const monthsInYear = 12;
    const yearStartValue = portfolioValue;
    let yearDividend = 0;

    for (let m = 0; m < monthsInYear; m++) {
      // 每月定投
      totalInvested += monthlyInvestment;
      portfolioValue += monthlyInvestment;

      // 月度收益
      portfolioValue *= (1 + monthlyRate);

      // 股息再投资
      const dividendThisMonth = portfolioValue * monthlyDividendRate;
      portfolioValue += dividendThisMonth;
      yearDividend += dividendThisMonth;
      cumulativeDividend += dividendThisMonth;
    }

    const annualReturn = portfolioValue - yearStartValue - (monthlyInvestment * monthsInYear);
    const totalGain = portfolioValue - totalInvested;

    yearlyData.push({
      year,
      investedTotal: Math.round(totalInvested * 100) / 100,
      portfolioValue: Math.round(portfolioValue * 100) / 100,
      annualReturn: Math.round(annualReturn * 100) / 100,
      dividendIncome: Math.round(yearDividend * 100) / 100,
      totalGain: Math.round(totalGain * 100) / 100,
      gainPercent: Math.round((totalGain / totalInvested) * 10000) / 100,
    });
  }

  const totalGain = portfolioValue - totalInvested;

  // 计算年化收益率 (CAGR)
  const annualizedReturn = totalInvested > 0 && years > 0
    ? (Math.pow(portfolioValue / totalInvested, 1 / years) - 1) * 100
    : 0;

  return {
    inputs: input,
    yearlyData,
    totalInvested: Math.round(totalInvested * 100) / 100,
    finalValue: Math.round(portfolioValue * 100) / 100,
    totalGain: Math.round(totalGain * 100) / 100,
    totalGainPercent: Math.round((totalGain / totalInvested) * 10000) / 100,
    dividendContribution: Math.round(cumulativeDividend * 100) / 100,
    annualizedReturn: Math.round(annualizedReturn * 10000) / 100,
  };
};

// ETF 历史收益率估算
const ETF_DEFAULT_RETURNS: Record<string, { name: string; expectedReturn: number; dividendYield: number }> = {
  'SPY': { name: 'SPDR S&P 500 ETF', expectedReturn: 10, dividendYield: 1.3 },
  'QQQ': { name: 'Invesco QQQ Trust', expectedReturn: 12, dividendYield: 0.6 },
  'VTI': { name: 'Vanguard Total Stock Market', expectedReturn: 10, dividendYield: 1.5 },
  'SCHD': { name: 'Schwab US Dividend Equity ETF', expectedReturn: 11, dividendYield: 3.5 },
  'VOO': { name: 'Vanguard S&P 500 ETF', expectedReturn: 10, dividendYield: 1.4 },
  'BND': { name: 'Vanguard Total Bond Market', expectedReturn: 4, dividendYield: 3.0 },
  'VXUS': { name: 'Vanguard Total International Stock', expectedReturn: 8, dividendYield: 2.5 },
  'JEPI': { name: 'JPMorgan Equity Premium Income', expectedReturn: 9, dividendYield: 7.5 },
  'JEPQ': { name: 'JPMorgan Nasdaq Equity Premium Income', expectedReturn: 10, dividendYield: 9.5 },
  'SPYD': { name: 'SPDR Portfolio S&P 500 High Dividend', expectedReturn: 9, dividendYield: 4.5 },
  'VYM': { name: 'Vanguard High Dividend Yield', expectedReturn: 10, dividendYield: 2.8 },
  'DGRO': { name: 'iShares Core Dividend Growth', expectedReturn: 10, dividendYield: 2.5 },
  'HDV': { name: 'iShares Core High Dividend', expectedReturn: 9, dividendYield: 3.2 },
  'VGT': { name: 'Vanguard Information Technology', expectedReturn: 13, dividendYield: 0.7 },
  'ARKK': { name: 'ARK Innovation ETF', expectedReturn: 15, dividendYield: 0.0 },
};

const DCACalculator: React.FC = () => {
  const [availableETFs, setAvailableETFs] = useState<ETFInfo[]>([]);
  const [etfLoading, setEtfLoading] = useState(true);
  const [selectedSymbol, setSelectedSymbol] = useState<string>('SPY');
  const [monthlyInvestment, setMonthlyInvestment] = useState(500);
  const [years, setYears] = useState(10);
  const [expectedReturn, setExpectedReturn] = useState(10);
  const [dividendYield, setDividendYield] = useState(1.3);
  const [result, setResult] = useState<DCAProjection | null>(null);

  useEffect(() => {
    const fetchETFs = async () => {
      try {
        const data = await request<{ success: boolean; data?: ETFInfo[] }>('/etf/list?pageSize=100');
        if (data.success && data.data) {
          setAvailableETFs(data.data);
        }
      } catch {
        // 使用默认 ETF 列表（离线可用）
      } finally {
        setEtfLoading(false);
      }
    };
    fetchETFs();
  }, []);

  const handleETFSymbolChange = (symbol: string) => {
    setSelectedSymbol(symbol);
    const defaults = ETF_DEFAULT_RETURNS[symbol];
    if (defaults) {
      setExpectedReturn(defaults.expectedReturn);
      setDividendYield(defaults.dividendYield);
    }
  };

  const handleCalculate = () => {
    if (monthlyInvestment <= 0) {
      message.error('每月定投金额必须大于 0');
      return;
    }
    if (years <= 0 || years > 50) {
      message.error('投资年限必须在 1-50 年之间');
      return;
    }
    const projection = calculateDCA({
      etfSymbol: selectedSymbol,
      monthlyInvestment,
      years,
      expectedReturn,
      dividendYield,
    });
    setResult(projection);
  };

  const chartData = result ? result.yearlyData.map(d => ({
    year: d.year,
    '累计投入': d.investedTotal,
    '组合价值': d.portfolioValue,
    '累计分红': result.yearlyData.slice(0, d.year).reduce((sum, yd) => sum + yd.dividendIncome, 0),
  })) : [];

  // ECharts 配置：面积图 = line + areaStyle
  const chartOption: EChartsOption = {
    tooltip: {
      trigger: 'axis',
      formatter(params) {
        const arr = Array.isArray(params) ? params : [params];
        const year = arr[0]?.name;
        let text = `第 ${year} 年末`;
        arr.forEach((p) => {
          const v = Number(p.value);
          text += `<br/>${p.marker}${p.seriesName}: $${v.toLocaleString('en-US', { minimumFractionDigits: 2 })}`;
        });
        return text;
      },
    },
    legend: {
      data: ['组合价值', '累计投入'],
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '10%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: chartData.map(d => d.year),
      boundaryGap: false,
      name: '年份',
      nameLocation: 'middle',
      nameGap: 30,
      splitLine: { show: true, lineStyle: { type: 'dashed' } },
    },
    yAxis: {
      type: 'value',
      name: '价值 (USD)',
      nameLocation: 'middle',
      nameGap: 50,
      nameRotate: 90,
      axisLabel: {
        formatter: (v: number) => `$${(v / 1000).toFixed(0)}k`,
      },
      splitLine: { lineStyle: { type: 'dashed' } },
    },
    series: [
      {
        name: '组合价值',
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        data: chartData.map(d => d['组合价值']),
        itemStyle: { color: '#52c41a' },
        lineStyle: { color: '#52c41a', width: 2 },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(82, 196, 26, 0.3)' },
              { offset: 1, color: 'rgba(82, 196, 26, 0)' },
            ],
          },
        },
      },
      {
        name: '累计投入',
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        data: chartData.map(d => d['累计投入']),
        itemStyle: { color: '#1890ff' },
        lineStyle: { color: '#1890ff', width: 2, type: 'dashed' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(24, 144, 255, 0.2)' },
              { offset: 1, color: 'rgba(24, 144, 255, 0)' },
            ],
          },
        },
      },
    ],
  };

  const columns = [
    { title: '年份', dataIndex: 'year', key: 'year', width: 60 },
    {
      title: '累计投入', dataIndex: 'investedTotal', key: 'investedTotal',
      render: (v: number) => `$${v.toLocaleString('en-US', { minimumFractionDigits: 2 })}`,
    },
    {
      title: '组合价值', dataIndex: 'portfolioValue', key: 'portfolioValue',
      render: (v: number) => `$${v.toLocaleString('en-US', { minimumFractionDigits: 2 })}`,
    },
    {
      title: '年度收益', dataIndex: 'annualReturn', key: 'annualReturn',
      render: (v: number) => (
        <span style={{ color: v >= 0 ? '#3f8600' : '#cf1322' }}>
          {v >= 0 ? '+' : ''}${v.toLocaleString('en-US', { minimumFractionDigits: 2 })}
        </span>
      ),
    },
    {
      title: '股息收入', dataIndex: 'dividendIncome', key: 'dividendIncome',
      render: (v: number) => `$${v.toLocaleString('en-US', { minimumFractionDigits: 2 })}`,
    },
    {
      title: '累积收益', dataIndex: 'totalGain', key: 'totalGain',
      render: (v: number) => (
        <span style={{ color: v >= 0 ? '#3f8600' : '#cf1322' }}>
          {v >= 0 ? '+' : ''}${v.toLocaleString('en-US', { minimumFractionDigits: 2 })}
        </span>
      ),
    },
    {
      title: '收益率', dataIndex: 'gainPercent', key: 'gainPercent',
      render: (v: number) => (
        <span style={{ color: v >= 0 ? '#3f8600' : '#cf1322' }}>
          {v >= 0 ? '+' : ''}{v.toFixed(2)}%
        </span>
      ),
    },
  ];

  return (
    <Layout>
      <div style={{ padding: '24px', maxWidth: 1200, margin: '0 auto' }}>
        <Title level={2}>
          <DollarOutlined /> 定投计算器
        </Title>
        <Paragraph type="secondary">
          计算每月定投 ETF 的长期复利效果。通过定期定额投资，利用时间复利效应实现财富增长。
        </Paragraph>

        <Row gutter={24}>
          <Col xs={24} lg={8}>
            <Card title="定投参数设置" style={{ marginBottom: 24 }}>
              <div style={{ marginBottom: 16 }}>
                <Text strong>
                  <RiseOutlined /> 选择 ETF
                </Text>
                <Select
                  style={{ width: '100%', marginTop: 8 }}
                  value={selectedSymbol}
                  onChange={handleETFSymbolChange}
                  loading={etfLoading}
                  showSearch
                  filterOption={(input, option) =>
                    (option?.label as string || '').toLowerCase().includes(input.toLowerCase())
                  }
                  options={[
                    ...Object.entries(ETF_DEFAULT_RETURNS).map(([k, v]) => ({
                      label: `${k} - ${v.name}`,
                      value: k,
                    })),
                    ...availableETFs
                      .filter(e => !ETF_DEFAULT_RETURNS[e.symbol])
                      .map(e => ({
                        label: `${e.symbol} - ${e.name}`,
                        value: e.symbol,
                      })),
                  ]}
                />
              </div>

              <div style={{ marginBottom: 16 }}>
                <Text strong>
                  <DollarOutlined /> 每月定投金额
                </Text>
                <InputNumber
                  style={{ width: '100%', marginTop: 8 }}
                  value={monthlyInvestment}
                  onChange={(v) => setMonthlyInvestment(v || 500)}
                  min={100}
                  max={100000}
                  step={100}
                  formatter={(v) => `$ ${v}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                  parser={(v) => parseFloat((v || '0').replace(/[$\s,]/g, ''))}
                />
              </div>

              <div style={{ marginBottom: 16 }}>
                <Text strong>
                  <CalendarOutlined /> 投资年限
                </Text>
                <Slider
                  min={1}
                  max={50}
                  value={years}
                  onChange={setYears}
                  marks={{ 1: '1年', 10: '10年', 20: '20年', 30: '30年', 50: '50年' }}
                />
              </div>

              <div style={{ marginBottom: 16 }}>
                <Text strong>预期年化收益率 ({expectedReturn.toFixed(1)}%)</Text>
                <Slider
                  min={1}
                  max={25}
                  step={0.5}
                  value={expectedReturn}
                  onChange={setExpectedReturn}
                  marks={{ 1: '1%', 10: '10%', 20: '20%', 25: '25%' }}
                />
              </div>

              <div style={{ marginBottom: 16 }}>
                <Text strong>股息率 ({dividendYield.toFixed(1)}%)</Text>
                <Slider
                  min={0}
                  max={12}
                  step={0.1}
                  value={dividendYield}
                  onChange={setDividendYield}
                  marks={{ 0: '0%', 4: '4%', 8: '8%', 12: '12%' }}
                />
              </div>

              <Button
                type="primary"
                size="large"
                onClick={handleCalculate}
                block
                icon={<RiseOutlined />}
              >
                计算定投收益
              </Button>
            </Card>
          </Col>

          <Col xs={24} lg={16}>
            {result ? (
              <>
                <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
                  <Col span={8}>
                    <Card>
                      <Statistic
                        title="累计投入"
                        value={result.totalInvested}
                        precision={2}
                        prefix="$"
                        valueStyle={{ fontSize: 24 }}
                      />
                    </Card>
                  </Col>
                  <Col span={8}>
                    <Card>
                      <Statistic
                        title="最终价值"
                        value={result.finalValue}
                        precision={2}
                        prefix="$"
                        valueStyle={{ color: '#3f8600', fontSize: 24 }}
                      />
                    </Card>
                  </Col>
                  <Col span={8}>
                    <Card>
                      <Statistic
                        title="总收益率"
                        value={result.totalGainPercent}
                        precision={2}
                        suffix="%"
                        valueStyle={{
                          color: result.totalGainPercent >= 0 ? '#3f8600' : '#cf1322',
                          fontSize: 24,
                        }}
                      />
                    </Card>
                  </Col>
                </Row>

                <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
                  <Col span={8}>
                    <Card size="small">
                      <Statistic
                        title="总收益"
                        value={result.totalGain}
                        precision={2}
                        prefix="$"
                        valueStyle={{ color: result.totalGain >= 0 ? '#3f8600' : '#cf1322' }}
                      />
                    </Card>
                  </Col>
                  <Col span={8}>
                    <Card size="small">
                      <Statistic
                        title="累计股息贡献"
                        value={result.dividendContribution}
                        precision={2}
                        prefix="$"
                      />
                    </Card>
                  </Col>
                  <Col span={8}>
                    <Card size="small">
                      <Statistic
                        title="年化收益率 (CAGR)"
                        value={result.annualizedReturn}
                        precision={2}
                        suffix="%"
                      />
                    </Card>
                  </Col>
                </Row>

                <Card title="资产增长曲线" style={{ marginBottom: 24 }}>
                  <ReactEChart option={chartOption} height={350} />
                </Card>

                <Card title="逐年明细">
                  <Table
                    dataSource={result.yearlyData}
                    columns={columns}
                    rowKey="year"
                    pagination={false}
                    size="small"
                    scroll={{ x: 800 }}
                  />
                </Card>
              </>
            ) : (
              <Card>
                <div style={{ textAlign: 'center', padding: '60px 20px' }}>
                  <DollarOutlined style={{ fontSize: 64, color: '#1890ff', marginBottom: 16 }} />
                  <Title level={4}>设置定投参数开始计算</Title>
                  <Paragraph type="secondary">
                    输入每月投资金额、选择 ETF 和投资年限，<br />
                    系统将自动计算复利增长效果。
                  </Paragraph>
                  <Paragraph type="secondary" style={{ fontSize: 12 }}>
                    示例：每月定投 $500 到 SPY，30 年后投入 $180,000，<br />
                    按历史年化 10% 计算，最终价值约 $1,130,000。
                  </Paragraph>
                </div>
              </Card>
            )}
          </Col>
        </Row>
      </div>
    </Layout>
  );
};

export default DCACalculator;

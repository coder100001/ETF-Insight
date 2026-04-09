import { useState, useMemo, useEffect } from 'react';
import styled from 'styled-components';
import { Table, App, Row, Col, Spin } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { ArrowLeftOutlined, InfoCircleOutlined } from '@ant-design/icons';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, Legend, ResponsiveContainer, ReferenceLine } from 'recharts';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import { etfAPI } from '../services/api';
import type { ETFData } from '../types';

const PageHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px 0;
  border-bottom: 1px solid ${theme.colors.border};
  h2 {
    margin: 0;
    font-size: ${theme.fonts.size['2xl']};
    color: ${theme.colors.textPrimary};
    display: flex;
    align-items: center;
    gap: 12px;
  }
`;

const ChartContainer = styled.div`
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.lg};
  padding: 24px;
  margin-bottom: 20px;
  box-shadow: ${theme.shadows.card};
  .chart-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    h3 { margin: 0; font-size: ${theme.fonts.size.lg}; color: ${theme.colors.textPrimary}; }
    .period-tabs { display: flex; gap: 8px; }
    button {
      padding: 6px 16px; border: 1px solid ${theme.colors.border}; border-radius: 20px;
      background: transparent; cursor: pointer; transition: all 0.3s; font-size: ${theme.fonts.size.sm};
      &:hover, &.active { background: ${theme.colors.primary}; color: white; border-color: ${theme.colors.primary}; }
    }
  }
`;

const SectionCard = styled.div`
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.lg};
  padding: 20px;
  margin-bottom: 16px;
  box-shadow: ${theme.shadows.card};
  .section-header {
    display: flex; justify-content: space-between; align-items: center;
    margin-bottom: 16px; padding-bottom: 12px; border-bottom: 1px solid ${theme.colors.border};
    h4 { margin: 0; font-size: ${theme.fonts.size.base}; font-weight: ${theme.fonts.weight.semibold}; color: ${theme.colors.textPrimary}; }
  }
`;

const StyledTable = styled(Table<ETFData>)`
  .ant-table-thead > tr > th { background: ${theme.colors.background}; font-weight: ${theme.fonts.weight.semibold}; color: ${theme.colors.textSecondary}; font-size: ${theme.fonts.size.sm}; border-bottom: 1px solid ${theme.colors.border}; }
  .ant-table-tbody > tr > td { border-bottom: 1px solid ${theme.colors.border}; }
  .ant-table-tbody > tr:hover > td { background: rgba(0,0,0,0.02); }
  .positive { color: ${theme.colors.success}; }
  .negative { color: ${theme.colors.danger}; }
`;

const ETFNameCell = styled.div`
  .symbol { font-weight: ${theme.fonts.weight.semibold}; color: ${theme.colors.textPrimary}; font-size: ${theme.fonts.size.base}; }
  .name { color: ${theme.colors.textMuted}; font-size: ${theme.fonts.size.xs}; margin-top: 2px; }
`;

const LoadingContainer = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
`;

const COLORS = ['#1890ff', '#faad14', '#13c2c2', '#f5222d', '#52c41a', '#722ed1'];

const PERIODS = [
  { key: '1m', label: '近1月' },
  { key: '3m', label: '近3月' },
  { key: '6m', label: '近6月' },
  { key: '1y', label: '近1年' },
];

const formatNumber = (num: number | null, decimals = 2): string => {
  if (num === null || isNaN(num)) return '--';
  if (Math.abs(num) >= 100000000) return `${(num / 100000000).toFixed(decimals)}亿`;
  if (Math.abs(num) >= 10000) return `${(num / 10000).toFixed(decimals)}万`;
  return num.toFixed(decimals);
};

const formatPercent = (value: number | null, showSign = true): string => {
  if (value === null || isNaN(value)) return '--';
  const prefix = showSign && value > 0 ? '+' : '';
  return `${prefix}${value.toFixed(2)}%`;
};

const getChangeClass = (value: number | null): string => {
  if (value === null || value === 0) return '';
  return value > 0 ? 'positive' : 'negative';
};

const ETFComparisonReport: React.FC = () => {
  const { message } = App.useApp();
  const [selectedPeriod, setSelectedPeriod] = useState('1y');
  const [etfData, setEtfData] = useState<ETFData[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchETFData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const fetchETFData = async () => {
    setLoading(true);
    try {
      const response = await etfAPI.getList();
      if (response.success && response.data) {
        setEtfData(response.data);
      } else {
        message.error('获取ETF数据失败');
      }
    } catch (error) {
      message.error('获取ETF数据失败: ' + (error as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const chartData = useMemo(() => {
    if (etfData.length === 0) return [];
    const days = 30;
    const dates: Record<string, string | number>[] = [];
    for (let i = 0; i < days; i++) {
      const date = new Date();
      date.setDate(date.getDate() - (days - i));
      dates.push({ date: date.toISOString().split('T')[0] });
    }
    etfData.forEach((etf) => {
      for (let i = 0; i < days; i++) {
        const progress = i / days;
        const simulatedChange = (etf.change_percent * progress) + (Math.random() - 0.5) * 2;
        dates[i][etf.symbol] = parseFloat(simulatedChange.toFixed(2));
      }
    });
    return dates;
  }, [etfData]);

  const renderNameCell = (record: ETFData) => (
    <ETFNameCell>
      <div className="symbol">{record.symbol}</div>
      <div className="name">{record.name}</div>
    </ETFNameCell>
  );

  const basicColumns: ColumnsType<ETFData> = [
    { title: 'ETF', key: 'name', width: 200, render: (_, r) => renderNameCell(r) },
    { title: '类别', dataIndex: 'category', key: 'category', align: 'center' },
    { title: '提供商', dataIndex: 'provider', key: 'provider', align: 'center' },
    { title: '当前价格', dataIndex: 'current_price', key: 'price', align: 'right', render: (v) => `$${Number(v)?.toFixed(2)}` },
    { title: '涨跌幅', dataIndex: 'change_percent', key: 'change', align: 'right', render: (v) => <span className={getChangeClass(Number(v))}>{formatPercent(Number(v))}</span> },
    { title: '成交量', dataIndex: 'volume', key: 'volume', align: 'right', render: (v) => formatNumber(Number(v)) },
  ];

  const metricColumns: ColumnsType<ETFData> = [
    { title: 'ETF', key: 'name', width: 200, render: (_, r) => renderNameCell(r) },
    { title: '年化波动率', dataIndex: 'volatility', key: 'volatility', align: 'right', render: (v) => `${Number(v)?.toFixed(2)}%` },
    { title: '夏普比率', dataIndex: 'sharpe_ratio', key: 'sharpe', align: 'right', render: (v) => Number(v)?.toFixed(2) },
    { title: '最大回撤', dataIndex: 'max_drawdown', key: 'maxDrawdown', align: 'right', render: (v) => <span className="negative">{Number(v)?.toFixed(2)}%</span> },
    { title: '年度收益', dataIndex: 'total_return', key: 'totalReturn', align: 'right', render: (v) => <span className={getChangeClass(Number(v))}>{formatPercent(Number(v))}</span> },
    { title: '股息率', dataIndex: 'dividend_yield', key: 'dividend', align: 'right', render: (v) => `${Number(v)?.toFixed(2)}%` },
    { title: '费率', dataIndex: 'expense_ratio', key: 'expense', align: 'right', render: (v) => `${Number(v)?.toFixed(2)}%` },
  ];

  if (loading) {
    return (
      <Layout>
        <PageHeader>
          <h2><ArrowLeftOutlined style={{ cursor: 'pointer' }} onClick={() => window.history.back()} />ETF对比</h2>
        </PageHeader>
        <LoadingContainer>
          <Spin size="large" tip="加载中..." />
        </LoadingContainer>
      </Layout>
    );
  }

  return (
    <Layout>
      <PageHeader>
        <h2><ArrowLeftOutlined style={{ cursor: 'pointer' }} onClick={() => window.history.back()} />ETF对比</h2>
      </PageHeader>

      <ChartContainer>
        <div className="chart-header">
          <h3>价格走势对比</h3>
          <div className="period-tabs">
            {PERIODS.map(p => (
              <button key={p.key} className={selectedPeriod === p.key ? 'active' : ''} onClick={() => setSelectedPeriod(p.key)}>{p.label}</button>
            ))}
          </div>
        </div>
        <ResponsiveContainer width="100%" height={350}>
          <LineChart data={chartData} margin={{ top: 5, right: 30, left: 0, bottom: 5 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#eee" />
            <XAxis dataKey="date" tick={{ fontSize: 11, fill: '#888' }} interval={Math.ceil(chartData.length / 6)} />
            <YAxis tick={{ fontSize: 11, fill: '#888' }} tickFormatter={(v: number) => `${v}%`} domain={['auto', 'auto']} />
            <RechartsTooltip 
              contentStyle={{ backgroundColor: '#fff', border: '1px solid #ddd', borderRadius: 8 }} 
              labelStyle={{ color: '#333' }} 
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              formatter={(value: any) => [`${Number(value).toFixed(2)}%`, '']} 
            />
            <Legend wrapperStyle={{ paddingTop: 20 }} />
            <ReferenceLine y={0} stroke="#666" strokeDasharray="3 3" />
            {etfData.map((etf, idx) => (
              <Line 
                key={String(etf.symbol)} 
                type="monotone" 
                dataKey={String(etf.symbol)} 
                name={String(etf.symbol)} 
                stroke={COLORS[idx % COLORS.length]} 
                strokeWidth={2} 
                dot={false} 
                activeDot={{ r: 4 }} 
              />
            ))}
          </LineChart>
        </ResponsiveContainer>
      </ChartContainer>

      <Row gutter={[16, 16]}>
        <Col span={24}>
          <SectionCard>
            <div className="section-header"><h4>基本信息</h4><InfoCircleOutlined style={{ color: theme.colors.textMuted }} /></div>
            <StyledTable dataSource={etfData} columns={basicColumns} rowKey="symbol" pagination={false} size="middle" />
          </SectionCard>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col span={24}>
          <SectionCard>
            <div className="section-header"><h4>关键指标</h4></div>
            <StyledTable dataSource={etfData} columns={metricColumns} rowKey="symbol" pagination={false} size="middle" />
          </SectionCard>
        </Col>
      </Row>
    </Layout>
  );
};

export default ETFComparisonReport;

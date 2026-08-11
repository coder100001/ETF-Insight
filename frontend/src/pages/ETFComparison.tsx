import { useState, useEffect, useMemo } from 'react';
import styled from 'styled-components';
import { Card, Table, Button, Select, App, Row, Col, Tag, Typography } from 'antd';
import { BarChartOutlined } from '@ant-design/icons';
import { FaBalanceScale, FaChartLine } from 'react-icons/fa';
import ReactEChart from '../components/ReactEChart';
import type { EChartsOption } from 'echarts';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import { etfAPI, request } from '../services/api';
import type { ETFData } from '../types';

const { Text } = Typography;

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
  flex-wrap: wrap;
  align-items: center;
`;

const MetricCard = styled.div<{ $color?: string }>`
  background: ${({ $color }) => $color || 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'};
  color: white;
  padding: 16px;
  border-radius: 12px;
  text-align: center;
  height: 100%;
`;

const MetricValue = styled.div`
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 4px;
`;

const MetricLabel = styled.div`
  font-size: 12px;
  opacity: 0.9;
`;

const SimilarCard = styled(Card)`
  cursor: pointer;
  transition: all 0.2s;
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  }
`;

const COLORS = ['#5470c6', '#91cc75', '#fac858', '#ee6666', '#73c0de', '#3ba272', '#fc8452', '#9a60b4'];

interface SimilarETF {
  symbol: string;
  name: string;
  score: number;
  strategy: string;
  focus: string;
  category: string;
  expense_ratio: number;
  correlation: number;
}

const ETFComparison: React.FC = () => {
  const { message } = App.useApp();
  const [etfs, setEtfs] = useState<ETFData[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedETFs, setSelectedETFs] = useState<string[]>(['SCHD', 'SPYD', 'JEPQ']);
  const [similarETFs, setSimilarETFs] = useState<SimilarETF[]>([]);
  const [similarLoading, setSimilarLoading] = useState(false);

  useEffect(() => {
    fetchETFData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const fetchETFData = async () => {
    setLoading(true);
    try {
      const response = await etfAPI.getList();
      if (response.success && response.data) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const formattedData: ETFData[] = response.data.map((item: any) => ({
          symbol: item.symbol,
          name: item.name,
          current_price: item.current_price || 0,
          previous_close: item.previous_close || 0,
          change: item.change || 0,
          change_percent: item.change_percent || 0,
          open_price: item.open_price || 0,
          high_price: item.high_price || 0,
          low_price: item.low_price || 0,
          volume: item.volume || 0,
          dividend_yield: item.dividend_yield || 0,
          volatility: item.volatility || 0,
          total_return: item.total_return || 0,
          max_drawdown: item.max_drawdown || 0,
          sharpe_ratio: item.sharpe_ratio || 0,
          expense_ratio: item.expense_ratio || 0,
          info: {
            focus: item.focus || '',
            strategy: item.strategy || '',
          },
        }));
        setEtfs(formattedData);
      }
    } catch (error: unknown) {
      message.error('获取ETF数据失败: ' + ((error instanceof Error ? error.message : '') || ''));
    } finally {
      setLoading(false);
    }
  };

  // 当选择的 ETF 改变时，获取相似 ETF 推荐（基于第一个选择的 ETF）
  useEffect(() => {
    if (selectedETFs.length > 0) {
      const primarySymbol = selectedETFs[0];
          fetchSimilarETFs(primarySymbol);
    }
  }, [selectedETFs]);

  const fetchSimilarETFs = async (symbol: string) => {
    setSimilarLoading(true);
    try {
      const response = await request<{ success: boolean; data?: SimilarETF[] }>(
        `/etf/${symbol}/similar?limit=5`
      );
      if (response.success && response.data) {
        setSimilarETFs(response.data);
      }
    } catch {
      // 静默失败
    } finally {
      setSimilarLoading(false);
    }
  };

  const filteredETFs = etfs.filter(etf => selectedETFs.includes(etf.symbol));

  // 雷达图数据
  const radarData = useMemo(() => {
    if (filteredETFs.length === 0) return [];
    const metrics = ['sharpe_ratio', 'total_return', 'dividend_yield', 'volatility', 'expense_ratio'];
    const maxValues: Record<string, number> = {
      sharpe_ratio: Math.max(...filteredETFs.map(e => e.sharpe_ratio || 0), 3),
      total_return: Math.max(...filteredETFs.map(e => Math.abs(e.total_return || 0)), 30),
      dividend_yield: Math.max(...filteredETFs.map(e => e.dividend_yield || 0), 10),
      volatility: Math.max(...filteredETFs.map(e => e.volatility || 0), 30),
      expense_ratio: Math.max(...filteredETFs.map(e => e.expense_ratio || 0), 2),
    };

    return metrics.map(metric => {
      const data: Record<string, number | string> = { metric };
      const labels: Record<string, string> = {
        sharpe_ratio: '夏普比率',
        total_return: '总收益',
        dividend_yield: '股息率',
        volatility: '波动率',
        expense_ratio: '费率',
      };
      data.label = labels[metric] || metric;
      filteredETFs.forEach(etf => {
        const val = etf[metric as keyof ETFData] as number || 0;
        // 对于费率和波动率，数值越小越好，取倒数
        if (metric === 'expense_ratio' || metric === 'volatility') {
          data[etf.symbol] = maxValues[metric] > 0 ? (1 - val / maxValues[metric]) * 100 : 0;
        } else {
          data[etf.symbol] = maxValues[metric] > 0 ? (val / maxValues[metric]) * 100 : 0;
        }
      });
      return data;
    });
  }, [filteredETFs]);

  // 柱状图数据 - 原始值对比
  const barChartData = useMemo(() => {
    if (filteredETFs.length === 0) return [];
    return [
      { metric: 'total_return', label: '总收益(%)', data: filteredETFs.map(e => ({ symbol: e.symbol, value: e.total_return || 0 })) },
      { metric: 'dividend_yield', label: '股息率(%)', data: filteredETFs.map(e => ({ symbol: e.symbol, value: e.dividend_yield || 0 })) },
      { metric: 'volatility', label: '波动率(%)', data: filteredETFs.map(e => ({ symbol: e.symbol, value: e.volatility || 0 })) },
      { metric: 'sharpe_ratio', label: '夏普比率', data: filteredETFs.map(e => ({ symbol: e.symbol, value: e.sharpe_ratio || 0 })) },
      { metric: 'expense_ratio', label: '费率(%)', data: filteredETFs.map(e => ({ symbol: e.symbol, value: e.expense_ratio || 0 })) },
    ];
  }, [filteredETFs]);

  const addSimilarETF = (symbol: string) => {
    if (!selectedETFs.includes(symbol)) {
      setSelectedETFs([...selectedETFs, symbol]);
    }
  };

  // 找出每个指标的最佳值
  const getBestValue = (metric: keyof ETFData, higherBetter: boolean): ETFData | null => {
    if (filteredETFs.length === 0) return null;
    return filteredETFs.reduce((best, curr) => {
      const bestVal = best[metric] as number || 0;
      const currVal = curr[metric] as number || 0;
      return higherBetter ? (currVal > bestVal ? curr : best) : (currVal < bestVal ? curr : best);
    }, filteredETFs[0]);
  };

  const metricLabels: Record<string, { label: string; higherBetter: boolean }> = {
    sharpe_ratio: { label: '夏普比率', higherBetter: true },
    total_return: { label: '总收益', higherBetter: true },
    dividend_yield: { label: '股息率', higherBetter: true },
    volatility: { label: '波动率', higherBetter: false },
    expense_ratio: { label: '费率', higherBetter: false },
  };

  const columns: import('antd/es/table').ColumnsType<ETFData> = [
    { title: 'ETF', dataIndex: 'symbol', key: 'symbol', fixed: 'left' as const,
      render: (text: string, record: ETFData) => (
        <div>
          <strong>{text}</strong>
          <br />
          <small style={{ color: theme.colors.textMuted }}>{record.name}</small>
        </div>
      ),
    },
    { title: '当前价格', dataIndex: 'current_price', key: 'current_price', align: 'center' as const,
      render: (v: number) => v ? `$${v.toFixed(2)}` : '-',
    },
    { title: '涨跌幅', dataIndex: 'change_percent', key: 'change_percent', align: 'center' as const,
      render: (v: number) => (
        <span style={{ color: v >= 0 ? theme.colors.success : theme.colors.danger }}>
          {v >= 0 ? '+' : ''}{v.toFixed(2)}%
        </span>
      ),
    },
    { title: '股息率', dataIndex: 'dividend_yield', key: 'dividend_yield', align: 'center' as const,
      render: (v: number) => highlightValue('dividend_yield', v),
    },
    { title: '总收益', dataIndex: 'total_return', key: 'total_return', align: 'center' as const,
      render: (v: number) => highlightValue('total_return', v),
    },
    { title: '波动率', dataIndex: 'volatility', key: 'volatility', align: 'center' as const,
      render: (v: number) => highlightValue('volatility', v),
    },
    { title: '夏普比率', dataIndex: 'sharpe_ratio', key: 'sharpe_ratio', align: 'center' as const,
      render: (v: number) => highlightValue('sharpe_ratio', v),
    },
    { title: '最大回撤', dataIndex: 'max_drawdown', key: 'max_drawdown', align: 'center' as const,
      render: (v: number) => (
        <span style={{ color: theme.colors.danger }}>{v.toFixed(2)}%</span>
      ),
    },
    { title: '费率', dataIndex: 'expense_ratio', key: 'expense_ratio', align: 'center' as const,
      render: (v: number) => highlightValue('expense_ratio', v),
    },
    { title: '策略', dataIndex: ['info', 'strategy'], key: 'strategy', align: 'center' as const, render: (v: string) => v ? <Tag>{v}</Tag> : '-', },
  ];

  const highlightValue = (metric: string, value: number) => {
    const cfg = metricLabels[metric];
    if (!cfg) return `${value.toFixed(2)}%`;
    const best = getBestValue(metric as keyof ETFData, cfg.higherBetter);
    if (best && value !== undefined) {
      const isBest = best[metric as keyof ETFData] === value;
      const formatted = metric === 'sharpe_ratio' ? value.toFixed(2) : `${value.toFixed(2)}%`;
      return isBest
        ? <Tag color="green" style={{ fontWeight: 'bold' }}>{formatted} ✓</Tag>
        : <span>{formatted}</span>;
    }
    return `${value.toFixed(2)}%`;
  };

  return (
    <Layout>
      <div style={{ padding: '24px' }}>
        <PageHeader>
          <h2><FaBalanceScale /> ETF对比分析</h2>
        </PageHeader>

        <FilterSection>
          <Select
            mode="multiple"
            placeholder="选择要对比的ETF"
            value={selectedETFs}
            onChange={setSelectedETFs}
            style={{ minWidth: 350 }}
            maxTagCount={5}
            options={etfs.map(etf => ({
              label: `${etf.symbol} - ${etf.name}`,
              value: etf.symbol,
            }))}
          />
          <Button type="primary" icon={<BarChartOutlined />}>
            开始对比
          </Button>
          {selectedETFs.length > 2 && (
            <Tag color="blue" style={{ marginLeft: 8 }}>
              已选择 {selectedETFs.length} 个 ETF
            </Tag>
          )}
        </FilterSection>

        {filteredETFs.length > 0 && (
          <>
            {/* 核心指标一览 */}
            <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
              {filteredETFs.slice(0, 4).map((etf, idx) => (
                <Col xs={12} sm={6} key={etf.symbol}>
                  <MetricCard $color={idx < COLORS.length ? `linear-gradient(135deg, ${COLORS[idx]} 0%, ${COLORS[idx]}dd 100%)` : undefined}>
                    <MetricLabel>{etf.symbol}</MetricLabel>
                    <MetricValue>${etf.current_price.toFixed(2)}</MetricValue>
                    <div style={{ fontSize: 12, marginTop: 4 }}>
                      收益: {etf.total_return != null ? `${etf.total_return >= 0 ? '+' : ''}${etf.total_return.toFixed(1)}%` : '-'}
                      {' | '}夏普: {etf.sharpe_ratio?.toFixed(2) ?? '-'}
                    </div>
                  </MetricCard>
                </Col>
              ))}
            </Row>

            <Row gutter={24}>
              {/* 雷达图 */}
              <Col xs={24} lg={12}>
                <Card title="多维度表现对比 (分数越高越好)" style={{ marginBottom: 24 }}>
                  <ReactEChart
                    height={320}
                    option={{
                      tooltip: {},
                      legend: { data: filteredETFs.map(e => e.symbol) },
                      radar: {
                        indicator: radarData.map(d => ({ name: d.label as string, max: 100 })),
                      },
                      series: [{
                        type: 'radar',
                        data: filteredETFs.map((etf, idx) => ({
                          value: radarData.map(d => d[etf.symbol] as number),
                          name: etf.symbol,
                          lineStyle: { color: COLORS[idx % COLORS.length] },
                          areaStyle: { color: COLORS[idx % COLORS.length], opacity: 0.15 },
                          itemStyle: { color: COLORS[idx % COLORS.length] },
                        })),
                      }],
                    }}
                  />
                </Card>
              </Col>

              {/* 收益/股息 / 波动率柱状图 */}
              <Col xs={24} lg={12}>
                <Card title="关键指标对比" style={{ marginBottom: 24 }}>
                  {barChartData.slice(0, 2).map(item => {
                    const option: EChartsOption = {
                      tooltip: {
                        formatter: (params) => {
                          const p = Array.isArray(params) ? params[0] : params;
                          return `${p.name}: ${Number(p.value ?? 0).toFixed(2)}`;
                        },
                      },
                      grid: { containLabel: true, left: '3%', right: '4%', top: '10%', bottom: '10%' },
                      xAxis: {
                        type: 'value',
                        axisLabel: { fontSize: 11 },
                        splitLine: { lineStyle: { type: 'dashed' } },
                      },
                      yAxis: {
                        type: 'category',
                        data: item.data.map(d => d.symbol),
                        axisLabel: { fontSize: 11 },
                        splitLine: { show: false },
                      },
                      series: [{
                        type: 'bar',
                        data: item.data.map((d, idx) => ({
                          value: d.value,
                          itemStyle: { color: COLORS[idx % COLORS.length] },
                        })),
                        barMaxWidth: 20,
                        itemStyle: { borderRadius: [0, 4, 4, 0] },
                      }],
                    };
                    return (
                      <div key={item.metric} style={{ marginBottom: 16 }}>
                        <Text strong style={{ display: 'block', marginBottom: 8 }}>{item.label}</Text>
                        <ReactEChart option={option} height={100} />
                      </div>
                    );
                  })}
                </Card>
              </Col>
            </Row>

            {/* 详情表格 */}
            <Card title="详细对比数据" style={{ marginBottom: 24 }}>
              <Table
                dataSource={filteredETFs}
                columns={columns}
                rowKey="symbol"
                pagination={false}
                bordered
                loading={loading}
                scroll={{ x: 1000 }}
              />
            </Card>
          </>
        )}

        {/* 相似 ETF 推荐 */}
        {selectedETFs.length > 0 && (
          <Card title={
            <span><FaChartLine style={{ marginRight: 8 }} />基于 {selectedETFs[0]} 的相似 ETF 推荐</span>
          } loading={similarLoading}>
            {similarETFs.length > 0 ? (
              <Row gutter={[16, 16]}>
                {similarETFs.map(etf => {
                  const corrLabel = etf.correlation > 0
                    ? `相关性 ${(etf.correlation).toFixed(1)}%`
                    : '相关性较低';
                  const scoreLabel = etf.score >= 60 ? '非常相似' : etf.score >= 40 ? '比较相似' : '部分相似';
                  return (
                    <Col xs={24} sm={12} md={8} key={etf.symbol}>
                      <SimilarCard
                        size="small"
                        hoverable
                        onClick={() => addSimilarETF(etf.symbol)}
                        title={
                          <span>
                            <strong>{etf.symbol}</strong>
                            <Tag style={{ marginLeft: 8 }} color="blue">{scoreLabel}</Tag>
                          </span>
                        }
                      >
                        <Text type="secondary" style={{ fontSize: 12 }}>{etf.name}</Text>
                        <div style={{ marginTop: 8, fontSize: 13 }}>
                          <Row gutter={[8, 4]}>
                            <Col span={12}><Text type="secondary">相似度:</Text> {etf.score.toFixed(0)}%</Col>
                            <Col span={12}><Text type="secondary">相关性:</Text> {corrLabel}</Col>
                            <Col span={12}><Text type="secondary">费率:</Text> {etf.expense_ratio.toFixed(2)}%</Col>
                            <Col span={12}><Text type="secondary">策略:</Text> {etf.strategy || '-'}</Col>
                          </Row>
                        </div>
                      </SimilarCard>
                    </Col>
                  );
                })}
              </Row>
            ) : (
              <Text type="secondary">暂无推荐的相似 ETF</Text>
            )}
          </Card>
        )}
      </div>
    </Layout>
  );
};

export default ETFComparison;

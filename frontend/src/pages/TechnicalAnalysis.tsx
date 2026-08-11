import React, { useState, useEffect, useMemo } from 'react';
import { Card, Row, Col, Select, DatePicker, Button, Spin, Alert, Statistic } from 'antd';
import ReactEChart from '../components/ReactEChart';
import type { EChartsOption } from 'echarts';
import { FundOutlined, LineChartOutlined, BarChartOutlined } from '@ant-design/icons';
import Layout from '../components/Layout';
import styled from 'styled-components';
import { theme } from '../styles/theme';
import { etfAPI } from '../services/api';
import dayjs from 'dayjs';

const { RangePicker } = DatePicker;
const { Option } = Select;

const PageContainer = styled.div`
  padding: ${theme.spacing.lg};
`;

const PageHeader = styled.div`
  margin-bottom: ${theme.spacing.lg};
`;

const PageTitle = styled.h1`
  font-size: ${theme.fonts.size['2xl']};
  font-weight: ${theme.fonts.weight.bold};
  color: ${theme.colors.textPrimary};
  margin-bottom: ${theme.spacing.sm};
`;

const FilterContainer = styled(Card)`
  margin-bottom: ${theme.spacing.lg};
`;

const ChartCard = styled(Card)`
  margin-bottom: ${theme.spacing.lg};
  .ant-card-head {
    border-bottom: 1px solid ${theme.colors.border};
  }
`;

const StatGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: ${theme.spacing.md};
  margin-bottom: ${theme.spacing.lg};
`;

const StatCard = styled.div`
  background: ${theme.colors.surface};
  padding: ${theme.spacing.md};
  border-radius: ${theme.borderRadius.md};
  border: 1px solid ${theme.colors.border};
`;

// 模拟数据类型
interface TechnicalIndicators {
  rsi: number;
  macd: number;
  macdSignal: number;
  bollingerUpper: number;
  bollingerMiddle: number;
  bollingerLower: number;
  sma20: number;
  sma50: number;
  ema12: number;
  ema26: number;
}

interface RadarData {
  item: string;
  value: number;
  symbol: string;
}

const TechnicalAnalysis: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [selectedETFs, setSelectedETFs] = useState<string[]>(['SPY', 'QQQ']);
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([
    dayjs().subtract(3, 'months'),
    dayjs()
  ]);
  const [radarData, setRadarData] = useState<RadarData[]>([]);
  const [indicators, setIndicators] = useState<Record<string, TechnicalIndicators>>({});

  // 获取技术指标数据
  const fetchTechnicalData = async () => {
    setLoading(true);
    try {
      const period = dateRange ? `${dateRange[0].format('YYYY-MM-DD')}_${dateRange[1].format('YYYY-MM-DD')}` : '1y';

      // 并行获取所有 ETF 的指标数据
      const results = await Promise.allSettled(
        selectedETFs.map(symbol => etfAPI.getMetrics(symbol, period))
      );

      const mockIndicators: Record<string, TechnicalIndicators> = {};

      results.forEach((result, index) => {
        const symbol = selectedETFs[index];
        if (result.status === 'fulfilled' && result.value?.success && result.value.data) {
          const metrics = result.value.data;
          mockIndicators[symbol] = {
            rsi: metrics.rsi || 50,
            macd: metrics.macd || 0,
            macdSignal: metrics.macd_signal || 0,
            bollingerUpper: metrics.bollinger_upper || 0,
            bollingerMiddle: metrics.bollinger_middle || 0,
            bollingerLower: metrics.bollinger_lower || 0,
            sma20: metrics.sma_20 || 0,
            sma50: metrics.sma_50 || 0,
            ema12: metrics.ema_12 || 0,
            ema26: metrics.ema_26 || 0,
          };
        } else {
          // 降级到模拟数据
          mockIndicators[symbol] = {
            rsi: 50 + Math.random() * 30,
            macd: Math.random() * 2 - 1,
            macdSignal: Math.random() * 1.5 - 0.75,
            bollingerUpper: 400 + Math.random() * 100,
            bollingerMiddle: 390 + Math.random() * 100,
            bollingerLower: 380 + Math.random() * 100,
            sma20: 395 + Math.random() * 100,
            sma50: 390 + Math.random() * 100,
            ema12: 398 + Math.random() * 100,
            ema26: 392 + Math.random() * 100,
          };
        }
      });

      setIndicators(mockIndicators);

      // 生成雷达图数据
      const radarChartData: RadarData[] = [];
      selectedETFs.forEach(symbol => {
        const data = mockIndicators[symbol];
        if (data) {
          radarChartData.push(
            { item: 'RSI', value: data.rsi, symbol },
            { item: 'MACD', value: Math.min(Math.max((data.macd + 2) * 25, 0), 100), symbol },
            { item: '布林带位置', value: data.bollingerUpper > data.bollingerLower ? ((data.bollingerMiddle - data.bollingerLower) / (data.bollingerUpper - data.bollingerLower)) * 100 : 50, symbol },
            { item: 'SMA趋势', value: data.sma20 > data.sma50 ? 70 : 30, symbol },
            { item: 'EMA趋势', value: data.ema12 > data.ema26 ? 75 : 25, symbol },
          );
        }
      });
      setRadarData(radarChartData);
    } catch (error) {
      console.error('Failed to fetch technical data:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTechnicalData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedETFs, dateRange]);

  // 雷达图配置（ECharts）
  // 将 @ant-design/charts 的扁平数据 [{ item, value, symbol }] 转换为 ECharts radar 所需的
  // indicator + 按 symbol 分组的 series 数据
  const radarOption: EChartsOption = useMemo(() => {
    const items = Array.from(new Set(radarData.map(d => d.item)));
    const symbols = Array.from(new Set(radarData.map(d => d.symbol)));

    // 每个维度固定 0-100（对应原 meta.value.min/max）
    const indicator = items.map(item => ({ name: item, min: 0, max: 100 }));

    const seriesData = symbols.map(symbol => ({
      name: symbol,
      // 按 indicator 顺序提取对应数值，缺失时补 0
      value: items.map(item => {
        const point = radarData.find(r => r.item === item && r.symbol === symbol);
        return point ? point.value : 0;
      }),
    }));

    return {
      tooltip: {},
      legend: {
        data: symbols,
        bottom: 0,
      },
      radar: {
        indicator,
        // 对应原 xAxis/yAxis 的 line/tickLine 关闭 + 实线网格
        axisLine: { lineStyle: { color: 'rgba(0, 0, 0, 0.2)' } },
        splitLine: { lineStyle: { color: 'rgba(0, 0, 0, 0.2)' } },
        splitArea: { show: false },
      },
      series: [
        {
          type: 'radar',
          data: seriesData,
          // 对应原 area.style.fillOpacity 与 point.size
          areaStyle: { opacity: 0.2 },
          symbolSize: 4,
        },
      ],
    };
  }, [radarData]);

  // 生成 MACD 图表数据
  const generateMACDData = () => {
    const data: Array<{ date: string; value: number; type: string }> = [];
    const dates = Array.from({ length: 30 }, (_, i) =>
      dayjs().subtract(29 - i, 'days').format('MM-DD')
    );

    selectedETFs.forEach(symbol => {
      const indicator = indicators[symbol];
      if (indicator) {
        dates.forEach((date, index) => {
          const baseValue = indicator.macd;
          data.push({
            date,
            value: baseValue + Math.sin(index * 0.3) * 0.5 + (Math.random() - 0.5) * 0.3,
            type: `${symbol} MACD`,
          });
          data.push({
            date,
            value: indicator.macdSignal + Math.sin(index * 0.3) * 0.4 + (Math.random() - 0.5) * 0.2,
            type: `${symbol} Signal`,
          });
        });
      }
    });

    return data;
  };

  // MACD 图表配置（ECharts）
  // 将 @ant-design/charts 的扁平数据 [{ date, value, type }] 转换为 ECharts line 所需的
  // xAxis 类目 + 按 type 分组的 series
  const macdOption: EChartsOption = useMemo(() => {
    const macdData = generateMACDData();
    const dates = Array.from(new Set(macdData.map(d => d.date)));
    const types = Array.from(new Set(macdData.map(d => d.type)));

    const series = types.map(type => ({
      name: type,
      type: 'line' as const,
      smooth: true,
      // 按 xAxis 顺序提取数值，缺失时补 null
      data: dates.map(date => {
        const point = macdData.find(d => d.date === date && d.type === type);
        return point ? point.value : null;
      }),
    }));

    return {
      tooltip: { trigger: 'axis' },
      legend: {
        data: types,
        top: 0,
      },
      xAxis: {
        type: 'category',
        data: dates,
        boundaryGap: false,
      },
      yAxis: {
        type: 'value',
      },
      series,
    };
    // generateMACDData 内部依赖 indicators 与 selectedETFs
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [indicators, selectedETFs]);

  return (
    <Layout>
      <PageContainer>
        <PageHeader>
          <PageTitle>
            <LineChartOutlined style={{ marginRight: theme.spacing.sm }} />
            技术分析
          </PageTitle>
        </PageHeader>

      <FilterContainer>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={12} md={8}>
            <div style={{ marginBottom: theme.spacing.xs }}>选择ETF</div>
            <Select
              mode="multiple"
              style={{ width: '100%' }}
              placeholder="选择要分析的ETF"
              value={selectedETFs}
              onChange={setSelectedETFs}
              maxTagCount={3}
            >
              <Option value="SPY">SPY (S&P 500)</Option>
              <Option value="QQQ">QQQ (Nasdaq 100)</Option>
              <Option value="IWM">IWM (Russell 2000)</Option>
              <Option value="VTI">VTI (Total Market)</Option>
              <Option value="VEA">VEA (Developed Markets)</Option>
              <Option value="VWO">VWO (Emerging Markets)</Option>
            </Select>
          </Col>
          <Col xs={24} sm={12} md={8}>
            <div style={{ marginBottom: theme.spacing.xs }}>时间范围</div>
            <RangePicker
              style={{ width: '100%' }}
              value={dateRange}
              onChange={(dates) => {
                if (dates && dates[0] && dates[1]) {
                  setDateRange([dates[0], dates[1]]);
                }
              }}
            />
          </Col>
          <Col xs={24} sm={24} md={8}>
            <Button
              type="primary"
              onClick={fetchTechnicalData}
              loading={loading}
              style={{ marginTop: 24 }}
              block
            >
              更新分析
            </Button>
          </Col>
        </Row>
      </FilterContainer>

      {loading ? (
        <div style={{ textAlign: 'center', padding: theme.spacing.xl }}>
          <Spin size="large" />
          <p>正在计算技术指标...</p>
        </div>
      ) : (
        <>
          {/* 指标概览 */}
          <StatGrid>
            {selectedETFs.map(symbol => {
              const indicator = indicators[symbol];
              if (!indicator) return null;

              return (
                <StatCard key={symbol}>
                  <div style={{ fontWeight: 'bold', marginBottom: theme.spacing.sm }}>
                    {symbol}
                  </div>
                  <Row gutter={[8, 8]}>
                    <Col span={12}>
                      <Statistic
                        title="RSI"
                        value={indicator.rsi}
                        precision={1}
                        styles={{
                          content: {
                            color: indicator.rsi > 70 ? '#cf1322' : indicator.rsi < 30 ? '#3f8600' : '#666'
                          }
                        }}
                      />
                    </Col>
                    <Col span={12}>
                      <Statistic
                        title="MACD"
                        value={indicator.macd}
                        precision={2}
                        styles={{ content: { color: indicator.macd > 0 ? '#3f8600' : '#cf1322' } }}
                      />
                    </Col>
                  </Row>
                </StatCard>
              );
            })}
          </StatGrid>

          {/* 雷达图 - 多因子对比 */}
          <Row gutter={[16, 16]}>
            <Col xs={24} lg={12}>
              <ChartCard
                title={
                  <span>
                    <FundOutlined style={{ marginRight: 8 }} />
                    多因子雷达图
                  </span>
                }
              >
                {radarData.length > 0 ? (
                  <ReactEChart option={radarOption} height={300} />
                ) : (
                  <Alert title="请选择ETF以查看雷达图" type="info" />
                )}
              </ChartCard>
            </Col>
            <Col xs={24} lg={12}>
              <ChartCard
                title={
                  <span>
                    <BarChartOutlined style={{ marginRight: 8 }} />
                    MACD 趋势
                  </span>
                }
              >
                <ReactEChart option={macdOption} height={300} />
              </ChartCard>
            </Col>
          </Row>

          {/* 布林带数据 */}
          <Row gutter={[16, 16]}>
            {selectedETFs.map(symbol => {
              const indicator = indicators[symbol];
              if (!indicator) return null;

              return (
                <Col xs={24} md={12} key={symbol}>
                  <ChartCard title={`${symbol} 布林带`}>
                    <Row gutter={[16, 16]}>
                      <Col span={8}>
                        <Statistic
                          title="上轨"
                          value={indicator.bollingerUpper}
                          precision={2}
                        />
                      </Col>
                      <Col span={8}>
                        <Statistic
                          title="中轨"
                          value={indicator.bollingerMiddle}
                          precision={2}
                        />
                      </Col>
                      <Col span={8}>
                        <Statistic
                          title="下轨"
                          value={indicator.bollingerLower}
                          precision={2}
                        />
                      </Col>
                    </Row>
                    <div style={{ marginTop: theme.spacing.md }}>
                      <Alert
                        title={
                          indicator.bollingerMiddle > indicator.bollingerLower &&
                          indicator.bollingerMiddle < indicator.bollingerUpper
                            ? '价格位于布林带中部，趋势稳定'
                            : indicator.bollingerMiddle >= indicator.bollingerUpper
                            ? '价格接近布林带上轨，可能超买'
                            : '价格接近布林带下轨，可能超卖'
                        }
                        type="info"
                        showIcon
                      />
                    </div>
                  </ChartCard>
                </Col>
              );
            })}
          </Row>
        </>
      )}
      </PageContainer>
    </Layout>
  );
};

export default TechnicalAnalysis;

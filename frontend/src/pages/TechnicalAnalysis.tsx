import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Select, DatePicker, Button, Spin, Alert, Statistic } from 'antd';
import { Radar, Line } from '@ant-design/charts';
import { FundOutlined, LineChartOutlined, BarChartOutlined } from '@ant-design/icons';
import Layout from '../components/Layout';
import styled from 'styled-components';
import { theme } from '../styles/theme';
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

  // 模拟获取技术指标数据
  const fetchTechnicalData = async () => {
    setLoading(true);
    try {
      // 这里应该调用实际 API
      // const response = await api.getTechnicalIndicators(selectedETFs, dateRange);

      // 模拟数据
      const mockIndicators: Record<string, TechnicalIndicators> = {
        'SPY': {
          rsi: 65.5,
          macd: 0.85,
          macdSignal: 0.72,
          bollingerUpper: 450.2,
          bollingerMiddle: 440.5,
          bollingerLower: 430.8,
          sma20: 441.2,
          sma50: 438.5,
          ema12: 442.1,
          ema26: 439.8,
        },
        'QQQ': {
          rsi: 72.3,
          macd: 1.25,
          macdSignal: 1.05,
          bollingerUpper: 380.5,
          bollingerMiddle: 370.2,
          bollingerLower: 359.9,
          sma20: 371.5,
          sma50: 368.2,
          ema12: 372.1,
          ema26: 369.5,
        },
      };

      setIndicators(mockIndicators);

      // 生成雷达图数据
      const radarChartData: RadarData[] = [];
      selectedETFs.forEach(symbol => {
        const data = mockIndicators[symbol];
        if (data) {
          // 归一化数据到 0-100 范围
          radarChartData.push(
            { item: 'RSI', value: data.rsi, symbol },
            { item: 'MACD', value: Math.min(Math.max((data.macd + 2) * 25, 0), 100), symbol },
            { item: '布林带位置', value: ((data.bollingerMiddle - data.bollingerLower) / (data.bollingerUpper - data.bollingerLower)) * 100, symbol },
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

  // 雷达图配置
  const radarConfig = {
    data: radarData,
    xField: 'item',
    yField: 'value',
    seriesField: 'symbol',
    meta: {
      value: {
        min: 0,
        max: 100,
      },
    },
    xAxis: {
      line: null,
      tickLine: null,
      grid: {
        line: {
          style: {
            lineDash: null,
          },
        },
      },
    },
    yAxis: {
      line: null,
      tickLine: null,
      grid: {
        line: {
          type: 'line',
          style: {
            lineDash: null,
          },
        },
      },
    },
    area: {
      style: {
        fillOpacity: 0.2,
      },
    },
    point: {
      size: 4,
    },
    legend: {
      position: 'bottom',
    },
  };

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

  const macdConfig = {
    data: generateMACDData(),
    xField: 'date',
    yField: 'value',
    seriesField: 'type',
    yAxis: {
      label: {
        formatter: (v: string) => `${v}`,
      },
    },
    legend: {
      position: 'top',
    },
    smooth: true,
    animation: {
      appear: {
        animation: 'path-in',
        duration: 1000,
      },
    },
  };

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
                  <Radar {...radarConfig} />
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
                <Line {...macdConfig} />
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

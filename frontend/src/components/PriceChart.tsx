import React, { useEffect, useRef } from 'react';
import * as echarts from 'echarts';
import styled from 'styled-components';
import { theme } from '../styles/theme';

const ChartContainer = styled.div`
  width: 100%;
  height: 400px;
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.lg};
  padding: ${theme.spacing.lg};
`;

interface PriceChartProps {
  data: {
    dates: string[];
    prices: number[];
    volumes?: number[];
  };
  symbol: string;
}

const PriceChart: React.FC<PriceChartProps> = ({ data, symbol }) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const chartInstance = useRef<echarts.ECharts | null>(null);

  useEffect(() => {
    if (!chartRef.current) return;

    // 初始化图表
    chartInstance.current = echarts.init(chartRef.current);

    const option: echarts.EChartsOption = {
      backgroundColor: 'transparent',
      title: {
        text: `${symbol} 价格走势`,
        left: 'center',
        textStyle: {
          color: theme.colors.textPrimary,
          fontSize: 16,
          fontWeight: 'bold',
        },
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'cross',
          label: {
            backgroundColor: theme.colors.primary,
          },
        },
        backgroundColor: theme.colors.surface,
        borderColor: theme.colors.border,
        textStyle: {
          color: theme.colors.textPrimary,
        },
      },
      legend: {
        data: ['价格', '成交量'],
        bottom: 0,
        textStyle: {
          color: theme.colors.textSecondary,
        },
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '15%',
        top: '15%',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        data: data.dates,
        boundaryGap: false,
        axisLine: {
          lineStyle: {
            color: theme.colors.border,
          },
        },
        axisLabel: {
          color: theme.colors.textSecondary,
        },
      },
      yAxis: [
        {
          type: 'value',
          name: '价格',
          position: 'left',
          axisLine: {
            show: true,
            lineStyle: {
              color: theme.colors.border,
            },
          },
          axisLabel: {
            color: theme.colors.textSecondary,
            formatter: '${value}',
          },
          splitLine: {
            lineStyle: {
              color: theme.colors.divider,
              type: 'dashed',
            },
          },
        },
        {
          type: 'value',
          name: '成交量',
          position: 'right',
          axisLine: {
            show: true,
            lineStyle: {
              color: theme.colors.border,
            },
          },
          axisLabel: {
            color: theme.colors.textSecondary,
          },
          splitLine: {
            show: false,
          },
        },
      ],
      dataZoom: [
        {
          type: 'inside',
          start: 0,
          end: 100,
        },
        {
          start: 0,
          end: 100,
          handleStyle: {
            color: theme.colors.primary,
          },
          textStyle: {
            color: theme.colors.textSecondary,
          },
        },
      ],
      series: [
        {
          name: '价格',
          type: 'line' as const,
          smooth: true,
          symbol: 'none',
          lineStyle: {
            width: 2,
            color: theme.colors.primary,
          },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: `${theme.colors.primary}40` },
              { offset: 1, color: `${theme.colors.primary}05` },
            ]),
          },
          data: data.prices,
        },
        ...(data.volumes
          ? [
              {
                name: '成交量',
                type: 'bar' as const,
                yAxisIndex: 1,
                itemStyle: {
                  color: `${theme.colors.textMuted}40`,
                },
                data: data.volumes,
              },
            ]
          : []),
      ],
    };

    chartInstance.current.setOption(option);

    // 响应式
    const handleResize = () => {
      chartInstance.current?.resize();
    };
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      chartInstance.current?.dispose();
    };
  }, [data, symbol]);

  return <ChartContainer ref={chartRef} />;
};

export default PriceChart;
import React, { useEffect, useRef } from 'react';
import * as echarts from 'echarts';
import styled from 'styled-components';
import { theme } from '../styles/theme';

const ChartContainer = styled.div`
  width: 100%;
  height: 350px;
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.lg};
  padding: ${theme.spacing.lg};
`;

interface SectorData {
  name: string;
  weight: number;
  value: number;
}

interface SectorBarChartProps {
  data: SectorData[];
  title?: string;
}

const SectorBarChart: React.FC<SectorBarChartProps> = ({ data, title = '行业分布' }) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const chartInstance = useRef<echarts.ECharts | null>(null);

  useEffect(() => {
    if (!chartRef.current || data.length === 0) return;

    chartInstance.current = echarts.init(chartRef.current);

    // 按权重排序
    const sortedData = [...data].sort((a, b) => b.weight - a.weight);

    const option: echarts.EChartsOption = {
      backgroundColor: 'transparent',
      title: {
        text: title,
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
          type: 'shadow',
        },
        backgroundColor: theme.colors.surface,
        borderColor: theme.colors.border,
        textStyle: {
          color: theme.colors.textPrimary,
        },
        formatter: (params: echarts.TooltipComponentFormatterCallbackParams) => {
          const dataIndex = Array.isArray(params) ? params[0].dataIndex : params.dataIndex;
          const item = sortedData[dataIndex as number];
          if (!item) return '';
          return `
            <div style="padding: 8px;">
              <strong>${item.name}</strong><br/>
              权重: ${item.weight.toFixed(2)}%<br/>
              市值: $${item.value.toLocaleString()}
            </div>
          `;
        },
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        top: '15%',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        data: sortedData.map(item => item.name),
        axisLabel: {
          color: theme.colors.textSecondary,
          rotate: 45,
          interval: 0,
        },
        axisLine: {
          lineStyle: {
            color: theme.colors.border,
          },
        },
      },
      yAxis: {
        type: 'value',
        name: '权重 (%)',
        nameTextStyle: {
          color: theme.colors.textSecondary,
        },
        axisLabel: {
          color: theme.colors.textSecondary,
          formatter: '{value}%',
        },
        axisLine: {
          lineStyle: {
            color: theme.colors.border,
          },
        },
        splitLine: {
          lineStyle: {
            color: theme.colors.divider,
            type: 'dashed',
          },
        },
      },
      series: [
        {
          name: '权重',
          type: 'bar',
          data: sortedData.map((item) => ({
            value: item.weight,
            itemStyle: {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: theme.colors.primary },
                { offset: 1, color: `${theme.colors.primary}60` },
              ]),
              borderRadius: [4, 4, 0, 0],
            },
          })),
          barWidth: '60%',
          label: {
            show: true,
            position: 'top',
            formatter: '{c}%',
            color: theme.colors.textSecondary,
            fontSize: 12,
          },
        },
      ],
    };

    chartInstance.current.setOption(option);

    const handleResize = () => {
      chartInstance.current?.resize();
    };
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      chartInstance.current?.dispose();
    };
  }, [data, title]);

  return <ChartContainer ref={chartRef} />;
};

export default SectorBarChart;

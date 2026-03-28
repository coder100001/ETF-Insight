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

interface HoldingData {
  symbol: string;
  name: string;
  weight: number;
  value: number;
}

interface HoldingPieChartProps {
  data: HoldingData[];
  title?: string;
}

const HoldingPieChart: React.FC<HoldingPieChartProps> = ({ data, title = '持仓分布' }) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const chartInstance = useRef<echarts.ECharts | null>(null);

  useEffect(() => {
    if (!chartRef.current || data.length === 0) return;

    chartInstance.current = echarts.init(chartRef.current);

    const colors = [
      theme.colors.primary,
      theme.colors.success,
      theme.colors.warning,
      theme.colors.danger,
      theme.colors.info,
      '#9C27B0',
      '#FF9800',
      '#00BCD4',
    ];

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
        trigger: 'item',
        backgroundColor: theme.colors.surface,
        borderColor: theme.colors.border,
        textStyle: {
          color: theme.colors.textPrimary,
        },
        formatter: (params: echarts.TooltipComponentFormatterCallbackParams) => {
          const dataIndex = Array.isArray(params) ? params[0].dataIndex : params.dataIndex;
          const item = data[dataIndex as number];
          if (!item) return '';
          return `
            <div style="padding: 8px;">
              <strong>${item.name}</strong><br/>
              代码: ${item.symbol}<br/>
              权重: ${item.weight.toFixed(2)}%<br/>
              市值: $${item.value.toLocaleString()}
            </div>
          `;
        },
      },
      legend: {
        orient: 'vertical',
        left: 'left',
        top: 'middle',
        textStyle: {
          color: theme.colors.textSecondary,
        },
      },
      series: [
        {
          name: '持仓分布',
          type: 'pie',
          radius: ['40%', '70%'],
          center: ['60%', '50%'],
          avoidLabelOverlap: false,
          itemStyle: {
            borderRadius: 8,
            borderColor: theme.colors.surface,
            borderWidth: 2,
          },
          label: {
            show: false,
            position: 'center',
          },
          emphasis: {
            label: {
              show: true,
              fontSize: 16,
              fontWeight: 'bold',
              color: theme.colors.textPrimary,
            },
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: 'rgba(0, 0, 0, 0.5)',
            },
          },
          labelLine: {
            show: false,
          },
          data: data.map((item, index) => ({
            value: item.weight,
            name: item.symbol,
            itemStyle: {
              color: colors[index % colors.length],
            },
          })),
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

export default HoldingPieChart;

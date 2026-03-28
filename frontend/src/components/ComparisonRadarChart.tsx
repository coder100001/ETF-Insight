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

interface ETFMetric {
  symbol: string;
  name: string;
  dividend_yield: number;
  volatility: number;
  sharpe_ratio: number;
  total_return: number;
  max_drawdown: number;
}

interface ComparisonRadarChartProps {
  data: ETFMetric[];
  title?: string;
}

const ComparisonRadarChart: React.FC<ComparisonRadarChartProps> = ({ data, title = 'ETF 对比分析' }) => {
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
    ];

    // 标准化数据到 0-100 范围
    const normalize = (value: number, min: number, max: number) => {
      if (max === min) return 50;
      return ((value - min) / (max - min)) * 100;
    };

    // 计算每个指标的最小值和最大值
    const metrics = ['dividend_yield', 'volatility', 'sharpe_ratio', 'total_return', 'max_drawdown'];
    const ranges = metrics.reduce((acc, metric) => {
      const values = data.map(d => d[metric as keyof ETFMetric] as number);
      acc[metric] = { min: Math.min(...values), max: Math.max(...values) };
      return acc;
    }, {} as Record<string, { min: number; max: number }>);

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
          const etf = data[dataIndex as number];
          if (!etf) return '';
          return `
            <div style="padding: 8px;">
              <strong>${etf.name}</strong><br/>
              股息率: ${etf.dividend_yield.toFixed(2)}%<br/>
              波动率: ${etf.volatility.toFixed(2)}%<br/>
              夏普比率: ${etf.sharpe_ratio.toFixed(2)}<br/>
              总收益: ${etf.total_return.toFixed(2)}%<br/>
              最大回撤: ${etf.max_drawdown.toFixed(2)}%
            </div>
          `;
        },
      },
      legend: {
        data: data.map(d => d.symbol),
        bottom: 0,
        textStyle: {
          color: theme.colors.textSecondary,
        },
      },
      radar: {
        indicator: [
          { name: '股息率', max: 100 },
          { name: '波动率', max: 100 },
          { name: '夏普比率', max: 100 },
          { name: '总收益', max: 100 },
          { name: '最大回撤', max: 100 },
        ],
        shape: 'polygon',
        splitNumber: 4,
        axisName: {
          color: theme.colors.textSecondary,
          fontSize: 12,
        },
        splitLine: {
          lineStyle: {
            color: theme.colors.divider,
          },
        },
        splitArea: {
          show: true,
          areaStyle: {
            color: ['transparent', `${theme.colors.primary}05`],
          },
        },
        axisLine: {
          lineStyle: {
            color: theme.colors.divider,
          },
        },
      },
      series: [
        {
          name: 'ETF 对比',
          type: 'radar',
          data: data.map((etf, index) => ({
            value: [
              normalize(etf.dividend_yield, ranges.dividend_yield.min, ranges.dividend_yield.max),
              normalize(etf.volatility, ranges.volatility.min, ranges.volatility.max),
              normalize(etf.sharpe_ratio, ranges.sharpe_ratio.min, ranges.sharpe_ratio.max),
              normalize(etf.total_return, ranges.total_return.min, ranges.total_return.max),
              normalize(etf.max_drawdown, ranges.max_drawdown.min, ranges.max_drawdown.max),
            ],
            name: etf.symbol,
            lineStyle: {
              color: colors[index % colors.length],
              width: 2,
            },
            areaStyle: {
              color: `${colors[index % colors.length]}20`,
            },
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

export default ComparisonRadarChart;

import { useEffect, useRef, useMemo } from 'react';
import echarts from "../lib/echarts";
import type { EChartsType } from "echarts/core";
import type { EChartsOption } from 'echarts';

interface ReactEChartProps {
  option: EChartsOption;
  height?: number | string;
  width?: number | string;
  className?: string;
  style?: React.CSSProperties;
  renderer?: 'canvas' | 'svg';
}

/**
 * 可复用的 ECharts 封装组件
 *
 * 特性：
 * - 自动初始化、更新、销毁 ECharts 实例
 * - 自动响应窗口大小变化
 * - 支持 canvas / svg 渲染器
 * - Safari 兼容性处理（禁用动画）
 *
 * @example
 * ```tsx
 * <ReactEChart
 *   option={{
 *     xAxis: { type: 'category', data: ['A', 'B', 'C'] },
 *     yAxis: { type: 'value' },
 *     series: [{ type: 'bar', data: [10, 20, 30] }],
 *   }}
 *   height={300}
 * />
 * ```
 */
const ReactEChart: React.FC<ReactEChartProps> = ({
  option,
  height = 300,
  width = '100%',
  className,
  style,
  renderer = 'canvas',
}) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const chartInstanceRef = useRef<EChartsType | null>(null);

  // Safari 检测（用于禁用动画以避免渲染问题）
  const isSafari = useMemo(() => {
    if (typeof navigator === 'undefined') return false;
    return /^((?!chrome|android).)*safari/i.test(navigator.userAgent);
  }, []);

  // 合并动画配置
  const mergedOption = useMemo<EChartsOption>(() => {
    if (isSafari) {
      return { ...option, animation: false };
    }
    return option;
  }, [option, isSafari]);

  // 初始化图表实例
  useEffect(() => {
    if (!chartRef.current) return;

    const chart = echarts.init(chartRef.current, undefined, { renderer });
    chartInstanceRef.current = chart;
    chart.setOption(mergedOption);

    const handleResize = () => chart.resize();
    window.addEventListener('resize', handleResize);

    // ResizeObserver 用于容器尺寸变化（如侧边栏折叠）
    const resizeObserver = new ResizeObserver(() => {
      chart.resize();
    });
    resizeObserver.observe(chartRef.current);

    return () => {
      window.removeEventListener('resize', handleResize);
      resizeObserver.disconnect();
      chart.dispose();
      chartInstanceRef.current = null;
    };
  }, [renderer]); // eslint-disable-line react-hooks/exhaustive-deps

  // 选项更新时刷新图表（增量 merge，保留 dataZoom/legend 等交互状态）
  useEffect(() => {
    if (chartInstanceRef.current && mergedOption) {
      chartInstanceRef.current.setOption(mergedOption);
    }
  }, [mergedOption]);

  return (
    <div
      ref={chartRef}
      className={className}
      style={{ width, height, ...style }}
    />
  );
};

export default ReactEChart;

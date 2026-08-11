/**
 * ECharts 按需引入（tree-shaking）
 *
 * 全量 `import * as echarts from 'echarts'` 打包约 1MB，此处只注册项目实际用到的
 * 图表类型和组件，显著减小包体积、加快首屏加载。
 *
 * 新增图表类型/组件时，先在这里注册，再在各页面使用。
 */
import * as echarts from 'echarts/core';
import {
  LineChart,
  BarChart,
  PieChart,
  RadarChart,
  ScatterChart,
} from 'echarts/charts';
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  MarkLineComponent,
  DataZoomComponent,
} from 'echarts/components';
import { CanvasRenderer } from 'echarts/renderers';

echarts.use([
  // 图表类型
  LineChart,
  BarChart,
  PieChart,
  RadarChart,
  ScatterChart,
  // 组件
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  MarkLineComponent,
  DataZoomComponent,
  // 渲染器
  CanvasRenderer,
]);

export default echarts;

export type { EChartsType } from "echarts/core";

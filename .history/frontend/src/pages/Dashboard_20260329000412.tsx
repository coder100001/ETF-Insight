import { useState, useEffect, useRef } from 'react';
import styled from 'styled-components';
import { Card, Table, Progress, Badge } from 'antd';
import { LineChartOutlined, ProjectOutlined } from '@ant-design/icons';
import * as echarts from 'echarts';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import type { WorkflowStat } from '../types';

const PageTitle = styled.h2`
  margin: 0 0 20px 0;
  font-size: ${theme.fonts.size['2xl']};
  color: ${theme.colors.textPrimary};
`;

const StatsRow = styled.div`
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 20px;

  @media (max-width: ${theme.breakpoints.lg}) {
    grid-template-columns: repeat(2, 1fr);
  }

  @media (max-width: ${theme.breakpoints.sm}) {
    grid-template-columns: 1fr;
  }
`;

const StatCard = styled.div<{ $borderColor?: string }>`
  background: ${theme.colors.surface};
  padding: 20px;
  border-radius: ${theme.borderRadius.md};
  box-shadow: ${theme.shadows.card};
  border-left: 4px solid ${props => props.$borderColor || theme.colors.primary};

  h3 {
    font-size: ${theme.fonts.size['3xl']};
    margin: 0 0 5px 0;
    color: ${props => props.$borderColor || theme.colors.primary};
  }

  p {
    color: ${theme.colors.textSecondary};
    margin: 0;
    font-size: ${theme.fonts.size.base};
  }
`;

const StyledCard = styled(Card)`
  margin-bottom: 20px;
  box-shadow: ${theme.shadows.card};

  .ant-card-head {
    background: ${theme.colors.background};
    border-bottom: 1px solid ${theme.colors.border};
  }

  .ant-card-body {
    padding: 20px;
  }
`;

const ChartWrapper = styled.div`
  height: 200px;
  width: 100%;
`;

// 模拟数据
const mockWorkflowStats: WorkflowStat[] = [
  { name: 'ETF数据更新', total: 156, success: 148, failed: 8, success_rate: 94.9, status: 'good' },
  { name: '汇率数据获取', total: 89, success: 85, failed: 4, success_rate: 95.5, status: 'good' },
  { name: '投资组合分析', total: 45, success: 38, failed: 7, success_rate: 84.4, status: 'warning' },
  { name: '数据备份', total: 30, success: 22, failed: 8, success_rate: 73.3, status: 'danger' },
];

const mockDailyStats = [
  { date: '03-22', total: 12, success: 11, failed: 1 },
  { date: '03-23', total: 15, success: 14, failed: 1 },
  { date: '03-24', total: 8, success: 7, failed: 1 },
  { date: '03-25', total: 18, success: 16, failed: 2 },
  { date: '03-26', total: 20, success: 19, failed: 1 },
  { date: '03-27', total: 14, success: 13, failed: 1 },
  { date: '03-28', total: 10, success: 9, failed: 1 },
];

const Dashboard: React.FC = () => {
  const chartRef = useRef<HTMLDivElement>(null);
  const [todayStats] = useState({
    total: 10,
    success: 9,
    failed: 1,
    running: 2,
  });

  useEffect(() => {
    if (chartRef.current) {
      const chart = echarts.init(chartRef.current);
      const option: echarts.EChartsOption = {
        tooltip: {
          trigger: 'axis',
        },
        legend: {
          data: ['成功', '失败'],
          bottom: 0,
        },
        grid: {
          left: '3%',
          right: '4%',
          bottom: '15%',
          top: '10%',
          containLabel: true,
        },
        xAxis: {
          type: 'category',
          data: mockDailyStats.map(d => d.date),
        },
        yAxis: {
          type: 'value',
        },
        series: [
          {
            name: '成功',
            type: 'line',
            smooth: true,
            data: mockDailyStats.map(d => d.success),
            itemStyle: { color: theme.colors.success },
          },
          {
            name: '失败',
            type: 'line',
            smooth: true,
            data: mockDailyStats.map(d => d.failed),
            itemStyle: { color: theme.colors.danger },
          },
        ],
      };
      chart.setOption(option);

      const handleResize = () => chart.resize();
      window.addEventListener('resize', handleResize);

      return () => {
        window.removeEventListener('resize', handleResize);
        chart.dispose();
      };
    }
  }, []);

  const columns = [
    {
      title: '工作流名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '总执行次数',
      dataIndex: 'total',
      key: 'total',
      align: 'center' as const,
    },
    {
      title: '成功次数',
      dataIndex: 'success',
      key: 'success',
      align: 'center' as const,
    },
    {
      title: '成功率',
      dataIndex: 'success_rate',
      key: 'success_rate',
      render: (rate: number) => (
        <Progress
          percent={rate}
          size="small"
          status={rate >= 90 ? 'success' : rate >= 70 ? 'normal' : 'exception'}
          format={(percent) => `${percent?.toFixed(1)}%`}
        />
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      align: 'center' as const,
      render: (status: string) => {
        const statusMap: { [key: string]: { text: string; color: string } } = {
          good: { text: '良好', color: 'success' },
          warning: { text: '一般', color: 'warning' },
          danger: { text: '需关注', color: 'error' },
        };
        const { text, color } = statusMap[status] || statusMap.good;
        return <Badge status={color as any} text={text} />;
      },
    },
  ];

  return (
    <Layout>
      <PageTitle>仪表板</PageTitle>

      <StatsRow>
        <StatCard $borderColor={theme.colors.primary}>
          <h3>{todayStats.total}</h3>
          <p>今日执行总数</p>
        </StatCard>
        <StatCard $borderColor={theme.colors.success}>
          <h3>{todayStats.success}</h3>
          <p>执行成功</p>
        </StatCard>
        <StatCard $borderColor={theme.colors.danger}>
          <h3>{todayStats.failed}</h3>
          <p>执行失败</p>
        </StatCard>
        <StatCard $borderColor={theme.colors.warning}>
          <h3>{todayStats.running}</h3>
          <p>正在运行</p>
        </StatCard>
      </StatsRow>

      <StyledCard
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <LineChartOutlined />
            <span>近7天执行趋势</span>
          </div>
        }
      >
        <ChartWrapper ref={chartRef} />
      </StyledCard>

      <StyledCard
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <ProjectOutlined />
            <span>工作流统计</span>
          </div>
        }
      >
        <Table
          dataSource={mockWorkflowStats}
          columns={columns}
          rowKey="name"
          pagination={false}
          locale={{
            emptyText: (
              <div style={{ textAlign: 'center', color: theme.colors.textMuted, padding: '40px' }}>
                暂无工作流统计数据
              </div>
            ),
          }}
        />
      </StyledCard>
    </Layout>
  );
};

export default Dashboard;

import { useState, useEffect, useRef, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { Card, Table, Badge, Spin, Row, Col, Statistic, Tag, Button } from 'antd';
import {
  LineChartOutlined, ProjectOutlined, FundOutlined,
  SwapOutlined, ReloadOutlined,
  ArrowUpOutlined, ArrowDownOutlined, WalletOutlined
} from '@ant-design/icons';
import echarts from "../lib/echarts";
import type { EChartsOption } from "echarts";
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import { portfolioConfigAPI, operationLogsAPI } from '../services/api';
import { useETFStore } from '../stores/etfStore';

interface LogEntry {
  id: number;
  log_type: string;
  timestamp: string;
  user: string;
  module: string;
  action_type: string;
  details: string;
  ip: string;
  status: string;
  status_code: number;
  error_message: string;
  duration_ms: number;
}

interface ETFItem {
  symbol: string;
  name?: string;
  current_price?: number;
  close_price?: number;
  price_change_percent?: number;
  change_percent?: number;
  market?: string;
}

interface ExchangeRateItem {
  from_currency: string;
  to_currency: string;
  rate: number;
  updated_at: string;
}

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
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

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
  height: 300px;
  width: 100%;
`;

const QuickActions = styled.div`
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 20px;
`;

const QuickActionCard = styled(Card)`
  flex: 1;
  min-width: 150px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    border-color: ${theme.colors.primary};
    box-shadow: 0 2px 8px rgba(24, 144, 255, 0.2);
  }

  .ant-card-body {
    padding: 16px;
    text-align: center;
  }
`;

const themeColors = {
  success: theme.colors.success,
  danger: theme.colors.danger,
  primary: theme.colors.primary,
  warning: theme.colors.warning,
};

interface DashboardStats {
  totalETFs: number;
  totalPortfolios: number;
  recentLogs: number;
  exchangeRates: number;
}

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const chartRef = useRef<HTMLDivElement>(null);
  const [loading, setLoading] = useState(false); // 初始 false：立即渲染 store 数据，不阻塞
  const [stats, setStats] = useState<DashboardStats>({
    totalETFs: 0,
    totalPortfolios: 0,
    recentLogs: 0,
    exchangeRates: 0,
  });
  const [recentLogs, setRecentLogs] = useState<LogEntry[]>([]);

  // 热门 ETF + 汇率数据来自全局 etfStore（一次获取，多页面共享，不再每次请求）
  const storeEtfs = useETFStore(state => state.etfList);
  const exchangeRates = useETFStore(state => state.exchangeRates);
  const initialize = useETFStore(state => state.initialize);
  const etfList = useMemo<ETFItem[]>(
    () => storeEtfs.slice(0, 10).map(etf => ({
      symbol: etf.symbol,
      name: etf.name,
      current_price: etf.current_price,
      change_percent: etf.change_percent,
    })),
    [storeEtfs]
  );

  useEffect(() => {
    initialize();
    fetchDashboardData();
  }, [initialize]);

  // 手动刷新：同时刷新 store 的 ETF 列表与汇率（force 强制重新获取）
  const refreshAll = () => {
    const store = useETFStore.getState();
    store.refreshETFList(true);
    store.fetchExchangeRates();
    fetchDashboardData();
  };

  // ETF 总数 + 汇率数量从全局 store 同步（store 一次获取，不重复请求）
  useEffect(() => {
    setStats(prev => ({ ...prev, totalETFs: storeEtfs.length, exchangeRates: exchangeRates.length }));
  }, [storeEtfs, exchangeRates]);

  const fetchDashboardData = async () => {
    setLoading(true);
    try {
      // ETF 列表和汇率由全局 etfStore 一次获取并缓存，此处只请求其余数据
      const [portfolioRes, logsRes] = await Promise.allSettled([
        portfolioConfigAPI.getAll(),
        operationLogsAPI.getLogs({ page: 1, page_size: 5 }),
      ]);

      // 处理组合配置
      if (portfolioRes.status === 'fulfilled' && portfolioRes.value?.success) {
        const portfolios = portfolioRes.value.data || [];
        setStats(prev => ({ ...prev, totalPortfolios: portfolios.length }));
      }

      // 处理操作日志（后端 OperationLog 字段映射为前端展示格式）
      if (logsRes.status === 'fulfilled' && logsRes.value?.success) {
        const logs = logsRes.value.data?.data || logsRes.value.data || [];
        const mapped = Array.isArray(logs)
          ? logs.slice(0, 5).map((log: Record<string, unknown>) => ({
              id: log.id as number,
              // 后端字段：operation_type / start_time / status(0:进行中,1:成功,2:失败)
              log_type: String(log.operation_type || log.log_type || '未知'),
              timestamp: String(log.start_time || log.timestamp || ''),
              status: log.status === 1 ? 'success' : log.status === 2 ? 'error' : log.status === 0 ? 'processing' : String(log.status || 'unknown'),
              action_type: String(log.operation_name || log.action_type || ''),
              user: String(log.operator || ''),
              module: '',
              details: String(log.details || ''),
              ip: '',
              status_code: 0,
              error_message: String(log.error_message || ''),
              duration_ms: log.duration_ms as number || 0,
            }))
          : [];
        setRecentLogs(mapped);
        setStats(prev => ({ ...prev, recentLogs: mapped.length }));
      }
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (chartRef.current && etfList.length > 0) {
      const isSafari = /^((?!chrome|android).)*safari/i.test(navigator.userAgent);
      const chart = echarts.init(chartRef.current, undefined, { renderer: 'canvas' });

      // 使用真实 ETF 数据生成图表
      const option: EChartsOption = {
        tooltip: {
          trigger: 'axis',
          formatter: '{b}: {c}',
        },
        grid: {
          left: '3%',
          right: '4%',
          bottom: '10%',
          top: '10%',
          containLabel: true,
        },
        xAxis: {
          type: 'category',
          data: etfList.map(etf => etf.symbol),
          axisLabel: {
            rotate: 45,
            fontSize: 10,
          },
        },
        yAxis: {
          type: 'value',
          name: '价格',
          axisLabel: {
            formatter: '${value}',
          },
        },
        animation: !isSafari,
        series: [
          {
            name: '当前价格',
            type: 'bar',
            data: etfList.map(etf => etf.current_price || etf.close_price || 0),
            itemStyle: {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                { offset: 0, color: themeColors.primary },
                { offset: 1, color: themeColors.success },
              ]),
            },
            barMaxWidth: 40,
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
  }, [etfList]);

  const logColumns = [
    {
      title: '操作',
      dataIndex: 'log_type',
      key: 'log_type',
      render: (text: string) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const statusMap: Record<string, { text: string; color: string }> = {
          'success': { text: '成功', color: 'success' },
          'error': { text: '失败', color: 'error' },
          'processing': { text: '进行中', color: 'processing' },
        };
        const { text, color } = statusMap[status] || { text: '未知', color: 'default' };
        return <Badge status={color as 'success' | 'error' | 'default' | 'processing' | 'warning'} text={text} />;
      },
    },
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      render: (text: string) => text ? new Date(text).toLocaleString('zh-CN') : '-',
    },
  ];

  const etfColumns = [
    {
      title: '代码',
      dataIndex: 'symbol',
      key: 'symbol',
      render: (text: string) => (
        <Button type="link" onClick={() => navigate(`/etf-detail/${text}`)}>
          {text}
        </Button>
      ),
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: '价格',
      dataIndex: 'current_price',
      key: 'current_price',
      render: (price: number, record: ETFItem) => {
        const p = price || record.close_price || 0;
        return p > 0 ? `$${p.toFixed(2)}` : '-';
      },
    },
    {
      title: '涨跌幅',
      dataIndex: 'change_percent',
      key: 'change_percent',
      render: (percent: number, record: ETFItem) => {
        const p = percent || record.change_percent || 0;
        if (p === 0) return <span>-</span>;
        const isUp = p > 0;
        return (
          <span style={{ color: isUp ? theme.colors.success : theme.colors.danger }}>
            {isUp ? <ArrowUpOutlined /> : <ArrowDownOutlined />}
            {Math.abs(p).toFixed(2)}%
          </span>
        );
      },
    },
  ];

  return (
    <Layout>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
        <PageTitle>仪表板</PageTitle>
        <Button
          icon={<ReloadOutlined />}
          onClick={refreshAll}
          loading={loading}
        >
          刷新数据
        </Button>
      </div>

      <Spin spinning={loading}>
        {/* 统计卡片 */}
        <StatsRow>
          <StatCard $borderColor={theme.colors.primary} onClick={() => navigate('/etf-dashboard')}>
            <h3>{stats.totalETFs}</h3>
            <p><FundOutlined /> ETF 基金数量</p>
          </StatCard>
          <StatCard $borderColor={theme.colors.success} onClick={() => navigate('/portfolio-config')}>
            <h3>{stats.totalPortfolios}</h3>
            <p><WalletOutlined /> 投资组合数</p>
          </StatCard>
          <StatCard $borderColor={theme.colors.warning} onClick={() => navigate('/operation-logs')}>
            <h3>{stats.recentLogs}</h3>
            <p><ProjectOutlined /> 近期操作</p>
          </StatCard>
          <StatCard $borderColor={theme.colors.danger} onClick={() => navigate('/exchange-rate')}>
            <h3>{stats.exchangeRates}</h3>
            <p><SwapOutlined /> 汇率对数</p>
          </StatCard>
        </StatsRow>

        {/* 快捷操作 */}
        <QuickActions>
          <QuickActionCard onClick={() => navigate('/portfolio-analysis')}>
            <Statistic
              title="组合分析"
              value="情景模拟"
              valueStyle={{ fontSize: 16, color: theme.colors.primary }}
            />
          </QuickActionCard>
          <QuickActionCard onClick={() => navigate('/portfolio-optimization')}>
            <Statistic
              title="组合优化"
              value="MPT/BL"
              valueStyle={{ fontSize: 16, color: theme.colors.success }}
            />
          </QuickActionCard>
          <QuickActionCard onClick={() => navigate('/risk-analysis')}>
            <Statistic
              title="风险分析"
              value="VaR/CVaR"
              valueStyle={{ fontSize: 16, color: theme.colors.warning }}
            />
          </QuickActionCard>
          <QuickActionCard onClick={() => navigate('/factor-analysis')}>
            <Statistic
              title="因子分析"
              value="三因子/五因子"
              valueStyle={{ fontSize: 16, color: theme.colors.danger }}
            />
          </QuickActionCard>
          <QuickActionCard onClick={() => navigate('/quantlib')}>
            <Statistic
              title="量化分析"
              value="期权/债券"
              valueStyle={{ fontSize: 16, color: '#722ed1' }}
            />
          </QuickActionCard>
        </QuickActions>

        <Row gutter={20}>
          {/* ETF 价格图表 */}
          <Col xs={24} lg={14}>
            <StyledCard
              title={
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <LineChartOutlined />
                  <span>ETF 价格概览</span>
                </div>
              }
              extra={
                <Button type="link" onClick={() => navigate('/etf-dashboard')}>
                  查看全部
                </Button>
              }
            >
              <ChartWrapper ref={chartRef} />
            </StyledCard>
          </Col>

          {/* 最近操作日志 */}
          <Col xs={24} lg={10}>
            <StyledCard
              title={
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <ProjectOutlined />
                  <span>最近操作</span>
                </div>
              }
              extra={
                <Button type="link" onClick={() => navigate('/operation-logs')}>
                  查看全部
                </Button>
              }
            >
              <Table
                dataSource={recentLogs}
                columns={logColumns}
                rowKey="id"
                pagination={false}
                size="small"
                locale={{ emptyText: '暂无操作记录' }}
              />
            </StyledCard>
          </Col>
        </Row>

        <Row gutter={20}>
          {/* 热门 ETF */}
          <Col xs={24} lg={12}>
            <StyledCard
              title={
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <FundOutlined />
                  <span>热门 ETF</span>
                </div>
              }
              extra={
                <Button type="link" onClick={() => navigate('/etf-dashboard')}>
                  查看全部
                </Button>
              }
            >
              <Table
                dataSource={etfList}
                columns={etfColumns}
                rowKey="symbol"
                pagination={false}
                size="small"
                locale={{ emptyText: '暂无 ETF 数据' }}
              />
            </StyledCard>
          </Col>

          {/* 汇率概览 */}
          <Col xs={24} lg={12}>
            <StyledCard
              title={
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <SwapOutlined />
                  <span>汇率概览</span>
                </div>
              }
              extra={
                <Button type="link" onClick={() => navigate('/exchange-rate')}>
                  查看全部
                </Button>
              }
            >
              <Table
                dataSource={exchangeRates}
                columns={[
                  {
                    title: '货币对',
                    dataIndex: 'currency_pair',
                    key: 'currency_pair',
                    render: (text: string, record: ExchangeRateItem) => text || `${record.from_currency}/${record.to_currency}`,
                  },
                  {
                    title: '汇率',
                    dataIndex: 'rate',
                    key: 'rate',
                    render: (rate: number) => rate ? rate.toFixed(4) : '-',
                  },
                  {
                    title: '更新时间',
                    dataIndex: 'updated_at',
                    key: 'updated_at',
                    render: (text: string) => text ? new Date(text).toLocaleString('zh-CN') : '-',
                  },
                ]}
                rowKey="id"
                pagination={false}
                size="small"
                locale={{ emptyText: '暂无汇率数据' }}
              />
            </StyledCard>
          </Col>
        </Row>
      </Spin>
    </Layout>
  );
};

export default Dashboard;

import { useState, useEffect, useMemo } from 'react';
import styled from 'styled-components';
import { Card, Table, Button, Select, App, Row, Col, Statistic, Tag } from 'antd';
import { BarChartOutlined, TrophyOutlined, SafetyOutlined, PercentageOutlined } from '@ant-design/icons';
import { FaBalanceScale } from 'react-icons/fa';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import { etfAPI, universalETFAPI } from '../services/api';
import type { ETFData } from '../types';

const PageHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;

  h2 {
    margin: 0;
    font-size: ${theme.fonts.size['2xl']};
    color: ${theme.colors.textPrimary};
    display: flex;
    align-items: center;
    gap: 10px;
  }
`;

const FilterSection = styled.div`
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
  padding: 16px;
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.md};
  box-shadow: ${theme.shadows.card};
`;

const StyledTable = styled(Table)`
  .ant-table-thead > tr > th {
    background: ${theme.colors.background};
    font-weight: ${theme.fonts.weight.semibold};
  }

  .ant-table-tbody > tr:hover > td {
    background: #f8f9fa;
  }
` as typeof Table;

const ComparisonSection = styled.div`
  margin-bottom: 20px;
`;

const HighlightRow = styled(Row)`
  margin-bottom: 16px;
`;

const HighlightCard = styled(Card)`
  text-align: center;
  box-shadow: ${theme.shadows.card};
  .ant-card-body {
    padding: 16px;
  }
`;

const WinnerTag = styled(Tag)`
  font-size: 14px;
  padding: 4px 12px;
  margin-top: 8px;
`;

interface ETFApiItem {
  symbol: string;
  name: string;
  current_price?: number;
  previous_close?: number;
  change?: number;
  change_percent?: number;
  open_price?: number;
  high_price?: number;
  low_price?: number;
  volume?: number;
  dividend_yield?: number;
  volatility?: number;
  total_return?: number;
  max_drawdown?: number;
  sharpe_ratio?: number;
  expense_ratio?: number;
  focus?: string;
  strategy?: string;
}

const ETFComparison: React.FC = () => {
  const { message } = App.useApp();
  const [etfs, setEtfs] = useState<ETFData[]>([]);
  const [loading, setLoading] = useState(false);
  const [compareLoading, setCompareLoading] = useState(false);
  const [selectedETFs, setSelectedETFs] = useState<string[]>(['SCHD', 'SPYD', 'JEPQ']);
  const [compared, setCompared] = useState(false);
  const [correlations, setCorrelations] = useState<Record<string, Record<string, number>> | null>(null);

  useEffect(() => {
    fetchETFData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const fetchETFData = async () => {
    setLoading(true);
    try {
      const response = await etfAPI.getList();
      if (response.success && response.data) {
        const formattedData: ETFData[] = response.data.map((item: ETFApiItem) => ({
          symbol: item.symbol,
          name: item.name,
          current_price: item.current_price || 0,
          previous_close: item.previous_close || 0,
          change: item.change || 0,
          change_percent: item.change_percent || 0,
          open_price: item.open_price || 0,
          high_price: item.high_price || 0,
          low_price: item.low_price || 0,
          volume: item.volume || 0,
          dividend_yield: item.dividend_yield || 0,
          volatility: item.volatility || 0,
          total_return: item.total_return || 0,
          max_drawdown: item.max_drawdown || 0,
          sharpe_ratio: item.sharpe_ratio || 0,
          expense_ratio: item.expense_ratio || 0,
          info: {
            focus: item.focus || '',
            strategy: item.strategy || '',
          },
        }));
        setEtfs(formattedData);
      } else {
        message.error('获取ETF数据失败');
      }
    } catch (error) {
      message.error('获取ETF数据失败: ' + (error as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const handleCompare = async () => {
    if (selectedETFs.length < 2) {
      message.warning('请至少选择2个ETF进行对比');
      return;
    }
    setCompareLoading(true);
    setCompared(false);
    setCorrelations(null);
    try {
      // 调用对比 API 获取相关性数据
      const response = await universalETFAPI.compare(selectedETFs);
      if (response.success && response.data) {
        setCorrelations(response.data.correlations || null);
      }
      setCompared(true);
      message.success(`已完成 ${selectedETFs.length} 个ETF的对比分析`);
    } catch {
      // API 调用失败时仍使用本地数据展示对比
      setCompared(true);
      message.info('已使用本地数据进行对比');
    } finally {
      setCompareLoading(false);
    }
  };

  const filteredETFs = etfs.filter(etf => selectedETFs.includes(etf.symbol));

  // 计算各指标最优 ETF
  const highlights = useMemo(() => {
    if (!compared || filteredETFs.length < 2) return null;
    const items = filteredETFs;
    const maxBy = (fn: (e: ETFData) => number) =>
      items.reduce((best, e) => (fn(e) > fn(best) ? e : best), items[0]);
    const minBy = (fn: (e: ETFData) => number) =>
      items.reduce((best, e) => (fn(e) < fn(best) ? e : best), items[0]);

    return {
      bestDividend: maxBy(e => e.dividend_yield ?? 0),
      bestSharpe: maxBy(e => e.sharpe_ratio ?? 0),
      lowestVolatility: minBy(e => e.volatility ?? Infinity),
      lowestDrawdown: minBy(e => Math.abs(e.max_drawdown ?? 0)),
      lowestExpense: minBy(e => e.expense_ratio ?? Infinity),
      highestReturn: maxBy(e => e.total_return ?? 0),
    };
  }, [compared, filteredETFs]);

  const columns: import('antd').TableProps<ETFData>['columns'] = [
    {
      title: 'ETF',
      dataIndex: 'symbol',
      key: 'symbol',
      render: (text, record) => (
        <div>
          <strong>{text}</strong>
          <br />
          <small style={{ color: theme.colors.textMuted }}>{record.name}</small>
        </div>
      ),
    },
    {
      title: '当前价格',
      dataIndex: 'current_price',
      key: 'current_price',
      align: 'center' as const,
      render: (value) => `$${(value as number).toFixed(2)}`,
    },
    {
      title: '今日涨跌',
      dataIndex: 'change_percent',
      key: 'change_percent',
      align: 'center' as const,
      render: (value) => (
        <span style={{ color: (value as number) >= 0 ? theme.colors.success : theme.colors.danger }}>
          {(value as number) >= 0 ? '+' : ''}{(value as number).toFixed(2)}%
        </span>
      ),
    },
    {
      title: '股息率',
      dataIndex: 'dividend_yield',
      key: 'dividend_yield',
      align: 'center' as const,
      render: (value) => value ? `${(value as number).toFixed(2)}%` : '-',
    },
    {
      title: '年化波动率',
      dataIndex: 'volatility',
      key: 'volatility',
      align: 'center' as const,
      render: (value) => `${(value as number).toFixed(2)}%`,
    },
    {
      title: '夏普比率',
      dataIndex: 'sharpe_ratio',
      key: 'sharpe_ratio',
      align: 'center' as const,
      render: (value) => (value as number).toFixed(2),
    },
    {
      title: '最大回撤',
      dataIndex: 'max_drawdown',
      key: 'max_drawdown',
      align: 'center' as const,
      render: (value) => (
        <span style={{ color: theme.colors.danger }}>{(value as number).toFixed(2)}%</span>
      ),
    },
    {
      title: '费率',
      dataIndex: 'expense_ratio',
      key: 'expense_ratio',
      align: 'center' as const,
      render: (value) => `${(value as number).toFixed(2)}%`,
    },
    {
      title: '策略',
      dataIndex: ['info', 'strategy'],
      key: 'strategy',
      align: 'center' as const,
    },
  ];

  return (
    <Layout>
      <PageHeader>
        <h2>
          <FaBalanceScale />
          ETF对比分析
        </h2>
      </PageHeader>

      <FilterSection>
        <Select
          mode="multiple"
          placeholder="选择要对比的ETF"
          value={selectedETFs}
          onChange={setSelectedETFs}
          style={{ minWidth: 300 }}
          options={etfs.map(etf => ({
            label: `${etf.symbol} - ${etf.name}`,
            value: etf.symbol,
          }))}
        />
        <Button type="primary" icon={<BarChartOutlined />} onClick={handleCompare} loading={compareLoading}>
          开始对比
        </Button>
      </FilterSection>

      {highlights && (
        <ComparisonSection>
          <HighlightRow gutter={16}>
            <Col span={8}>
              <HighlightCard>
                <Statistic
                  title="最高股息率"
                  value={`${(highlights.bestDividend.dividend_yield ?? 0).toFixed(2)}%`}
                  prefix={<PercentageOutlined />}
                  valueStyle={{ color: theme.colors.success }}
                />
                <WinnerTag color="green">
                  <TrophyOutlined /> {highlights.bestDividend.symbol}
                </WinnerTag>
              </HighlightCard>
            </Col>
            <Col span={8}>
              <HighlightCard>
                <Statistic
                  title="最佳夏普比率"
                  value={(highlights.bestSharpe.sharpe_ratio ?? 0).toFixed(2)}
                  prefix={<BarChartOutlined />}
                  valueStyle={{ color: theme.colors.primary }}
                />
                <WinnerTag color="blue">
                  <TrophyOutlined /> {highlights.bestSharpe.symbol}
                </WinnerTag>
              </HighlightCard>
            </Col>
            <Col span={8}>
              <HighlightCard>
                <Statistic
                  title="最低波动率"
                  value={`${(highlights.lowestVolatility.volatility ?? 0).toFixed(2)}%`}
                  prefix={<SafetyOutlined />}
                  valueStyle={{ color: theme.colors.info }}
                />
                <WinnerTag color="cyan">
                  <TrophyOutlined /> {highlights.lowestVolatility.symbol}
                </WinnerTag>
              </HighlightCard>
            </Col>
          </HighlightRow>
          <HighlightRow gutter={16}>
            <Col span={8}>
              <HighlightCard>
                <Statistic
                  title="最小回撤"
                  value={`${(highlights.lowestDrawdown.max_drawdown ?? 0).toFixed(2)}%`}
                  prefix={<SafetyOutlined />}
                  valueStyle={{ color: theme.colors.warning }}
                />
                <WinnerTag color="orange">
                  <TrophyOutlined /> {highlights.lowestDrawdown.symbol}
                </WinnerTag>
              </HighlightCard>
            </Col>
            <Col span={8}>
              <HighlightCard>
                <Statistic
                  title="最低费率"
                  value={`${(highlights.lowestExpense.expense_ratio ?? 0).toFixed(2)}%`}
                  prefix={<PercentageOutlined />}
                  valueStyle={{ color: '#722ed1' }}
                />
                <WinnerTag color="purple">
                  <TrophyOutlined /> {highlights.lowestExpense.symbol}
                </WinnerTag>
              </HighlightCard>
            </Col>
            <Col span={8}>
              <HighlightCard>
                <Statistic
                  title="最高总回报"
                  value={`${(highlights.highestReturn.total_return ?? 0).toFixed(2)}%`}
                  prefix={<TrophyOutlined />}
                  valueStyle={{ color: theme.colors.success }}
                />
                <WinnerTag color="green">
                  <TrophyOutlined /> {highlights.highestReturn.symbol}
                </WinnerTag>
              </HighlightCard>
            </Col>
          </HighlightRow>

          {correlations && Object.keys(correlations).length > 0 && (
            <Card
              title="相关性矩阵"
              style={{ boxShadow: theme.shadows.card, marginTop: 16 }}
              size="small"
            >
              <Table
                dataSource={Object.entries(correlations).map(([symbol, corrs]) => ({
                  symbol,
                  ...corrs,
                }))}
                columns={[
                  { title: '', dataIndex: 'symbol', key: 'symbol', fixed: 'left' as const, width: 80 },
                  ...Object.keys(correlations).map(sym => ({
                    title: sym,
                    dataIndex: sym,
                    key: sym,
                    align: 'center' as const,
                    width: 100,
                    render: (val: number) => {
                      if (val === undefined || val === null) return '-';
                      const color = val >= 0.7 ? theme.colors.danger : val >= 0.4 ? theme.colors.warning : theme.colors.success;
                      return <span style={{ color }}>{val.toFixed(2)}</span>;
                    },
                  })),
                ]}
                pagination={false}
                bordered
                size="small"
                scroll={{ x: 'max-content' }}
                rowKey="symbol"
              />
            </Card>
          )}
        </ComparisonSection>
      )}

      <Card style={{ boxShadow: theme.shadows.card }}>
        <StyledTable
          dataSource={filteredETFs}
          columns={columns}
          rowKey="symbol"
          pagination={false}
          bordered
          loading={loading}
        />
      </Card>
    </Layout>
  );
};

export default ETFComparison;

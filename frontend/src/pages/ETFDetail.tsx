
import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import styled from 'styled-components';
import { Card, Button, Badge, Row, Col, Statistic, App } from 'antd';
import { ArrowLeftOutlined, BarChartOutlined } from '@ant-design/icons';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import { etfAPI } from '../services/api';
import type { ETFData } from '../types';
import PriceChart from '../components/PriceChart';

const PageHeader = styled.div`
  display: flex;
  align-items: center;
  gap: 16px;
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

const ETFHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 20px;
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.md};
  box-shadow: ${theme.shadows.card};
`;

const ETFInfo = styled.div`
  h1 {
    margin: 0;
    font-size: ${theme.fonts.size['3xl']};
    color: ${theme.colors.textPrimary};
  }

  p {
    margin: 8px 0 0 0;
    color: ${theme.colors.textSecondary};
  }
`;

const PriceInfo = styled.div`
  text-align: right;

  .price {
    font-size: ${theme.fonts.size['3xl']};
    font-weight: bold;
    color: ${theme.colors.textPrimary};
  }

  .change {
    font-size: ${theme.fonts.size.lg};
    margin-top: 4px;
  }
`;

const ChartPlaceholder = styled.div`
  height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: ${theme.colors.background};
  border-radius: ${theme.borderRadius.md};
  color: ${theme.colors.textSecondary};
  margin-bottom: 20px;
`;

const InfoCard = styled(Card)`
  margin-bottom: 20px;
  box-shadow: ${theme.shadows.card};

  .ant-card-head {
    background: ${theme.colors.background};
    border-bottom: 1px solid ${theme.colors.border};
  }
`;

const InfoGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
`;

const InfoItem = styled.div`
  display: flex;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid ${theme.colors.border};

  &:last-child {
    border-bottom: none;
  }

  .label {
    color: ${theme.colors.textSecondary};
  }

  .value {
    font-weight: ${theme.fonts.weight.semibold};
  }
`;

interface ChartData {
  dates: string[];
  prices: number[];
  volumes: number[];
}

const ETFDetail: React.FC = () => {
  const { message } = App.useApp();
  const { symbol } = useParams<{ symbol: string }>();
  const [etf, setEtf] = useState<ETFData | null>(null);
  const [chartData, setChartData] = useState<ChartData | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (symbol) {
      fetchETFData(symbol);
    }
  }, [symbol]);

  const fetchETFData = async (sym: string) => {
    setLoading(true);
    try {
      // 同时获取实时数据、指标数据和历史数据
      const [realtimeResponse, metricsResponse, historyResponse] = await Promise.all([
        etfAPI.getRealtimeData(sym),
        etfAPI.getMetrics(sym, '1y'),
        etfAPI.getHistory(sym, '1y'),
      ]);

      if (realtimeResponse.success && realtimeResponse.data) {
        const realtimeData = realtimeResponse.data;
        const metricsData = metricsResponse.success && metricsResponse.data ? metricsResponse.data : {};

        setEtf({
          symbol: realtimeData.symbol,
          name: realtimeData.name,
          current_price: realtimeData.current_price || 0,
          previous_close: realtimeData.previous_close || 0,
          change: realtimeData.change || 0,
          change_percent: realtimeData.change_percent || 0,
          open_price: realtimeData.open_price || 0,
          high_price: realtimeData.high_price || 0,
          low_price: realtimeData.low_price || 0,
          volume: realtimeData.volume || 0,
          dividend_yield: metricsData.dividend_yield || realtimeData.dividend_yield || 0,
          volatility: metricsData.volatility || realtimeData.volatility || 0,
          total_return: metricsData.total_return || realtimeData.total_return || 0,
          max_drawdown: metricsData.max_drawdown || realtimeData.max_drawdown || 0,
          sharpe_ratio: metricsData.sharpe_ratio || realtimeData.sharpe_ratio || 0,
          expense_ratio: metricsData.expense_ratio || realtimeData.expense_ratio || 0,
          info: {
            focus: realtimeData.focus || '',
            strategy: realtimeData.strategy || '',
            description: realtimeData.description || '',
          },
        });

        // 处理历史数据用于图表
        if (historyResponse.success && historyResponse.data && Array.isArray(historyResponse.data)) {
          const history = historyResponse.data;
          setChartData({
            dates: history.map((item: any) => item.date?.split('T')[0] || ''),
            prices: history.map((item: any) => item.close_price || 0),
            volumes: history.map((item: any) => item.volume || 0),
          });
        }
      } else {
        message.error('获取ETF数据失败');
      }
    } catch (error) {
      message.error('获取ETF数据失败: ' + (error as Error).message);
    } finally {
      setLoading(false);
    }
  };

  if (!etf) {
    return (
      <Layout>
        <PageHeader>
          <Link to="/etf-dashboard">
            <Button icon={<ArrowLeftOutlined />}>返回</Button>
          </Link>
          <h2>ETF详情</h2>
        </PageHeader>
        <Card style={{ textAlign: 'center', padding: '40px' }} loading={loading}>
          {!loading && <p>未找到ETF: {symbol}</p>}
        </Card>
      </Layout>
    );
  }

  const isUp = etf.change_percent >= 0;

  return (
    <Layout>
      <PageHeader>
        <Link to="/etf-dashboard">
          <Button icon={<ArrowLeftOutlined />}>返回</Button>
        </Link>
        <h2>
          <BarChartOutlined />
          ETF详情
        </h2>
      </PageHeader>

      <ETFHeader>
        <ETFInfo>
          <h1>{etf.symbol}</h1>
          <p>{etf.name}</p>
          <Badge
            count={etf.info.focus}
            style={{ backgroundColor: theme.colors.primary, marginTop: 8 }}
          />
        </ETFInfo>
        <PriceInfo>
          <div className="price">${etf.current_price.toFixed(2)}</div>
          <div
            className="change"
            style={{ color: isUp ? theme.colors.success : theme.colors.danger }}
          >
            {isUp ? '+' : ''}{etf.change.toFixed(2)} ({isUp ? '+' : ''}{etf.change_percent.toFixed(2)}%)
          </div>
        </PriceInfo>
      </ETFHeader>

      {chartData && chartData.dates.length > 0 ? (
        <PriceChart data={chartData} symbol={etf.symbol} />
      ) : (
        <ChartPlaceholder>
          <div style={{ textAlign: 'center' }}>
            <BarChartOutlined style={{ fontSize: 48, marginBottom: 16 }} />
            <p>暂无历史价格数据</p>
          </div>
        </ChartPlaceholder>
      )}

      <Row gutter={[20, 20]}>
        <Col xs={24} lg={12}>
          <InfoCard title="基本信息">
            <InfoGrid>
              <InfoItem>
                <span className="label">全称</span>
                <span className="value">{etf.name}</span>
              </InfoItem>
              <InfoItem>
                <span className="label">策略类型</span>
                <span className="value">{etf.info.strategy}</span>
              </InfoItem>
              <InfoItem>
                <span className="label">管理费率</span>
                <span className="value">{etf.expense_ratio}%</span>
              </InfoItem>
              <InfoItem>
                <span className="label">投资焦点</span>
                <span className="value">{etf.info.focus}</span>
              </InfoItem>
            </InfoGrid>
            {etf.info.description && (
              <p style={{ marginTop: 16, color: theme.colors.textSecondary }}>
                {etf.info.description}
              </p>
            )}
          </InfoCard>
        </Col>

        <Col xs={24} lg={12}>
          <InfoCard title="关键指标">
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <Statistic
                  title="股息率"
                  value={etf.dividend_yield || 0}
                  precision={2}
                  suffix="%"
                  valueStyle={{ color: theme.colors.warning }}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="年化波动率"
                  value={etf.volatility}
                  precision={2}
                  suffix="%"
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="夏普比率"
                  value={etf.sharpe_ratio}
                  precision={2}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="最大回撤"
                  value={etf.max_drawdown}
                  precision={2}
                  suffix="%"
                  valueStyle={{ color: theme.colors.danger }}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="年度收益"
                  value={etf.total_return}
                  precision={2}
                  suffix="%"
                  valueStyle={{ color: etf.total_return >= 0 ? theme.colors.success : theme.colors.danger }}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="成交量"
                  value={(etf.volume / 1000000).toFixed(2)}
                  suffix="M"
                />
              </Col>
            </Row>
          </InfoCard>
        </Col>
      </Row>
    </Layout>
  );
};

export default ETFDetail;

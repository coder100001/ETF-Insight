
import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import styled from 'styled-components';
import { Card, Button, Badge, Row, Col, Statistic, App } from 'antd';
import { ArrowLeftOutlined, BarChartOutlined, LineChartOutlined } from '@ant-design/icons';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import { etfAPI } from '../services/api';
import type { ETFData } from '../types';

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

// 模拟ETF数据
const mockETFData: { [key: string]: ETFData } = {
  SCHD: {
    symbol: 'SCHD',
    name: 'Schwab US Dividend Equity ETF',
    current_price: 30.44,
    previous_close: 31.67,
    change: -1.23,
    change_percent: -3.88,
    open_price: 30.35,
    high_price: 30.59,
    low_price: 30.20,
    volume: 8500000,
    dividend_yield: 3.45,
    volatility: 15.2,
    total_return: 12.5,
    max_drawdown: -8.3,
    sharpe_ratio: 1.2,
    expense_ratio: 0.06,
    info: {
      focus: '美股高股息',
      strategy: '质量因子筛选',
      description: 'SCHD追踪道琼斯美国股息100指数，投资于具有至少10年连续分红历史的高质量股息股票。',
    },
  },
  SPYD: {
    symbol: 'SPYD',
    name: 'SPDR S&P 500 High Dividend ETF',
    current_price: 47.85,
    previous_close: 48.14,
    change: -0.29,
    change_percent: -0.60,
    open_price: 47.71,
    high_price: 48.09,
    low_price: 47.47,
    volume: 6200000,
    dividend_yield: 4.12,
    volatility: 16.8,
    total_return: 8.3,
    max_drawdown: -10.5,
    sharpe_ratio: 0.9,
    expense_ratio: 0.07,
    info: {
      focus: 'S&P高股息',
      strategy: '股息率加权',
      description: 'SPYD投资于S&P 500指数中股息率最高的80只股票，采用股息率加权。',
    },
  },
  JEPQ: {
    symbol: 'JEPQ',
    name: 'JPMorgan Nasdaq Equity Premium Income ETF',
    current_price: 57.20,
    previous_close: 57.51,
    change: -0.31,
    change_percent: -0.54,
    open_price: 57.03,
    high_price: 57.49,
    low_price: 56.74,
    volume: 4800000,
    dividend_yield: 11.2,
    volatility: 18.5,
    total_return: 15.8,
    max_drawdown: -12.1,
    sharpe_ratio: 1.1,
    expense_ratio: 0.35,
    info: {
      focus: '纳斯达克备兑',
      strategy: '期权收益增强',
      description: 'JEPQ通过持有纳斯达克股票并卖出看涨期权来产生收入。',
    },
  },
  JEPI: {
    symbol: 'JEPI',
    name: 'JPMorgan Equity Premium Income ETF',
    current_price: 58.90,
    previous_close: 59.31,
    change: -0.41,
    change_percent: -0.69,
    open_price: 58.72,
    high_price: 59.19,
    low_price: 58.43,
    volume: 9200000,
    dividend_yield: 9.8,
    volatility: 14.2,
    total_return: 11.2,
    max_drawdown: -7.5,
    sharpe_ratio: 1.3,
    expense_ratio: 0.35,
    info: {
      focus: '美股备兑',
      strategy: '期权收益增强',
      description: 'JEPI通过卖出标普500指数的备兑看涨期权来产生月收入。',
    },
  },
  VYM: {
    symbol: 'VYM',
    name: 'Vanguard High Dividend Yield ETF',
    current_price: 154.50,
    previous_close: 155.37,
    change: -0.87,
    change_percent: -0.56,
    open_price: 154.04,
    high_price: 155.27,
    low_price: 153.26,
    volume: 3800000,
    dividend_yield: 2.95,
    volatility: 13.8,
    total_return: 9.6,
    max_drawdown: -9.2,
    sharpe_ratio: 1.0,
    expense_ratio: 0.06,
    info: {
      focus: '美股高股息',
      strategy: '股息率筛选',
      description: 'VYM追踪富时高股息收益率指数，投资于股息率高于平均水平的大型美国股票。',
    },
  },
};

const ETFDetail: React.FC = () => {
  const { message } = App.useApp();
  const { symbol } = useParams<{ symbol: string }>();
  const [etf, setEtf] = useState<ETFData | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (symbol) {
      fetchETFData(symbol);
    }
  }, [symbol]);

  const fetchETFData = async (sym: string) => {
    setLoading(true);
    try {
      const response = await etfAPI.getRealtimeData(sym);
      if (response.success && response.data) {
        const data = response.data;
        setEtf({
          symbol: data.symbol,
          name: data.name,
          current_price: data.current_price || 0,
          previous_close: data.previous_close || 0,
          change: data.change || 0,
          change_percent: data.change_percent || 0,
          open_price: data.open_price || 0,
          high_price: data.high_price || 0,
          low_price: data.low_price || 0,
          volume: data.volume || 0,
          dividend_yield: data.dividend_yield || 0,
          volatility: data.volatility || 0,
          total_return: data.total_return || 0,
          max_drawdown: data.max_drawdown || 0,
          sharpe_ratio: data.sharpe_ratio || 0,
          expense_ratio: data.expense_ratio || 0,
          info: {
            focus: data.focus || '',
            strategy: data.strategy || '',
            description: data.description || '',
          },
        });
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

      <ChartPlaceholder>
        <div style={{ textAlign: 'center' }}>
          <LineChartOutlined style={{ fontSize: 48, marginBottom: 16 }} />
          <p>价格走势图 - 需要集成 Chart.js 或 ECharts</p>
        </div>
      </ChartPlaceholder>

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

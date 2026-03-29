import { useState } from 'react';
import styled from 'styled-components';
import {
  Row, Col, Card, InputNumber, Slider, Button, Table,
  Progress, Statistic, Space, App, Tabs
} from 'antd';
import {
  PieChartOutlined, LineChartOutlined, CalculatorOutlined,
  SyncOutlined, DollarOutlined, PercentageOutlined,
  WalletOutlined
} from '@ant-design/icons';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import { etfAPI } from '../services/api';
import type { PortfolioResult, PortfolioHolding, UserConfig } from '../types';
import HoldingPieChart from '../components/HoldingPieChart';

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

const ConfigCard = styled(Card)`
  margin-bottom: 20px;
  box-shadow: ${theme.shadows.card};

  .ant-card-head {
    background: ${theme.colors.background};
    border-bottom: 1px solid ${theme.colors.border};
  }
`;

const AllocationItem = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;

  .symbol {
    width: 60px;
    font-weight: ${theme.fonts.weight.semibold};
    color: ${theme.colors.textPrimary};
  }

  .slider {
    flex: 1;
  }

  .input {
    width: 80px;
  }
`;

const SummaryRow = styled.div`
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 16px;
  margin-bottom: 20px;

  @media (max-width: ${theme.breakpoints.xl}) {
    grid-template-columns: repeat(3, 1fr);
  }

  @media (max-width: ${theme.breakpoints.md}) {
    grid-template-columns: repeat(2, 1fr);
  }
`;

const SummaryCard = styled.div<{ $color?: string }>`
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.md};
  padding: 16px;
  box-shadow: ${theme.shadows.card};
  border-left: 4px solid ${props => props.$color || theme.colors.primary};

  .label {
    font-size: ${theme.fonts.size.sm};
    color: ${theme.colors.textSecondary};
    margin-bottom: 4px;
  }

  .value {
    font-size: ${theme.fonts.size.xl};
    font-weight: ${theme.fonts.weight.bold};
    color: ${props => props.$color || theme.colors.textPrimary};
  }

  .change {
    font-size: ${theme.fonts.size.sm};
    margin-top: 4px;
    color: ${theme.colors.textMuted};
  }
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

const PresetButton = styled(Button)`
  margin-right: 8px;
  margin-bottom: 8px;
`;

// 模拟投资组合数据
const mockPortfolioResult: PortfolioResult = {
  total_investment: 100000,
  total_value: 108450.50,
  total_return: 8450.50,
  total_return_percent: 8.45,
  annual_dividend_before_tax: 4850.25,
  annual_dividend_after_tax: 4365.23,
  dividend_tax: 485.02,
  tax_rate: 10,
  weighted_dividend_yield: 4.85,
  total_return_with_dividend: 12815.73,
  total_return_with_dividend_percent: 12.82,
  holdings: [
    {
      symbol: 'SCHD',
      name: 'Schwab US Dividend Equity ETF',
      weight: 40,
      investment: 40000,
      current_price: 30.44,
      shares: 1314.06,
      current_value: 40000,
      capital_gain: 0,
      capital_gain_percent: 0,
      total_return: 12.5,
      volatility: 15.2,
      dividend_yield: 3.45,
      annual_dividend_before_tax: 1380,
      annual_dividend_after_tax: 1242,
    },
    {
      symbol: 'SPYD',
      name: 'SPDR S&P 500 High Dividend ETF',
      weight: 30,
      investment: 30000,
      current_price: 47.85,
      shares: 626.96,
      current_value: 30000,
      capital_gain: 0,
      capital_gain_percent: 0,
      total_return: 8.3,
      volatility: 16.8,
      dividend_yield: 4.12,
      annual_dividend_before_tax: 1236,
      annual_dividend_after_tax: 1112.40,
    },
    {
      symbol: 'JEPQ',
      name: 'JPMorgan Nasdaq Equity Premium Income ETF',
      weight: 30,
      investment: 30000,
      current_price: 57.20,
      shares: 524.48,
      current_value: 30000,
      capital_gain: 0,
      capital_gain_percent: 0,
      total_return: 15.8,
      volatility: 18.5,
      dividend_yield: 11.2,
      annual_dividend_before_tax: 3360,
      annual_dividend_after_tax: 3024,
    },
  ],
};

const PortfolioAnalysis: React.FC = () => {
  const { message } = App.useApp();
  const [calculating, setCalculating] = useState(false);
  const [portfolio, setPortfolio] = useState<PortfolioResult>(mockPortfolioResult);

  // 配置状态
  const [config, setConfig] = useState<UserConfig>({
    total_investment: 100000,
    allocation: {
      SCHD: 0.4,
      SPYD: 0.3,
      JEPQ: 0.3,
      JEPI: 0,
      VYM: 0,
    },
    tax_rate: 10,
  });

  // 计算总配比
  const totalAllocation = Object.values(config.allocation).reduce((sum, val) => sum + val, 0);

  const handleAllocationChange = (symbol: string, value: number) => {
    setConfig(prev => ({
      ...prev,
      allocation: {
        ...prev.allocation,
        [symbol]: value / 100,
      },
    }));
  };

  const handleCalculate = async () => {
    if (Math.abs(totalAllocation - 1) > 0.001) {
      message.warning('配比总和必须等于100%');
      return;
    }

    setCalculating(true);
    try {
      // 调用后端 API
      const allocation: Record<string, number> = {};
      Object.entries(config.allocation).forEach(([symbol, weight]) => {
        if (weight > 0) {
          allocation[symbol] = Math.round(weight * 100);
        }
      });

      const response = await etfAPI.getPortfolioAnalysis(
        allocation,
        config.total_investment,
        config.tax_rate / 100
      );

      if (response.success && response.data) {
        // 转换后端数据为前端格式
        const backendData = response.data;
        const annualDividendBeforeTax = backendData.annual_dividend_before_tax || 0;
        const taxRate = backendData.tax_rate / 100 || (config.tax_rate / 100);
        const dividendTax = backendData.dividend_tax || (annualDividendBeforeTax * taxRate);
        const annualDividendAfterTax = backendData.annual_dividend || (annualDividendBeforeTax - dividendTax);
        
        setPortfolio({
          total_investment: config.total_investment,
          total_value: backendData.total_value || config.total_investment,
          total_return: backendData.total_return || 0,
          total_return_percent: backendData.total_return_pct || 0,
          annual_dividend_before_tax: annualDividendBeforeTax,
          annual_dividend_after_tax: annualDividendAfterTax,
          dividend_tax: dividendTax,
          tax_rate: taxRate * 100,
          weighted_dividend_yield: backendData.dividend_yield || 0,
          total_return_with_dividend: backendData.after_tax_return || 0,
          total_return_with_dividend_percent: (backendData.after_tax_return || 0) / config.total_investment * 100,
          holdings: (backendData.holdings || []).map((h: PortfolioHolding) => ({
            symbol: h.symbol,
            name: h.name || h.symbol + ' ETF',
            weight: h.weight,
            investment: h.investment,
            current_price: h.current_price || 0,
            shares: h.shares || 0,
            current_value: h.current_value || h.investment,
            capital_gain: h.capital_gain || 0,
            capital_gain_percent: h.capital_gain_percent || 0,
            total_return: h.total_return || 0,
            volatility: h.volatility || 0,
            dividend_yield: h.dividend_yield || 0,
            annual_dividend_before_tax: h.annual_dividend_before_tax || 0,
            annual_dividend_after_tax: h.annual_dividend_after_tax || 0,
          })),
        });
        message.success('计算完成');
      } else {
        message.error('获取数据失败');
      }
    } catch (error) {
      message.error('计算失败: ' + (error as Error).message);
    } finally {
      setCalculating(false);
    }
  };

  const applyPreset = (preset: string) => {
    const presets: { [key: string]: { [key: string]: number } } = {
      '433': { SCHD: 0.4, SPYD: 0.3, JEPQ: 0.3, JEPI: 0, VYM: 0 },
      '442': { SCHD: 0.4, SPYD: 0.4, JEPQ: 0.2, JEPI: 0, VYM: 0 },
      'balanced': { SCHD: 0.3, SPYD: 0.2, JEPQ: 0.15, JEPI: 0.2, VYM: 0.15 },
      'conservative': { SCHD: 0.4, SPYD: 0.2, JEPQ: 0.1, JEPI: 0.2, VYM: 0.1 },
    };

    if (presets[preset]) {
      setConfig(prev => ({
        ...prev,
        allocation: presets[preset],
      }));
      message.success('已应用预设配置');
    }
  };

  return (
    <Layout>
      <PageHeader>
        <h2>
          <WalletOutlined />
          投资组合分析
        </h2>
        <Button
          type="primary"
          icon={<SyncOutlined spin={calculating} />}
          onClick={handleCalculate}
          loading={calculating}
        >
          计算分析
        </Button>
      </PageHeader>

      {/* 配置区域 */}
      <ConfigCard
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <CalculatorOutlined />
            <span>组合配置</span>
          </div>
        }
      >
        <Row gutter={[24, 24]}>
          <Col xs={24} lg={12}>
            <div style={{ marginBottom: 16 }}>
              <label style={{ display: 'block', marginBottom: 8, fontWeight: 500 }}>
                投资金额
              </label>
              <InputNumber
                style={{ width: '100%' }}
                prefix={<DollarOutlined />}
                value={config.total_investment}
                onChange={(value) => setConfig(prev => ({ ...prev, total_investment: value || 0 }))}
                formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                parser={(value) => Number(value!.replace(/\$\s?|(,*)/g, ''))}
                min={1000}
                step={1000}
                size="large"
              />
            </div>

            <div style={{ marginBottom: 16 }}>
              <label style={{ display: 'block', marginBottom: 8, fontWeight: 500 }}>
                ETF配比
              </label>

              {Object.entries(config.allocation).map(([symbol, weight]) => (
                <AllocationItem key={symbol}>
                  <span className="symbol">{symbol}</span>
                  <Slider
                    className="slider"
                    value={weight * 100}
                    onChange={(value) => handleAllocationChange(symbol, value)}
                    min={0}
                    max={100}
                    step={1}
                  />
                  <Space.Compact>
                    <InputNumber
                      className="input"
                      value={weight * 100}
                      onChange={(value) => handleAllocationChange(symbol, value || 0)}
                      min={0}
                      max={100}
                      precision={0}
                    />
                    <span style={{ padding: '0 8px', display: 'flex', alignItems: 'center' }}>%</span>
                  </Space.Compact>
                </AllocationItem>
              ))}

              <div style={{
                textAlign: 'right',
                padding: '8px 0',
                color: Math.abs(totalAllocation - 1) < 0.001 ? theme.colors.success : theme.colors.danger,
                fontWeight: 500,
              }}>
                总配比: {(totalAllocation * 100).toFixed(0)}%
                {Math.abs(totalAllocation - 1) > 0.001 && ' (需要调整为100%)'}
              </div>
            </div>
          </Col>

          <Col xs={24} lg={12}>
            <div style={{ marginBottom: 16 }}>
              <label style={{ display: 'block', marginBottom: 8, fontWeight: 500 }}>
                快速预设
              </label>
              <div>
                <PresetButton onClick={() => applyPreset('433')}>
                  4:3:3 组合
                </PresetButton>
                <PresetButton onClick={() => applyPreset('442')}>
                  4:4:2 组合
                </PresetButton>
                <PresetButton onClick={() => applyPreset('balanced')}>
                  平衡型
                </PresetButton>
                <PresetButton onClick={() => applyPreset('conservative')}>
                  稳健型
                </PresetButton>
              </div>
            </div>

            <div>
              <label style={{ display: 'block', marginBottom: 8, fontWeight: 500 }}>
                股息税率
              </label>
              <Space.Compact style={{ width: '100%' }}>
                <InputNumber
                  style={{ flex: 1 }}
                  prefix={<PercentageOutlined />}
                  value={config.tax_rate}
                  onChange={(value) => setConfig(prev => ({ ...prev, tax_rate: value || 0 }))}
                  min={0}
                  max={50}
                  precision={1}
                />
                <span style={{ padding: '0 8px', display: 'flex', alignItems: 'center' }}>%</span>
              </Space.Compact>
            </div>
          </Col>
        </Row>
      </ConfigCard>

      {/* 汇总数据 */}
      <SummaryRow>
        <SummaryCard>
          <div className="label">总投资</div>
          <div className="value">${portfolio.total_investment.toLocaleString()}</div>
        </SummaryCard>
        <SummaryCard $color={theme.colors.success}>
          <div className="label">当前价值</div>
          <div className="value" style={{ color: theme.colors.success }}>
            ${portfolio.total_value.toLocaleString()}
          </div>
        </SummaryCard>
        <SummaryCard $color={portfolio.total_return >= 0 ? theme.colors.success : theme.colors.danger}>
          <div className="label">资本利得</div>
          <div className="value" style={{ color: portfolio.total_return >= 0 ? theme.colors.success : theme.colors.danger }}>
            ${portfolio.total_return.toLocaleString()}
          </div>
          <div className="change">{portfolio.total_return_percent.toFixed(2)}%</div>
        </SummaryCard>
        <SummaryCard $color={theme.colors.warning}>
          <div className="label">税前年股息</div>
          <div className="value" style={{ color: theme.colors.warning }}>
            ${portfolio.annual_dividend_before_tax.toLocaleString()}
          </div>
        </SummaryCard>
        <SummaryCard $color={theme.colors.danger}>
          <div className="label">股息税 ({portfolio.tax_rate}%)</div>
          <div className="value" style={{ color: theme.colors.danger }}>
            -${portfolio.dividend_tax.toLocaleString()}
          </div>
        </SummaryCard>
        <SummaryCard $color={theme.colors.success}>
          <div className="label">税后年股息</div>
          <div className="value" style={{ color: theme.colors.success }}>
            ${portfolio.annual_dividend_after_tax.toLocaleString()}
          </div>
          <div className="change">收益率 {portfolio.weighted_dividend_yield.toFixed(2)}%</div>
        </SummaryCard>
      </SummaryRow>

      {/* 持仓明细表格 */}
      <Card
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <PieChartOutlined />
            <span>持仓明细</span>
          </div>
        }
        style={{ marginBottom: 20, boxShadow: theme.shadows.card }}
      >
        <StyledTable
          dataSource={portfolio.holdings}
          rowKey="symbol"
          pagination={false}
          size="middle"
        >
          <Table.Column
            title="ETF"
            dataIndex="symbol"
            key="symbol"
            render={(text: string, record: PortfolioHolding) => (
              <div>
                <strong>{text}</strong>
                <br />
                <small style={{ color: theme.colors.textMuted }}>{record.name}</small>
              </div>
            )}
          />
          <Table.Column
            title="配比"
            dataIndex="weight"
            key="weight"
            render={(value: number) => <Progress percent={value} size="small" showInfo={false} />}
          />
          <Table.Column
            title="投资金额"
            dataIndex="investment"
            key="investment"
            align="right"
            render={(value: number) => `$${value.toLocaleString()}`}
          />
          <Table.Column
            title="当前价格"
            dataIndex="current_price"
            key="current_price"
            align="right"
            render={(value: number) => `$${(value as number).toFixed(2)}`}
          />
          <Table.Column
            title="持有份数"
            dataIndex="shares"
            key="shares"
            align="right"
            render={(value: number) => (value as number).toFixed(2)}
          />
          <Table.Column
            title="当前价值"
            dataIndex="current_value"
            key="current_value"
            align="right"
            render={(value: number) => `$${(value as number).toLocaleString()}`}
          />
          <Table.Column
            title="资本利得"
            dataIndex="capital_gain"
            key="capital_gain"
            align="right"
            render={(value: number) => {
              const numValue = value as number;
              return (
                <span style={{ color: numValue >= 0 ? theme.colors.success : theme.colors.danger }}>
                  ${numValue.toLocaleString()}
                </span>
              );
            }}
          />
          <Table.Column
            title="股息率"
            dataIndex="dividend_yield"
            key="dividend_yield"
            align="right"
            render={(value?: number) => value ? `${(value as number).toFixed(2)}%` : '-'}
          />
          <Table.Column
            title="税后年股息"
            dataIndex="annual_dividend_after_tax"
            key="annual_dividend_after_tax"
            align="right"
            render={(value: number) => {
              const numValue = value as number;
              return (
                <span style={{ color: theme.colors.success }}>${numValue.toLocaleString()}</span>
              );
            }}
          />
        </StyledTable>
      </Card>

      {/* 持仓分布图表 */}
      {portfolio.holdings.length > 0 && (
        <Card
          title={
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <PieChartOutlined />
              <span>持仓可视化</span>
            </div>
          }
          style={{ boxShadow: theme.shadows.card, marginTop: 20 }}
        >
          <Tabs
            items={[
              {
                key: 'pie',
                label: (
                  <span>
                    <PieChartOutlined />
                    持仓分布
                  </span>
                ),
                children: (
                  <HoldingPieChart
                    data={portfolio.holdings.map(h => ({
                      symbol: h.symbol,
                      name: h.name,
                      weight: h.weight,
                      value: h.current_value,
                    }))}
                    title=""
                  />
                ),
              },
            ]}
          />
        </Card>
      )}

      {/* 综合收益 */}
      <Card
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <LineChartOutlined />
            <span>综合收益分析</span>
          </div>
        }
        style={{ boxShadow: theme.shadows.card }}
      >
        <Row gutter={[24, 24]}>
          <Col xs={24} md={8}>
            <Statistic
              title="综合总收益"
              value={portfolio.total_return_with_dividend}
              precision={2}
              prefix="$"
              styles={{
                content: {
                  color: portfolio.total_return_with_dividend >= 0 ? theme.colors.success : theme.colors.danger,
                  fontSize: 32,
                }
              }}
            />
            <div style={{ marginTop: 8, color: theme.colors.textSecondary }}>
              包含资本利得 + 税后股息
            </div>
          </Col>
          <Col xs={24} md={8}>
            <Statistic
              title="综合收益率"
              value={portfolio.total_return_with_dividend_percent}
              precision={2}
              suffix="%"
              styles={{
                content: {
                  color: portfolio.total_return_with_dividend_percent >= 0 ? theme.colors.success : theme.colors.danger,
                  fontSize: 32,
                }
              }}
            />
            <div style={{ marginTop: 8, color: theme.colors.textSecondary }}>
              年化收益率
            </div>
          </Col>
          <Col xs={24} md={8}>
            <div style={{ padding: '20px', background: theme.colors.background, borderRadius: 8 }}>
              <h4 style={{ margin: '0 0 12px 0' }}>收益构成</h4>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                <span>资本利得:</span>
                <span style={{ color: portfolio.total_return >= 0 ? theme.colors.success : theme.colors.danger }}>
                  ${portfolio.total_return.toLocaleString()}
                </span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>税后股息:</span>
                <span style={{ color: theme.colors.success }}>
                  ${portfolio.annual_dividend_after_tax.toLocaleString()}
                </span>
              </div>
            </div>
          </Col>
        </Row>
      </Card>
    </Layout>
  );
};

export default PortfolioAnalysis;

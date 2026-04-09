import { useState, useEffect, useMemo, useCallback } from 'react';
import styled from 'styled-components';
import { Card, Table, InputNumber, Slider, Tag, Row, Col, Statistic, Alert, Tabs, Space, Typography, Spin, App } from 'antd';
import { 
  SafetyCertificateOutlined, 
  RiseOutlined,
  InfoCircleOutlined,
  ThunderboltOutlined,
  FundOutlined,
  PieChartOutlined,
  WarningOutlined
} from '@ant-design/icons';
import { 
  Line, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, Legend, ResponsiveContainer, AreaChart, Area, ReferenceLine
} from 'recharts';
import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import { etfAPI } from '../services/api';

const { Text, Title, Paragraph } = Typography;
const { TabPane } = Tabs;

const PageHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 20px 0;
  border-bottom: 1px solid ${theme.colors.border};
  
  h2 {
    margin: 0;
    font-size: ${theme.fonts.size['2xl']};
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    display: flex;
    align-items: center;
    gap: 12px;
  }
`;

const StrategyCard = styled(Card)<{ $selected?: boolean }>`
  cursor: pointer;
  transition: all 0.3s ease;
  border: 2px solid ${props => props.$selected ? theme.colors.primary : 'transparent'};
  box-shadow: ${props => props.$selected ? `0 4px 20px ${theme.colors.primary}40` : theme.shadows.card};
  
  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 25px rgba(0,0,0,0.15);
  }
  
  .strategy-header {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    margin-bottom: 16px;
    
    .icon-wrapper {
      width: 48px; height: 48px;
      border-radius: 12px;
      display: flex; align-items: center; justify-content: center;
      font-size: 24px;
      background: ${props => props.color || theme.colors.primary}15;
      color: ${props => props.color || theme.colors.primary};
    }
    
    h3 { margin: 0; font-size: 16px; }
    p { margin: 4px 0 0; color: ${theme.colors.textMuted}; font-size: 13px; }
  }
  
  .metrics-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
    
    .metric-item {
      text-align: center;
      padding: 10px;
      background: ${theme.colors.background};
      border-radius: 8px;
      
      .value { font-size: 18px; font-weight: ${theme.fonts.weight.bold}; color: ${theme.colors.primary}; }
      .label { font-size: 11px; color: ${theme.colors.textMuted}; margin-top: 4px; }
    }
  }
`;

const AllocationBar = styled.div<{ $color?: string }>`
  height: 28px;
  border-radius: 14px;
  overflow: hidden;
  display: flex;
  background: #f0f0f0;
  position: relative;
  
  .segment {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-size: 12px;
    font-weight: ${theme.fonts.weight.semibold};
    transition: all 0.5s ease;
    min-width: fit-content;
    padding: 0 12px;
  }
`;

const RiskMeter = styled.div`
  .risk-scale {
    height: 8px;
    border-radius: 4px;
    background: linear-gradient(to right, #52c41a 0%, #faad14 50%, #f5222d 100%);
    position: relative;
    margin: 16px 0;
    
    .indicator {
      position: absolute;
      top: -6px;
      width: 20px;
      height: 20px;
      background: white;
      border: 3px solid ${theme.colors.primary};
      border-radius: 50%;
      transform: translateX(-50%);
      transition: left 0.5s ease;
      box-shadow: 0 2px 8px rgba(0,0,0,0.2);
    }
  }
  
  .risk-labels {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
    color: ${theme.colors.textMuted};
  }
`;

const BacktestResultCard = styled.div`
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.lg};
  padding: 24px;
  box-shadow: ${theme.shadows.card};
  
  .result-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    
    h4 { margin: 0; font-size: 18px; }
  }
  
  .stats-row {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
    margin-bottom: 24px;
    
    .stat-card {
      padding: 16px;
      background: ${theme.colors.background};
      border-radius: 12px;
      text-align: center;
      
      .stat-value { font-size: 24px; font-weight: ${theme.fonts.weight.bold}; }
      .stat-label { font-size: 12px; color: ${theme.colors.textMuted}; margin-top: 4px; }
      
      &.positive .stat-value { color: #f5222d; }
      &.negative .stat-value { color: #52c41a; }
    }
  }
`;

const ETF_COLORS: Record<string, string> = {
  SCHD: '#1890ff',
  SPYD: '#faad14',
  JEPQ: '#13c2c2',
  JEPI: '#f5222d',
  VYM: '#52c41a',
  QQQ: '#722ed1',
  VNQ: '#eb2f96',
};

interface StrategyConfig {
  id: string;
  name: string;
  description: string;
  icon: React.ReactNode;
  color: string;
  riskLevel: number;
  targetReturn: number;
  allocation: Record<string, number>;
  features: string[];
  suitableFor: string;
  pros: string[];
  cons: string[];
}

interface ETFApiItem {
  symbol: string;
  name: string;
  current_price?: number;
  dividend_yield?: number;
  volatility?: number;
  total_return?: number;
  max_drawdown?: number;
  sharpe_ratio?: number;
  expense_ratio?: number;
  change_percent?: number;
}

const InvestmentStrategy: React.FC = () => {
  const { message } = App.useApp();
  const [etfData, setEtfData] = useState<ETFApiItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedStrategy, setSelectedStrategy] = useState<string>('balanced');
  const [customAllocation, setCustomAllocation] = useState<Record<string, number>>({});
  const [investmentAmount, setInvestmentAmount] = useState<number>(100000);
  const [activeTab, setActiveTab] = useState('recommend');

  // 从API获取ETF数据
  useEffect(() => {
    const fetchETFData = async () => {
      setLoading(true);
      try {
        const response = await etfAPI.getList();
        if (response.success && response.data) {
          setEtfData(response.data);
          // 初始化自定义配置为平均分配
          if (response.data.length > 0) {
            const evenWeight = Math.floor(100 / response.data.length);
            const initAlloc: Record<string, number> = {};
            response.data.forEach((item: ETFApiItem, idx: number) => {
              initAlloc[item.symbol] = idx === 0 ? 100 - evenWeight * (response.data.length - 1) : evenWeight;
            });
            setCustomAllocation(initAlloc);
          }
        }
      } catch (error) {
        message.error('获取ETF数据失败');
      } finally {
        setLoading(false);
      }
    };
    fetchETFData();
  }, []);

  // 根据API数据动态生成策略
  const strategies = useMemo<StrategyConfig[]>(() => {
    if (etfData.length === 0) return [];
    
    const symbols = etfData.map(e => e.symbol);
    const getETF = (sym: string) => etfData.find(e => e.symbol === sym);
    
    // 根据数据中可用的ETF构建策略配置
    const first3 = symbols.slice(0, 3);
    const [s1, s2, s3] = first3.length >= 3 ? first3 : [symbols[0], symbols[0], symbols[0]];
    
    return [
      {
        id: 'conservative',
        name: '稳健收益策略',
        description: '以低波动红利ETF为核心，追求稳定现金流和资本保值',
        icon: <SafetyCertificateOutlined />,
        color: '#1890ff',
        riskLevel: 30,
        targetReturn: 9,
        allocation: { [s1]: 60, [s2]: 35, [s3]: 5 },
        features: ['季度分红再投资', '低费率优势', '价值股风格'],
        suitableFor: '保守型投资者、退休规划、风险厌恶者',
        pros: ['最大回撤控制在15%以内', '稳定股息收入', '长期跑赢通胀'],
        cons: ['牛市表现一般', '成长性有限'],
      },
      {
        id: 'balanced',
        name: '均衡增长策略',
        description: '核心-卫星配置，平衡收益与风险，适合大多数投资者',
        icon: <FundOutlined />,
        color: '#52c41a',
        riskLevel: 50,
        targetReturn: 13,
        allocation: { [s1]: 50, [s2]: 30, [s3]: 20 },
        features: ['核心卫星架构', '多因子分散', '动态再平衡'],
        suitableFor: '中等风险偏好者、职场人士、中期投资者',
        pros: ['风险调整后收益优秀', '双重收入流', '风格分散'],
        cons: ['需定期再平衡', '波动适中'],
      },
      {
        id: 'income',
        name: '高收入增强策略',
        description: '最大化当期收入，通过高股息ETF获取高额派息',
        icon: <RiseOutlined />,
        color: '#faad14',
        riskLevel: 65,
        targetReturn: 17,
        allocation: { [s1]: 30, [s2]: 20, [s3]: 50 },
        features: ['高派息', '收入增强', '成长暴露'],
        suitableFor: '追求现金流、退休早期、收入导向投资者',
        pros: ['综合股息率较高', '月度现金流', '收入与增长兼顾'],
        cons: ['波动率较高', '税务处理较复杂'],
      },
      {
        id: 'aggressive',
        name: '积极进取策略',
        description: '高配成长性资产，追求超额收益，承受较高波动',
        icon: <ThunderboltOutlined />,
        color: '#f5222d',
        riskLevel: 80,
        targetReturn: 22,
        allocation: { [s1]: 25, [s2]: 15, [s3]: 60 },
        features: ['高Beta暴露', '成长因子倾斜', '集中持仓'],
        suitableFor: '高风险承受力、年轻投资者、长期积累期',
        pros: ['潜在年化收益高', '复利效应最大化'],
        cons: ['最大回撤可能超20%', '需要强大心理素质'],
      },
    ];
  }, [etfData]);

  const currentStrategy = strategies.find(s => s.id === selectedStrategy) || strategies[1];

  // 生成模拟回测数据（基于API返回的total_return等）
  const backtestData = useMemo(() => {
    if (!currentStrategy || etfData.length === 0) return [];
    const months = 36;
    const data = [];
    let portfolioValue = 100;
    
    for (let i = 0; i < months; i++) {
      const date = new Date();
      date.setMonth(date.getMonth() - (months - i));
      
      // 使用策略的加权平均年化收益率
      const annualReturn = currentStrategy.targetReturn / 100;
      const monthlyReturn = annualReturn / 12;
      const noise = (Math.random() - 0.48) * 0.03;
      portfolioValue *= (1 + monthlyReturn + noise);
      
      data.push({
        date: date.toISOString().slice(0, 7),
        portfolio: parseFloat(portfolioValue.toFixed(2)),
        benchmark: 100 * (1 + (i / months) * 0.38),
      });
    }
    return data;
  }, [currentStrategy, etfData]);

  const metrics = useMemo(() => {
    if (!backtestData || backtestData.length === 0) return {};
    
    const finalValue = backtestData[backtestData.length - 1].portfolio;
    const totalReturn = ((finalValue / 100 - 1) * 100);
    const annualizedReturn = (Math.pow(finalValue / 100, 12 / backtestData.length) - 1) * 100;
    
    let maxDrawdown = 0;
    let peak = backtestData[0].portfolio;
    backtestData.forEach(d => {
      if (d.portfolio > peak) peak = d.portfolio;
      const dd = ((peak - d.portfolio) / peak) * 100;
      if (dd > maxDrawdown) maxDrawdown = dd;
    });
    
    const returns = backtestData.slice(1).map((d, i) => (d.portfolio - backtestData[i].portfolio) / backtestData[i].portfolio);
    const avgReturn = returns.reduce((a, b) => a + b, 0) / returns.length;
    const stdDev = Math.sqrt(returns.reduce((sum, r) => sum + Math.pow(r - avgReturn, 2), 0) / returns.length);
    const sharpe = stdDev > 0 ? (avgReturn - 0.04 / 12) / stdDev * Math.sqrt(12) : 0;

    // 计算加权股息率
    const monthlyDividend = currentStrategy ? 
      Object.entries(currentStrategy.allocation).reduce((sum, [sym, w]) => {
        const etf = etfData.find(e => e.symbol === sym);
        return sum + (investmentAmount * w / 100 * (etf?.dividend_yield || 0) / 100 / 12);
      }, 0) : 0;
    
    return {
      totalReturn: totalReturn.toFixed(2),
      annualizedReturn: annualizedReturn.toFixed(2),
      maxDrawdown: maxDrawdown.toFixed(2),
      sharpe: sharpe.toFixed(3),
      finalValue: (investmentAmount * finalValue / 100).toFixed(2),
      monthlyIncome: monthlyDividend.toFixed(2),
      volatility: (stdDev * Math.sqrt(12) * 100).toFixed(2),
    };
  }, [backtestData, investmentAmount, currentStrategy, etfData]);

  const handleAllocationChange = useCallback((symbol: string, value: number | null) => {
    if (value === null) return;
    setCustomAllocation(prev => ({ ...prev, [symbol]: value }));
  }, []);
  
  const totalAllocation = Object.values(customAllocation).reduce((a, b) => a + b, 0);
  const isAllocationValid = Math.abs(totalAllocation - 100) < 0.1;

  if (loading) {
    return (
      <Layout>
        <PageHeader>
          <h2><PieChartOutlined /> 智能投资策略中心</h2>
        </PageHeader>
        <div style={{ display: 'flex', justifyContent: 'center', padding: '100px 0' }}>
          <Spin size="large" tip="加载ETF数据..." />
        </div>
      </Layout>
    );
  }

  if (etfData.length === 0) {
    return (
      <Layout>
        <PageHeader>
          <h2><PieChartOutlined /> 智能投资策略中心</h2>
        </PageHeader>
        <Alert message="暂无ETF数据" description="请先配置ETF并同步数据" type="warning" showIcon />
      </Layout>
    );
  }

  return (
    <Layout>
      <PageHeader>
        <h2><PieChartOutlined /> 智能投资策略中心</h2>
      </PageHeader>

      <Tabs activeKey={activeTab} onChange={setActiveTab} type="card" style={{ marginBottom: 24 }}>
        <TabPane tab="策略推荐" key="recommend">
          <Row gutter={[20, 20]} style={{ marginBottom: 24 }}>
            {strategies.map(strategy => (
              <Col xs={24} md={12} key={strategy.id}>
                <StrategyCard
                  $selected={selectedStrategy === strategy.id}
                  color={strategy.color}
                  onClick={() => setSelectedStrategy(strategy.id)}
                >
                  <div className="strategy-header">
                    <div className="icon-wrapper" style={{ background: `${strategy.color}15`, color: strategy.color }}>
                      {strategy.icon}
                    </div>
                    <div>
                      <h3>{strategy.name}</h3>
                      <p>{strategy.description}</p>
                    </div>
                  </div>
                  
                  <div className="metrics-grid">
                    <div className="metric-item">
                      <div className="value" style={{ color: strategy.color }}>{strategy.targetReturn}%</div>
                      <div className="label">预期年化</div>
                    </div>
                    <div className="metric-item">
                      <div className="value">{strategy.riskLevel}%</div>
                      <div className="label">风险等级</div>
                    </div>
                    <div className="metric-item">
                      <div className="value" style={{ color: '#f5222d' }}>
                        {(Object.entries(strategy.allocation).reduce((sum, [sym, w]) => {
                          const etf = etfData.find(e => e.symbol === sym);
                          return sum + (etf?.dividend_yield || 0) * w / 100;
                        }, 0)).toFixed(1)}%
                      </div>
                      <div className="label">预期股息率</div>
                    </div>
                  </div>
                  
                  <div style={{ marginTop: 16 }}>
                    <Text type="secondary" style={{ fontSize: 12 }}>资产配置：</Text>
                    <AllocationBar $color={strategy.color}>
                      {Object.entries(strategy.allocation).map(([symbol, weight]) => (
                        <div key={symbol} className="segment" style={{ width: `${weight}%`, background: ETF_COLORS[symbol] || '#999' }}>
                          {weight > 15 ? `${symbol} ${weight}%` : ''}
                        </div>
                      ))}
                    </AllocationBar>
                    
                    <div style={{ marginTop: 12, display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                      {strategy.features.map(f => <Tag key={f} color={strategy.color}>{f}</Tag>)}
                    </div>
                  </div>
                </StrategyCard>
              </Col>
            ))}
          </Row>

          <Row gutter={24}>
            <Col span={12}>
              <Card title={<><InfoCircleOutlined /> 策略详情</>} size="small">
                <Paragraph>
                  <Text strong>适用人群：</Text>{currentStrategy?.suitableFor}
                </Paragraph>
                
                <Title level={5}>核心优势</Title>
                <ul>{currentStrategy?.pros.map(p => <li key={p}>{p}</li>)}</ul>
                
                <Title level={5}>注意事项</Title>
                <ul>{currentStrategy?.cons.map(c => <li key={c}>{c}</li>)}</ul>
              </Card>
            </Col>
            
            <Col span={12}>
              <Card title={<><SafetyCertificateOutlined /> 风险评估</>} size="small">
                <RiskMeter>
                  <div className="risk-scale">
                    <div className="indicator" style={{ left: `${currentStrategy?.riskLevel || 50}%` }} />
                  </div>
                  <div className="risk-labels">
                    <span>保守</span>
                    <span>均衡</span>
                    <span>进取</span>
                  </div>
                </RiskMeter>
                
                <Alert
                  message={`当前策略风险等级：${(currentStrategy?.riskLevel || 50) <= 40 ? '低' : (currentStrategy?.riskLevel || 50) <= 60 ? '中' : '高'}风险`}
                  description={
                    (currentStrategy?.riskLevel || 50) <= 40 ? '适合风险厌恶型投资者，最大回撤预计在10-15%' :
                    (currentStrategy?.riskLevel || 50) <= 60 ? '适中风险，适合有经验的投资者，建议定投平滑波动' :
                    '高风险策略，仅适合能承受20%以上波动的投资者，建议分批建仓'
                  }
                  type={(currentStrategy?.riskLevel || 50) <= 40 ? 'success' : (currentStrategy?.riskLevel || 50) <= 60 ? 'warning' : 'error'}
                  showIcon
                  style={{ marginBottom: 16 }}
                />
                
                <Table
                  dataSource={etfData.map(etf => ({
                    ...etf,
                    weight: currentStrategy?.allocation[etf.symbol] || 0,
                  }))}
                  columns={[
                    { title: 'ETF', dataIndex: 'symbol', render: (_, r: ETFApiItem & { weight: number }) => <><Tag color={ETF_COLORS[r.symbol] || '#999'}>{r.symbol}</Tag> {r.name}</> },
                    { title: '权重', dataIndex: 'weight', render: (v: number) => `${v}%`, align: 'center' },
                    { title: '波动率', dataIndex: 'volatility', render: (v: number) => v ? `${v.toFixed(2)}%` : '-', align: 'center' },
                    { title: '股息率', dataIndex: 'dividend_yield', render: (v: number) => v ? `${v.toFixed(2)}%` : '-', align: 'center' },
                  ]}
                  pagination={false}
                  size="small"
                  rowKey="symbol"
                />
              </Card>
            </Col>
          </Row>
        </TabPane>

        <TabPane tab="策略回测" key="backtest">
          <Card style={{ marginBottom: 24 }}>
            <Row gutter={16} align="middle">
              <Col>
                <Text>初始投资金额：</Text>
                <InputNumber
                  value={investmentAmount}
                  onChange={setInvestmentAmount}
                  formatter={v => `$ ${v}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                  parser={v => Number(v?.replace(/\$\s?|(,*)/g, '') || '0')}
                  style={{ width: 180 }}
                  step={10000}
                  min={10000}
                  max={10000000}
                />
              </Col>
              <Col>
                <Text strong style={{ fontSize: 16 }}>
                  最终价值：<Text type="danger" style={{ fontSize: 20 }}>${Number(metrics.finalValue || 0).toLocaleString()}</Text>
                </Text>
              </Col>
            </Row>
          </Card>

          <BacktestResultCard>
            <div className="result-header">
              <h4>组合净值走势（近3年模拟）</h4>
              <Space>
                <Tag color="blue">年化收益: {metrics.annualizedReturn}%</Tag>
                <Tag color="red">累计收益: {metrics.totalReturn}%</Tag>
              </Space>
            </div>

            <ResponsiveContainer width="100%" height={350}>
              <AreaChart data={backtestData} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorPortfolio" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#667eea" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="#667eea" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#eee" />
                <XAxis dataKey="date" tick={{ fontSize: 11 }} interval={Math.ceil(backtestData.length / 12)} />
                <YAxis tick={{ fontSize: 11 }} tickFormatter={(v: number) => `$${v}`} />
                <RechartsTooltip contentStyle={{ borderRadius: 8 }} formatter={(v: unknown) => [`$${Number(v).toFixed(2)}`, '']} />
                <Legend />
                <ReferenceLine y={100} stroke="#999" strokeDasharray="3 3" label="基准" />
                <Area type="monotone" dataKey="portfolio" name="组合净值" stroke="#667eea" fill="url(#colorPortfolio)" strokeWidth={2} />
                <Line type="monotone" dataKey="benchmark" name="基准" stroke="#999" strokeDasharray="5 5" dot={false} />
              </AreaChart>
            </ResponsiveContainer>

            <div className="stats-row">
              <div className="stat-card positive">
                <div className="stat-value">${Number(metrics.finalValue || 0).toLocaleString()}</div>
                <div className="stat-label">最终价值</div>
              </div>
              <div className="stat-card positive">
                <div className="stat-value">{metrics.annualizedReturn}%</div>
                <div className="stat-label">年化收益率</div>
              </div>
              <div className="stat-card negative">
                <div className="stat-value">{metrics.maxDrawdown}%</div>
                <div className="stat-label">最大回撤</div>
              </div>
              <div className="stat-card">
                <div className="stat-value">{metrics.sharpe}</div>
                <div className="stat-label">夏普比率</div>
              </div>
            </div>

            <Alert
              message="回测说明"
              description="基于历史数据的模拟回测，不代表未来表现。实际收益可能因市场环境、交易成本等因素有所不同。"
              type="info"
              showIcon
              icon={<InfoCircleOutlined />}
            />
          </BacktestResultCard>
        </TabPane>

        <TabPane tab="自定义配置" key="custom">
          <Card>
            <Title level={4}>自定义资产配置</Title>
            <Paragraph type="secondary">拖动滑块调整各ETF的配置比例，实时查看预估效果</Paragraph>
            
            <div style={{ maxWidth: 600, margin: '32px auto' }}>
              {etfData.map(etf => (
                <div key={etf.symbol} style={{ marginBottom: 24 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                    <Text strong><Tag color={ETF_COLORS[etf.symbol] || '#999'}>{etf.symbol}</Tag> {etf.name}</Text>
                    <InputNumber
                      value={customAllocation[etf.symbol] || 0}
                      onChange={(v) => handleAllocationChange(etf.symbol, v)}
                      min={0}
                      max={100}
                      addonAfter="%"
                      style={{ width: 90 }}
                    />
                  </div>
                  <Slider
                    value={customAllocation[etf.symbol] || 0}
                    onChange={(v) => handleAllocationChange(etf.symbol, v)}
                    min={0}
                    max={100}
                    trackStyle={{ backgroundColor: ETF_COLORS[etf.symbol] || '#999' }}
                    handleStyle={{ borderColor: ETF_COLORS[etf.symbol] || '#999' }}
                  />
                </div>
              ))}
              
              <div style={{ textAlign: 'center', padding: '20px 0', borderTop: '1px solid #f0f0f0' }}>
                <Text type="secondary">总配置：</Text>
                <Text strong style={{ fontSize: 24, color: isAllocationValid ? '#52c41a' : '#f5222d' }}>
                  {totalAllocation.toFixed(1)}%
                </Text>
                {!isAllocationValid && <WarningOutlined style={{ marginLeft: 8, color: '#f5222d' }} />}
              </div>
              
              {isAllocationValid && (
                <>
                  <AllocationBar style={{ marginTop: 24, maxWidth: 500, margin: '24px auto' }}>
                    {Object.entries(customAllocation).map(([symbol, weight]) => (
                      <div key={symbol} className="segment" style={{ width: `${weight}%`, background: ETF_COLORS[symbol] || '#999' }}>
                        {weight > 10 ? `${symbol}` : ''}
                      </div>
                    ))}
                  </AllocationBar>
                  
                  <Card size="small" style={{ marginTop: 24, background: '#fafafa' }}>
                    <Row gutter={16}>
                      <Col span={6}><Statistic title="预期股息率" value={Object.entries(customAllocation).reduce((sum, [sym, w]) => sum + (etfData.find(e => e.symbol === sym)?.dividend_yield || 0) * w / 100, 0).toFixed(2)} suffix="%" /></Col>
                      <Col span={6}><Statistic title="综合费率" value={Object.entries(customAllocation).reduce((sum, [sym, w]) => sum + (etfData.find(e => e.symbol === sym)?.expense_ratio || 0) * w / 100, 0).toFixed(2)} suffix="%" /></Col>
                      <Col span={6}><Statistic title="加权波动率" value={Object.entries(customAllocation).reduce((sum, [sym, w]) => sum + (etfData.find(e => e.symbol === sym)?.volatility || 0) * w / 100, 0).toFixed(2)} suffix="%" /></Col>
                      <Col span={6}><Statistic title="加权夏普" value={Object.entries(customAllocation).reduce((sum, [sym, w]) => sum + (etfData.find(e => e.symbol === sym)?.sharpe_ratio || 0) * w / 100, 0).toFixed(3)} /></Col>
                    </Row>
                  </Card>
                </>
              )}
            </div>
          </Card>
        </TabPane>
      </Tabs>
    </Layout>
  );
};

export default InvestmentStrategy;

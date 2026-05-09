# ETF-Insight v2.7 前后端一体化实施方案

**版本**: v2.7
**创建日期**: 2026-04-25
**文档状态**: 📝 待实施
**目标**: 实现因子择时、Alpha观点生成、Black-Litterman优化、风险预算管理的完整前后端集成

---

## 📋 目录

1. [实施概述](#实施概述)
2. [前端页面改造方案](#前端页面改造方案)
3. [后端API接口规范](#后端api接口规范)
4. [数据交互协议](#数据交互协议)
5. [前后端联调方案](#前后端联调方案)
6. [兼容性处理策略](#兼容性处理策略)
7. [分阶段实施计划](#分阶段实施计划)
8. [Kimi2.6接口接入指南](#kimi26接口接入指南)

---

## 实施概述

### 核心目标

将v2.7新增的数据层能力（因子择时、Alpha观点、Black-Litterman、风险预算）完整集成到前端界面，实现两个闭环：

1. **闭环一**：因子择时 → Alpha观点生成 → Black-Litterman优化 → 组合权重调整
2. **闭环二**：风险预算配置 → CVaR计算 → 风险贡献分解 → 组合优化

### 技术栈

**前端**：
- React 18 + TypeScript
- Ant Design 5.x
- Recharts（图表库）
- Styled Components
- React Query（数据缓存）

**后端**：
- Go 1.26+
- Gin Framework
- GORM
- PostgreSQL/SQLite

**AI集成**：
- Kimi2.6 API（Moonshot AI）
- 用于智能观点生成和市场分析

---

## 前端页面改造方案

### 1. 新增页面组件

#### 1.1 因子择时信号页面 (`FactorTiming.tsx`)

**位置**: `/frontend/src/pages/FactorTiming.tsx`

**功能模块**：
```typescript
interface FactorTimingPageProps {
  // 因子选择器
  factorSelector: {
    factors: ['Mkt-RF', 'SMB', 'HML', 'RMW', 'CMA'];
    selectedFactor: string;
    onFactorChange: (factor: string) => void;
  };

  // 信号展示
  signalDisplay: {
    currentSignal: FactorTimingSignal;
    signalHistory: FactorTimingSignal[];
    visualization: 'chart' | 'table' | 'card';
  };

  // 观点生成
  viewGeneration: {
    onGenerateView: (signal: FactorTimingSignal) => Promise<AlphaView>;
    autoGenerate: boolean;
    targetAsset: string;
  };
}
```

**页面布局**：
```
┌─────────────────────────────────────────────────────────┐
│  因子择时信号分析                                          │
├─────────────────────────────────────────────────────────┤
│  [因子选择: Mkt-RF ▼]  [回看天数: 60]  [计算信号]         │
├─────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  MA斜率      │  │  Z-Score     │  │  百分位数    │  │
│  │  +0.0023     │  │  +1.85       │  │  85.2%       │  │
│  │  ↗ 上升趋势  │  │  偏高        │  │  历史高位    │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
├─────────────────────────────────────────────────────────┤
│  信号强度: [强正向]  预期收益: +2.3%  置信度: 75%        │
├─────────────────────────────────────────────────────────┤
│  [生成Alpha观点]  [资产: SPY ▼]  [自动应用]             │
└─────────────────────────────────────────────────────────┘
```

**核心组件代码**：
```typescript
import React, { useState, useEffect } from 'react';
import { Card, Select, Button, Row, Col, Statistic, Tag, message } from 'antd';
import { ArrowUpOutlined, ArrowDownOutlined, MinusOutlined } from '@ant-design/icons';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import styled from 'styled-components';
import { factorTimingAPI } from '../services/api';

const Container = styled.div`
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
`;

const SignalCard = styled(Card)`
  .ant-statistic-title {
    font-size: 14px;
    color: #8c8c8c;
  }
  .ant-statistic-content {
    font-size: 24px;
    font-weight: 600;
  }
`;

const FactorTiming: React.FC = () => {
  const [selectedFactor, setSelectedFactor] = useState('Mkt-RF');
  const [lookbackDays, setLookbackDays] = useState(60);
  const [signal, setSignal] = useState<FactorTimingSignal | null>(null);
  const [signalHistory, setSignalHistory] = useState<FactorTimingSignal[]>([]);
  const [loading, setLoading] = useState(false);
  const [targetAsset, setTargetAsset] = useState('SPY');

  const calculateSignal = async () => {
    setLoading(true);
    try {
      const result = await factorTimingAPI.calculateSignal(selectedFactor, lookbackDays);
      setSignal(result);
      message.success('信号计算成功');
    } catch (error) {
      message.error('信号计算失败: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const generateView = async () => {
    if (!signal) {
      message.warning('请先计算信号');
      return;
    }

    try {
      const view = await factorTimingAPI.generateView(signal, targetAsset);
      message.success(`Alpha观点已生成: ${targetAsset} 预期收益 ${view.view_return.toFixed(2)}%`);
    } catch (error) {
      message.error('观点生成失败: ' + error.message);
    }
  };

  const getSignalIcon = (strength: SignalStrength) => {
    switch (strength) {
      case 'strong_positive':
        return <ArrowUpOutlined style={{ color: '#52c41a', fontSize: 32 }} />;
      case 'weak_positive':
        return <ArrowUpOutlined style={{ color: '#73d13d', fontSize: 28 }} />;
      case 'neutral':
        return <MinusOutlined style={{ color: '#8c8c8c', fontSize: 24 }} />;
      case 'weak_negative':
        return <ArrowDownOutlined style={{ color: '#ff7875', fontSize: 28 }} />;
      case 'strong_negative':
        return <ArrowDownOutlined style={{ color: '#ff4d4f', fontSize: 32 }} />;
    }
  };

  return (
    <Container>
      <h1>因子择时信号分析</h1>

      <Card style={{ marginBottom: 24 }}>
        <Row gutter={16} align="middle">
          <Col>
            <span>选择因子: </span>
            <Select value={selectedFactor} onChange={setSelectedFactor} style={{ width: 120 }}>
              <Select.Option value="Mkt-RF">市场因子</Select.Option>
              <Select.Option value="SMB">规模因子</Select.Option>
              <Select.Option value="HML">价值因子</Select.Option>
              <Select.Option value="RMW">盈利因子</Select.Option>
              <Select.Option value="CMA">投资因子</Select.Option>
            </Select>
          </Col>
          <Col>
            <span>回看天数: </span>
            <InputNumber value={lookbackDays} onChange={setLookbackDays} min={30} max={252} />
          </Col>
          <Col>
            <Button type="primary" onClick={calculateSignal} loading={loading}>
              计算信号
            </Button>
          </Col>
        </Row>
      </Card>

      {signal && (
        <>
          <Row gutter={16} style={{ marginBottom: 24 }}>
            <Col span={8}>
              <SignalCard>
                <Statistic
                  title="MA斜率 (60日)"
                  value={signal.ma_slope_60}
                  precision={4}
                  valueStyle={{ color: signal.ma_slope_60 > 0 ? '#3f8600' : '#cf1322' }}
                />
                <Tag color={signal.ma_slope_60 > 0 ? 'green' : 'red'}>
                  {signal.ma_slope_60 > 0 ? '上升趋势' : '下降趋势'}
                </Tag>
              </SignalCard>
            </Col>
            <Col span={8}>
              <SignalCard>
                <Statistic
                  title="Z-Score"
                  value={signal.z_score}
                  precision={2}
                  valueStyle={{ color: Math.abs(signal.z_score) > 1.5 ? '#1890ff' : '#8c8c8c' }}
                />
                <Tag color={Math.abs(signal.z_score) > 2 ? 'blue' : 'default'}>
                  {Math.abs(signal.z_score) > 2 ? '极端值' : '正常范围'}
                </Tag>
              </SignalCard>
            </Col>
            <Col span={8}>
              <SignalCard>
                <Statistic
                  title="百分位数"
                  value={signal.percentile}
                  suffix="%"
                  precision={1}
                />
                <Tag color={signal.percentile > 80 ? 'red' : signal.percentile < 20 ? 'green' : 'default'}>
                  {signal.percentile > 80 ? '历史高位' : signal.percentile < 20 ? '历史低位' : '中等水平'}
                </Tag>
              </SignalCard>
            </Col>
          </Row>

          <Card title="信号强度" style={{ marginBottom: 24 }}>
            <Row gutter={16} align="middle">
              <Col span={4}>
                {getSignalIcon(signal.signal_strength)}
              </Col>
              <Col span={6}>
                <Statistic title="预期收益" value={signal.expected_return} precision={2} suffix="%" />
              </Col>
              <Col span={6}>
                <Statistic title="置信度" value={signal.confidence} precision={0} suffix="%" />
              </Col>
              <Col span={8}>
                <Tag color={
                  signal.signal_strength === 'strong_positive' ? 'green' :
                  signal.signal_strength === 'weak_positive' ? 'lime' :
                  signal.signal_strength === 'neutral' ? 'default' :
                  signal.signal_strength === 'weak_negative' ? 'orange' : 'red'
                } style={{ fontSize: 16, padding: '4px 12px' }}>
                  {signal.signal_strength.replace('_', ' ').toUpperCase()}
                </Tag>
              </Col>
            </Row>
          </Card>

          <Card title="生成Alpha观点" style={{ marginBottom: 24 }}>
            <Row gutter={16} align="middle">
              <Col>
                <span>目标资产: </span>
                <Select value={targetAsset} onChange={setTargetAsset} style={{ width: 150 }}>
                  <Select.Option value="SPY">SPY (标普500)</Select.Option>
                  <Select.Option value="QQQ">QQQ (纳斯达克100)</Select.Option>
                  <Select.Option value="IWM">IWM (罗素2000)</Select.Option>
                  <Select.Option value="EEM">EEM (新兴市场)</Select.Option>
                </Select>
              </Col>
              <Col>
                <Button type="primary" onClick={generateView}>
                  生成Alpha观点
                </Button>
              </Col>
            </Row>
          </Card>
        </>
      )}
    </Container>
  );
};

export default FactorTiming;
```

#### 1.2 Alpha观点管理页面 (`AlphaViews.tsx`)

**位置**: `/frontend/src/pages/AlphaViews.tsx`

**功能模块**：
```typescript
interface AlphaViewsPageProps {
  // 观点列表
  viewList: {
    views: AlphaView[];
    filters: {
      status: 'active' | 'expired' | 'validated';
      asset: string;
      method: 'factor_timing' | 'momentum' | 'mean_reversion';
    };
    pagination: {
      page: number;
      pageSize: number;
      total: number;
    };
  };

  // 观点详情
  viewDetail: {
    selectedView: AlphaView;
    performance: AlphaViewPerformance;
    relatedSignals: FactorTimingSignal[];
  };

  // 观点操作
  viewActions: {
    onCreate: (view: CreateAlphaViewRequest) => Promise<void>;
    onUpdate: (id: number, view: UpdateAlphaViewRequest) => Promise<void>;
    onDeactivate: (id: number) => Promise<void>;
    onValidate: (id: number) => Promise<void>;
  };
}
```

**页面布局**：
```
┌─────────────────────────────────────────────────────────┐
│  Alpha观点管理                                           │
├─────────────────────────────────────────────────────────┤
│  [状态: 全部 ▼]  [资产: 全部 ▼]  [方法: 全部 ▼]  [搜索]  │
├─────────────────────────────────────────────────────────┤
│  ┌───────────────────────────────────────────────────┐  │
│  │ 观点列表                                            │  │
│  │ ┌─────┬──────┬────────┬──────┬──────┬────────┐  │  │
│  │ │ ID  │ 资产 │ 类型   │ 收益 │ 置信 │ 状态   │  │  │
│  │ ├─────┼──────┼────────┼──────┼──────┼────────┤  │  │
│  │ │ 1   │ SPY  │ 绝对   │ +2.3%│ 75%  │ Active │  │  │
│  │ │ 2   │ QQQ  │ 相对   │ +1.5%│ 60%  │ Active │  │  │
│  │ │ 3   │ IWM  │ 绝对   │ -1.2%│ 80%  │Expired │  │  │
│  │ └─────┴──────┴────────┴──────┴──────┴────────┘  │  │
│  └───────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────┤
│  [创建新观点]  [批量导入]  [导出CSV]                      │
└─────────────────────────────────────────────────────────┘
```

#### 1.3 Black-Litterman配置页面 (`BlackLittermanConfig.tsx`)

**位置**: `/frontend/src/pages/BlackLittermanConfig.tsx`

**功能模块**：
```typescript
interface BlackLittermanConfigPageProps {
  // 配置管理
  configManagement: {
    configs: BlackLittermanConfig[];
    selectedConfig: BlackLittermanConfig | null;
    onCreate: (config: CreateBLConfigRequest) => Promise<void>;
    onUpdate: (id: number, config: UpdateBLConfigRequest) => Promise<void>;
  };

  // 观点集成
  viewIntegration: {
    availableViews: AlphaView[];
    selectedViews: AlphaView[];
    onViewSelect: (views: AlphaView[]) => void;
  };

  // 计算结果
  calculationResult: {
    posteriorReturns: BLPosteriorReturn;
    efficientFrontier: EfficientFrontierPoint[];
    onCalculate: () => Promise<void>;
  };
}
```

**页面布局**：
```
┌─────────────────────────────────────────────────────────┐
│  Black-Litterman模型配置                                 │
├─────────────────────────────────────────────────────────┤
│  [配置名称: 默认配置]  [风险厌恶系数: 2.5]                │
│  [先验类型: 市值加权 ▼]  [Omega方法: Idzorek ▼]         │
├─────────────────────────────────────────────────────────┤
│  市场权重配置                                            │
│  ┌───────────────────────────────────────────────────┐  │
│  │ SPY: 40%  QQQ: 30%  IWM: 20%  EEM: 10%           │  │
│  │ [可视化饼图]                                       │  │
│  └───────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────┤
│  Alpha观点选择                                           │
│  ┌───────────────────────────────────────────────────┐  │
│  │ ☑ SPY: +2.3% (置信度: 75%)                        │  │
│  │ ☑ QQQ: +1.5% (置信度: 60%)                        │  │
│  │ ☐ IWM: -1.2% (置信度: 80%)                        │  │
│  └───────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────┤
│  [计算后验收益]  [应用到组合优化]                         │
└─────────────────────────────────────────────────────────┘
```

#### 1.4 风险预算管理页面 (`RiskBudget.tsx`)

**位置**: `/frontend/src/pages/RiskBudget.tsx`

**功能模块**：
```typescript
interface RiskBudgetPageProps {
  // 预算配置
  budgetConfig: {
    config: RiskBudgetConfig;
    onUpdate: (config: UpdateRiskBudgetRequest) => Promise<void>;
  };

  // CVaR计算
  cvarCalculation: {
    params: {
      confidenceLevel: number;
      timeHorizon: number;
      method: 'historical' | 'parametric' | 'monte_carlo';
    };
    result: {
      cvar: number;
      var: number;
      riskContributions: RiskContribution[];
    };
    onCalculate: () => Promise<void>;
  };

  // 蒙特卡洛模拟
  monteCarloSimulation: {
    params: {
      simulations: number;
      timeSteps: number;
    };
    result: MonteCarloSimulation;
    onRun: () => Promise<void>;
  };
}
```

### 2. 现有页面改造

#### 2.1 组合优化页面增强 (`PortfolioOptimization.tsx`)

**新增功能**：
- 集成Black-Litterman优化器
- 风险预算约束选项
- 实时风险贡献展示

**改造代码片段**：
```typescript
// 在PortfolioOptimization.tsx中新增BL优化选项卡
<TabPane tab="Black-Litterman优化" key="black-litterman">
  <BlackLittermanOptimizer
    assets={selectedETFs}
    views={activeViews}
    onOptimize={handleBLOptimization}
  />
</TabPane>

// 新增风险预算约束组件
<RiskBudgetConstraints
  budgetConfig={riskBudgetConfig}
  onChange={setRiskBudgetConfig}
/>
```

#### 2.2 因子分析页面增强 (`FactorAnalysis.tsx`)

**新增功能**：
- 因子择时信号展示
- 一键生成Alpha观点
- 观点历史表现追踪

**改造代码片段**：
```typescript
// 在FactorAnalysis.tsx中新增择时信号卡片
<Card title="因子择时信号" style={{ marginBottom: 16 }}>
  <FactorTimingSignalCard
    factor={selectedFactor}
    onGenerateView={handleGenerateViewFromSignal}
  />
</Card>
```

---

## 后端API接口规范

### 1. 因子择时API

#### 1.1 计算因子择时信号

**接口**: `POST /api/factor/timing/calculate`

**请求参数**:
```json
{
  "factor_name": "Mkt-RF",
  "lookback_days": 60,
  "calculation_date": "2026-04-25"
}
```

**响应数据**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "factor_name": "Mkt-RF",
    "signal_date": "2026-04-25T00:00:00Z",
    "ma_slope_60": 0.0023,
    "z_score": 1.85,
    "percentile": 85.2,
    "signal_strength": "strong_positive",
    "signal_score": 4,
    "expected_return": 0.023,
    "confidence": 75.0,
    "created_at": "2026-04-25T10:30:00Z"
  }
}
```

**Go实现**:
```go
// handlers/factor_timing_handler.go
func (h *FactorTimingHandler) CalculateTimingSignal(c *gin.Context) {
    var req CalculateTimingSignalRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
        return
    }

    signal, err := h.factorService.CalculateTimingSignal(req.FactorName, req.LookbackDays)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    signal,
    })
}
```

#### 1.2 获取因子择时信号历史

**接口**: `GET /api/factor/timing/history`

**请求参数**:
```
factor_name=Mkt-RF&start_date=2026-01-01&end_date=2026-04-25
```

**响应数据**:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "factor_name": "Mkt-RF",
      "signal_date": "2026-04-25T00:00:00Z",
      "signal_strength": "strong_positive",
      "expected_return": 0.023,
      "confidence": 75.0
    },
    // ... more signals
  ],
  "total": 120
}
```

### 2. Alpha观点API

#### 2.1 创建Alpha观点

**接口**: `POST /api/alpha-views`

**请求参数**:
```json
{
  "portfolio_id": 1,
  "asset_symbol": "SPY",
  "view_return": 0.023,
  "confidence": 75.0,
  "view_type": "absolute",
  "view_method": "factor_timing",
  "valid_until": "2026-05-25T00:00:00Z",
  "source_factor": "Mkt-RF",
  "factor_loading": 1.2
}
```

**响应数据**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "portfolio_id": 1,
    "asset_symbol": "SPY",
    "view_return": 0.023,
    "confidence": 75.0,
    "view_type": "absolute",
    "view_method": "factor_timing",
    "generated_at": "2026-04-25T10:30:00Z",
    "valid_until": "2026-05-25T00:00:00Z",
    "status": "active",
    "source_factor": "Mkt-RF",
    "factor_loading": 1.2,
    "created_at": "2026-04-25T10:30:00Z"
  }
}
```

**Go实现**:
```go
// handlers/alpha_view_handler.go
func (h *AlphaViewHandler) CreateAlphaView(c *gin.Context) {
    var req CreateAlphaViewRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
        return
    }

    view := &models.AlphaView{
        PortfolioID:   req.PortfolioID,
        AssetSymbol:   req.AssetSymbol,
        ViewReturn:    decimal.NewFromFloat(req.ViewReturn),
        Confidence:    decimal.NewFromFloat(req.Confidence),
        ViewType:      models.ViewType(req.ViewType),
        ViewMethod:    models.ViewMethod(req.ViewMethod),
        ValidUntil:    req.ValidUntil,
        SourceFactor:  req.SourceFactor,
        FactorLoading: decimal.NewFromFloat(req.FactorLoading),
        Status:        models.ViewStatusActive,
        GeneratedAt:   time.Now(),
        CreatedAt:     time.Now(),
    }

    if err := h.alphaService.CreateAlphaView(view); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    view,
    })
}
```

#### 2.2 从因子信号生成观点

**接口**: `POST /api/alpha-views/generate-from-signal`

**请求参数**:
```json
{
  "signal_id": 1,
  "target_asset": "SPY"
}
```

**响应数据**:
```json
{
  "success": true,
  "data": {
    "id": 2,
    "asset_symbol": "SPY",
    "view_return": 0.023,
    "confidence": 75.0,
    "view_type": "absolute",
    "view_method": "factor_timing",
    "source_factor": "Mkt-RF"
  }
}
```

#### 2.3 获取活跃观点列表

**接口**: `GET /api/alpha-views/active`

**请求参数**:
```
asset_symbol=SPY&method=factor_timing
```

**响应数据**:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "asset_symbol": "SPY",
      "view_return": 0.023,
      "confidence": 75.0,
      "status": "active",
      "generated_at": "2026-04-25T10:30:00Z"
    }
  ]
}
```

### 3. Black-Litterman API

#### 3.1 创建BL配置

**接口**: `POST /api/black-litterman/configs`

**请求参数**:
```json
{
  "portfolio_id": 1,
  "risk_aversion": 2.5,
  "prior_type": "market_cap",
  "prior_weights": [0.4, 0.3, 0.2, 0.1],
  "omega_method": "Idzorek"
}
```

**响应数据**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "portfolio_id": 1,
    "risk_aversion": 2.5,
    "prior_type": "market_cap",
    "omega_method": "Idzorek",
    "is_active": true,
    "created_at": "2026-04-25T10:30:00Z"
  }
}
```

#### 3.2 计算后验收益

**接口**: `POST /api/black-litterman/calculate`

**请求参数**:
```json
{
  "config_id": 1,
  "view_ids": [1, 2, 3]
}
```

**响应数据**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "config_id": 1,
    "calculation_date": "2026-04-25T10:30:00Z",
    "posterior_returns": [0.025, 0.018, 0.012, 0.008],
    "posterior_weights": [0.42, 0.31, 0.18, 0.09],
    "posterior_cov": "...",
    "num_views": 3,
    "view_impact": 0.015
  }
}
```

**Go实现**:
```go
// handlers/black_litterman_handler.go
func (h *BlackLittermanHandler) CalculatePosteriorReturns(c *gin.Context) {
    var req CalculatePosteriorRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
        return
    }

    // 获取配置
    config, err := h.blService.GetConfig(req.ConfigID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Config not found"})
        return
    }

    // 获取观点
    var views []models.AlphaView
    for _, viewID := range req.ViewIDs {
        view, err := h.alphaService.GetAlphaView(viewID)
        if err != nil {
            continue
        }
        views = append(views, *view)
    }

    // 计算后验收益
    posterior, err := h.blService.CalculatePosteriorReturns(req.ConfigID, views)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    posterior,
    })
}
```

### 4. 风险预算API

#### 4.1 创建风险预算配置

**接口**: `POST /api/risk-budget/configs`

**请求参数**:
```json
{
  "portfolio_id": 1,
  "cvar_limit": 0.05,
  "confidence_level": 0.95,
  "time_horizon": 10,
  "method": "monte_carlo",
  "risk_budgets": {
    "SPY": 0.4,
    "QQQ": 0.3,
    "IWM": 0.2,
    "EEM": 0.1
  }
}
```

**响应数据**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "portfolio_id": 1,
    "cvar_limit": 0.05,
    "confidence_level": 0.95,
    "time_horizon": 10,
    "method": "monte_carlo",
    "created_at": "2026-04-25T10:30:00Z"
  }
}
```

#### 4.2 计算CVaR和风险贡献

**接口**: `POST /api/risk-budget/calculate-cvar`

**请求参数**:
```json
{
  "config_id": 1,
  "weights": [0.4, 0.3, 0.2, 0.1],
  "returns_data": [[...], [...], ...]
}
```

**响应数据**:
```json
{
  "success": true,
  "data": {
    "cvar": 0.048,
    "var": 0.035,
    "risk_contributions": [
      {
        "asset_symbol": "SPY",
        "weight": 0.4,
        "marginal_risk": 0.012,
        "risk_contribution": 0.019,
        "percentage_contribution": 39.6
      },
      {
        "asset_symbol": "QQQ",
        "weight": 0.3,
        "marginal_risk": 0.010,
        "risk_contribution": 0.014,
        "percentage_contribution": 29.2
      }
    ]
  }
}
```

#### 4.3 运行蒙特卡洛模拟

**接口**: `POST /api/risk-budget/monte-carlo`

**请求参数**:
```json
{
  "config_id": 1,
  "simulations": 10000,
  "time_steps": 252
}
```

**响应数据**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "config_id": 1,
    "simulation_date": "2026-04-25T10:30:00Z",
    "num_simulations": 10000,
    "time_steps": 252,
    "mean_return": 0.08,
    "std_return": 0.15,
    "percentile_5": -0.18,
    "percentile_95": 0.35,
    "created_at": "2026-04-25T10:30:00Z"
  }
}
```

---

## 数据交互协议

### 1. 请求/响应格式规范

#### 1.1 统一响应格式

**成功响应**:
```json
{
  "success": true,
  "data": { ... },
  "message": "操作成功"
}
```

**错误响应**:
```json
{
  "success": false,
  "error": "错误描述",
  "error_code": "INVALID_PARAMETER",
  "details": {
    "field": "factor_name",
    "reason": "因子名称不能为空"
  }
}
```

#### 1.2 分页查询格式

**请求参数**:
```
?page=1&page_size=20&sort_by=created_at&order=desc
```

**响应格式**:
```json
{
  "success": true,
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

### 2. 数据类型定义

#### 2.1 TypeScript类型定义

**文件位置**: `/frontend/src/types/factor.ts`

```typescript
export type SignalStrength =
  | 'strong_positive'
  | 'weak_positive'
  | 'neutral'
  | 'weak_negative'
  | 'strong_negative';

export interface FactorTimingSignal {
  id: number;
  factor_name: string;
  signal_date: string;
  ma_slope_60: number;
  z_score: number;
  percentile: number;
  signal_strength: SignalStrength;
  signal_score: number;
  expected_return: number;
  confidence: number;
  created_at: string;
}

export type ViewType = 'absolute' | 'relative';
export type ViewMethod = 'factor_timing' | 'momentum' | 'mean_reversion';
export type ViewStatus = 'active' | 'expired' | 'validated';

export interface AlphaView {
  id: number;
  portfolio_id: number;
  asset_symbol: string;
  view_return: number;
  confidence: number;
  view_type: ViewType;
  view_method: ViewMethod;
  generated_at: string;
  valid_until: string;
  status: ViewStatus;
  source_factor: string;
  factor_loading: number;
  created_at: string;
  updated_at: string;
  performance?: AlphaViewPerformance;
}

export interface AlphaViewPerformance {
  id: number;
  view_id: number;
  actual_return: number;
  prediction_error: number;
  is_validated: boolean;
  validation_date: string;
  is_correct: boolean;
  rolling_win_rate: number;
  created_at: string;
}

export type PriorType = 'equal_weight' | 'min_variance' | 'market_cap';
export type OmegaMethod = 'Idzorek' | 'HeLitterman';

export interface BlackLittermanConfig {
  id: number;
  portfolio_id: number;
  risk_aversion: number;
  prior_type: PriorType;
  prior_weights: string;
  implied_returns: string;
  omega_method: OmegaMethod;
  omega_matrix: string;
  is_active: boolean;
  last_calculated: string;
  created_at: string;
  updated_at: string;
}

export interface BLPosteriorReturn {
  id: number;
  config_id: number;
  calculation_date: string;
  posterior_returns: string;
  posterior_weights: string;
  posterior_cov: string;
  num_views: number;
  view_impact: number;
  created_at: string;
}

export type RiskMethod = 'historical' | 'parametric' | 'monte_carlo';

export interface RiskBudgetConfig {
  id: number;
  portfolio_id: number;
  cvar_limit: number;
  confidence_level: number;
  time_horizon: number;
  method: RiskMethod;
  risk_budgets: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface RiskContribution {
  id: number;
  config_id: number;
  calculation_date: string;
  asset_symbol: string;
  weight: number;
  marginal_risk: number;
  risk_contribution: number;
  percentage_contribution: number;
  created_at: string;
}

export interface MonteCarloSimulation {
  id: number;
  config_id: number;
  simulation_date: string;
  num_simulations: number;
  time_steps: number;
  mean_return: number;
  std_return: number;
  percentile_5: number;
  percentile_95: number;
  simulation_data: string;
  created_at: string;
}
```

### 3. API服务封装

**文件位置**: `/frontend/src/services/api.ts`

```typescript
// 因子择时API
export const factorTimingAPI = {
  calculateSignal: async (factorName: string, lookbackDays: number): Promise<FactorTimingSignal> => {
    return request<FactorTimingSignal>('/factor/timing/calculate', {
      method: 'POST',
      body: JSON.stringify({
        factor_name: factorName,
        lookback_days: lookbackDays,
      }),
    });
  },

  getSignalHistory: async (factorName: string, startDate: string, endDate: string): Promise<FactorTimingSignal[]> => {
    return request<FactorTimingSignal[]>(
      `/factor/timing/history?factor_name=${factorName}&start_date=${startDate}&end_date=${endDate}`
    );
  },

  generateView: async (signal: FactorTimingSignal, targetAsset: string): Promise<AlphaView> => {
    return request<AlphaView>('/alpha-views/generate-from-signal', {
      method: 'POST',
      body: JSON.stringify({
        signal_id: signal.id,
        target_asset: targetAsset,
      }),
    });
  },
};

// Alpha观点API
export const alphaViewAPI = {
  create: async (view: CreateAlphaViewRequest): Promise<AlphaView> => {
    return request<AlphaView>('/alpha-views', {
      method: 'POST',
      body: JSON.stringify(view),
    });
  },

  getActive: async (assetSymbol?: string, method?: ViewMethod): Promise<AlphaView[]> => {
    const params = new URLSearchParams();
    if (assetSymbol) params.append('asset_symbol', assetSymbol);
    if (method) params.append('method', method);

    return request<AlphaView[]>(`/alpha-views/active?${params.toString()}`);
  },

  getById: async (id: number): Promise<AlphaView> => {
    return request<AlphaView>(`/alpha-views/${id}`);
  },

  update: async (id: number, view: UpdateAlphaViewRequest): Promise<AlphaView> => {
    return request<AlphaView>(`/alpha-views/${id}`, {
      method: 'PUT',
      body: JSON.stringify(view),
    });
  },

  deactivate: async (id: number): Promise<void> => {
    return request<void>(`/alpha-views/${id}/deactivate`, {
      method: 'POST',
    });
  },

  validate: async (id: number, actualReturn: number): Promise<AlphaViewPerformance> => {
    return request<AlphaViewPerformance>(`/alpha-views/${id}/validate`, {
      method: 'POST',
      body: JSON.stringify({ actual_return: actualReturn }),
    });
  },
};

// Black-Litterman API
export const blackLittermanAPI = {
  createConfig: async (config: CreateBLConfigRequest): Promise<BlackLittermanConfig> => {
    return request<BlackLittermanConfig>('/black-litterman/configs', {
      method: 'POST',
      body: JSON.stringify(config),
    });
  },

  calculate: async (configId: number, viewIds: number[]): Promise<BLPosteriorReturn> => {
    return request<BLPosteriorReturn>('/black-litterman/calculate', {
      method: 'POST',
      body: JSON.stringify({
        config_id: configId,
        view_ids: viewIds,
      }),
    });
  },

  getPosteriorReturns: async (configId: number): Promise<BLPosteriorReturn> => {
    return request<BLPosteriorReturn>(`/black-litterman/configs/${configId}/posterior`);
  },
};

// 风险预算API
export const riskBudgetAPI = {
  createConfig: async (config: CreateRiskBudgetRequest): Promise<RiskBudgetConfig> => {
    return request<RiskBudgetConfig>('/risk-budget/configs', {
      method: 'POST',
      body: JSON.stringify(config),
    });
  },

  calculateCVaR: async (configId: number, weights: number[]): Promise<CVaRResult> => {
    return request<CVaRResult>('/risk-budget/calculate-cvar', {
      method: 'POST',
      body: JSON.stringify({
        config_id: configId,
        weights: weights,
      }),
    });
  },

  runMonteCarlo: async (configId: number, simulations: number, timeSteps: number): Promise<MonteCarloSimulation> => {
    return request<MonteCarloSimulation>('/risk-budget/monte-carlo', {
      method: 'POST',
      body: JSON.stringify({
        config_id: configId,
        simulations: simulations,
        time_steps: timeSteps,
      }),
    });
  },
};
```

---

## 前后端联调方案

### 1. 开发环境配置

#### 1.1 后端配置

**文件**: `/backend/.env`
```env
# 服务器配置
SERVER_PORT=8080
SERVER_MODE=debug

# 数据库配置
DB_TYPE=sqlite
DB_PATH=./data/etf-insight.db

# CORS配置
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization

# Kimi API配置
KIMI_API_KEY=your_kimi_api_key
KIMI_API_BASE_URL=https://api.moonshot.cn/v1
```

#### 1.2 前端配置

**文件**: `/frontend/.env.development`
```env
VITE_API_BASE_URL=http://localhost:8080/api
VITE_WS_URL=ws://localhost:8080/ws
VITE_KIMI_API_KEY=your_kimi_api_key
```

### 2. API路由注册

**文件**: `/backend/routes/routes.go`

```go
func SetupRouter() *gin.Engine {
    r := gin.Default()

    // CORS中间件
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

    // API路由组
    api := r.Group("/api")
    {
        // 因子择时路由
        factorTiming := api.Group("/factor/timing")
        {
            handler := handlers.NewFactorTimingHandler()
            factorTiming.POST("/calculate", handler.CalculateTimingSignal)
            factorTiming.GET("/history", handler.GetSignalHistory)
        }

        // Alpha观点路由
        alphaViews := api.Group("/alpha-views")
        {
            handler := handlers.NewAlphaViewHandler()
            alphaViews.POST("", handler.CreateAlphaView)
            alphaViews.GET("/active", handler.GetActiveViews)
            alphaViews.GET("/:id", handler.GetAlphaView)
            alphaViews.PUT("/:id", handler.UpdateAlphaView)
            alphaViews.POST("/:id/deactivate", handler.DeactivateView)
            alphaViews.POST("/:id/validate", handler.ValidateView)
            alphaViews.POST("/generate-from-signal", handler.GenerateFromSignal)
        }

        // Black-Litterman路由
        bl := api.Group("/black-litterman")
        {
            handler := handlers.NewBlackLittermanHandler()
            bl.POST("/configs", handler.CreateConfig)
            bl.GET("/configs/:id", handler.GetConfig)
            bl.PUT("/configs/:id", handler.UpdateConfig)
            bl.POST("/calculate", handler.CalculatePosteriorReturns)
            bl.GET("/configs/:id/posterior", handler.GetPosteriorReturns)
        }

        // 风险预算路由
        riskBudget := api.Group("/risk-budget")
        {
            handler := handlers.NewRiskBudgetHandler()
            riskBudget.POST("/configs", handler.CreateConfig)
            riskBudget.GET("/configs/:id", handler.GetConfig)
            riskBudget.PUT("/configs/:id", handler.UpdateConfig)
            riskBudget.POST("/calculate-cvar", handler.CalculateCVaR)
            riskBudget.POST("/monte-carlo", handler.RunMonteCarlo)
            riskBudget.GET("/configs/:id/contributions", handler.GetRiskContributions)
        }
    }

    return r
}
```

### 3. 联调测试流程

#### 3.1 单元测试

**后端单元测试**:
```bash
cd backend
go test ./handlers -v -run TestFactorTimingHandler
go test ./handlers -v -run TestAlphaViewHandler
go test ./handlers -v -run TestBlackLittermanHandler
go test ./handlers -v -run TestRiskBudgetHandler
```

**前端单元测试**:
```bash
cd frontend
npm test -- FactorTiming.test.tsx
npm test -- AlphaViews.test.tsx
npm test -- BlackLittermanConfig.test.tsx
npm test -- RiskBudget.test.tsx
```

#### 3.2 集成测试

**测试脚本**: `/scripts/integration-test.sh`

```bash
#!/bin/bash

echo "Starting integration tests..."

# 启动后端服务
cd backend
go run main.go &
BACKEND_PID=$!
sleep 5

# 启动前端服务
cd ../frontend
npm run dev &
FRONTEND_PID=$!
sleep 5

# 运行集成测试
npm run test:integration

# 清理
kill $BACKEND_PID
kill $FRONTEND_PID

echo "Integration tests completed."
```

#### 3.3 API测试用例

**文件**: `/tests/api/factor_timing_api_test.http`

```http
### 计算因子择时信号
POST http://localhost:8080/api/factor/timing/calculate
Content-Type: application/json

{
  "factor_name": "Mkt-RF",
  "lookback_days": 60
}

### 获取信号历史
GET http://localhost:8080/api/factor/timing/history?factor_name=Mkt-RF&start_date=2026-01-01&end_date=2026-04-25

### 创建Alpha观点
POST http://localhost:8080/api/alpha-views
Content-Type: application/json

{
  "portfolio_id": 1,
  "asset_symbol": "SPY",
  "view_return": 0.023,
  "confidence": 75.0,
  "view_type": "absolute",
  "view_method": "factor_timing",
  "valid_until": "2026-05-25T00:00:00Z",
  "source_factor": "Mkt-RF"
}

### 计算Black-Litterman后验收益
POST http://localhost:8080/api/black-litterman/calculate
Content-Type: application/json

{
  "config_id": 1,
  "view_ids": [1, 2, 3]
}

### 计算CVaR
POST http://localhost:8080/api/risk-budget/calculate-cvar
Content-Type: application/json

{
  "config_id": 1,
  "weights": [0.4, 0.3, 0.2, 0.1]
}
```

---

## 兼容性处理策略

### 1. 向后兼容

#### 1.1 API版本管理

**策略**: 使用URL路径版本控制

```
/api/v1/factor/analyze          # 旧版API
/api/v2/factor/timing/calculate # 新版API
```

**实现**:
```go
// routes/routes.go
func SetupRouter() *gin.Engine {
    r := gin.Default()

    // v1 API (旧版)
    v1 := r.Group("/api/v1")
    {
        v1.POST("/factor/analyze", legacyHandler.AnalyzeFactorExposure)
        v1.POST("/portfolio/optimize", legacyHandler.OptimizePortfolio)
    }

    // v2 API (新版)
    v2 := r.Group("/api/v2")
    {
        v2.POST("/factor/timing/calculate", newHandler.CalculateTimingSignal)
        v2.POST("/alpha-views", newHandler.CreateAlphaView)
        v2.POST("/black-litterman/calculate", newHandler.CalculatePosteriorReturns)
    }

    // 默认路由指向最新版本
    api := r.Group("/api")
    {
        api.Any("/*action", func(c *gin.Context) {
            c.Redirect(http.StatusMovedPermanently, "/api/v2"+c.Param("action"))
        })
    }

    return r
}
```

#### 1.2 数据库迁移兼容

**策略**: 渐进式迁移，保留旧表结构

```sql
-- 迁移脚本：保留旧表，创建新表
CREATE TABLE factor_data_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    factor_name VARCHAR(20) NOT NULL,
    date DATE NOT NULL,
    value DECIMAL(10,6) NOT NULL,
    data_source VARCHAR(50),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(factor_name, date)
);

-- 数据迁移
INSERT INTO factor_data_new (factor_name, date, value, data_source, created_at)
SELECT factor_name, date, value, 'legacy', created_at
FROM old_factor_data;

-- 重命名表（在确认无问题后执行）
-- DROP TABLE old_factor_data;
-- ALTER TABLE factor_data_new RENAME TO factor_data;
```

### 2. 前端兼容

#### 2.1 特性检测

```typescript
// 检测API是否支持新功能
export const checkAPIFeature = async (feature: string): Promise<boolean> => {
  try {
    const response = await fetch(`${API_BASE_URL}/features/${feature}`);
    return response.ok;
  } catch {
    return false;
  }
};

// 使用示例
const hasFactorTiming = await checkAPIFeature('factor-timing');
if (hasFactorTiming) {
  // 使用新功能
} else {
  // 降级到旧功能
}
```

#### 2.2 渐进式增强

```typescript
// FactorAnalysis.tsx
const FactorAnalysis: React.FC = () => {
  const [hasTimingFeature, setHasTimingFeature] = useState(false);

  useEffect(() => {
    checkAPIFeature('factor-timing').then(setHasTimingFeature);
  }, []);

  return (
    <Container>
      {/* 原有功能 */}
      <Card title="因子暴露分析">
        {/* ... */}
      </Card>

      {/* 新功能：仅在支持时显示 */}
      {hasTimingFeature && (
        <Card title="因子择时信号">
          <FactorTimingSignalCard />
        </Card>
      )}
    </Container>
  );
};
```

### 3. 浏览器兼容

#### 3.1 Polyfill配置

**文件**: `/frontend/vite.config.ts`

```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  build: {
    target: 'es2015',
    polyfillModulePreload: true,
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-react': ['react', 'react-dom'],
          'vendor-antd': ['antd', '@ant-design/icons'],
          'vendor-charts': ['recharts'],
        },
      },
    },
  },
  optimizeDeps: {
    include: [
      'react',
      'react-dom',
      'antd',
      'recharts',
    ],
  },
});
```

#### 3.2 CSS兼容

```typescript
// 使用styled-components自动添加前缀
import styled from 'styled-components';

const FlexContainer = styled.div`
  display: -webkit-box;
  display: -ms-flexbox;
  display: flex;
  -webkit-box-align: center;
  -ms-flex-align: center;
  align-items: center;
`;
```

---

## 分阶段实施计划

### 阶段一：基础设施准备（1周）

**时间**: 2026-05-02 ~ 2026-05-08

**目标**: 完成数据库迁移、API基础框架搭建

**任务清单**:
- [ ] 执行数据库迁移脚本
- [ ] 创建API Handler骨架
- [ ] 配置开发环境
- [ ] 编写API文档
- [ ] 设置测试环境

**交付物**:
- ✅ 数据库表结构就绪
- ✅ API路由注册完成
- ✅ Swagger文档生成
- ✅ 单元测试框架搭建

### 阶段二：因子择时模块（1周）

**时间**: 2026-05-09 ~ 2026-05-15

**目标**: 实现因子择时信号的完整前后端功能

**后端任务**:
- [ ] 实现 `FactorTimingHandler`
- [ ] 编写单元测试
- [ ] 集成Kimi API（可选）
- [ ] 性能优化

**前端任务**:
- [ ] 创建 `FactorTiming.tsx` 页面
- [ ] 实现信号可视化组件
- [ ] 集成API调用
- [ ] 编写组件测试

**联调任务**:
- [ ] 前后端接口联调
- [ ] 集成测试
- [ ] Bug修复

**交付物**:
- ✅ 因子择时API上线
- ✅ 前端页面可访问
- ✅ 测试覆盖率 > 80%

### 阶段三：Alpha观点模块（1周）

**时间**: 2026-05-16 ~ 2026-05-22

**目标**: 实现Alpha观点生成和管理功能

**后端任务**:
- [ ] 实现 `AlphaViewHandler`
- [ ] 实现观点验证逻辑
- [ ] 编写单元测试

**前端任务**:
- [ ] 创建 `AlphaViews.tsx` 页面
- [ ] 实现观点列表组件
- [ ] 实现观点创建表单
- [ ] 集成因子择时信号

**联调任务**:
- [ ] 观点生成流程测试
- [ ] 观点验证流程测试
- [ ] 性能测试

**交付物**:
- ✅ Alpha观点CRUD API
- ✅ 观点管理页面
- ✅ 观点性能追踪

### 阶段四：Black-Litterman模块（1.5周）

**时间**: 2026-05-23 ~ 2026-06-02

**目标**: 实现Black-Litterman模型集成

**后端任务**:
- [ ] 实现 `BlackLittermanHandler`
- [ ] 实现矩阵运算优化
- [ ] 编写单元测试

**前端任务**:
- [ ] 创建 `BlackLittermanConfig.tsx` 页面
- [ ] 实现配置管理组件
- [ ] 实现观点选择组件
- [ ] 集成组合优化

**联调任务**:
- [ ] BL计算流程测试
- [ ] 优化结果验证
- [ ] 性能优化

**交付物**:
- ✅ BL配置API
- ✅ BL计算API
- ✅ 前端配置页面
- ✅ 优化结果展示

### 阶段五：风险预算模块（1.5周）

**时间**: 2026-06-03 ~ 2026-06-13

**目标**: 实现风险预算和CVaR计算功能

**后端任务**:
- [ ] 实现 `RiskBudgetHandler`
- [ ] 实现蒙特卡洛模拟
- [ ] 实现风险贡献计算
- [ ] 编写单元测试

**前端任务**:
- [ ] 创建 `RiskBudget.tsx` 页面
- [ ] 实现CVaR可视化
- [ ] 实现风险贡献图表
- [ ] 集成组合优化

**联调任务**:
- [ ] CVaR计算测试
- [ ] 蒙特卡洛模拟测试
- [ ] 风险预算优化测试

**交付物**:
- ✅ 风险预算配置API
- ✅ CVaR计算API
- ✅ 前端管理页面
- ✅ 风险可视化

### 阶段六：集成测试与优化（1周）

**时间**: 2026-06-14 ~ 2026-06-20

**目标**: 完成端到端测试和性能优化

**测试任务**:
- [ ] 端到端测试
- [ ] 性能测试
- [ ] 压力测试
- [ ] 安全测试

**优化任务**:
- [ ] API性能优化
- [ ] 前端性能优化
- [ ] 数据库查询优化
- [ ] 缓存策略实施

**文档任务**:
- [ ] API文档完善
- [ ] 用户手册编写
- [ ] 部署文档更新

**交付物**:
- ✅ 测试报告
- ✅ 性能报告
- ✅ 完整文档
- ✅ 生产环境部署

---

## Kimi2.6接口接入指南

### 1. Kimi API集成

#### 1.1 配置

**文件**: `/backend/config/kimi.go`

```go
package config

type KimiConfig struct {
    APIKey     string `env:"KIMI_API_KEY" envDefault:""`
    BaseURL    string `env:"KIMI_API_BASE_URL" envDefault:"https://api.moonshot.cn/v1"`
    Model      string `env:"KIMI_MODEL" envDefault:"moonshot-v1-8k"`
    MaxTokens  int    `env:"KIMI_MAX_TOKENS" envDefault:"2048"`
    Temperature float64 `env:"KIMI_TEMPERATURE" envDefault:"0.7"`
}
```

#### 1.2 服务封装

**文件**: `/backend/services/kimi_service.go`

```go
package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "etf-insight/config"
)

type KimiService struct {
    config *config.KimiConfig
    client *http.Client
}

func NewKimiService(cfg *config.KimiConfig) *KimiService {
    return &KimiService{
        config: cfg,
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatRequest struct {
    Model       string        `json:"model"`
    Messages    []ChatMessage `json:"messages"`
    Temperature float64       `json:"temperature"`
    MaxTokens   int           `json:"max_tokens"`
}

type ChatResponse struct {
    ID      string `json:"id"`
    Object  string `json:"object"`
    Created int64  `json:"created"`
    Model   string `json:"model"`
    Choices []struct {
        Index   int `json:"index"`
        Message struct {
            Role    string `json:"role"`
            Content string `json:"content"`
        } `json:"message"`
        FinishReason string `json:"finish_reason"`
    } `json:"choices"`
    Usage struct {
        PromptTokens     int `json:"prompt_tokens"`
        CompletionTokens int `json:"completion_tokens"`
        TotalTokens      int `json:"total_tokens"`
    } `json:"usage"`
}

func (s *KimiService) Chat(messages []ChatMessage) (string, error) {
    req := ChatRequest{
        Model:       s.config.Model,
        Messages:    messages,
        Temperature: s.config.Temperature,
        MaxTokens:   s.config.MaxTokens,
    }

    body, err := json.Marshal(req)
    if err != nil {
        return "", fmt.Errorf("failed to marshal request: %w", err)
    }

    httpReq, err := http.NewRequest("POST", s.config.BaseURL+"/chat/completions", bytes.NewBuffer(body))
    if err != nil {
        return "", fmt.Errorf("failed to create request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+s.config.APIKey)

    resp, err := s.client.Do(httpReq)
    if err != nil {
        return "", fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("failed to read response: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
    }

    var chatResp ChatResponse
    if err := json.Unmarshal(respBody, &chatResp); err != nil {
        return "", fmt.Errorf("failed to unmarshal response: %w", err)
    }

    if len(chatResp.Choices) == 0 {
        return "", fmt.Errorf("no choices in response")
    }

    return chatResp.Choices[0].Message.Content, nil
}

func (s *KimiService) GenerateMarketAnalysis(factorData string) (string, error) {
    messages := []ChatMessage{
        {
            Role: "system",
            Content: "你是一个专业的量化分析师，擅长分析因子数据和市场趋势。请基于提供的因子数据，给出专业的市场分析和投资建议。",
        },
        {
            Role: "user",
            Content: fmt.Sprintf("请分析以下因子数据，并给出市场趋势判断和投资建议：\n\n%s", factorData),
        },
    }

    return s.Chat(messages)
}

func (s *KimiService) GenerateAlphaView(signal *FactorTimingSignal) (string, error) {
    messages := []ChatMessage{
        {
            Role: "system",
            Content: "你是一个量化投资专家，擅长从因子择时信号中提取Alpha观点。请基于信号强度和置信度，生成专业的投资观点。",
        },
        {
            Role: "user",
            Content: fmt.Sprintf(
                "因子择时信号：\n- 因子: %s\n- MA斜率: %.4f\n- Z-Score: %.2f\n- 百分位数: %.1f%%\n- 信号强度: %s\n- 预期收益: %.2f%%\n- 置信度: %.1f%%\n\n请生成Alpha观点和投资建议。",
                signal.FactorName,
                signal.MASlope60,
                signal.ZScore,
                signal.Percentile,
                signal.SignalStrength,
                signal.ExpectedReturn*100,
                signal.Confidence,
            ),
        },
    }

    return s.Chat(messages)
}
```

#### 1.3 API集成

**文件**: `/backend/handlers/kimi_handler.go`

```go
package handlers

import (
    "net/http"

    "etf-insight/services"

    "github.com/gin-gonic/gin"
)

type KimiHandler struct {
    kimiService *services.KimiService
}

func NewKimiHandler(kimiService *services.KimiService) *KimiHandler {
    return &KimiHandler{
        kimiService: kimiService,
    }
}

func (h *KimiHandler) GenerateMarketAnalysis(c *gin.Context) {
    var req struct {
        FactorData string `json:"factor_data" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error":   "参数错误: " + err.Error(),
        })
        return
    }

    analysis, err := h.kimiService.GenerateMarketAnalysis(req.FactorData)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error":   "分析失败: " + err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "analysis": analysis,
        },
    })
}

func (h *KimiHandler) GenerateAlphaView(c *gin.Context) {
    var req struct {
        SignalID uint `json:"signal_id" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error":   "参数错误: " + err.Error(),
        })
        return
    }

    // 获取信号
    signal, err := h.factorService.GetTimingSignal(req.SignalID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "success": false,
            "error":   "信号不存在",
        })
        return
    }

    // 生成AI观点
    aiView, err := h.kimiService.GenerateAlphaView(signal)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error":   "观点生成失败: " + err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "signal":  signal,
            "ai_view": aiView,
        },
    })
}
```

### 2. 前端集成

#### 2.1 Kimi API调用

**文件**: `/frontend/src/services/kimiAPI.ts`

```typescript
export const kimiAPI = {
  generateMarketAnalysis: async (factorData: string): Promise<string> => {
    const response = await request<{ analysis: string }>('/kimi/market-analysis', {
      method: 'POST',
      body: JSON.stringify({ factor_data: factorData }),
    });
    return response.analysis;
  },

  generateAlphaView: async (signalId: number): Promise<{ signal: FactorTimingSignal; ai_view: string }> => {
    return request<{ signal: FactorTimingSignal; ai_view: string }>('/kimi/alpha-view', {
      method: 'POST',
      body: JSON.stringify({ signal_id: signalId }),
    });
  },
};
```

#### 2.2 UI集成

```typescript
// FactorTiming.tsx
import { kimiAPI } from '../services/kimiAPI';

const FactorTiming: React.FC = () => {
  const [aiAnalysis, setAiAnalysis] = useState<string>('');
  const [aiLoading, setAiLoading] = useState(false);

  const generateAIAnalysis = async () => {
    if (!signal) return;

    setAiLoading(true);
    try {
      const analysis = await kimiAPI.generateMarketAnalysis(
        `因子: ${signal.factor_name}, Z-Score: ${signal.z_score}, 信号强度: ${signal.signal_strength}`
      );
      setAiAnalysis(analysis);
    } catch (error) {
      message.error('AI分析失败: ' + error.message);
    } finally {
      setAiLoading(false);
    }
  };

  return (
    <Container>
      {/* ... 其他组件 ... */}

      <Card title="AI市场分析" style={{ marginTop: 24 }}>
        <Button
          type="primary"
          onClick={generateAIAnalysis}
          loading={aiLoading}
          icon={<RobotOutlined />}
        >
          生成AI分析
        </Button>
        {aiAnalysis && (
          <Alert
            message="AI分析结果"
            description={aiAnalysis}
            type="info"
            style={{ marginTop: 16 }}
          />
        )}
      </Card>
    </Container>
  );
};
```

---

## 附录

### A. 数据库Schema

详见：[001_add_factor_tables.sql](../../backend/migrations/001_add_factor_tables.sql)

### B. API文档

详见：Swagger UI - `http://localhost:8080/swagger/index.html`

### C. 测试覆盖率报告

```bash
# 后端测试覆盖率
cd backend
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# 前端测试覆盖率
cd frontend
npm test -- --coverage --watchAll=false
```

### D. 部署清单

- [ ] 数据库迁移脚本执行
- [ ] 环境变量配置
- [ ] API服务部署
- [ ] 前端构建部署
- [ ] Nginx配置
- [ ] SSL证书配置
- [ ] 监控告警配置

---

**文档维护者**: ETF-Insight Team
**最后更新**: 2026-04-25
**版本**: v1.0

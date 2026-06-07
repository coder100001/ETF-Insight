# 组合分析数据准确性提升设计

> **版本**: v1.0 | **日期**: 2026-06-06 | **状态**: 待审查

---

## 1. 问题陈述

组合分析模块存在 3 类共 12 个数据准确性问题，导致分析结果与真实市场偏差显著：

### 1.1 数据一致性问题

| # | 问题 | 偏差 | 位置 |
|---|------|------|------|
| P1 | 无风险利率不统一（2%/4%/4.5%三处不同） | Sharpe 偏差 10-20% | `portfolio_analytics.go:149`, `portfolio_optimizer.go:85`, `portfolio_handler_risk.go:118` |
| P2 | 股息率全部硬编码（8个ETF固定值） | 与真实市场偏差 30-50% | `portfolio_analytics.go:453-470` |
| P3 | 回测年化用365天而非252交易日 | 年化收益低估 ~5% | `backtest_handler.go` |

### 1.2 计算精度问题

| # | 问题 | 偏差 | 位置 |
|---|------|------|------|
| P4 | CVaR 参数法用 `VaR * 1.2` 近似 | 尾部风险偏差 20-30% | `risk_models.go:123` |
| P5 | CVaR CDF 用 `1-confidence` 近似 | 极端分位数偏差大 | `portfolio_analytics.go:561` |
| P6 | 方差用总体公式（除N非N-1） | 小样本波动率系统性低估 | `portfolio_analytics.go:426`, `risk_models.go:345` |
| P7 | BL 后验收益用加权平均替代矩阵公式 | 结果偏离理论值 | `black_litterman.go:339-367` |
| P8 | BL 权重用 softmax 替代均值方差优化 | 权重分配不精确 | `black_litterman.go:419-430` |

### 1.3 数据源缺失

| # | 问题 | 影响 | 位置 |
|---|------|------|------|
| P9 | ETF 持仓数据接口未实现 | 重叠度/穿透分析无真实数据 | `finage_provider.go` `GetETFHoldings()` |
| P10 | 因子数据为合成随机数 | Fama-French 分析不可靠 | `factor_data_service.go` `SeedSampleFactorData()` |
| P11 | 无实时股息数据获取 | 分红预测全靠猜 | `portfolio_analytics.go` `getEstimatedDividendYield()` |
| P12 | 无 FRED 国债收益率接入 | 无风险利率不反映市场 | 多处硬编码 |

---

## 2. 设计目标

| 目标 | 衡量标准 |
|------|----------|
| 数据一致性 | 全项目无风险利率唯一来源，股息率从 API 获取 |
| 计算精度 | CVaR 偏差 <5%，BL 后验收益与理论值偏差 <3% |
| 可维护性 | 金融常量集中管理，修改一处全局生效 |
| 向后兼容 | API 接口不变，前端无需改动 |

---

## 3. 架构设计

### 3.1 总体架构

```
┌─────────────────────────────────────────────────────┐
│                   配置层 (Phase 1)                   │
│  FinancialConfig (无风险利率/交易日/币种)             │
│  DividendService (股息率查询/API获取/缓存)           │
└──────────────────┬──────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────┐
│                   计算层 (Phase 2)                   │
│  StatisticsUtil (方差/标准差/CDF)                    │
│  VaRCalculator (VaR/CVaR 精确计算)                  │
│  BlackLittermanSolver (完整矩阵公式)                │
└──────────────────┬──────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────┐
│                   数据层 (Phase 3)                   │
│  FactorDataLoader (Kenneth French / Alpha Vantage)  │
│  ETFHoldingsProvider (持仓数据)                      │
│  TreasuryRateProvider (FRED 国债收益率)              │
└─────────────────────────────────────────────────────┘
```

### 3.2 模块依赖关系

```
FinancialConfig ◄── 所有计算模块
       ▲
       │
DividendService ◄── PortfolioAnalytics, ScenarioAnalysis
       ▲
       │
TreasuryRateProvider (可选，Phase 3)
```

---

## 4. Phase 1：数据一致性修复

### 4.1 创建统一金融配置

**新增文件**: `backend/config/financial.go`

```go
package config

import "sync"

type FinancialConfig struct {
    // 无风险利率 (年化，小数形式)
    // 默认值: 当前10年期美债收益率，可通过管理接口动态更新
    RiskFreeRate float64

    // 每年交易日数
    TradingDaysYear int

    // 默认币种
    DefaultCurrency string
}

var (
    financialConfig *FinancialConfig
    configOnce      sync.Once
)

func GetFinancialConfig() *FinancialConfig {
    configOnce.Do(func() {
        financialConfig = &FinancialConfig{
            RiskFreeRate:    0.0435, // 4.35% 2026年6月美债利率
            TradingDaysYear: 252,
            DefaultCurrency: "USD",
        }
    })
    return financialConfig
}

// SetRiskFreeRate 允许运行时更新无风险利率
func SetRiskFreeRate(rate float64) {
    GetFinancialConfig().RiskFreeRate = rate
}
```

**改动范围**:

| 文件 | 当前值 | 改为 |
|------|--------|------|
| `portfolio_analytics.go:149` | `0.045` | `config.GetFinancialConfig().RiskFreeRate` |
| `portfolio_analytics.go:330` | `0.045` | 同上 |
| `portfolio_analytics.go:870` | `0.045` | 同上 |
| `portfolio_optimizer.go:85` | `0.04` | 同上 |
| `portfolio_handler_risk.go:118` | `0.02` | 同上 |
| `mpt_optimizer.go` | `0.045` | 同上 |
| `black_litterman.go:21` | `0.045` | 同上 |
| `scenario_analysis.go` | `0.045` | 同上 |
| `backtest_handler.go` | `365` | `config.GetFinancialConfig().TradingDaysYear` |
| `risk_budget_service.go` | `0.045` | 同上 |

### 4.2 股息率服务

**新增文件**: `backend/services/dividend_service.go`

```go
type DividendService struct {
    db       *gorm.DB
    provider datasource.DataSourceProvider
    cache    map[string]CachedYield  // symbol → {yield, expiry}
    mu       sync.RWMutex
}

type CachedYield struct {
    Yield    float64
    FetchedAt time.Time
    TTL      time.Duration // 默认 24 小时
}

// GetDividendYield 查询顺序：
// 1. 内存缓存 (TTL 24h)
// 2. 数据库 ETFDividend 表
// 3. Finage API 获取
// 4. 硬编码兜底值 (仅在全部失败时)
func (s *DividendService) GetDividendYield(symbol string) (float64, error)
```

**改动**: `portfolio_analytics.go` 的 `getEstimatedDividendYield()` 改为调用 `DividendService.GetDividendYield()`。

### 4.3 新增 API 端点

| 端点 | 方法 | 用途 |
|------|------|------|
| `GET /api/config/financial` | GET | 获取当前金融配置 |
| `PUT /api/config/financial` | PUT | 更新金融配置（利率等） |

---

## 5. Phase 2：计算精度修复

### 5.1 统计工具函数

**新增文件**: `backend/services/statistics/util.go`

```go
package statistics

import "math"

// SampleVariance 样本方差 (除以 N-1)
func SampleVariance(values []float64) float64 {
    if len(values) < 2 {
        return 0
    }
    mean := Mean(values)
    sum := 0.0
    for _, v := range values {
        diff := v - mean
        sum += diff * diff
    }
    return sum / float64(len(values)-1) // N-1
}

// SampleStdDev 样本标准差
func SampleStdDev(values []float64) float64 {
    return math.Sqrt(SampleVariance(values))
}

// NormalCDF 标准正态分布 CDF (精确实现)
// 使用 Abramowitz & Stegun 近似，误差 < 7.5e-8
func NormalCDF(x float64) float64 {
    if x < -8 {
        return 0
    }
    if x > 8 {
        return 1
    }
    // Horner 形式的有理逼近
    const (
        a1 = 0.254829592
        a2 = -0.284496736
        a3 = 1.421413741
        a4 = -1.453152027
        a5 = 1.061405429
        p  = 0.3275911
    )
    sign := 1.0
    if x < 0 {
        sign = -1.0
        x = -x
    }
    t := 1.0 / (1.0 + p*x)
    y := 1.0 - (((((a5*t+a4)*t)+a3)*t+a2)*t+a1)*t*math.Exp(-x*x/2)
    return 0.5 * (1.0 + sign*y)
}

// NormalPDF 标准正态分布 PDF
func NormalPDF(x float64) float64 {
    return (1.0 / math.Sqrt(2*math.Pi)) * math.Exp(-x*x/2)
}

// NormalQuantile 标准正态分布分位数 (Beasley-Springer-Moro 算法)
func NormalQuantile(p float64) float64 {
    // ... 精确实现
}
```

### 5.2 CVaR 精确计算

**修改文件**: `backend/services/risk_models.go`

```go
// 修改前 (line 123):
cvarValue := varValue.Mul(decimal.NewFromFloat(1.2))

// 修改后:
// CVaR = μ - σ × φ(z) / (1 - α)
zFloat := zScore.InexactFloat64()
phiZ := statistics.NormalPDF(zFloat)
oneMinusAlpha := 1.0 - confidence
cvarDecimal := mean.Sub(stdDev.Mul(decimal.NewFromFloat(phiZ / oneMinusAlpha)))
cvarValue := cvarDecimal.Neg() // 转为正数表示损失
```

**修改文件**: `backend/services/portfolio_analytics.go`

```go
// 修改前 (line 561):
Phi := 1 - confidence

// 修改后:
Phi := statistics.NormalCDF(zScore)
```

### 5.3 方差公式修复

**修改文件**: `backend/services/portfolio_analytics.go:426`

```go
// 修改前:
return sum / float64(len(values))

// 修改后:
n := len(values)
if n < 2 {
    return 0
}
return sum / float64(n-1)
```

**修改文件**: `backend/services/risk_models.go:345` 同理。

### 5.4 Black-Litterman 完整矩阵公式

**修改文件**: `backend/services/optimization/black_litterman.go`

核心修改：将 `calculatePosteriorReturns()` 从加权平均改为标准 BL 公式：

```
E[R] = [(τΣ)⁻¹ + P'Ω⁻¹P]⁻¹ × [(τΣ)⁻¹Π + P'Ω⁻¹Q]
```

需要实现的矩阵运算：
- 矩阵求逆（高斯消元，已有部分实现）
- 矩阵乘法
- 矩阵转置

```go
func (o *BlackLittermanOptimizer) calculatePosteriorReturns(
    Pi []float64,      // 先验收益
    Sigma [][]float64, // 协方差矩阵
    P [][]float64,     // 观点矩阵
    Q []float64,       // 观点收益
    Omega [][]float64, // 观点不确定性
    tau float64,       // 不确定性参数
) []float64 {
    n := len(Pi)
    k := len(Q)

    // 1. 计算 (τΣ)⁻¹
    tauSigma := scalarMultiply(Sigma, tau)
    tauSigmaInv := matrixInverse(tauSigma)

    // 2. 计算 P'Ω⁻¹P
    Pt := transpose(P)
    OmegaInv := matrixInverse(Omega)
    PtOmegaInv := matrixMultiply(Pt, OmegaInv)
    PtOmegaInvP := matrixMultiply(PtOmegaInv, P)

    // 3. 计算 [(τΣ)⁻¹ + P'Ω⁻¹P]
    M := matrixAdd(tauSigmaInv, PtOmegaInvP)

    // 4. 计算 M⁻¹
    MInv := matrixInverse(M)

    // 5. 计算 (τΣ)⁻¹Π
    tauSigmaInvPi := matrixVectorMultiply(tauSigmaInv, Pi)

    // 6. 计算 P'Ω⁻¹Q
    PtOmegaInvQ := matrixVectorMultiply(PtOmegaInv, Q)

    // 7. 计算 (τΣ)⁻¹Π + P'Ω⁻¹Q
    rhs := vectorAdd(tauSigmaInvPi, PtOmegaInvQ)

    // 8. 最终: M⁻¹ × rhs
    return matrixVectorMultiply(MInv, rhs)
}
```

### 5.5 BL 权重优化

**修改文件**: `backend/services/optimization/black_litterman.go`

```go
// 修改前: softmax 风格
expReturns[i] = math.Exp(mu[i] * 10)

// 修改后: 均值方差优化
// w* = (1/δ) × Σ⁻¹ × μ
// 其中 δ = 风险厌恶系数 (默认 2.5)
func (o *BlackLittermanOptimizer) optimizeWeights(
    posteriorReturns []float64,
    Sigma [][]float64,
    riskAversion float64,
    constraint *BlackLittermanConstraint,
) []float64 {
    SigmaInv := matrixInverse(Sigma)
    excessReturns := make([]float64, len(posteriorReturns))
    for i, r := range posteriorReturns {
        excessReturns[i] = r - o.RiskFreeRate
    }
    weights := matrixVectorMultiply(SigmaInv, excessReturns)
    // 归一化
    sum := 0.0
    for _, w := range weights {
        sum += w
    }
    for i := range weights {
        weights[i] /= sum
    }
    return o.applyConstraints(weights, symbols, constraint)
}
```

---

## 6. Phase 3：数据源接入

### 6.1 ETF 持仓数据

**新增文件**: `backend/services/etf_holdings_service.go`

```go
type ETFHoldingsService struct {
    db       *gorm.DB
    provider datasource.DataSourceProvider
}

// 数据来源优先级:
// 1. 数据库 ETFConstituent 表 (缓存)
// 2. Finage API (如已实现)
// 3. 备用: iShares/Vanguard 官网爬取 (未来)
// 4. 兜底: 返回空 + 标记为 estimated

func (s *ETFHoldingsService) GetHoldings(symbol string) ([]ETFHolding, error)
func (s *ETFHoldingsService) SyncHoldings(symbol string) error
func (s *ETFHoldingsService) GetOverlap(etf1, etf2 string) (float64, error)
```

**优先级说明**: Finage 的 ETF 持仓 API 是付费功能。如果不可用，可以：
- 方案 A: 接入 Alpha Vantage 的 holdings API（免费额度有限）
- 方案 B: 从 iShares/Vanguard 官网定期爬取主要 ETF 的持仓
- 方案 C: 手动维护主要 ETF（SCHD/JEPI/JEPQ/QQQ 等）的 Top 10 持仓

**推荐**: 方案 C 作为 Phase 3 的第一步，覆盖 10-15 个核心 ETF 的 Top 25 持仓，够用且可控。

### 6.2 真实因子数据

**修改文件**: `backend/services/factor/fama_french.go`

```go
// LoadFactorDataFromFrench 接入 Kenneth French 数据库
// URL: https://mba.tuck.dartmouth.edu/pages/faculty/ken.french/data_library.html
// 数据: Fama-French 3/5 因子，日度/月度
func (s *FamaFrenchService) LoadFactorDataFromFrench(
    startDate, endDate time.Time,
    frequency string, // "daily" or "monthly"
) error {
    // 1. 下载 CSV
    // 2. 解析为 FactorData 结构
    // 3. 存入数据库
    // 4. 替换 SeedSampleFactorData 的合成数据
}
```

**备用方案**: 如果 French 网站访问不稳定，使用 Alpha Vantage 的 `FACTORS` API。

### 6.3 股息历史数据

**修改文件**: `backend/services/dividend_service.go` (Phase 1 已创建)

```go
// SyncDividendHistory 从 Finage 获取历史分红记录
func (s *DividendService) SyncDividendHistory(symbol string, years int) error

// GetTrailing12MonthYield 计算过去12个月实际股息率
func (s *DividendService) GetTrailing12MonthYield(symbol string) (float64, error)
```

### 6.4 国债收益率（可选）

**新增文件**: `backend/services/treasury_rate_service.go`

```go
// FRED API: https://api.stlouisfed.org/fred/series/observations
// Series: DGS10 (10年期国债收益率)
// 需要 FRED API Key，免费申请
type TreasuryRateService struct {
    apiKey string
    cache  float64
}

func (s *TreasuryRateService) Get10YearRate() (float64, error)
```

---

## 7. 测试计划

### 7.1 单元测试

| 模块 | 测试重点 | 测试用例数 |
|------|----------|-----------|
| `statistics/util.go` | NormalCDF 精度、SampleVariance 对比 | 15+ |
| `VaRCalculator` | CVaR 精确值 vs 蒙特卡洛验证 | 10+ |
| `BlackLittermanSolver` | 后验收益与 Python `pyportfolioopt` 对比 | 8+ |
| `DividendService` | 缓存命中/过期/API 失败降级 | 8+ |

### 7.2 精度验证

使用 Python 作为参照实现验证：

```python
# CVaR 验证
from scipy import stats
cvar = mean - std * stats.norm.pdf(z) / (1 - confidence)

# BL 验证
from pypfopt.black_litterman import BlackLittermanModel
bl = BlackLittermanModel(cov_matrix, pi, P, Q, omega)
posterior_returns = bl.bl_returns()
```

### 7.3 集成测试

- 修改后所有现有测试必须通过
- 新增测试覆盖所有修改的函数

---

## 8. 迁移策略

### 8.1 向后兼容

- API 接口不变，前端无需改动
- 配置变更通过 `FinancialConfig` 管理，不改变 API 签名
- CVaR/BL 计算精度提升是内部优化，对外表现为更准确的数值

### 8.2 灰度方案

由于是计算精度修复，不需要灰度。但建议：
1. Phase 1 完成后运行全量测试
2. Phase 2 完成后对比新旧计算结果差异
3. Phase 3 完成后验证数据源可用性

---

## 9. 风险与缓解

| 风险 | 概率 | 影响 | 缓解 |
|------|------|------|------|
| BL 矩阵求逆数值不稳定 | 中 | 计算失败 | 添加 Tikhonov 正则化（已有经验） |
| FRED/French API 不可达 | 低 | 无法获取实时数据 | 配置兜底值 + 定时重试 |
| 修改方差公式影响现有测试 | 高 | 测试失败 | 逐个修复测试断言 |
| CVaR 精确计算比近似慢 | 低 | 性能下降 | 差异在 μs 级别，可忽略 |

---

## 10. 实施顺序

```
Phase 1.1: FinancialConfig (1天)
    ↓
Phase 1.2: DividendService (1天)
    ↓
Phase 1.3: 统一引用 + API 端点 (1天)
    ↓
Phase 2.1: statistics/util.go (0.5天)
    ↓
Phase 2.2: CVaR 修复 (0.5天)
    ↓
Phase 2.3: 方差公式修复 (0.5天)
    ↓
Phase 2.4: BL 后验收益 (1天)
    ↓
Phase 2.5: BL 权重优化 (0.5天)
    ↓
Phase 3.1: ETF 持仓数据 (1天)
    ↓
Phase 3.2: 因子数据接入 (1天)
    ↓
Phase 3.3: 股息历史数据 (0.5天)
    ↓
Phase 3.4: 国债收益率 (0.5天，可选)
```

**总计**: ~9-10 天

---

## 附录 A：受影响文件清单

| 文件 | Phase | 改动类型 |
|------|-------|----------|
| `backend/config/financial.go` | 1 | **新增** |
| `backend/services/dividend_service.go` | 1,3 | **新增** |
| `backend/services/portfolio_analytics.go` | 1,2 | 修改 |
| `backend/services/risk_models.go` | 1,2 | 修改 |
| `backend/services/scenario_analysis.go` | 1 | 修改 |
| `backend/services/portfolio_optimizer.go` | 1 | 修改 |
| `backend/services/optimization/mpt_optimizer.go` | 1 | 修改 |
| `backend/services/optimization/black_litterman.go` | 1,2 | 修改 |
| `backend/services/optimization/risk_parity.go` | 1 | 修改 |
| `backend/services/risk_budget_service.go` | 1 | 修改 |
| `backend/services/factor/fama_french.go` | 3 | 修改 |
| `backend/services/factor_data_service.go` | 3 | 修改 |
| `backend/handlers/portfolio_handler.go` | 1 | 修改 |
| `backend/handlers/portfolio_handler_risk.go` | 1 | 修改 |
| `backend/handlers/backtest_handler.go` | 1 | 修改 |
| `backend/services/statistics/util.go` | 2 | **新增** |
| `backend/services/treasury_rate_service.go` | 3 | **新增** (可选) |
| `backend/services/etf_holdings_service.go` | 3 | **新增** |

## 附录 B：金融公式参考

### CVaR (Expected Shortfall)
```
CVaR_α = μ - σ × φ(z_α) / (1 - α)

其中:
- μ: 收益率均值
- σ: 收益率标准差
- z_α: 标准正态分布的 α 分位数
- φ: 标准正态 PDF
- α: 置信水平 (如 0.95)
```

### Black-Litterman 后验收益
```
E[R] = [(τΣ)⁻¹ + P'Ω⁻¹P]⁻¹ × [(τΣ)⁻¹Π + P'Ω⁻¹Q]

其中:
- τ: 不确定性参数 (通常 0.025 ~ 0.05)
- Σ: 协方差矩阵
- P: 观点矩阵 (k × n)
- Q: 观点收益向量 (k × 1)
- Ω: 观点不确定性矩阵 (k × k)
- Π: 市场隐含均衡收益
```

### BL 最优权重
```
w* = (δΣ)⁻¹ × E[R]

其中:
- δ: 风险厌恶系数 (通常 2.0 ~ 3.0)
- Σ: 后验协方差矩阵
- E[R]: 后验收益向量
```

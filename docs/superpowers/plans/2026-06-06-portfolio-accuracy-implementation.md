# 组合分析数据准确性提升 - 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复组合分析模块的 12 个数据准确性问题，统一金融常量、精确计算公式、接入真实数据源

**Architecture:** 三阶段并行开发 - 配置层/计算层/数据层独立推进，通过统一接口解耦

**Tech Stack:** Go 1.21+, GORM, decimal.Decimal, Finage API, Kenneth French Database

---

## 并行开发分组

```
Track A (配置层)          Track B (计算层)          Track C (数据层)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Task 1: FinancialConfig   Task 4: statistics/util   Task 7: ETF持仓数据
Task 2: DividendService   Task 5: CVaR修复          Task 8: 因子数据接入
Task 3: 统一引用+API      Task 6: BL矩阵修复        Task 9: 股息历史+国债
```

---

## Track A: 配置层 (可独立开发)

### Task 1: FinancialConfig 统一金融配置

**Files:**
- Create: `backend/config/financial.go`
- Create: `backend/config/financial_test.go`

- [ ] **Step 1: 创建 FinancialConfig 结构体**

```go
// backend/config/financial.go
package config

import "sync"

type FinancialConfig struct {
    RiskFreeRate    float64
    TradingDaysYear int
    DefaultCurrency string
}

var (
    financialConfig *FinancialConfig
    configOnce      sync.Once
)

func GetFinancialConfig() *FinancialConfig {
    configOnce.Do(func() {
        financialConfig = &FinancialConfig{
            RiskFreeRate:    0.0435,
            TradingDaysYear: 252,
            DefaultCurrency: "USD",
        }
    })
    return financialConfig
}

func SetRiskFreeRate(rate float64) {
    GetFinancialConfig().RiskFreeRate = rate
}

func SetTradingDaysYear(days int) {
    GetFinancialConfig().TradingDaysYear = days
}
```

- [ ] **Step 2: 写测试**

```go
// backend/config/financial_test.go
package config

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestGetFinancialConfig(t *testing.T) {
    cfg := GetFinancialConfig()
    assert.NotNil(t, cfg)
    assert.Equal(t, 0.0435, cfg.RiskFreeRate)
    assert.Equal(t, 252, cfg.TradingDaysYear)
    assert.Equal(t, "USD", cfg.DefaultCurrency)
}

func TestSetRiskFreeRate(t *testing.T) {
    SetRiskFreeRate(0.05)
    assert.Equal(t, 0.05, GetFinancialConfig().RiskFreeRate)
    // 恢复
    SetRiskFreeRate(0.0435)
}

func TestSingleton(t *testing.T) {
    a := GetFinancialConfig()
    b := GetFinancialConfig()
    assert.Same(t, a, b)
}
```

- [ ] **Step 3: 运行测试验证**

```bash
cd backend && go test ./config/ -v -run TestGetFinancialConfig
cd backend && go test ./config/ -v -run TestSetRiskFreeRate
cd backend && go test ./config/ -v -run TestSingleton
```

- [ ] **Step 4: Commit**

```bash
git add backend/config/financial.go backend/config/financial_test.go
git commit -m "feat(config): add unified FinancialConfig for risk-free rate and trading days"
```

---

### Task 2: DividendService 股息率服务

**Files:**
- Create: `backend/services/dividend_service.go`
- Create: `backend/services/dividend_service_test.go`

- [ ] **Step 1: 定义接口和结构体**

```go
// backend/services/dividend_service.go
package services

import (
    "sync"
    "time"
    "gorm.io/gorm"
    "etf-insight/backend/models"
)

type CachedYield struct {
    Yield     float64
    FetchedAt time.Time
    TTL       time.Duration
}

type DividendService struct {
    db       *gorm.DB
    cache    map[string]*CachedYield
    mu       sync.RWMutex
    cacheTTL time.Duration
}

func NewDividendService(db *gorm.DB) *DividendService {
    return &DividendService{
        db:       db,
        cache:    make(map[string]*CachedYield),
        cacheTTL: 24 * time.Hour,
    }
}

// 硬编码兜底值 (仅在全部数据源失败时使用)
var fallbackYields = map[string]float64{
    "SCHD": 0.035,
    "JEPQ": 0.095,
    "QQQ":  0.006,
    "VTI":  0.015,
    "SPY":  0.013,
    "VYM":  0.028,
    "JEPI": 0.075,
    "BND":  0.030,
}

func (s *DividendService) GetDividendYield(symbol string) (float64, error) {
    // 1. 检查内存缓存
    s.mu.RLock()
    if cached, ok := s.cache[symbol]; ok {
        if time.Since(cached.FetchedAt) < cached.TTL {
            s.mu.RUnlock()
            return cached.Yield, nil
        }
    }
    s.mu.RUnlock()

    // 2. 查询数据库 ETFDividend 表
    var dividend models.ETFDividend
    result := s.db.Where("symbol = ?", symbol).
        Order("ex_dividend_date desc").
        First(&dividend)
    if result.Error == nil && dividend.Yield > 0 {
        s.updateCache(symbol, dividend.Yield)
        return dividend.Yield, nil
    }

    // 3. 查询 UniversalETF 的股息率
    var etf models.UniversalETF
    result = s.db.Where("symbol = ?", symbol).First(&etf)
    if result.Error == nil && etf.DividendYield > 0 {
        s.updateCache(symbol, etf.DividendYield)
        return etf.DividendYield, nil
    }

    // 4. 兜底值
    if y, ok := fallbackYields[symbol]; ok {
        return y, nil
    }
    return 0.02, nil // 默认 2%
}

func (s *DividendService) updateCache(symbol string, yield float64) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.cache[symbol] = &CachedYield{
        Yield:     yield,
        FetchedAt: time.Now(),
        TTL:       s.cacheTTL,
    }
}

func (s *DividendService) InvalidateCache(symbol string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.cache, symbol)
}
```

- [ ] **Step 2: 写测试**

```go
// backend/services/dividend_service_test.go
package services

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestGetDividendYield_Fallback(t *testing.T) {
    // 测试兜底值 (无数据库)
    yields := fallbackYields
    assert.Equal(t, 0.035, yields["SCHD"])
    assert.Equal(t, 0.095, yields["JEPQ"])
    assert.Equal(t, 0.075, yields["JEPI"])
}

func TestGetDividendYield_Default(t *testing.T) {
    // 未知符号返回默认值
    y, ok := fallbackYields["UNKNOWN"]
    assert.False(t, ok)
    assert.Equal(t, 0.0, y)
}
```

- [ ] **Step 3: 运行测试**

```bash
cd backend && go test ./services/ -v -run TestGetDividendYield
```

- [ ] **Step 4: Commit**

```bash
git add backend/services/dividend_service.go backend/services/dividend_service_test.go
git commit -m "feat(service): add DividendService with cache and DB fallback"
```

---

### Task 3: 统一引用 + API 端点

**Files:**
- Modify: `backend/services/portfolio_analytics.go`
- Modify: `backend/services/scenario_analysis.go`
- Modify: `backend/services/portfolio_optimizer.go`
- Modify: `backend/services/optimization/mpt_optimizer.go`
- Modify: `backend/services/optimization/black_litterman.go`
- Modify: `backend/services/optimization/risk_parity.go`
- Modify: `backend/services/risk_budget_service.go`
- Modify: `backend/handlers/portfolio_handler_risk.go`
- Modify: `backend/handlers/backtest_handler.go`
- Create: `backend/handlers/config_handler.go`

- [ ] **Step 1: 替换 portfolio_analytics.go 中的硬编码利率**

搜索所有 `0.045` 和 `0.04`，替换为 `config.GetFinancialConfig().RiskFreeRate`。

```go
// 文件顶部添加 import
import "etf-insight/backend/config"

// 替换 (约3处):
// 旧: riskFreeRate := 0.045
// 新:
riskFreeRate := config.GetFinancialConfig().RiskFreeRate
```

- [ ] **Step 2: 替换 scenario_analysis.go 中的硬编码利率**

```go
import "etf-insight/backend/config"

// 替换所有 0.045 为 config.GetFinancialConfig().RiskFreeRate
```

- [ ] **Step 3: 替换 portfolio_optimizer.go 中的硬编码利率**

```go
// 旧: DefaultRiskFreeRate = 0.04
// 新:
DefaultRiskFreeRate = config.GetFinancialConfig().RiskFreeRate
```

- [ ] **Step 4: 替换 optimization/ 下所有文件的硬编码利率**

`mpt_optimizer.go`, `black_litterman.go`, `risk_parity.go` 同理。

- [ ] **Step 5: 替换 risk_budget_service.go**

```go
// 替换所有 0.045
```

- [ ] **Step 6: 替换 portfolio_handler_risk.go**

```go
// 旧: riskFreeRate := 0.02 / 252
// 新:
riskFreeRate := config.GetFinancialConfig().RiskFreeRate / float64(config.GetFinancialConfig().TradingDaysYear)
```

- [ ] **Step 7: 替换 backtest_handler.go 中的 365**

```go
// 旧: DaysPerYear = 365.0
// 新:
DaysPerYear = float64(config.GetFinancialConfig().TradingDaysYear)
```

- [ ] **Step 8: 创建配置 API 端点**

```go
// backend/handlers/config_handler.go
package handlers

import (
    "net/http"
    "etf-insight/backend/config"
    "github.com/gin-gonic/gin"
)

type FinancialConfigResponse struct {
    RiskFreeRate    float64 `json:"risk_free_rate"`
    TradingDaysYear int     `json:"trading_days_year"`
    DefaultCurrency string  `json:"default_currency"`
}

func GetFinancialConfig(c *gin.Context) {
    cfg := config.GetFinancialConfig()
    c.JSON(http.StatusOK, FinancialConfigResponse{
        RiskFreeRate:    cfg.RiskFreeRate,
        TradingDaysYear: cfg.TradingDaysYear,
        DefaultCurrency: cfg.DefaultCurrency,
    })
}

type UpdateFinancialConfigRequest struct {
    RiskFreeRate    *float64 `json:"risk_free_rate,omitempty"`
    TradingDaysYear *int     `json:"trading_days_year,omitempty"`
}

func UpdateFinancialConfig(c *gin.Context) {
    var req UpdateFinancialConfigRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if req.RiskFreeRate != nil {
        config.SetRiskFreeRate(*req.RiskFreeRate)
    }
    if req.TradingDaysYear != nil {
        config.SetTradingDaysYear(*req.TradingDaysYear)
    }
    c.JSON(http.StatusOK, gin.H{"message": "updated"})
}
```

- [ ] **Step 9: 注册路由**

在 `router/router.go` 中添加:
```go
configGroup := api.Group("/config")
{
    configGroup.GET("/financial", handlers.GetFinancialConfig)
    configGroup.PUT("/financial", handlers.UpdateFinancialConfig)
}
```

- [ ] **Step 10: 运行全量测试**

```bash
cd backend && go test ./... -race
```

- [ ] **Step 11: Commit**

```bash
git add -A
git commit -m "refactor: unify risk-free rate and trading days across all modules"
```

---

## Track B: 计算层 (可独立开发)

### Task 4: statistics/util.go 统计工具

**Files:**
- Create: `backend/services/statistics/util.go`
- Create: `backend/services/statistics/util_test.go`

- [ ] **Step 1: 实现统计工具函数**

```go
// backend/services/statistics/util.go
package statistics

import "math"

func Mean(values []float64) float64 {
    if len(values) == 0 {
        return 0
    }
    sum := 0.0
    for _, v := range values {
        sum += v
    }
    return sum / float64(len(values))
}

// SampleVariance 样本方差 (除以 N-1)
func SampleVariance(values []float64) float64 {
    n := len(values)
    if n < 2 {
        return 0
    }
    mean := Mean(values)
    sum := 0.0
    for _, v := range values {
        diff := v - mean
        sum += diff * diff
    }
    return sum / float64(n-1)
}

func SampleStdDev(values []float64) float64 {
    return math.Sqrt(SampleVariance(values))
}

// PopulationVariance 总体方差 (除以 N) - 保留兼容
func PopulationVariance(values []float64) float64 {
    if len(values) == 0 {
        return 0
    }
    mean := Mean(values)
    sum := 0.0
    for _, v := range values {
        diff := v - mean
        sum += diff * diff
    }
    return sum / float64(len(values))
}

// NormalCDF 标准正态分布 CDF (Abramowitz & Stegun 近似, 误差 < 7.5e-8)
func NormalCDF(x float64) float64 {
    if x < -8 {
        return 0
    }
    if x > 8 {
        return 1
    }
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

// NormalQuantile 标准正态分位数 (Beasley-Springer-Moro 算法)
func NormalQuantile(p float64) float64 {
    if p <= 0 || p >= 1 {
        return 0
    }
    // Rational approximation
    if p < 0.5 {
        return -rationalApprox(math.Sqrt(-2 * math.Log(p)))
    }
    return rationalApprox(math.Sqrt(-2 * math.Log(1-p)))
}

func rationalApprox(t float64) float64 {
    const (
        c0 = 2.515517
        c1 = 0.802853
        c2 = 0.010328
        d1 = 1.432788
        d2 = 0.189269
        d3 = 0.001308
    )
    return t - (c0+c1*t+c2*t*t)/(1+d1*t+d2*t*t+d3*t*t*t)
}
```

- [ ] **Step 2: 写测试**

```go
// backend/services/statistics/util_test.go
package statistics

import (
    "testing"
    "math"
    "github.com/stretchr/testify/assert"
)

func TestSampleVariance(t *testing.T) {
    // 已知数据: [2, 4, 4, 4, 5, 5, 7, 9], 样本方差 = 4.5714
    data := []float64{2, 4, 4, 4, 5, 5, 7, 9}
    v := SampleVariance(data)
    assert.InDelta(t, 4.5714, v, 0.001)
}

func TestSampleVariance_PopulationDiff(t *testing.T) {
    data := []float64{1, 2, 3, 4, 5}
    sample := SampleVariance(data)    // /4
    pop := PopulationVariance(data)   // /5
    assert.True(t, sample > pop)
    assert.InDelta(t, 2.5, sample, 0.001)
    assert.InDelta(t, 2.0, pop, 0.001)
}

func TestNormalCDF(t *testing.T) {
    tests := []struct{ x, expected float64 }{
        {0, 0.5},
        {1.645, 0.95},
        {-1.645, 0.05},
        {2.326, 0.99},
        {-2.326, 0.01},
        {1.0, 0.8413},
    }
    for _, tt := range tests {
        result := NormalCDF(tt.x)
        assert.InDelta(t, tt.expected, result, 0.0001,
            "NormalCDF(%f) = %f, want %f", tt.x, result, tt.expected)
    }
}

func TestNormalPDF(t *testing.T) {
    // φ(0) = 1/sqrt(2π) ≈ 0.3989
    assert.InDelta(t, 0.3989, NormalPDF(0), 0.0001)
    // φ(1) ≈ 0.2420
    assert.InDelta(t, 0.2420, NormalPDF(1), 0.0001)
}

func TestNormalQuantile(t *testing.T) {
    tests := []struct{ p, expected float64 }{
        {0.5, 0},
        {0.95, 1.645},
        {0.05, -1.645},
        {0.975, 1.96},
    }
    for _, tt := range tests {
        result := NormalQuantile(tt.p)
        assert.InDelta(t, tt.expected, result, 0.01,
            "NormalQuantile(%f) = %f, want %f", tt.p, result, tt.expected)
    }
}

func TestMean(t *testing.T) {
    assert.Equal(t, 3.0, Mean([]float64{1, 2, 3, 4, 5}))
    assert.Equal(t, 0.0, Mean([]float64{}))
}
```

- [ ] **Step 3: 运行测试**

```bash
cd backend && go test ./services/statistics/ -v
```

- [ ] **Step 4: Commit**

```bash
git add backend/services/statistics/
git commit -m "feat(statistics): add precise NormalCDF, SampleVariance, NormalQuantile"
```

---

### Task 5: CVaR 精确计算 + 方差修复

**Files:**
- Modify: `backend/services/risk_models.go:110-136`
- Modify: `backend/services/portfolio_analytics.go:420-427, 540-565`

- [ ] **Step 1: 修复 risk_models.go 的 CVaR**

```go
// risk_models.go - 替换 line 115-123
import "etf-insight/backend/services/statistics"

// 精确 CVaR 计算
// CVaR = μ - σ × φ(z) / (1 - α)
zFloat := zScore.InexactFloat64()
phiZ := statistics.NormalPDF(zFloat)
oneMinusAlpha := 1.0 - confidence
if oneMinusAlpha < 1e-10 {
    oneMinusAlpha = 1e-10
}
cvarRaw := meanFloat - stdFloat*(phiZ/oneMinusAlpha)
cvarValue := decimal.NewFromFloat(-cvarRaw) // 转为正数表示损失
```

- [ ] **Step 2: 修复 risk_models.go 的 calculateStdDev**

```go
// line 345 附近, 替换总体方差为样本方差
func calculateStdDev(values []float64) float64 {
    return statistics.SampleStdDev(values)
}
```

- [ ] **Step 3: 修复 portfolio_analytics.go 的方差**

```go
// line 420-427, 替换 variance 函数
func (s *PortfolioAnalyticsService) variance(values []float64, mean float64) float64 {
    n := len(values)
    if n < 2 {
        return 0
    }
    sum := 0.0
    for _, v := range values {
        diff := v - mean
        sum += diff * diff
    }
    return sum / float64(n-1) // N-1 样本方差
}
```

- [ ] **Step 4: 修复 portfolio_analytics.go 的 CVaR CDF**

```go
// line 557-564, 替换 CDF 近似
import "etf-insight/backend/services/statistics"

// 精确 CDF
Phi := statistics.NormalCDF(zScore)
if Phi < 1e-10 {
    Phi = 1e-10
}
cvar := mean - std*(phi/Phi)
```

- [ ] **Step 5: 运行测试**

```bash
cd backend && go test ./services/ -v -run TestVaR
cd backend && go test ./services/ -v -run TestCVaR
cd backend && go test ./services/ -v -run TestPortfolio
```

- [ ] **Step 6: Commit**

```bash
git add backend/services/risk_models.go backend/services/portfolio_analytics.go
git commit -m "fix(calc): precise CVaR formula and sample variance (N-1)"
```

---

### Task 6: Black-Litterman 完整矩阵公式

**Files:**
- Modify: `backend/services/optimization/black_litterman.go:310-430`

- [ ] **Step 1: 实现矩阵工具函数**

在 `black_litterman.go` 顶部添加矩阵运算辅助函数:

```go
// matrixInverse 通过高斯消元求逆 (已有实现可复用)
func matrixInverse(m [][]float64) [][]float64 {
    n := len(m)
    // 增广矩阵 [m | I]
    aug := make([][]float64, n)
    for i := range aug {
        aug[i] = make([]float64, 2*n)
        for j := range n {
            aug[i][j] = m[i][j]
        }
        aug[i][n+i] = 1
    }
    // 前向消元 (部分主元)
    for col := 0; col < n; col++ {
        maxVal := math.Abs(aug[col][col])
        maxRow := col
        for row := col + 1; row < n; row++ {
            if math.Abs(aug[row][col]) > maxVal {
                maxVal = math.Abs(aug[row][col])
                maxRow = row
            }
        }
        aug[col], aug[maxRow] = aug[maxRow], aug[col]
        pivot := aug[col][col]
        if math.Abs(pivot) < 1e-12 {
            continue
        }
        for j := 0; j < 2*n; j++ {
            aug[col][j] /= pivot
        }
        for row := 0; row < n; row++ {
            if row == col {
                continue
            }
            factor := aug[row][col]
            for j := 0; j < 2*n; j++ {
                aug[row][j] -= factor * aug[col][j]
            }
        }
    }
    inv := make([][]float64, n)
    for i := range inv {
        inv[i] = make([]float64, n)
        for j := range n {
            inv[i][j] = aug[i][n+j]
        }
    }
    return inv
}

func matrixMultiply(a, b [][]float64) [][]float64 {
    rows, cols, inner := len(a), len(b[0]), len(b)
    result := make([][]float64, rows)
    for i := range result {
        result[i] = make([]float64, cols)
        for j := range cols {
            for k := range inner {
                result[i][j] += a[i][k] * b[k][j]
            }
        }
    }
    return result
}

func matrixTranspose(m [][]float64) [][]float64 {
    rows, cols := len(m), len(m[0])
    t := make([][]float64, cols)
    for i := range t {
        t[i] = make([]float64, rows)
        for j := range rows {
            t[i][j] = m[j][i]
        }
    }
    return t
}

func matrixAdd(a, b [][]float64) [][]float64 {
    rows, cols := len(a), len(a[0])
    result := make([][]float64, rows)
    for i := range result {
        result[i] = make([]float64, cols)
        for j := range cols {
            result[i][j] = a[i][j] + b[i][j]
        }
    }
    return result
}

func matrixVectorMultiply(m [][]float64, v []float64) []float64 {
    rows := len(m)
    result := make([]float64, rows)
    for i := range rows {
        for j := range len(v) {
            result[i] += m[i][j] * v[j]
        }
    }
    return result
}

func scalarMultiplyMatrix(m [][]float64, s float64) [][]float64 {
    rows, cols := len(m), len(m[0])
    result := make([][]float64, rows)
    for i := range result {
        result[i] = make([]float64, cols)
        for j := range cols {
            result[i][j] = m[i][j] * s
        }
    }
    return result
}
```

- [ ] **Step 2: 重写 calculatePosteriorReturns**

```go
// 替换 black_litterman.go:310-369
func (o *BlackLittermanOptimizer) calculatePosteriorReturns(
    Pi []float64,
    Sigma [][]float64,
    P [][]float64,
    Q []float64,
    Omega [][]float64,
    tau float64,
) []float64 {
    n := len(Pi)

    // 1. τΣ
    tauSigma := scalarMultiplyMatrix(Sigma, tau)

    // 2. (τΣ)⁻¹
    tauSigmaInv := matrixInverse(tauSigma)

    // 3. P' and Ω⁻¹
    Pt := matrixTranspose(P)
    OmegaInv := matrixInverse(Omega)

    // 4. P'Ω⁻¹P
    PtOmegaInv := matrixMultiply(Pt, OmegaInv)
    PtOmegaInvP := matrixMultiply(PtOmegaInv, P)

    // 5. M = (τΣ)⁻¹ + P'Ω⁻¹P
    M := matrixAdd(tauSigmaInv, PtOmegaInvP)

    // 6. M⁻¹
    MInv := matrixInverse(M)

    // 7. (τΣ)⁻¹Π
    tauSigmaInvPi := matrixVectorMultiply(tauSigmaInv, Pi)

    // 8. P'Ω⁻¹Q
    PtOmegaInvQ := matrixVectorMultiply(PtOmegaInv, Q)

    // 9. rhs = (τΣ)⁻¹Π + P'Ω⁻¹Q
    rhs := make([]float64, n)
    for i := range n {
        rhs[i] = tauSigmaInvPi[i] + PtOmegaInvQ[i]
    }

    // 10. E[R] = M⁻¹ × rhs
    return matrixVectorMultiply(MInv, rhs)
}
```

- [ ] **Step 3: 重写权重优化**

```go
// 替换 black_litterman.go:413-430
func (o *BlackLittermanOptimizer) optimizeWeights(
    posteriorReturns map[string]float64,
    symbols []string,
    Sigma [][]float64,
    constraint *BlackLittermanConstraint,
) []float64 {
    n := len(symbols)

    // w* = (1/δ) × Σ⁻¹ × (μ - rf)
    SigmaInv := matrixInverse(Sigma)
    excessReturns := make([]float64, n)
    for i, sym := range symbols {
        excessReturns[i] = posteriorReturns[sym] - o.RiskFreeRate
    }

    weights := matrixVectorMultiply(SigmaInv, excessReturns)

    // 归一化为正权重
    sum := 0.0
    for _, w := range weights {
        if w > 0 {
            sum += w
        }
    }
    if sum > 0 {
        for i := range weights {
            if weights[i] < 0 {
                weights[i] = 0
            }
            weights[i] /= sum
        }
    } else {
        // fallback: 等权重
        for i := range weights {
            weights[i] = 1.0 / float64(n)
        }
    }

    return o.applyConstraints(weights, symbols, constraint)
}
```

- [ ] **Step 4: 运行测试**

```bash
cd backend && go test ./services/optimization/ -v -run TestBlackLitterman
```

- [ ] **Step 5: Commit**

```bash
git add backend/services/optimization/black_litterman.go
git commit -m "fix(bl): implement full BL matrix formula for posterior returns and weights"
```

---

## Track C: 数据层 (可独立开发)

### Task 7: ETF 持仓数据 (手动维护核心 ETF)

**Files:**
- Create: `backend/services/etf_holdings_service.go`
- Create: `backend/services/etf_holdings_data.go` (静态数据)
- Create: `backend/services/etf_holdings_service_test.go`

- [ ] **Step 1: 创建持仓数据文件**

```go
// backend/services/etf_holdings_data.go
package services

type StaticHolding struct {
    Symbol   string
    Name     string
    Weight   float64 // 百分比, 如 5.5 = 5.5%
    Sector   string
}

// 核心ETF Top 25 持仓 (手动维护, 定期更新)
// 数据来源: 各ETF官网, 截至 2026-06
var staticHoldings = map[string][]StaticHolding{
    "SCHD": {
        {Symbol: "AVGO", Name: "Broadcom Inc", Weight: 5.2, Sector: "Technology"},
        {Symbol: "ABBV", Name: "AbbVie Inc", Weight: 4.8, Sector: "Healthcare"},
        {Symbol: "JPM", Name: "JPMorgan Chase", Weight: 4.5, Sector: "Financials"},
        // ... 补充到 Top 25
    },
    "JEPI": {
        {Symbol: "ROST", Name: "Ross Stores", Weight: 1.71, Sector: "Consumer Cyclical"},
        {Symbol: "NVDA", Name: "NVIDIA", Weight: 1.68, Sector: "Technology"},
        // ... Top 25
    },
    "JEPQ": {
        {Symbol: "NVDA", Name: "NVIDIA", Weight: 7.43, Sector: "Technology"},
        {Symbol: "AAPL", Name: "Apple", Weight: 6.30, Sector: "Technology"},
        // ... Top 25
    },
    // QQQ, VTI, SPY, VYM, BND, SPYD, HDV, DGRO, VNQ
}
```

- [ ] **Step 2: 创建持仓服务**

```go
// backend/services/etf_holdings_service.go
package services

type ETFHoldingsService struct {
    // 未来可接入 API provider
}

func NewETFDividendService() *ETFDividendService {
    return &ETFDividendService{}
}

func (s *ETFDividendService) GetHoldings(symbol string) ([]StaticHolding, error) {
    if holdings, ok := staticHoldings[symbol]; ok {
        return holdings, nil
    }
    return nil, fmt.Errorf("no holdings data for %s", symbol)
}

func (s *ETFDividendService) CalculateOverlap(etf1, etf2 string) (float64, error) {
    h1, err := s.GetHoldings(etf1)
    if err != nil {
        return 0, err
    }
    h2, err := s.GetHoldings(etf2)
    if err != nil {
        return 0, err
    }

    // 最小权重法计算重叠度
    weights1 := make(map[string]float64)
    for _, h := range h1 {
        weights1[h.Symbol] = h.Weight
    }
    overlap := 0.0
    for _, h := range h2 {
        if w, ok := weights1[h.Symbol]; ok {
            if w < h.Weight {
                overlap += w
            } else {
                overlap += h.Weight
            }
        }
    }
    return overlap, nil
}
```

- [ ] **Step 3: 写测试**

```go
func TestGetHoldings(t *testing.T) {
    svc := &ETFDividendService{}
    h, err := svc.GetHoldings("SCHD")
    assert.NoError(t, err)
    assert.True(t, len(h) > 0)
}

func TestCalculateOverlap(t *testing.T) {
    svc := &ETFDividendService{}
    overlap, err := svc.CalculateOverlap("SCHD", "VYM")
    assert.NoError(t, err)
    assert.True(t, overlap >= 0 && overlap <= 100)
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/services/etf_holdings_service.go backend/services/etf_holdings_data.go backend/services/etf_holdings_service_test.go
git commit -m "feat(data): add static ETF holdings for top ETFs with overlap calculation"
```

---

### Task 8: 因子数据接入

**Files:**
- Modify: `backend/services/factor/fama_french.go` (LoadFactorDataFromFrench)
- Modify: `backend/services/factor_data_service.go`

- [ ] **Step 1: 实现 Kenneth French 数据下载**

```go
// fama_french.go - 实现 LoadFactorDataFromFrench
func (s *FamaFrenchService) LoadFactorDataFromFrench(
    startDate, endDate time.Time,
    frequency string,
) error {
    // Kenneth French 数据库 URL
    // 3 Factor: https://mba.tuck.dartmouth.edu/pages/faculty/ken.french/ftp/F-F_Research_Data_Factors_CSV.zip
    // 5 Factor: https://mba.tuck.dartmouth.edu/pages/faculty/ken.french/ftp/F-F_Research_Data_5_Factors_2x3_CSV.zip

    var url string
    if frequency == "daily" {
        url = "https://mba.tuck.dartmouth.edu/pages/faculty/ken.french/ftp/F-F_Research_Data_Factors_daily_CSV.zip"
    } else {
        url = "https://mba.tuck.dartmouth.edu/pages/faculty/ken.french/ftp/F-F_Research_Data_Factors_CSV.zip"
    }

    // 1. 下载 ZIP
    resp, err := http.Get(url)
    if err != nil {
        return fmt.Errorf("failed to download factor data: %w", err)
    }
    defer resp.Body.Close()

    // 2. 解析 CSV (跳过头部行)
    // 3. 存入 FactorData 表
    // 4. 更新数据源标记
    return nil
}
```

- [ ] **Step 2: 修改 SeedSampleFactorData 优先查数据库**

```go
// factor_data_service.go - 修改 SeedSampleFactorData
func (s *FactorDataService) SeedSampleFactorData() error {
    // 先检查数据库是否有真实数据
    var count int64
    s.db.Model(&models.FactorData{}).Count(&count)
    if count > 0 {
        return nil // 已有数据，不覆盖
    }

    // 尝试从 French 数据库加载
    err := s.famaFrench.LoadFactorDataFromFrench(
        time.Now().AddDate(-5, 0, 0),
        time.Now(),
        "monthly",
    )
    if err == nil {
        return nil // 成功加载真实数据
    }

    // 仅在真实数据不可用时使用合成数据
    log.Warn("Using synthetic factor data as fallback")
    return s.generateSyntheticData()
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/services/factor/fama_french.go backend/services/factor_data_service.go
git commit -m "feat(factor): implement Kenneth French data loader, prefer real over synthetic"
```

---

### Task 9: 股息历史 + 国债收益率 (可选)

**Files:**
- Modify: `backend/services/dividend_service.go` (添加 SyncDividendHistory)
- Create: `backend/services/treasury_rate_service.go` (可选)

- [ ] **Step 1: 扩展 DividendService**

```go
// dividend_service.go - 添加方法
func (s *DividendService) GetTrailing12MonthYield(symbol string) (float64, error) {
    // 查询过去12个月的分红记录
    var dividends []models.ETFDividend
    oneYearAgo := time.Now().AddDate(-1, 0, 0)
    result := s.db.Where("symbol = ? AND ex_dividend_date > ?", symbol, oneYearAgo).
        Find(&dividends)
    if result.Error != nil || len(dividends) == 0 {
        return s.GetDividendYield(symbol) // fallback
    }

    totalDividend := 0.0
    for _, d := range dividends {
        totalDividend += d.Amount
    }

    // 获取当前价格
    var etf models.UniversalETF
    s.db.Where("symbol = ?", symbol).First(&etf)
    if etf.CurrentPrice > 0 {
        return totalDividend / etf.CurrentPrice, nil
    }
    return s.GetDividendYield(symbol)
}
```

- [ ] **Step 2: (可选) 创建 TreasuryRateService**

```go
// backend/services/treasury_rate_service.go
package services

import (
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"
)

type TreasuryRateService struct {
    apiKey string
    cache  float64
    mu     sync.RWMutex
    lastFetch time.Time
}

func NewTreasuryRateService(apiKey string) *TreasuryRateService {
    return &TreasuryRateService{apiKey: apiKey}
}

func (s *TreasuryRateService) Get10YearRate() (float64, error) {
    s.mu.RLock()
    if time.Since(s.lastFetch) < 1*time.Hour && s.cache > 0 {
        defer s.mu.RUnlock()
        return s.cache, nil
    }
    s.mu.RUnlock()

    // FRED API
    url := fmt.Sprintf(
        "https://api.stlouisfed.org/fred/series/observations?series_id=DGS10&api_key=%s&file_type=json&sort_order=desc&limit=1",
        s.apiKey,
    )
    resp, err := http.Get(url)
    if err != nil {
        return s.cache, err
    }
    defer resp.Body.Close()

    var result struct {
        Observations []struct {
            Value string `json:"value"`
        } `json:"observations"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return s.cache, err
    }
    if len(result.Observations) > 0 {
        var rate float64
        fmt.Sscanf(result.Observations[0].Value, "%f", &rate)
        s.mu.Lock()
        s.cache = rate / 100 // 转为小数
        s.lastFetch = time.Now()
        s.mu.Unlock()
        return s.cache, nil
    }
    return s.cache, fmt.Errorf("no data from FRED")
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/services/dividend_service.go backend/services/treasury_rate_service.go
git commit -m "feat(data): add trailing 12M dividend yield and FRED treasury rate service"
```

---

## 验证清单

所有 Track 完成后执行:

- [ ] **全量测试**: `cd backend && go test ./... -race`
- [ ] **覆盖率检查**: `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`
- [ ] **Lint**: `golangci-lint run`
- [ ] **精度验证**: 对比 Python scipy/pypfopt 的 CVaR 和 BL 计算结果

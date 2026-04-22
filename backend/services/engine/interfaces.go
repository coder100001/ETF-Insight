package engine

import (
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
)

// ================================
// 基础引擎接口
// ================================

// Engine 通用引擎接口
type Engine interface {
	Name() string        // 引擎名称
	Description() string // 引擎描述
	Version() string     // 引擎版本
	IsAvailable() bool   // 是否可用
	HealthCheck() error  // 健康检查
}

// ConfigurableEngine 可配置引擎接口
type ConfigurableEngine interface {
	Engine
	Configure(config interface{}) error // 配置引擎
	GetConfig() interface{}             // 获取当前配置
	ResetConfig() error                 // 重置配置
}

// ================================
// 组合分析引擎接口
// ================================

// PortfolioAnalysisEngine 组合分析引擎
type PortfolioAnalysisEngine interface {
	Engine

	// 组合收益计算
	CalculateReturns(portfolio *models.Portfolio, positions []*models.PortfolioPosition, prices []*models.Price) (*PortfolioReturns, error)
	CalculateRollingReturns(portfolio *models.Portfolio, positions []*models.PortfolioPosition, prices []*models.Price, windowSize int) ([]RollingReturn, error)

	// 风险指标计算
	CalculateRiskMetrics(portfolio *models.Portfolio, positions []*models.PortfolioPosition, prices []*models.Price) (*RiskMetrics, error)
	CalculateValueAtRisk(portfolio *models.Portfolio, positions []*models.PortfolioPosition, prices []*models.Price, confidenceLevel decimal.Decimal, horizonDays int) (*ValueAtRisk, error)

	// 情景分析
	RunScenarioAnalysis(portfolio *models.Portfolio, positions []*models.PortfolioPosition, scenarios []Scenario) ([]ScenarioResult, error)
	RunMonteCarloSimulation(portfolio *models.Portfolio, positions []*models.PortfolioPosition, numSimulations int, horizonDays int) (*MonteCarloResult, error)

	// 压力测试
	RunStressTest(portfolio *models.Portfolio, positions []*models.PortfolioPosition, stressScenarios []StressScenario) ([]StressTestResult, error)
}

// PortfolioReturns 组合收益计算结果
type PortfolioReturns struct {
	TotalReturn     decimal.Decimal // 总收益率
	AnnualReturn    decimal.Decimal // 年化收益率
	DailyReturn     decimal.Decimal // 日收益率
	WeeklyReturn    decimal.Decimal // 周收益率
	MonthlyReturn   decimal.Decimal // 月收益率
	QuarterlyReturn decimal.Decimal // 季度收益率
	YTDReturn       decimal.Decimal // 年初至今收益率
	MaxReturn       decimal.Decimal // 最大收益率
	MinReturn       decimal.Decimal // 最小收益率
	AvgReturn       decimal.Decimal // 平均收益率
	StdDevReturn    decimal.Decimal // 收益率标准差
	SharpeRatio     decimal.Decimal // 夏普比率
	SortinoRatio    decimal.Decimal // 索提诺比率
	CalmarRatio     decimal.Decimal // 卡尔玛比率
	TreynorRatio    decimal.Decimal // 特雷诺比率
}

// RollingReturn 滚动收益率
type RollingReturn struct {
	StartDate       time.Time       // 开始日期
	EndDate         time.Time       // 结束日期
	Return          decimal.Decimal // 收益率
	BenchmarkReturn decimal.Decimal // 基准收益率
	ExcessReturn    decimal.Decimal // 超额收益率
}

// RiskMetrics 风险指标
type RiskMetrics struct {
	Volatility        decimal.Decimal // 波动率（标准差）
	Beta              decimal.Decimal // Beta值
	Alpha             decimal.Decimal // Alpha值
	R2                decimal.Decimal // R平方值
	TrackingError     decimal.Decimal // 跟踪误差
	InformationRatio  decimal.Decimal // 信息比率
	MaxDrawdown       decimal.Decimal // 最大回撤
	DownsideDeviation decimal.Decimal // 下行偏差
	Skewness          decimal.Decimal // 偏度
	Kurtosis          decimal.Decimal // 峰度
	VaR95             decimal.Decimal // 95% VaR
	CVaR95            decimal.Decimal // 95% CVaR
}

// ValueAtRisk VaR计算结果
type ValueAtRisk struct {
	Value           decimal.Decimal // VaR值
	ConfidenceLevel decimal.Decimal // 置信水平
	HorizonDays     int             // 时间范围（天）
	Method          string          // 计算方法：historical/parametric/montecarlo
	Components      []VaRComponent  // 成分VaR
}

// VaRComponent 成分VaR
type VaRComponent struct {
	AssetID        uint            // 资产ID
	Symbol         string          // 资产代码
	Weight         decimal.Decimal // 权重
	VaR            decimal.Decimal // 贡献VaR
	MarginalVaR    decimal.Decimal // 边际VaR
	IncrementalVaR decimal.Decimal // 增量VaR
}

// Scenario 情景定义
type Scenario struct {
	Name         string          // 情景名称
	Description  string          // 情景描述
	MarketReturn decimal.Decimal // 市场收益率
	Volatility   decimal.Decimal // 市场波动率
	Correlation  decimal.Decimal // 相关性变化
	Duration     int             // 持续时间（天）
}

// ScenarioResult 情景分析结果
type ScenarioResult struct {
	Scenario        Scenario        // 情景定义
	PortfolioReturn decimal.Decimal // 组合收益率
	PortfolioValue  decimal.Decimal // 组合价值
	VaR             decimal.Decimal // VaR值
	CVaR            decimal.Decimal // CVaR值
	MaxDrawdown     decimal.Decimal // 最大回撤
	RiskMetrics     RiskMetrics     // 风险指标
}

// MonteCarloResult 蒙特卡洛模拟结果
type MonteCarloResult struct {
	NumSimulations     int                // 模拟次数
	HorizonDays        int                // 时间范围
	MeanReturn         decimal.Decimal    // 平均收益率
	MedianReturn       decimal.Decimal    // 中位数收益率
	StdDevReturn       decimal.Decimal    // 收益率标准差
	ConfidenceInterval []ConfidenceBound  // 置信区间
	ProbabilityDist    []ProbabilityPoint // 概率分布
}

// ConfidenceBound 置信区间边界
type ConfidenceBound struct {
	ConfidenceLevel decimal.Decimal // 置信水平
	LowerBound      decimal.Decimal // 下限
	UpperBound      decimal.Decimal // 上限
}

// ProbabilityPoint 概率分布点
type ProbabilityPoint struct {
	Return      decimal.Decimal // 收益率
	Probability decimal.Decimal // 概率
}

// StressScenario 压力测试情景
type StressScenario struct {
	Name         string                     // 情景名称
	Description  string                     // 情景描述
	MarketShock  decimal.Decimal            // 市场冲击（百分比）
	SectorShocks map[string]decimal.Decimal // 行业冲击
	Duration     int                        // 持续时间（天）
}

// StressTestResult 压力测试结果
type StressTestResult struct {
	Scenario       StressScenario  // 压力情景
	PortfolioLoss  decimal.Decimal // 组合损失
	MaxDrawdown    decimal.Decimal // 最大回撤
	RecoveryPeriod int             // 恢复期（天）
	RiskMetrics    RiskMetrics     // 风险指标
}

// ================================
// ETF分析引擎接口
// ================================

// ETFAnalysisEngine ETF分析引擎
type ETFAnalysisEngine interface {
	Engine

	// 持仓分析
	CalculateOverlap(etf1, etf2 *models.Asset, holdings1, holdings2 []*models.Holding) (*ETFOverlap, error)
	CalculateSectorExposure(holdings []*models.Holding) (map[string]decimal.Decimal, error)
	CalculateCountryExposure(holdings []*models.Holding) (map[string]decimal.Decimal, error)
	CalculateFactorExposure(holdings []*models.Holding) (map[string]decimal.Decimal, error)

	// 成本分析
	CalculateTotalExpenseRatio(etf *models.Asset, holdings []*models.Holding) (decimal.Decimal, error)
	CalculateTrackingError(etf *models.Asset, benchmark *models.Asset, prices []*models.Price) (decimal.Decimal, error)
	CalculatePremiumDiscount(etf *models.Asset, nav, marketPrice decimal.Decimal) (decimal.Decimal, error)

	// 流动性分析
	CalculateLiquidityMetrics(etf *models.Asset, prices []*models.Price) (*LiquidityMetrics, error)
	CalculateBidAskSpread(etf *models.Asset, bid, ask decimal.Decimal) (decimal.Decimal, error)

	// 比较分析
	CompareETFs(etfs []*models.Asset, holdingsMap map[uint][]*models.Holding, pricesMap map[uint][]*models.Price) ([]ETFComparison, error)
}

// ETFOverlap ETF重叠度分析
type ETFOverlap struct {
	ETF1ID          uint            // ETF1 ID
	ETF2ID          uint            // ETF2 ID
	ETF1Symbol      string          // ETF1 代码
	ETF2Symbol      string          // ETF2 代码
	TotalOverlap    decimal.Decimal // 总重叠度（%）
	CommonHoldings  []CommonHolding // 共同持仓
	UniqueHoldings1 []UniqueHolding // ETF1独有持仓
	UniqueHoldings2 []UniqueHolding // ETF2独有持仓
	OverlapMetrics  OverlapMetrics  // 重叠度指标
}

// CommonHolding 共同持仓
type CommonHolding struct {
	AssetID      uint            // 资产ID
	Symbol       string          // 资产代码
	Name         string          // 资产名称
	WeightETF1   decimal.Decimal // ETF1中的权重
	WeightETF2   decimal.Decimal // ETF2中的权重
	WeightDiff   decimal.Decimal // 权重差异
	Contribution decimal.Decimal // 重叠度贡献
}

// UniqueHolding 独有持仓
type UniqueHolding struct {
	AssetID     uint            // 资产ID
	Symbol      string          // 资产代码
	Name        string          // 资产名称
	Weight      decimal.Decimal // 权重
	MarketValue decimal.Decimal // 市值
}

// OverlapMetrics 重叠度指标
type OverlapMetrics struct {
	WeightedOverlap   decimal.Decimal // 加权重叠度
	UnweightedOverlap decimal.Decimal // 非加权重叠度
	JaccardIndex      decimal.Decimal // 雅卡尔指数
	CosineSimilarity  decimal.Decimal // 余弦相似度
	SectorOverlap     decimal.Decimal // 行业重叠度
	CountryOverlap    decimal.Decimal // 国家重叠度
	FactorOverlap     decimal.Decimal // 因子重叠度
}

// LiquidityMetrics 流动性指标
type LiquidityMetrics struct {
	AvgVolume        int64           // 平均成交量
	AvgTurnover      decimal.Decimal // 平均成交额
	VolumeStdDev     decimal.Decimal // 成交量标准差
	TurnoverStdDev   decimal.Decimal // 成交额标准差
	VolumeToShares   decimal.Decimal // 成交量与流通股比
	BidAskSpread     decimal.Decimal // 买卖价差
	SpreadPercentage decimal.Decimal // 价差百分比
	ImpactCost       decimal.Decimal // 冲击成本
	MarketDepth      decimal.Decimal // 市场深度
}

// ETFComparison ETF对比结果
type ETFComparison struct {
	ETF1ID           uint            // ETF1 ID
	ETF2ID           uint            // ETF2 ID
	ETF1Symbol       string          // ETF1 代码
	ETF2Symbol       string          // ETF2 代码
	Correlation      decimal.Decimal // 相关性
	Beta             decimal.Decimal // Beta值
	TrackingError    decimal.Decimal // 跟踪误差
	ExpenseRatioDiff decimal.Decimal // 费率差异
	ReturnDiff1Y     decimal.Decimal // 1年收益差异
	ReturnDiff3Y     decimal.Decimal // 3年收益差异
	VolatilityDiff   decimal.Decimal // 波动率差异
	SharpeRatioDiff  decimal.Decimal // 夏普比率差异
	MaxDrawdownDiff  decimal.Decimal // 最大回撤差异
	SectorOverlap    decimal.Decimal // 行业重叠度
	CountryOverlap   decimal.Decimal // 国家重叠度
	FactorOverlap    decimal.Decimal // 因子重叠度
}

// ================================
// 因子分析引擎接口
// ================================

// FactorAnalysisEngine 因子分析引擎
type FactorAnalysisEngine interface {
	Engine

	// 因子暴露度计算
	CalculateFactorExposure(portfolio *models.Portfolio, positions []*models.PortfolioPosition) (map[string]decimal.Decimal, error)
	CalculateFactorReturns(factors []string, prices []*models.Price, factorDefinitions map[string]FactorDefinition) (map[string]decimal.Decimal, error)
	CalculateFactorAttribution(portfolio *models.Portfolio, positions []*models.PortfolioPosition, factorReturns map[string]decimal.Decimal) (*FactorAttribution, error)

	// 多因子模型
	RunFamaFrench3Factor(portfolio *models.Portfolio, positions []*models.PortfolioPosition, marketPrices []*models.Price) (*FamaFrenchResult, error)
	RunFamaFrench5Factor(portfolio *models.Portfolio, positions []*models.PortfolioPosition, marketPrices []*models.Price) (*FamaFrenchResult, error)
	RunCarhart4Factor(portfolio *models.Portfolio, positions []*models.PortfolioPosition, marketPrices []*models.Price) (*CarhartResult, error)

	// 风险因子分解
	DecomposeRisk(portfolio *models.Portfolio, positions []*models.PortfolioPosition, factors []string) (*RiskDecomposition, error)
	CalculateActiveRisk(portfolio *models.Portfolio, benchmark *models.Portfolio, positions []*models.PortfolioPosition) (*ActiveRiskAnalysis, error)
}

// FactorDefinition 因子定义
type FactorDefinition struct {
	Name        string                     // 因子名称
	Description string                     // 因子描述
	Type        string                     // 因子类型：market/size/value/profitability/investment/momentum
	Calculation string                     // 计算方法
	Weights     map[string]decimal.Decimal // 权重映射
}

// FactorAttribution 因子归因分析
type FactorAttribution struct {
	PortfolioID    uint                       // 组合ID
	PeriodStart    time.Time                  // 期间开始
	PeriodEnd      time.Time                  // 期间结束
	TotalReturn    decimal.Decimal            // 总收益率
	FactorReturns  map[string]decimal.Decimal // 因子收益率
	FactorExposure map[string]decimal.Decimal // 因子暴露度
	Attribution    map[string]decimal.Decimal // 因子归因
	Residual       decimal.Decimal            // 残差
	R2             decimal.Decimal            // R平方
	AdjustedR2     decimal.Decimal            // 调整后R平方
}

// FamaFrenchResult Fama-French模型结果
type FamaFrenchResult struct {
	Alpha         decimal.Decimal            // Alpha值
	MarketBeta    decimal.Decimal            // 市场因子Beta
	SMBBeta       decimal.Decimal            // 规模因子Beta
	HMLBeta       decimal.Decimal            // 价值因子Beta
	RMWBeta       decimal.Decimal            // 盈利能力因子Beta（5因子）
	CMABeta       decimal.Decimal            // 投资因子Beta（5因子）
	R2            decimal.Decimal            // R平方
	AdjustedR2    decimal.Decimal            // 调整后R平方
	StandardError decimal.Decimal            // 标准误差
	TStatistics   map[string]decimal.Decimal // T统计量
	PValues       map[string]decimal.Decimal // P值
}

// CarhartResult Carhart模型结果
type CarhartResult struct {
	FamaFrenchResult                 // 继承Fama-French结果
	MomentumBeta     decimal.Decimal // 动量因子Beta
}

// RiskDecomposition 风险分解
type RiskDecomposition struct {
	TotalRisk           decimal.Decimal            // 总风险
	FactorRisk          decimal.Decimal            // 因子风险
	SpecificRisk        decimal.Decimal            // 特质风险
	FactorRiskBreakdown map[string]decimal.Decimal // 因子风险分解
	MarginalRisk        map[string]decimal.Decimal // 边际风险
	RiskContribution    map[string]decimal.Decimal // 风险贡献度
}

// ActiveRiskAnalysis 主动风险分析
type ActiveRiskAnalysis struct {
	ActiveReturn         decimal.Decimal            // 主动收益率
	TrackingError        decimal.Decimal            // 跟踪误差
	InformationRatio     decimal.Decimal            // 信息比率
	ActiveShare          decimal.Decimal            // 主动份额
	FactorActiveExposure map[string]decimal.Decimal // 因子主动暴露度
	ActiveRiskBreakdown  *RiskDecomposition         // 主动风险分解
}

// ================================
// 优化引擎接口
// ================================

// PortfolioOptimizationEngine 组合优化引擎
type PortfolioOptimizationEngine interface {
	Engine
	ConfigurableEngine

	// 优化目标
	OptimizeForMaxSharpe(assets []*models.Asset, returns []decimal.Decimal, covMatrix [][]decimal.Decimal, constraints OptimizationConstraints) (*OptimizationResult, error)
	OptimizeForMinVariance(assets []*models.Asset, returns []decimal.Decimal, covMatrix [][]decimal.Decimal, constraints OptimizationConstraints) (*OptimizationResult, error)
	OptimizeForMaxReturn(assets []*models.Asset, returns []decimal.Decimal, covMatrix [][]decimal.Decimal, constraints OptimizationConstraints, targetRisk decimal.Decimal) (*OptimizationResult, error)
	OptimizeForRiskParity(assets []*models.Asset, covMatrix [][]decimal.Decimal, constraints OptimizationConstraints) (*OptimizationResult, error)

	// 有效前沿
	CalculateEfficientFrontier(assets []*models.Asset, returns []decimal.Decimal, covMatrix [][]decimal.Decimal, constraints OptimizationConstraints, numPoints int) ([]EfficientFrontierPoint, error)

	// Black-Litterman模型
	RunBlackLitterman(marketWeights []decimal.Decimal, returns []decimal.Decimal, covMatrix [][]decimal.Decimal, views []View, tau decimal.Decimal) (*BlackLittermanResult, error)
}

// OptimizationConstraints 优化约束
type OptimizationConstraints struct {
	MinWeights       []decimal.Decimal // 最小权重
	MaxWeights       []decimal.Decimal // 最大权重
	SumToOne         bool              // 权重和为1
	LongOnly         bool              // 仅多头
	GroupConstraints []GroupConstraint // 组约束
	TurnoverLimit    decimal.Decimal   // 换手率限制
	CardinalityLimit int               // 资产数量限制
}

// GroupConstraint 组约束
type GroupConstraint struct {
	GroupName    string          // 组名称
	AssetIndices []int           // 资产索引
	MinWeight    decimal.Decimal // 组最小权重
	MaxWeight    decimal.Decimal // 组最大权重
}

// OptimizationResult 优化结果
type OptimizationResult struct {
	Weights              []decimal.Decimal // 最优权重
	ExpectedReturn       decimal.Decimal   // 预期收益率
	ExpectedRisk         decimal.Decimal   // 预期风险
	SharpeRatio          decimal.Decimal   // 夏普比率
	Converged            bool              // 是否收敛
	Iterations           int               // 迭代次数
	ObjectiveValue       decimal.Decimal   // 目标函数值
	ConstraintsSatisfied bool              // 约束是否满足
}

// EfficientFrontierPoint 有效前沿点
type EfficientFrontierPoint struct {
	Return      decimal.Decimal   // 预期收益率
	Risk        decimal.Decimal   // 预期风险
	Weights     []decimal.Decimal // 权重
	SharpeRatio decimal.Decimal   // 夏普比率
}

// View Black-Litterman观点
type View struct {
	AssetIndices []int           // 相关资产索引
	ViewReturn   decimal.Decimal // 观点收益率
	Uncertainty  decimal.Decimal // 不确定性
	Type         string          // 观点类型：absolute/relative
}

// BlackLittermanResult Black-Litterman结果
type BlackLittermanResult struct {
	PriorWeights       []decimal.Decimal // 先验权重
	PosteriorWeights   []decimal.Decimal // 后验权重
	PriorReturns       []decimal.Decimal // 先验收益率
	PosteriorReturns   []decimal.Decimal // 后验收益率
	ViewImpliedReturns []decimal.Decimal // 观点隐含收益率
	Confidence         []decimal.Decimal // 置信度
}

package services

import (
	"fmt"
	"math"
	"sort"

	"etf-insight/config"
	"etf-insight/models"
	"etf-insight/utils"

	"github.com/shopspring/decimal"
)

type PortfolioOptimizer struct {
	analysisService *ETFAnalysisService
}

func NewPortfolioOptimizer(analysisService *ETFAnalysisService) *PortfolioOptimizer {
	return &PortfolioOptimizer{
		analysisService: analysisService,
	}
}

type OptimizationType string

const (
	OptimizationTypeMaxSharpe     OptimizationType = "max_sharpe"
	OptimizationTypeMinVolatility OptimizationType = "min_volatility"
	OptimizationTypeEqualWeight   OptimizationType = "equal_weight"
)

type PortfolioOptimizationRequest struct {
	Symbols          []string                `json:"symbols" binding:"required,min=2,max=20"`
	OptimizationType OptimizationType        `json:"optimization_type"`
	RiskAversion     decimal.Decimal         `json:"risk_aversion"`
	TargetReturn     decimal.Decimal         `json:"target_return"`
	RiskFreeRate     decimal.Decimal         `json:"risk_free_rate"`
	Constraints      OptimizationConstraints `json:"constraints"`
}

type OptimizationConstraints struct {
	MaxWeightPerAsset decimal.Decimal `json:"max_weight_per_asset"`
	MinWeightPerAsset decimal.Decimal `json:"min_weight_per_asset"`
	AllowShort        bool            `json:"allow_short"`
}

type PortfolioOptimizationResult struct {
	Weights            map[string]decimal.Decimal `json:"weights"`
	ExpectedReturn     decimal.Decimal            `json:"expected_return"`
	ExpectedVolatility decimal.Decimal            `json:"expected_volatility"`
	SharpeRatio        decimal.Decimal            `json:"sharpe_ratio"`
	OptimizationType   OptimizationType           `json:"optimization_type"`
	EfficientFrontier  []FrontierPoint            `json:"efficient_frontier,omitempty"`
	RiskFreeRate       decimal.Decimal            `json:"risk_free_rate"`
}

type FrontierPoint struct {
	ExpectedReturn decimal.Decimal            `json:"expected_return"`
	Volatility     decimal.Decimal            `json:"volatility"`
	Weights        map[string]decimal.Decimal `json:"weights"`
}

func (o *PortfolioOptimizer) Optimize(request PortfolioOptimizationRequest) (*PortfolioOptimizationResult, error) {
	if err := o.validateRequest(request); err != nil {
		return nil, err
	}

	defaultConstraints := OptimizationConstraints{
		MaxWeightPerAsset: decimal.NewFromFloat(0.4),
		MinWeightPerAsset: decimal.NewFromFloat(0.05),
		AllowShort:        false,
	}
	if request.Constraints.MaxWeightPerAsset.IsZero() {
		request.Constraints.MaxWeightPerAsset = defaultConstraints.MaxWeightPerAsset
	}
	if request.Constraints.MinWeightPerAsset.IsZero() {
		request.Constraints.MinWeightPerAsset = defaultConstraints.MinWeightPerAsset
	}

	if request.RiskAversion.IsZero() {
		request.RiskAversion = decimal.NewFromFloat(2.0)
	}
	if request.RiskFreeRate.IsZero() {
		request.RiskFreeRate = decimal.NewFromFloat(config.GetFinancialConfig().RiskFreeRate)
	}

	symbolReturns, err := o.getHistoricalReturns(request.Symbols)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical returns: %w", err)
	}

	meanReturns := o.calculateMeanReturns(symbolReturns)
	covMatrix := o.calculateCovarianceMatrix(symbolReturns)

	switch request.OptimizationType {
	case OptimizationTypeMaxSharpe:
		return o.maximizeSharpeRatio(request, meanReturns, covMatrix)
	case OptimizationTypeMinVolatility:
		return o.minimizeVolatility(request, meanReturns, covMatrix)
	case OptimizationTypeEqualWeight:
		return o.equalWeightAllocation(request, meanReturns, covMatrix)
	default:
		return o.maximizeSharpeRatio(request, meanReturns, covMatrix)
	}
}

func (o *PortfolioOptimizer) validateRequest(request PortfolioOptimizationRequest) error {
	if len(request.Symbols) < 2 {
		return fmt.Errorf("at least 2 symbols required for portfolio optimization")
	}
	if len(request.Symbols) > 20 {
		return fmt.Errorf("maximum 20 symbols allowed for portfolio optimization")
	}
	if request.OptimizationType == "" {
		request.OptimizationType = OptimizationTypeMaxSharpe
	}
	return nil
}

func (o *PortfolioOptimizer) getHistoricalReturns(symbols []string) (map[string][]decimal.Decimal, error) {
	returns := make(map[string][]decimal.Decimal)

	if models.DB == nil {
		return nil, ErrDatabaseNotInitialized
	}

	for _, symbol := range symbols {
		var prices []models.ETFData
		if err := models.DB.Where("symbol = ?", symbol).Order("date DESC").Limit(252).Find(&prices).Error; err != nil {
			continue
		}

		if len(prices) < 30 {
			defaultYield := getDividendYieldByCategory(symbol, "")
			dailyReturn := defaultYield.Div(decimal.NewFromInt(252))
			dailyReturns := make([]decimal.Decimal, 252)
			for i := range dailyReturns {
				dailyReturns[i] = dailyReturn
			}
			returns[symbol] = dailyReturns
			continue
		}

		var dailyReturns []decimal.Decimal
		for i := 1; i < len(prices); i++ {
			prevClose := prices[i-1].ClosePrice
			if prevClose.IsPositive() {
				dailyReturn := prices[i].ClosePrice.Sub(prevClose).Div(prevClose)
				dailyReturns = append(dailyReturns, dailyReturn)
			}
		}
		returns[symbol] = dailyReturns
	}

	return returns, nil
}

func (o *PortfolioOptimizer) calculateMeanReturns(returns map[string][]decimal.Decimal) map[string]decimal.Decimal {
	meanReturns := make(map[string]decimal.Decimal)

	for symbol, dailyReturns := range returns {
		if len(dailyReturns) == 0 {
			meanReturns[symbol] = decimal.Zero
			continue
		}

		avgDaily := utils.CalculateMean(dailyReturns)
		annualized := avgDaily.Mul(decimal.NewFromInt(252))
		meanReturns[symbol] = annualized
	}

	return meanReturns
}

func (o *PortfolioOptimizer) calculateCovarianceMatrix(returns map[string][]decimal.Decimal) map[string]map[string]decimal.Decimal {
	symbols := make([]string, 0, len(returns))
	for symbol := range returns {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	covMatrix := make(map[string]map[string]decimal.Decimal)
	for _, s1 := range symbols {
		covMatrix[s1] = make(map[string]decimal.Decimal)
		for _, s2 := range symbols {
			covMatrix[s1][s2] = o.calculateCovariance(returns[s1], returns[s2])
		}
	}

	return covMatrix
}

func (o *PortfolioOptimizer) calculateCovariance(returns1, returns2 []decimal.Decimal) decimal.Decimal {
	cov := utils.CalculateCovariance(returns1, returns2)
	return cov.Mul(decimal.NewFromInt(252))
}

func (o *PortfolioOptimizer) maximizeSharpeRatio(request PortfolioOptimizationRequest, meanReturns map[string]decimal.Decimal, covMatrix map[string]map[string]decimal.Decimal) (*PortfolioOptimizationResult, error) {
	if len(request.Symbols) == 0 {
		return nil, fmt.Errorf("cannot optimize empty portfolio: no symbols provided")
	}

	riskFreeRate := request.RiskFreeRate
	weights := o.gradientDescentOptimization(meanReturns, covMatrix, riskFreeRate, request.Constraints)

	if len(weights) == 0 {
		return nil, fmt.Errorf("optimization failed: no weights returned")
	}

	weightsDecimal := make(map[string]decimal.Decimal)
	for i, symbol := range request.Symbols {
		weightsDecimal[symbol] = weights[i]
	}

	expectedReturn := o.calculatePortfolioReturn(weightsDecimal, meanReturns)
	expectedVolatility := o.calculatePortfolioVolatility(weightsDecimal, covMatrix)
	sharpeRatio := o.calculateSharpeRatio(expectedReturn, expectedVolatility, riskFreeRate)

	return &PortfolioOptimizationResult{
		Weights:            weightsDecimal,
		ExpectedReturn:     expectedReturn,
		ExpectedVolatility: expectedVolatility,
		SharpeRatio:        sharpeRatio,
		OptimizationType:   OptimizationTypeMaxSharpe,
		RiskFreeRate:       riskFreeRate,
	}, nil
}

func (o *PortfolioOptimizer) minimizeVolatility(request PortfolioOptimizationRequest, meanReturns map[string]decimal.Decimal, covMatrix map[string]map[string]decimal.Decimal) (*PortfolioOptimizationResult, error) {
	n := len(request.Symbols)
	weights := make([]decimal.Decimal, n)
	stepSize := decimal.NewFromFloat(0.001)
	tolerance := decimal.NewFromFloat(0.0001)
	maxIterations := 1000

	for i := range weights {
		weights[i] = decimal.NewFromFloat(1.0 / float64(n))
	}

	for range maxIterations {
		gradients := o.calculateVolatilityGradients(weights, covMatrix)
		norm := decimal.Zero
		for _, g := range gradients {
			norm = norm.Add(g.Mul(g))
		}
		norm = decimal.NewFromFloat(math.Sqrt(norm.InexactFloat64()))
		if norm.LessThan(tolerance) {
			break
		}

		normalizedGradients := make([]decimal.Decimal, n)
		for i, g := range gradients {
			normalizedGradients[i] = g.Div(norm)
		}

		for i := range weights {
			weights[i] = weights[i].Sub(stepSize.Mul(normalizedGradients[i]))
		}

		weights = o.applyConstraints(weights, request.Constraints)
	}

	weightsDecimal := make(map[string]decimal.Decimal)
	for i, symbol := range request.Symbols {
		weightsDecimal[symbol] = weights[i]
	}

	expectedReturn := o.calculatePortfolioReturn(weightsDecimal, meanReturns)
	expectedVolatility := o.calculatePortfolioVolatility(weightsDecimal, covMatrix)
	sharpeRatio := o.calculateSharpeRatio(expectedReturn, expectedVolatility, request.RiskFreeRate)

	return &PortfolioOptimizationResult{
		Weights:            weightsDecimal,
		ExpectedReturn:     expectedReturn,
		ExpectedVolatility: expectedVolatility,
		SharpeRatio:        sharpeRatio,
		OptimizationType:   OptimizationTypeMinVolatility,
		RiskFreeRate:       request.RiskFreeRate,
	}, nil
}

func (o *PortfolioOptimizer) equalWeightAllocation(request PortfolioOptimizationRequest, meanReturns map[string]decimal.Decimal, covMatrix map[string]map[string]decimal.Decimal) (*PortfolioOptimizationResult, error) {
	n := len(request.Symbols)
	weights := decimal.NewFromFloat(1.0 / float64(n))

	weightsDecimal := make(map[string]decimal.Decimal)
	for _, symbol := range request.Symbols {
		weightsDecimal[symbol] = weights
	}

	expectedReturn := o.calculatePortfolioReturn(weightsDecimal, meanReturns)
	expectedVolatility := o.calculatePortfolioVolatility(weightsDecimal, covMatrix)
	sharpeRatio := o.calculateSharpeRatio(expectedReturn, expectedVolatility, request.RiskFreeRate)

	return &PortfolioOptimizationResult{
		Weights:            weightsDecimal,
		ExpectedReturn:     expectedReturn,
		ExpectedVolatility: expectedVolatility,
		SharpeRatio:        sharpeRatio,
		OptimizationType:   OptimizationTypeEqualWeight,
		RiskFreeRate:       request.RiskFreeRate,
	}, nil
}

func (o *PortfolioOptimizer) gradientDescentOptimization(meanReturns map[string]decimal.Decimal, covMatrix map[string]map[string]decimal.Decimal, riskFreeRate decimal.Decimal, constraints OptimizationConstraints) []decimal.Decimal {
	n := len(meanReturns)
	weights := make([]decimal.Decimal, n)
	for i := range weights {
		weights[i] = decimal.NewFromFloat(1.0 / float64(n))
	}

	stepSize := decimal.NewFromFloat(0.01)
	tolerance := decimal.NewFromFloat(0.0001)
	maxIterations := 500

	for range maxIterations {
		negSharpeGradients := o.calculateNegativeSharpeGradients(weights, meanReturns, covMatrix, riskFreeRate)

		norm := decimal.Zero
		for _, g := range negSharpeGradients {
			norm = norm.Add(g.Mul(g))
		}
		norm = decimal.NewFromFloat(math.Sqrt(norm.InexactFloat64()))

		if norm.LessThan(tolerance) {
			break
		}

		normalizedGradients := make([]decimal.Decimal, n)
		for i, g := range negSharpeGradients {
			normalizedGradients[i] = g.Div(norm)
		}

		for i := range weights {
			weights[i] = weights[i].Sub(stepSize.Mul(normalizedGradients[i]))
		}

		weights = o.applyConstraints(weights, constraints)
	}

	return weights
}

func (o *PortfolioOptimizer) calculateNegativeSharpeGradients(weights []decimal.Decimal, meanReturns map[string]decimal.Decimal, covMatrix map[string]map[string]decimal.Decimal, riskFreeRate decimal.Decimal) []decimal.Decimal {
	n := len(weights)
	symbols := make([]string, 0, n)
	for symbol := range meanReturns {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	weightsMap := make(map[string]decimal.Decimal)
	for i, symbol := range symbols {
		weightsMap[symbol] = weights[i]
	}

	expectedReturn := o.calculatePortfolioReturn(weightsMap, meanReturns)
	volatility := o.calculatePortfolioVolatility(weightsMap, covMatrix)
	sharpe := o.calculateSharpeRatio(expectedReturn, volatility, riskFreeRate)

	if volatility.IsZero() || sharpe.IsZero() {
		return make([]decimal.Decimal, n)
	}

	gradients := make([]decimal.Decimal, n)
	for i, symbol := range symbols {
		returnGrad := meanReturns[symbol]

		covSum := decimal.Zero
		for j, otherSymbol := range symbols {
			cov := covMatrix[symbol][otherSymbol]
			covSum = covSum.Add(cov.Mul(weights[j]))
		}

		gradReturn := returnGrad.Sub(expectedReturn).Div(volatility)
		gradVol := covSum.Div(volatility)

		gradients[i] = gradReturn.Sub(gradVol.Mul(sharpe)).Neg()
	}

	return gradients
}

func (o *PortfolioOptimizer) calculateVolatilityGradients(weights []decimal.Decimal, covMatrix map[string]map[string]decimal.Decimal) []decimal.Decimal {
	n := len(weights)
	symbols := make([]string, 0, n)
	for symbol := range covMatrix {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	volatility := o.calculatePortfolioVolatilityFromArray(weights, covMatrix, symbols)
	if volatility.IsZero() {
		return make([]decimal.Decimal, n)
	}

	gradients := make([]decimal.Decimal, n)
	for i := range weights {
		covSum := decimal.Zero
		for j := range weights {
			covSum = covSum.Add(covMatrix[symbols[i]][symbols[j]].Mul(weights[j]))
		}
		gradients[i] = covSum.Div(volatility)
	}

	return gradients
}

func (o *PortfolioOptimizer) calculatePortfolioReturn(weights map[string]decimal.Decimal, meanReturns map[string]decimal.Decimal) decimal.Decimal {
	if len(weights) == 0 {
		return decimal.Zero
	}

	totalReturn := decimal.Zero
	for symbol, weight := range weights {
		if ret, ok := meanReturns[symbol]; ok {
			totalReturn = totalReturn.Add(ret.Mul(weight))
		}
	}
	return totalReturn
}

func (o *PortfolioOptimizer) calculatePortfolioVolatility(weights map[string]decimal.Decimal, covMatrix map[string]map[string]decimal.Decimal) decimal.Decimal {
	if len(weights) == 0 {
		return decimal.Zero
	}
	symbols := make([]string, 0, len(weights))
	for symbol := range weights {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	variance := decimal.Zero
	for _, s1 := range symbols {
		for _, s2 := range symbols {
			w1 := weights[s1]
			w2 := weights[s2]
			cov := covMatrix[s1][s2]
			variance = variance.Add(w1.Mul(w2).Mul(cov))
		}
	}

	if variance.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}

	return decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64()))
}

func (o *PortfolioOptimizer) calculatePortfolioVolatilityFromArray(weights []decimal.Decimal, covMatrix map[string]map[string]decimal.Decimal, symbols []string) decimal.Decimal {
	variance := decimal.Zero
	for i, s1 := range symbols {
		for j, s2 := range symbols {
			w1 := weights[i]
			w2 := weights[j]
			cov := covMatrix[s1][s2]
			variance = variance.Add(w1.Mul(w2).Mul(cov))
		}
	}

	if variance.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}

	return decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64()))
}

func (o *PortfolioOptimizer) calculateSharpeRatio(expectedReturn, volatility, riskFreeRate decimal.Decimal) decimal.Decimal {
	if volatility.IsZero() || volatility.IsNegative() {
		return decimal.Zero
	}
	excessReturn := expectedReturn.Sub(riskFreeRate)
	return excessReturn.Div(volatility)
}

func (o *PortfolioOptimizer) applyConstraints(weights []decimal.Decimal, constraints OptimizationConstraints) []decimal.Decimal {
	// 应用权重约束
	for i, w := range weights {
		if w.GreaterThan(constraints.MaxWeightPerAsset) {
			weights[i] = constraints.MaxWeightPerAsset
		} else if w.LessThan(constraints.MinWeightPerAsset) {
			weights[i] = constraints.MinWeightPerAsset
		} else if !constraints.AllowShort && w.LessThan(decimal.Zero) {
			weights[i] = decimal.Zero
		}
	}

	return utils.NormalizeWeights(weights)
}

func (o *PortfolioOptimizer) GetEfficientFrontier(request PortfolioOptimizationRequest) ([]FrontierPoint, error) {
	if err := o.validateRequest(request); err != nil {
		return nil, err
	}

	if request.RiskFreeRate.IsZero() {
		request.RiskFreeRate = decimal.NewFromFloat(config.GetFinancialConfig().RiskFreeRate)
	}

	defaultConstraints := OptimizationConstraints{
		MaxWeightPerAsset: decimal.NewFromFloat(0.4),
		MinWeightPerAsset: decimal.NewFromFloat(0.05),
		AllowShort:        false,
	}
	if request.Constraints.MaxWeightPerAsset.IsZero() {
		request.Constraints.MaxWeightPerAsset = defaultConstraints.MaxWeightPerAsset
	}
	if request.Constraints.MinWeightPerAsset.IsZero() {
		request.Constraints.MinWeightPerAsset = defaultConstraints.MinWeightPerAsset
	}

	symbolReturns, err := o.getHistoricalReturns(request.Symbols)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical returns: %w", err)
	}

	meanReturns := o.calculateMeanReturns(symbolReturns)
	covMatrix := o.calculateCovarianceMatrix(symbolReturns)

	return o.GenerateEfficientFrontier(request, meanReturns, covMatrix)
}

func (o *PortfolioOptimizer) GenerateEfficientFrontier(request PortfolioOptimizationRequest, meanReturns map[string]decimal.Decimal, covMatrix map[string]map[string]decimal.Decimal) ([]FrontierPoint, error) {
	minReturn := decimal.Zero
	maxReturn := decimal.Zero
	for _, ret := range meanReturns {
		if ret.LessThan(minReturn) || minReturn.IsZero() {
			minReturn = ret
		}
		if ret.GreaterThan(maxReturn) {
			maxReturn = ret
		}
	}

	numPoints := 20
	step := maxReturn.Sub(minReturn).Div(decimal.NewFromInt(int64(numPoints)))

	frontier := make([]FrontierPoint, 0, numPoints)
	for i := 0; i <= numPoints; i++ {
		targetReturn := minReturn.Add(step.Mul(decimal.NewFromInt(int64(i))))

		weights := o.minimizeVolatilityForTarget(request, targetReturn, meanReturns, covMatrix)
		if weights == nil {
			continue
		}

		weightsMap := make(map[string]decimal.Decimal)
		for j, symbol := range request.Symbols {
			weightsMap[symbol] = weights[j]
		}

		volatility := o.calculatePortfolioVolatility(weightsMap, covMatrix)

		frontier = append(frontier, FrontierPoint{
			ExpectedReturn: targetReturn,
			Volatility:     volatility,
			Weights:        weightsMap,
		})
	}

	return frontier, nil
}

func (o *PortfolioOptimizer) minimizeVolatilityForTarget(request PortfolioOptimizationRequest, targetReturn decimal.Decimal, meanReturns map[string]decimal.Decimal, covMatrix map[string]map[string]decimal.Decimal) []decimal.Decimal {
	n := len(request.Symbols)
	weights := make([]decimal.Decimal, n)
	for i := range weights {
		weights[i] = decimal.NewFromFloat(1.0 / float64(n))
	}

	stepSize := decimal.NewFromFloat(0.005)
	tolerance := decimal.NewFromFloat(0.0001)
	maxIterations := 300

	symbols := make([]string, 0, n)
	for symbol := range meanReturns {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	for range maxIterations {
		weightsMap := make(map[string]decimal.Decimal)
		for i, symbol := range symbols {
			weightsMap[symbol] = weights[i]
		}

		currentReturn := o.calculatePortfolioReturn(weightsMap, meanReturns)
		returnDiff := currentReturn.Sub(targetReturn)

		gradients := o.calculateVolatilityGradients(weights, covMatrix)

		returnPenalty := decimal.NewFromFloat(100.0)
		returnGradients := make([]decimal.Decimal, n)
		for i, symbol := range symbols {
			returnGradients[i] = meanReturns[symbol].Mul(returnPenalty)
		}

		for i := range weights {
			gradients[i] = gradients[i].Add(returnGradients[i].Mul(returnDiff))
		}

		norm := decimal.Zero
		for _, g := range gradients {
			norm = norm.Add(g.Mul(g))
		}
		norm = decimal.NewFromFloat(math.Sqrt(norm.InexactFloat64()))

		if norm.LessThan(tolerance) {
			break
		}

		normalizedGradients := make([]decimal.Decimal, n)
		for i, g := range gradients {
			normalizedGradients[i] = g.Div(norm)
		}

		for i := range weights {
			weights[i] = weights[i].Sub(stepSize.Mul(normalizedGradients[i]))
		}

		weights = o.applyConstraints(weights, request.Constraints)
	}

	return weights
}

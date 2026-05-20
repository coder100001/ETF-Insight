package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"etf-insight/models"
	"etf-insight/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var (
	ErrInsufficientReturns = errors.New("insufficient returns data")
)

type SimulationError struct {
	Step   string
	Reason string
	Input  interface{}
}

func (e *SimulationError) Error() string {
	return fmt.Sprintf("simulation error at step '%s': %s (input: %v)", e.Step, e.Reason, e.Input)
}

type SimulationWarnings struct {
	InvalidPaths   int
	ClampedReturns int
	ZeroVolatility bool
}

// generateNormalRandom generates a standard normal random variate using Box-Muller transform
// with retry mechanism to handle edge cases (log(0), overflow, etc.)
func generateNormalRandom() (float64, error) {
	maxAttempts := 3
	for attempt := 0; attempt < maxAttempts; attempt++ {
		u1 := rand.Float64()
		u2 := rand.Float64()

		if u1 <= 1e-10 || u1 >= 1.0 {
			continue
		}

		logU1 := math.Log(u1)
		if math.IsInf(logU1, 0) {
			continue
		}

		z := math.Sqrt(-2*logU1) * math.Cos(2*math.Pi*u2)

		if math.IsNaN(z) || math.IsInf(z, 0) || math.Abs(z) > 6 {
			continue
		}

		return z, nil
	}

	return 0, errors.New("failed to generate valid normal random number after multiple attempts")
}

func decimalPow3(d decimal.Decimal) decimal.Decimal {
	return d.Mul(d).Mul(d)
}

type RiskBudgetService struct {
	db *gorm.DB
}

func NewRiskBudgetService(db *gorm.DB) *RiskBudgetService {
	return &RiskBudgetService{db: db}
}

func (s *RiskBudgetService) CreateConfig(config *models.RiskBudgetConfig) error {
	if err := s.validateConfig(config); err != nil {
		return err
	}

	return s.db.Create(config).Error
}

func (s *RiskBudgetService) GetConfig(id uint) (*models.RiskBudgetConfig, error) {
	var config models.RiskBudgetConfig
	err := s.db.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *RiskBudgetService) UpdateConfig(config *models.RiskBudgetConfig) error {
	return s.db.Save(config).Error
}

func (s *RiskBudgetService) validateConfig(config *models.RiskBudgetConfig) error {
	if config.PortfolioID == 0 {
		return errors.New("portfolio_id is required")
	}

	if config.CVaRConfidence.IsZero() {
		config.CVaRConfidence = decimal.NewFromFloat(0.95)
	}
	if config.CVaRConfidence.LessThan(decimal.NewFromFloat(0.8)) ||
		config.CVaRConfidence.GreaterThan(decimal.NewFromFloat(0.999)) {
		return errors.New("CVaR confidence level should be between 0.80 and 0.999")
	}

	if config.VaRConfidence.IsZero() {
		config.VaRConfidence = decimal.NewFromFloat(0.95)
	}
	if config.VaRConfidence.LessThan(decimal.NewFromFloat(0.8)) ||
		config.VaRConfidence.GreaterThan(decimal.NewFromFloat(0.999)) {
		return errors.New("VaR confidence level should be between 0.80 and 0.999")
	}

	if config.StockCVaRBudget.LessThan(decimal.Zero) ||
		config.StockCVaRBudget.GreaterThan(decimal.NewFromInt(1)) {
		return errors.New("stock CVaR budget must be between 0 and 1")
	}

	if config.BondCVaRBudget.LessThan(decimal.Zero) ||
		config.BondCVaRBudget.GreaterThan(decimal.NewFromInt(1)) {
		return errors.New("bond CVaR budget must be between 0 and 1")
	}

	if config.CommodityCVaRBudget.LessThan(decimal.Zero) ||
		config.CommodityCVaRBudget.GreaterThan(decimal.NewFromInt(1)) {
		return errors.New("commodity CVaR budget must be between 0 and 1")
	}

	if config.CashCVaRBudget.LessThan(decimal.Zero) ||
		config.CashCVaRBudget.GreaterThan(decimal.NewFromInt(1)) {
		return errors.New("cash CVaR budget must be between 0 and 1")
	}

	totalBudget := config.StockCVaRBudget.Add(config.BondCVaRBudget).
		Add(config.CommodityCVaRBudget).Add(config.CashCVaRBudget)
	if totalBudget.GreaterThan(decimal.NewFromFloat(1.1)) {
		return errors.New("total CVaR budget should not exceed 110%")
	}

	if !config.MaxDrawdown.IsZero() {
		if config.MaxDrawdown.GreaterThan(decimal.Zero) {
			return errors.New("max drawdown must be negative or zero")
		}
		if config.MaxDrawdown.LessThan(decimal.NewFromInt(-1)) {
			return errors.New("max drawdown must be greater than -100%")
		}
	}

	return nil
}

func (s *RiskBudgetService) CalculateCVaR(
	returns []decimal.Decimal,
	confidenceLevel decimal.Decimal,
	useParametric bool,
) (decimal.Decimal, decimal.Decimal, error) {
	if len(returns) < 10 {
		return decimal.Zero, decimal.Zero, ErrInsufficientReturns
	}

	if useParametric {
		return s.calculateParametricCVaR(returns, confidenceLevel)
	}
	return s.calculateHistoricalCVaR(returns, confidenceLevel)
}

func (s *RiskBudgetService) calculateHistoricalCVaR(
	returns []decimal.Decimal,
	confidenceLevel decimal.Decimal,
) (decimal.Decimal, decimal.Decimal, error) {
	sortedReturns := make([]decimal.Decimal, len(returns))
	copy(sortedReturns, returns)
	utils.SortDecimals(sortedReturns)

	index := int(decimal.NewFromInt(int64(len(sortedReturns))).Mul(
		decimal.NewFromInt(1).Sub(confidenceLevel),
	).IntPart())

	mean := utils.CalculateMean(sortedReturns)
	stdDev := utils.CalculatePopulationStdDev(sortedReturns, mean)

	zScore := decimal.NewFromFloat(-1.645)
	if confidenceLevel.Equal(decimal.NewFromFloat(0.99)) {
		zScore = decimal.NewFromFloat(-2.33)
	}

	varVaR := mean.Add(zScore.Mul(stdDev))

	var cvarSum decimal.Decimal
	count := 0
	for i := range index {
		cvarSum = cvarSum.Add(sortedReturns[i])
		count++
	}

	varCVaR := cvarSum.Div(decimal.NewFromInt(int64(count)))

	return varVaR, varCVaR, nil
}

func (s *RiskBudgetService) calculateParametricCVaR(
	returns []decimal.Decimal,
	confidenceLevel decimal.Decimal,
) (decimal.Decimal, decimal.Decimal, error) {
	mean := utils.CalculateMean(returns)
	stdDev := utils.CalculatePopulationStdDev(returns, mean)

	zScore := decimal.NewFromFloat(-1.645)
	if confidenceLevel.Equal(decimal.NewFromFloat(0.99)) {
		zScore = decimal.NewFromFloat(-2.33)
	}

	varVaR := mean.Add(zScore.Mul(stdDev))

	zScoreCVaR := decimal.NewFromFloat(-2.06)
	if confidenceLevel.Equal(decimal.NewFromFloat(0.99)) {
		zScoreCVaR = decimal.NewFromFloat(-2.67)
	}

	varCVaR := mean.Add(zScoreCVaR.Mul(stdDev))

	return varVaR, varCVaR, nil
}

func (s *RiskBudgetService) calculateMonteCarloCVaR(
	returns []decimal.Decimal,
	confidenceLevel decimal.Decimal,
) (decimal.Decimal, decimal.Decimal, error) {
	mean := utils.CalculateMean(returns)
	stdDev := utils.CalculatePopulationStdDev(returns, mean)

	if stdDev.IsZero() || stdDev.LessThan(decimal.NewFromFloat(1e-6)) {
		return decimal.Zero, decimal.Zero, errors.New("insufficient volatility for Monte Carlo simulation")
	}

	numSimulations := 10000
	simulatedReturns := make([]decimal.Decimal, numSimulations)

	for i := range numSimulations {
		z, err := generateNormalRandom()
		if err != nil {
			z = 0
		}
		simulatedReturn := mean.InexactFloat64() + stdDev.InexactFloat64()*z
		simulatedReturns[i] = decimal.NewFromFloat(simulatedReturn)
	}

	utils.SortDecimals(simulatedReturns)

	index := int(decimal.NewFromInt(int64(numSimulations)).Mul(
		decimal.NewFromInt(1).Sub(confidenceLevel),
	).IntPart())

	if index < 0 || index >= len(simulatedReturns) {
		index = 0
	}

	varVaR := simulatedReturns[index]

	if index <= 0 {
		return varVaR, decimal.Zero, nil
	}

	var cvarSum decimal.Decimal
	for i := range index {
		cvarSum = cvarSum.Add(simulatedReturns[i])
	}
	varCVaR := cvarSum.Div(decimal.NewFromInt(int64(index)))

	return varVaR, varCVaR, nil
}

func (s *RiskBudgetService) CalculateMonteCarloCVaR(
	returns []decimal.Decimal,
	confidenceLevel decimal.Decimal,
) (decimal.Decimal, decimal.Decimal, error) {
	if len(returns) < 10 {
		return decimal.Zero, decimal.Zero, ErrInsufficientReturns
	}
	return s.calculateMonteCarloCVaR(returns, confidenceLevel)
}

func (s *RiskBudgetService) CalculateRiskContributions(
	weights []decimal.Decimal,
	returnsMatrix [][]decimal.Decimal,
	confidenceLevel decimal.Decimal,
) ([]models.RiskContribution, error) {
	if len(weights) == 0 || len(returnsMatrix) == 0 {
		return nil, ErrInsufficientReturns
	}

	n := len(weights)
	contributions := make([]models.RiskContribution, n)

	portfolioReturns := make([]decimal.Decimal, len(returnsMatrix[0]))
	for i := range portfolioReturns {
		portfolioReturns[i] = decimal.Zero
		for j := range n {
			portfolioReturns[i] = portfolioReturns[i].Add(
				weights[j].Mul(returnsMatrix[j][i]),
			)
		}
	}

	_, portfolioCVaR, err := s.CalculateCVaR(
		portfolioReturns,
		confidenceLevel,
		false,
	)
	if err != nil {
		return nil, err
	}

	for i := range n {
		delta := decimal.NewFromFloat(0.01)
		newWeights := make([]decimal.Decimal, n)
		copy(newWeights, weights)
		newWeights[i] = newWeights[i].Add(delta)

		total := decimal.Zero
		for _, w := range newWeights {
			total = total.Add(w)
		}
		for j := range newWeights {
			newWeights[j] = newWeights[j].Div(total)
		}

		newPortfolioReturns := make([]decimal.Decimal, len(returnsMatrix[0]))
		for k := range newPortfolioReturns {
			newPortfolioReturns[k] = decimal.Zero
			for j := range n {
				newPortfolioReturns[k] = newPortfolioReturns[k].Add(
					newWeights[j].Mul(returnsMatrix[j][k]),
				)
			}
		}

		_, newCVaR, err := s.CalculateCVaR(
			newPortfolioReturns,
			confidenceLevel,
			false,
		)
		if err != nil {
			return nil, err
		}

		marginalRisk := newCVaR.Sub(portfolioCVaR).Div(delta)
		riskContribution := weights[i].Mul(marginalRisk)
		percentageContribution := riskContribution.Div(portfolioCVaR.Abs()).Mul(decimal.NewFromInt(100))

		contributions[i] = models.RiskContribution{
			Weight:           weights[i],
			MarginalCVaR:     marginalRisk,
			CVaRContribution: riskContribution,
			CVaRPercentage:   percentageContribution,
			CalculationDate:  time.Now(),
			CreatedAt:        time.Now(),
		}
	}

	return contributions, nil
}

func (s *RiskBudgetService) RunMonteCarloSimulation(
	configID uint,
	numSimulations int,
	timeSteps int,
	returns []decimal.Decimal,
) (*models.MonteCarloSimulation, error) {
	if err := s.validateSimulationInputs(numSimulations, timeSteps, returns); err != nil {
		return nil, err
	}

	// 过滤并验证收益率数据
	validReturns, invalidCount := s.filterValidReturns(returns)
	if len(validReturns) < 20 {
		return nil, &SimulationError{
			Step:   "input_validation",
			Reason: fmt.Sprintf("insufficient valid returns: need at least 20, got %d (filtered %d invalid)", len(validReturns), invalidCount),
			Input:  len(returns),
		}
	}

	// 计算统计参数，处理边界情况
	mean, stdDev, err := s.calculateSimulationStats(validReturns)
	if err != nil {
		return nil, err
	}

	warnings := &SimulationWarnings{}
	if stdDev.IsZero() || stdDev.LessThan(decimal.NewFromFloat(1e-6)) {
		stdDev = decimal.NewFromFloat(1e-6)
		warnings.ZeroVolatility = true
	}

	dt := decimal.NewFromFloat(1.0 / 252.0)
	drift := mean.Mul(dt)
	diffusion := stdDev.Mul(decimalSqrt(dt))

	finalReturns := make([]decimal.Decimal, numSimulations)

	for i := range numSimulations {
		price := decimal.NewFromInt(100)
		validPath := true

		for step := 0; step < timeSteps; step++ {
			z, err := generateNormalRandom()
			if err != nil {
				z = 0
				validPath = false
			}

			noise := diffusion.Mul(decimal.NewFromFloat(z))
			return_ := drift.Add(noise)

			if return_.LessThan(decimal.NewFromFloat(-0.9)) {
				return_ = decimal.NewFromFloat(-0.9)
				validPath = false
				warnings.ClampedReturns++
			}

			price = price.Mul(decimal.NewFromInt(1).Add(return_))

			if price.LessThanOrEqual(decimal.Zero) {
				price = decimal.NewFromFloat(0.01)
				validPath = false
			}
		}

		finalReturns[i] = price.Sub(decimal.NewFromInt(100)).Div(decimal.NewFromInt(100))

		if !validPath {
			warnings.InvalidPaths++
		}
	}

	utils.SortDecimals(finalReturns)

	meanReturn := utils.CalculateMean(finalReturns)
	stdReturn := utils.CalculatePopulationStdDev(finalReturns, meanReturn)

	if math.IsNaN(meanReturn.InexactFloat64()) || math.IsInf(meanReturn.InexactFloat64(), 0) {
		return nil, errors.New("simulation produced invalid mean return")
	}
	if math.IsNaN(stdReturn.InexactFloat64()) || math.IsInf(stdReturn.InexactFloat64(), 0) {
		return nil, errors.New("simulation produced invalid standard deviation")
	}
	if meanReturn.LessThan(decimal.NewFromFloat(-2.0)) || meanReturn.GreaterThan(decimal.NewFromFloat(2.0)) {
		return nil, fmt.Errorf("simulation produced unrealistic mean return: %.2f%%",
			meanReturn.Mul(decimal.NewFromInt(100)).InexactFloat64())
	}

	percentile5Index := int(float64(numSimulations) * 0.05)
	percentile95Index := int(float64(numSimulations) * 0.95)

	if percentile5Index < 0 {
		percentile5Index = 0
	}
	if percentile5Index >= len(finalReturns) {
		percentile5Index = len(finalReturns) - 1
	}
	if percentile95Index < 0 {
		percentile95Index = 0
	}
	if percentile95Index >= len(finalReturns) {
		percentile95Index = len(finalReturns) - 1
	}

	percentile5 := finalReturns[percentile5Index]
	percentile95 := finalReturns[percentile95Index]

	if percentile5.GreaterThan(percentile95) {
		percentile5, percentile95 = percentile95, percentile5
	}

	simulationDataJSON, _ := json.Marshal(finalReturns)

	warningsJSON, _ := json.Marshal(warnings)

	simulation := &models.MonteCarloSimulation{
		PortfolioID:      configID,
		SimulationDate:   time.Now(),
		NumPaths:         numSimulations,
		TimeSteps:        timeSteps,
		ConfidenceLevel:  decimal.NewFromFloat(0.95),
		MeanReturn:       meanReturn,
		StdDev:           stdReturn,
		VaR95:            percentile5,
		CVaR95:           percentile95,
		SimulationResult: string(simulationDataJSON),
		CreatedAt:        time.Now(),
	}

	simulation.Warnings = string(warningsJSON)

	if err := s.db.Create(simulation).Error; err != nil {
		return nil, err
	}

	return simulation, nil
}

func (s *RiskBudgetService) validateSimulationInputs(numSimulations int, timeSteps int, returns []decimal.Decimal) error {
	if numSimulations < 100 {
		return &SimulationError{
			Step:   "param_validation",
			Reason: fmt.Sprintf("number of simulations too small: got %d, minimum 100 required", numSimulations),
			Input:  numSimulations,
		}
	}
	if numSimulations > 100000 {
		return &SimulationError{
			Step:   "param_validation",
			Reason: fmt.Sprintf("number of simulations too large: got %d, maximum 100000", numSimulations),
			Input:  numSimulations,
		}
	}
	if timeSteps < 1 {
		return &SimulationError{
			Step:   "param_validation",
			Reason: fmt.Sprintf("time steps must be at least 1, got %d", timeSteps),
			Input:  timeSteps,
		}
	}
	if timeSteps > 2520 {
		return &SimulationError{
			Step:   "param_validation",
			Reason: fmt.Sprintf("time steps too large: got %d, maximum 2520 (10 years)", timeSteps),
			Input:  timeSteps,
		}
	}
	if len(returns) == 0 {
		return &SimulationError{
			Step:   "input_validation",
			Reason: "returns data is empty",
			Input:  0,
		}
	}
	if len(returns) < 20 {
		return &SimulationError{
			Step:   "input_validation",
			Reason: fmt.Sprintf("insufficient returns data: need at least 20 data points, got %d", len(returns)),
			Input:  len(returns),
		}
	}
	return nil
}

func (s *RiskBudgetService) filterValidReturns(returns []decimal.Decimal) ([]decimal.Decimal, int) {
	validReturns := make([]decimal.Decimal, 0, len(returns))
	invalidCount := 0

	for _, r := range returns {
		f := r.InexactFloat64()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			invalidCount++
			continue
		}
		if r.LessThan(decimal.NewFromFloat(-2.0)) || r.GreaterThan(decimal.NewFromFloat(2.0)) {
			invalidCount++
			continue
		}
		validReturns = append(validReturns, r)
	}

	return validReturns, invalidCount
}

func (s *RiskBudgetService) calculateSimulationStats(returns []decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	var sum decimal.Decimal
	for _, r := range returns {
		sum = sum.Add(r)
	}
	mean := sum.Div(decimal.NewFromInt(int64(len(returns))))

	if math.IsNaN(mean.InexactFloat64()) || math.IsInf(mean.InexactFloat64(), 0) {
		return decimal.Zero, decimal.Zero, errors.New("computed NaN or Inf for mean return")
	}

	variance := decimal.Zero
	for _, r := range returns {
		diff := r.Sub(mean)
		variance = variance.Add(diff.Mul(diff))
	}
	variance = variance.Div(decimal.NewFromInt(int64(len(returns))))

	if variance.LessThan(decimal.Zero) {
		return decimal.Zero, decimal.Zero, errors.New("negative variance detected due to numerical precision issues")
	}

	stdDev := decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64()))

	if math.IsNaN(stdDev.InexactFloat64()) || math.IsInf(stdDev.InexactFloat64(), 0) {
		return decimal.Zero, decimal.Zero, errors.New("computed NaN or Inf for standard deviation")
	}

	return mean, stdDev, nil
}

func (s *RiskBudgetService) GetSimulation(portfolioID uint) (*models.MonteCarloSimulation, error) {
	var simulation models.MonteCarloSimulation
	err := s.db.Where("portfolio_id = ?", portfolioID).
		Order("simulation_date DESC").
		First(&simulation).Error
	if err != nil {
		return nil, err
	}
	return &simulation, nil
}

func (s *RiskBudgetService) SaveRiskContributions(
	simulationID uint,
	contributions []models.RiskContribution,
) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for i := range contributions {
			contributions[i].SimulationID = simulationID
			if err := tx.Create(&contributions[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *RiskBudgetService) GetRiskContributions(simulationID uint) ([]models.RiskContribution, error) {
	var contributions []models.RiskContribution
	err := s.db.Where("simulation_id = ?", simulationID).
		Order("calculation_date DESC").
		Find(&contributions).Error
	return contributions, err
}

func (s *RiskBudgetService) OptimizeRiskBudget(
	returnsMatrix [][]decimal.Decimal,
	targetBudgets []decimal.Decimal,
	confidenceLevel decimal.Decimal,
	maxIterations int,
) ([]decimal.Decimal, []models.RiskContribution, error) {
	if len(returnsMatrix) == 0 || len(targetBudgets) == 0 {
		return nil, nil, ErrInsufficientReturns
	}

	n := len(returnsMatrix)
	if len(targetBudgets) != n {
		return nil, nil, errors.New("target budgets length must match number of assets")
	}

	weights := make([]decimal.Decimal, n)
	for i := range n {
		weights[i] = decimal.NewFromFloat(1.0 / float64(n))
	}

	bestWeights := make([]decimal.Decimal, n)
	copy(bestWeights, weights)

	bestError := decimal.NewFromFloat(math.MaxFloat64)

	learningRate := decimal.NewFromFloat(0.01)

	for range maxIterations {
		contributions, err := s.CalculateRiskContributions(weights, returnsMatrix, confidenceLevel)
		if err != nil {
			return nil, nil, err
		}

		totalError := decimal.Zero
		for i, c := range contributions {
			error := c.CVaRPercentage.Div(decimal.NewFromInt(100)).Sub(targetBudgets[i])
			totalError = totalError.Add(error.Mul(error))
		}

		if totalError.LessThan(bestError) {
			bestError = totalError
			copy(bestWeights, weights)
		}

		if totalError.LessThan(decimal.NewFromFloat(1e-6)) {
			break
		}

		portfolioReturns := make([]decimal.Decimal, len(returnsMatrix[0]))
		for i := range portfolioReturns {
			portfolioReturns[i] = decimal.Zero
			for j := range n {
				portfolioReturns[i] = portfolioReturns[i].Add(
					weights[j].Mul(returnsMatrix[j][i]),
				)
			}
		}

		_, portfolioCVaR, err := s.CalculateCVaR(portfolioReturns, confidenceLevel, false)
		if err != nil {
			return nil, nil, err
		}

		for i := range n {
			delta := decimal.NewFromFloat(0.001)
			newWeights := make([]decimal.Decimal, n)
			copy(newWeights, weights)
			newWeights[i] = newWeights[i].Add(delta)

			sum := decimal.Zero
			for _, w := range newWeights {
				sum = sum.Add(w)
			}
			for j := range newWeights {
				newWeights[j] = newWeights[j].Div(sum)
			}

			newPortfolioReturns := make([]decimal.Decimal, len(returnsMatrix[0]))
			for k := range newPortfolioReturns {
				newPortfolioReturns[k] = decimal.Zero
				for j := range n {
					newPortfolioReturns[k] = newPortfolioReturns[k].Add(
						newWeights[j].Mul(returnsMatrix[j][k]),
					)
				}
			}

			_, newCVaR, err := s.CalculateCVaR(newPortfolioReturns, confidenceLevel, false)
			if err != nil {
				continue
			}

			marginalCVaR := newCVaR.Sub(portfolioCVaR).Div(delta)
			currentRC := weights[i].Mul(marginalCVaR)
			targetRC := targetBudgets[i].Mul(portfolioCVaR)
			gradient := currentRC.Sub(targetRC).Mul(decimal.NewFromInt(2))

			weights[i] = weights[i].Sub(learningRate.Mul(gradient))
			if weights[i].LessThan(decimal.Zero) {
				weights[i] = decimal.Zero
			}
		}

		sum := decimal.Zero
		for _, w := range weights {
			sum = sum.Add(w)
		}
		for i := range weights {
			weights[i] = weights[i].Div(sum)
		}
	}

	finalContributions, err := s.CalculateRiskContributions(bestWeights, returnsMatrix, confidenceLevel)
	if err != nil {
		return nil, nil, err
	}

	return bestWeights, finalContributions, nil
}

func (s *RiskBudgetService) CalculatePortfolioSkewness(
	returnsMatrix [][]decimal.Decimal,
	weights []decimal.Decimal,
) (decimal.Decimal, error) {
	if len(returnsMatrix) == 0 || len(weights) == 0 {
		return decimal.Zero, ErrInsufficientReturns
	}

	n := len(weights)
	portfolioReturns := make([]decimal.Decimal, len(returnsMatrix[0]))
	for i := range portfolioReturns {
		portfolioReturns[i] = decimal.Zero
		for j := range n {
			portfolioReturns[i] = portfolioReturns[i].Add(
				weights[j].Mul(returnsMatrix[j][i]),
			)
		}
	}

	mean := utils.CalculateMean(portfolioReturns)
	stdDev := utils.CalculatePopulationStdDev(portfolioReturns, mean)

	if stdDev.IsZero() {
		return decimal.Zero, nil
	}

	skewness := utils.CalculateSkewness(portfolioReturns, mean, stdDev)

	return skewness, nil
}

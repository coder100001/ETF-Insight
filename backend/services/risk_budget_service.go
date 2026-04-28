package services

import (
	"encoding/json"
	"errors"
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
	for i := 0; i < index; i++ {
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

	numSimulations := 10000
	simulatedReturns := make([]decimal.Decimal, numSimulations)

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < numSimulations; i++ {
		u1 := rand.Float64()
		u2 := rand.Float64()
		z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
		simulatedReturn := mean.InexactFloat64() + stdDev.InexactFloat64()*z
		simulatedReturns[i] = decimal.NewFromFloat(simulatedReturn)
	}

	utils.SortDecimals(simulatedReturns)

	index := int(decimal.NewFromInt(int64(numSimulations)).Mul(
		decimal.NewFromInt(1).Sub(confidenceLevel),
	).IntPart())

	varVaR := simulatedReturns[index]

	var cvarSum decimal.Decimal
	for i := 0; i < index; i++ {
		cvarSum = cvarSum.Add(simulatedReturns[i])
	}
	varCVaR := cvarSum.Div(decimal.NewFromInt(int64(index)))

	return varVaR, varCVaR, nil
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
		for j := 0; j < n; j++ {
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

	for i := 0; i < n; i++ {
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
			for j := 0; j < n; j++ {
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
	if numSimulations <= 0 || timeSteps <= 0 {
		return nil, errors.New("invalid simulation parameters")
	}

	var sum decimal.Decimal
	for _, r := range returns {
		sum = sum.Add(r)
	}
	mean := sum.Div(decimal.NewFromInt(int64(len(returns))))

	variance := decimal.Zero
	for _, r := range returns {
		diff := r.Sub(mean)
		variance = variance.Add(diff.Mul(diff))
	}
	variance = variance.Div(decimal.NewFromInt(int64(len(returns))))
	stdDev := decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64()))

	dt := decimal.NewFromFloat(1.0 / 252.0)
	drift := mean.Mul(dt)
	diffusion := stdDev.Mul(decimalSqrt(dt))

	rand.Seed(time.Now().UnixNano())
	finalReturns := make([]decimal.Decimal, numSimulations)

	for i := 0; i < numSimulations; i++ {
		price := decimal.NewFromInt(100)
		for j := 0; j < timeSteps; j++ {
			u1 := rand.Float64()
			u2 := rand.Float64()
			z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)

			noise := diffusion.Mul(decimal.NewFromFloat(z))
			return_ := drift.Add(noise)
			price = price.Mul(decimal.NewFromInt(1).Add(return_))
		}
		finalReturns[i] = price.Sub(decimal.NewFromInt(100)).Div(decimal.NewFromInt(100))
	}

	utils.SortDecimals(finalReturns)

	meanReturn := utils.CalculateMean(finalReturns)
	stdReturn := utils.CalculatePopulationStdDev(finalReturns, meanReturn)

	percentile5Index := int(float64(numSimulations) * 0.05)
	percentile95Index := int(float64(numSimulations) * 0.95)

	percentile5 := finalReturns[percentile5Index]
	percentile95 := finalReturns[percentile95Index]

	simulationDataJSON, _ := json.Marshal(finalReturns)

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

	if err := s.db.Create(simulation).Error; err != nil {
		return nil, err
	}

	return simulation, nil
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
	for i := 0; i < n; i++ {
		weights[i] = decimal.NewFromFloat(1.0 / float64(n))
	}

	bestWeights := make([]decimal.Decimal, n)
	copy(bestWeights, weights)

	bestError := decimal.NewFromFloat(math.MaxFloat64)

	learningRate := decimal.NewFromFloat(0.01)

	for iter := 0; iter < maxIterations; iter++ {
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
			for j := 0; j < n; j++ {
				portfolioReturns[i] = portfolioReturns[i].Add(
					weights[j].Mul(returnsMatrix[j][i]),
				)
			}
		}

		_, portfolioCVaR, err := s.CalculateCVaR(portfolioReturns, confidenceLevel, false)
		if err != nil {
			return nil, nil, err
		}

		for i := 0; i < n; i++ {
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
				for j := 0; j < n; j++ {
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
		for j := 0; j < n; j++ {
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

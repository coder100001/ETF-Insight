package services

import (
	"errors"
	"testing"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRiskTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(
		&models.RiskBudgetConfig{},
		&models.MonteCarloSimulation{},
		&models.RiskContribution{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func cleanupRiskTestDB(db *gorm.DB) {
	sqlDB, _ := db.DB()
	sqlDB.Close()
}

func TestRiskBudgetService_CreateConfig(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	config := &models.RiskBudgetConfig{
		PortfolioID:         1,
		StockCVaRBudget:     decimal.NewFromFloat(0.05),
		BondCVaRBudget:      decimal.NewFromFloat(0.03),
		CommodityCVaRBudget: decimal.NewFromFloat(0.02),
		CashCVaRBudget:      decimal.NewFromFloat(0.01),
		CVaRConfidence:      decimal.NewFromFloat(0.95),
		IsActive:            true,
		EffectiveDate:       time.Now(),
	}

	err := service.CreateConfig(config)
	if err != nil {
		t.Errorf("CreateConfig failed: %v", err)
	}

	if config.ID == 0 {
		t.Error("Config ID should not be zero after creation")
	}
}

func TestRiskBudgetService_GetConfig(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	config := &models.RiskBudgetConfig{
		PortfolioID:         1,
		StockCVaRBudget:     decimal.NewFromFloat(0.05),
		BondCVaRBudget:      decimal.NewFromFloat(0.03),
		CommodityCVaRBudget: decimal.NewFromFloat(0.02),
		CashCVaRBudget:      decimal.NewFromFloat(0.01),
		CVaRConfidence:      decimal.NewFromFloat(0.95),
		IsActive:            true,
		EffectiveDate:       time.Now(),
	}

	_ = service.CreateConfig(config)

	retrieved, err := service.GetConfig(config.ID)
	if err != nil {
		t.Errorf("GetConfig failed: %v", err)
	}

	if retrieved.PortfolioID != config.PortfolioID {
		t.Errorf("Expected PortfolioID %d, got %d", config.PortfolioID, retrieved.PortfolioID)
	}
}

func TestRiskBudgetService_CalculateHistoricalCVaR(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	returns := []decimal.Decimal{}
	for i := range 100 {
		ret := decimal.NewFromFloat(0.001 * float64(i%10-5))
		returns = append(returns, ret)
	}

	confidenceLevel := decimal.NewFromFloat(0.95)

	varVaR, varCVaR, err := service.CalculateCVaR(returns, confidenceLevel, false)
	if err != nil {
		t.Errorf("CalculateCVaR failed: %v", err)
	}

	if varVaR.IsZero() {
		t.Error("VaR should not be zero")
	}

	if varCVaR.IsZero() {
		t.Error("CVaR should not be zero")
	}

	t.Logf("Historical VaR: %s, CVaR: %s", varVaR.String(), varCVaR.String())
}

func TestRiskBudgetService_CalculateParametricCVaR(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	returns := []decimal.Decimal{}
	for i := range 100 {
		ret := decimal.NewFromFloat(0.001 * float64(i%10-5))
		returns = append(returns, ret)
	}

	confidenceLevel := decimal.NewFromFloat(0.95)

	varVaR, varCVaR, err := service.CalculateCVaR(returns, confidenceLevel, true)
	if err != nil {
		t.Errorf("CalculateCVaR failed: %v", err)
	}

	if varVaR.IsZero() {
		t.Error("VaR should not be zero")
	}

	if varCVaR.IsZero() {
		t.Error("CVaR should not be zero")
	}

	t.Logf("Parametric VaR: %s, CVaR: %s", varVaR.String(), varCVaR.String())
}

func TestRiskBudgetService_CalculateCVaR_InsufficientData(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	returns := []decimal.Decimal{
		decimal.NewFromFloat(0.01),
		decimal.NewFromFloat(-0.02),
		decimal.NewFromFloat(0.03),
	}

	confidenceLevel := decimal.NewFromFloat(0.95)

	_, _, err := service.CalculateCVaR(returns, confidenceLevel, false)
	if err == nil {
		t.Error("Expected error for insufficient data")
	}

	if err != ErrInsufficientReturns {
		t.Errorf("Expected ErrInsufficientReturns, got %v", err)
	}
}

func TestRiskBudgetService_CalculateRiskContributions(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	weights := []decimal.Decimal{
		decimal.NewFromFloat(0.4),
		decimal.NewFromFloat(0.3),
		decimal.NewFromFloat(0.3),
	}

	returnsMatrix := make([][]decimal.Decimal, 3)
	for i := range 3 {
		returnsMatrix[i] = make([]decimal.Decimal, 100)
		for j := range 100 {
			returnsMatrix[i][j] = decimal.NewFromFloat(0.001 * float64((j+i*10)%20-10))
		}
	}

	confidenceLevel := decimal.NewFromFloat(0.95)

	contributions, err := service.CalculateRiskContributions(weights, returnsMatrix, confidenceLevel)
	if err != nil {
		t.Errorf("CalculateRiskContributions failed: %v", err)
	}

	if len(contributions) != 3 {
		t.Errorf("Expected 3 contributions, got %d", len(contributions))
	}

	for i, c := range contributions {
		if c.Weight.IsZero() {
			t.Errorf("Contribution %d weight should not be zero", i)
		}
		t.Logf("Asset %d: Weight=%s, MarginalCVaR=%s, CVaRContribution=%s, CVaRPercentage=%s",
			i, c.Weight.String(), c.MarginalCVaR.String(),
			c.CVaRContribution.String(), c.CVaRPercentage.String())
	}
}

func TestRiskBudgetService_RunMonteCarloSimulation(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	returns := []decimal.Decimal{}
	for i := range 100 {
		ret := decimal.NewFromFloat(0.001 * float64(i%10-5))
		returns = append(returns, ret)
	}

	numSimulations := 1000
	timeSteps := 252

	simulation, err := service.RunMonteCarloSimulation(1, numSimulations, timeSteps, returns)
	if err != nil {
		t.Errorf("RunMonteCarloSimulation failed: %v", err)
	}

	if simulation.ID == 0 {
		t.Error("Simulation ID should not be zero after creation")
	}

	if simulation.NumPaths != numSimulations {
		t.Errorf("Expected NumPaths %d, got %d", numSimulations, simulation.NumPaths)
	}

	if simulation.TimeSteps != timeSteps {
		t.Errorf("Expected TimeSteps %d, got %d", timeSteps, simulation.TimeSteps)
	}

	t.Logf("Simulation: MeanReturn=%s, StdDev=%s, VaR95=%s, CVaR95=%s",
		simulation.MeanReturn.String(), simulation.StdDev.String(),
		simulation.VaR95.String(), simulation.CVaR95.String())
}

func TestRiskBudgetService_GetSimulation(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	returns := []decimal.Decimal{}
	for i := range 100 {
		ret := decimal.NewFromFloat(0.001 * float64(i%10-5))
		returns = append(returns, ret)
	}

	created, _ := service.RunMonteCarloSimulation(1, 1000, 252, returns)

	retrieved, err := service.GetSimulation(1)
	if err != nil {
		t.Errorf("GetSimulation failed: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, retrieved.ID)
	}
}

func TestRiskBudgetService_SaveRiskContributions(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	contributions := []models.RiskContribution{
		{
			AssetSymbol:      "SPY",
			Weight:           decimal.NewFromFloat(0.4),
			CVaRContribution: decimal.NewFromFloat(0.02),
			MarginalCVaR:     decimal.NewFromFloat(0.05),
			CVaRPercentage:   decimal.NewFromFloat(40.0),
			CalculationDate:  time.Now(),
			CreatedAt:        time.Now(),
		},
		{
			AssetSymbol:      "TLT",
			Weight:           decimal.NewFromFloat(0.3),
			CVaRContribution: decimal.NewFromFloat(0.01),
			MarginalCVaR:     decimal.NewFromFloat(0.03),
			CVaRPercentage:   decimal.NewFromFloat(30.0),
			CalculationDate:  time.Now(),
			CreatedAt:        time.Now(),
		},
	}

	err := service.SaveRiskContributions(1, contributions)
	if err != nil {
		t.Errorf("SaveRiskContributions failed: %v", err)
	}

	for _, c := range contributions {
		if c.ID == 0 {
			t.Error("Contribution ID should not be zero after creation")
		}
	}
}

func TestRiskBudgetService_GetRiskContributions(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	contributions := []models.RiskContribution{
		{
			AssetSymbol:      "SPY",
			Weight:           decimal.NewFromFloat(0.4),
			CVaRContribution: decimal.NewFromFloat(0.02),
			MarginalCVaR:     decimal.NewFromFloat(0.05),
			CVaRPercentage:   decimal.NewFromFloat(40.0),
			CalculationDate:  time.Now(),
			CreatedAt:        time.Now(),
		},
	}

	_ = service.SaveRiskContributions(1, contributions)

	retrieved, err := service.GetRiskContributions(1)
	if err != nil {
		t.Errorf("GetRiskContributions failed: %v", err)
	}

	if len(retrieved) != 1 {
		t.Errorf("Expected 1 contribution, got %d", len(retrieved))
	}

	if retrieved[0].AssetSymbol != "SPY" {
		t.Errorf("Expected AssetSymbol SPY, got %s", retrieved[0].AssetSymbol)
	}
}

func TestRiskBudgetService_OptimizeRiskBudget(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	returnsMatrix := make([][]decimal.Decimal, 3)
	for i := range 3 {
		returnsMatrix[i] = make([]decimal.Decimal, 100)
		for j := range 100 {
			returnsMatrix[i][j] = decimal.NewFromFloat(0.001 * float64((j+i*10)%20-10))
		}
	}

	targetBudgets := []decimal.Decimal{
		decimal.NewFromFloat(0.4),
		decimal.NewFromFloat(0.35),
		decimal.NewFromFloat(0.25),
	}

	confidenceLevel := decimal.NewFromFloat(0.95)

	weights, contributions, err := service.OptimizeRiskBudget(returnsMatrix, targetBudgets, confidenceLevel, 100)
	if err != nil {
		t.Errorf("OptimizeRiskBudget failed: %v", err)
	}

	if len(weights) != 3 {
		t.Errorf("Expected 3 weights, got %d", len(weights))
	}

	if len(contributions) != 3 {
		t.Errorf("Expected 3 contributions, got %d", len(contributions))
	}

	sum := decimal.Zero
	for _, w := range weights {
		sum = sum.Add(w)
	}
	if !sum.Equal(decimal.NewFromInt(1)) {
		t.Errorf("Weights should sum to 1, got %s", sum.String())
	}

	for i, c := range contributions {
		t.Logf("Asset %d: Weight=%s, CVaRPercentage=%s", i, weights[i].String(), c.CVaRPercentage.String())
	}
}

func TestRiskBudgetService_OptimizeRiskBudget_InvalidInputs(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	_, _, err := service.OptimizeRiskBudget(nil, nil, decimal.NewFromFloat(0.95), 100)
	if err == nil {
		t.Error("Expected error for nil inputs")
	}
	if err != ErrInsufficientReturns {
		t.Errorf("Expected ErrInsufficientReturns, got %v", err)
	}

	returnsMatrix := make([][]decimal.Decimal, 2)
	targetBudgets := []decimal.Decimal{decimal.NewFromFloat(0.5)}

	_, _, err = service.OptimizeRiskBudget(returnsMatrix, targetBudgets, decimal.NewFromFloat(0.95), 100)
	if err == nil {
		t.Error("Expected error for mismatched lengths")
	}
}

func TestRiskBudgetService_CalculatePortfolioSkewness(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	returnsMatrix := make([][]decimal.Decimal, 3)
	for i := range 3 {
		returnsMatrix[i] = make([]decimal.Decimal, 100)
		for j := range 100 {
			returnsMatrix[i][j] = decimal.NewFromFloat(0.001 * float64((j+i*10)%20-10))
		}
	}

	weights := []decimal.Decimal{
		decimal.NewFromFloat(0.4),
		decimal.NewFromFloat(0.3),
		decimal.NewFromFloat(0.3),
	}

	skewness, err := service.CalculatePortfolioSkewness(returnsMatrix, weights)
	if err != nil {
		t.Errorf("CalculatePortfolioSkewness failed: %v", err)
	}

	t.Logf("Portfolio Skewness: %s", skewness.String())
}

func TestRiskBudgetService_CalculatePortfolioSkewness_InsufficientData(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	_, err := service.CalculatePortfolioSkewness(nil, nil)
	if err == nil {
		t.Error("Expected error for nil inputs")
	}
	if err != ErrInsufficientReturns {
		t.Errorf("Expected ErrInsufficientReturns, got %v", err)
	}
}

func TestRunMonteCarloSimulation_InvalidNumSimulations(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)
	returns := generateTestReturns(100)

	_, err := service.RunMonteCarloSimulation(1, 50, 252, returns)
	if err == nil {
		t.Fatal("Expected error for numSimulations < 100")
	}
	var simErr *SimulationError
	if !errors.As(err, &simErr) {
		t.Errorf("Expected SimulationError, got %T: %v", err, err)
	}
	if simErr.Step != "param_validation" {
		t.Errorf("Expected step 'param_validation', got %q", simErr.Step)
	}
}

func TestRunMonteCarloSimulation_ExcessiveNumSimulations(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)
	returns := generateTestReturns(100)

	_, err := service.RunMonteCarloSimulation(1, 200000, 252, returns)
	if err == nil {
		t.Fatal("Expected error for numSimulations > 100000")
	}
	var simErr *SimulationError
	if !errors.As(err, &simErr) {
		t.Errorf("Expected SimulationError, got %T: %v", err, err)
	}
}

func TestRunMonteCarloSimulation_InvalidTimeSteps(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)
	returns := generateTestReturns(100)

	_, err := service.RunMonteCarloSimulation(1, 1000, 0, returns)
	if err == nil {
		t.Fatal("Expected error for timeSteps < 1")
	}

	_, err = service.RunMonteCarloSimulation(1, 1000, 5000, returns)
	if err == nil {
		t.Fatal("Expected error for timeSteps > 2520")
	}
}

func TestRunMonteCarloSimulation_EmptyReturns(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	_, err := service.RunMonteCarloSimulation(1, 1000, 252, []decimal.Decimal{})
	if err == nil {
		t.Fatal("Expected error for empty returns")
	}
	var simErr *SimulationError
	if !errors.As(err, &simErr) {
		t.Errorf("Expected SimulationError, got %T: %v", err, err)
	}
}

func TestRunMonteCarloSimulation_InsufficientReturns(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)
	returns := generateTestReturns(15)

	_, err := service.RunMonteCarloSimulation(1, 1000, 252, returns)
	if err == nil {
		t.Fatal("Expected error for insufficient returns (< 20)")
	}
}

func TestRunMonteCarloSimulation_WithExtremeReturns(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)
	returns := generateTestReturns(100)
	returns[0] = decimal.NewFromFloat(3.0)
	returns[1] = decimal.NewFromFloat(-3.0)

	simulation, err := service.RunMonteCarloSimulation(1, 1000, 252, returns)
	if err != nil {
		t.Errorf("Should handle extreme returns by filtering: %v", err)
	}
	if simulation.ID == 0 {
		t.Error("Simulation ID should not be zero")
	}
}

func TestRunMonteCarloSimulation_ZeroVolatility(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	returns := make([]decimal.Decimal, 20)
	for i := range 20 {
		returns[i] = decimal.NewFromFloat(0.001)
	}

	simulation, err := service.RunMonteCarloSimulation(1, 1000, 252, returns)
	if err != nil {
		t.Errorf("Should handle zero volatility gracefully: %v", err)
	}
	if simulation.ID == 0 {
		t.Error("Simulation ID should not be zero")
	}
}

func TestRunMonteCarloSimulation_WithFilteredReturns(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	returns := generateTestReturns(25)
	returns[0] = decimal.NewFromFloat(5.0)
	returns[1] = decimal.NewFromFloat(-5.0)

	simulation, err := service.RunMonteCarloSimulation(1, 1000, 252, returns)
	if err != nil {
		t.Errorf("Should filter out extreme returns: %v", err)
	}

	if simulation.MeanReturn.IsZero() && simulation.StdDev.IsZero() {
		t.Error("Simulation results should not be all zeros after filtering")
	}
}

func TestRunMonteCarloSimulation_MinimalValidInput(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	returns := generateTestReturns(20)

	simulation, err := service.RunMonteCarloSimulation(1, 100, 1, returns)
	if err != nil {
		t.Errorf("Should accept minimal valid input: %v", err)
	}
	if simulation.NumPaths != 100 {
		t.Errorf("Expected NumPaths 100, got %d", simulation.NumPaths)
	}
	if simulation.TimeSteps != 1 {
		t.Errorf("Expected TimeSteps 1, got %d", simulation.TimeSteps)
	}
}

func TestRunMonteCarloSimulation_ResultsInRange(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)
	returns := generateTestReturns(100)

	simulation, err := service.RunMonteCarloSimulation(1, 1000, 252, returns)
	if err != nil {
		t.Errorf("RunMonteCarloSimulation failed: %v", err)
	}

	if simulation.VaR95.GreaterThan(decimal.NewFromFloat(0.5)) || simulation.VaR95.LessThan(decimal.NewFromFloat(-1.0)) {
		t.Errorf("VaR95 seems unreasonable: %s", simulation.VaR95.String())
	}

	if simulation.MeanReturn.GreaterThan(decimal.NewFromFloat(2.0)) || simulation.MeanReturn.LessThan(decimal.NewFromFloat(-2.0)) {
		t.Errorf("MeanReturn seems unreasonable: %s", simulation.MeanReturn.String())
	}

	t.Logf("Simulation: MeanReturn=%s, StdDev=%s, VaR95=%s, CVaR95=%s",
		simulation.MeanReturn.String(), simulation.StdDev.String(),
		simulation.VaR95.String(), simulation.CVaR95.String())
}

func TestRiskBudgetService_CalculateMonteCarloCVaR(t *testing.T) {
	db := setupRiskTestDB(t)
	defer cleanupRiskTestDB(db)

	service := NewRiskBudgetService(db)

	returns := make([]decimal.Decimal, 0, 200)
	for i := range 200 {
		ret := decimal.NewFromFloat(0.001 * float64(i%10-5))
		returns = append(returns, ret)
	}

	varVaR, varCVaR, err := service.CalculateMonteCarloCVaR(returns, decimal.NewFromFloat(0.95))
	if err != nil {
		t.Errorf("CalculateMonteCarloCVaR failed: %v", err)
	}

	if varVaR.IsZero() {
		t.Error("VaR should not be zero")
	}
	if varCVaR.IsZero() {
		t.Error("CVaR should not be zero")
	}

	t.Logf("Monte Carlo VaR95: %s, CVaR95: %s", varVaR.String(), varCVaR.String())
}

func generateTestReturns(n int) []decimal.Decimal {
	returns := make([]decimal.Decimal, 0, n)
	for i := range n {
		ret := decimal.NewFromFloat(0.001 * float64(i%10-5))
		returns = append(returns, ret)
	}
	return returns
}

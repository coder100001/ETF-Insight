package services

import (
	"testing"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAlphaTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(
		&models.AlphaView{},
		&models.AlphaViewPerformance{},
		&models.BlackLittermanConfig{},
		&models.BLPosteriorReturn{},
		&models.FactorTimingSignal{},
		&models.Portfolio{},
		&models.PortfolioPosition{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func cleanupAlphaTestDB(db *gorm.DB) {
	sqlDB, _ := db.DB()
	sqlDB.Close()
}

func TestAlphaViewService_CreateAlphaView(t *testing.T) {
	db := setupAlphaTestDB(t)
	defer cleanupAlphaTestDB(db)

	factorService := NewFactorDataService(db)
	service := NewAlphaViewService(db, factorService)

	view := &models.AlphaView{
		PortfolioID:  1,
		AssetSymbol:  "SPY",
		ViewReturn:   decimal.NewFromFloat(0.05),
		Confidence:   decimal.NewFromFloat(75.0),
		ViewType:     models.ViewTypeAbsolute,
		ViewMethod:   models.ViewMethodFactorTiming,
		GeneratedAt:  time.Now(),
		ValidUntil:   time.Now().AddDate(0, 0, 30),
		Status:       models.ViewStatusActive,
		SourceFactor: "Mkt-RF",
		CreatedAt:    time.Now(),
	}

	err := service.CreateAlphaView(view)
	if err != nil {
		t.Errorf("CreateAlphaView failed: %v", err)
	}

	if view.ID == 0 {
		t.Error("View ID should not be zero after creation")
	}
}

func TestAlphaViewService_CreateAlphaView_InvalidViewType(t *testing.T) {
	db := setupAlphaTestDB(t)
	defer cleanupAlphaTestDB(db)

	factorService := NewFactorDataService(db)
	service := NewAlphaViewService(db, factorService)

	view := &models.AlphaView{
		PortfolioID: 1,
		AssetSymbol: "SPY",
		ViewReturn:  decimal.NewFromFloat(0.05),
		Confidence:  decimal.NewFromFloat(75.0),
		ViewType:    models.ViewType("invalid"),
		ViewMethod:  models.ViewMethodFactorTiming,
		GeneratedAt: time.Now(),
		ValidUntil:  time.Now().AddDate(0, 0, 30),
		Status:      models.ViewStatusActive,
		CreatedAt:   time.Now(),
	}

	err := service.CreateAlphaView(view)
	if err == nil {
		t.Error("Expected error for invalid view type")
	}
	if err != ErrInvalidViewType {
		t.Errorf("Expected ErrInvalidViewType, got %v", err)
	}
}

func TestAlphaViewService_CreateAlphaView_InvalidConfidence(t *testing.T) {
	db := setupAlphaTestDB(t)
	defer cleanupAlphaTestDB(db)

	factorService := NewFactorDataService(db)
	service := NewAlphaViewService(db, factorService)

	view := &models.AlphaView{
		PortfolioID: 1,
		AssetSymbol: "SPY",
		ViewReturn:  decimal.NewFromFloat(0.05),
		Confidence:  decimal.NewFromFloat(150.0),
		ViewType:    models.ViewTypeAbsolute,
		ViewMethod:  models.ViewMethodFactorTiming,
		GeneratedAt: time.Now(),
		ValidUntil:  time.Now().AddDate(0, 0, 30),
		Status:      models.ViewStatusActive,
		CreatedAt:   time.Now(),
	}

	err := service.CreateAlphaView(view)
	if err == nil {
		t.Error("Expected error for invalid confidence")
	}
	if err != ErrInvalidConfidence {
		t.Errorf("Expected ErrInvalidConfidence, got %v", err)
	}
}

func TestAlphaViewService_CreateAlphaView_ExpiredView(t *testing.T) {
	db := setupAlphaTestDB(t)
	defer cleanupAlphaTestDB(db)

	factorService := NewFactorDataService(db)
	service := NewAlphaViewService(db, factorService)

	view := &models.AlphaView{
		PortfolioID: 1,
		AssetSymbol: "SPY",
		ViewReturn:  decimal.NewFromFloat(0.05),
		Confidence:  decimal.NewFromFloat(75.0),
		ViewType:    models.ViewTypeAbsolute,
		ViewMethod:  models.ViewMethodFactorTiming,
		GeneratedAt: time.Now(),
		ValidUntil:  time.Now().AddDate(0, 0, -1),
		Status:      models.ViewStatusActive,
		CreatedAt:   time.Now(),
	}

	err := service.CreateAlphaView(view)
	if err == nil {
		t.Error("Expected error for expired view")
	}
	if err != ErrViewExpired {
		t.Errorf("Expected ErrViewExpired, got %v", err)
	}
}

func TestAlphaViewService_GetAlphaView(t *testing.T) {
	db := setupAlphaTestDB(t)
	defer cleanupAlphaTestDB(db)

	factorService := NewFactorDataService(db)
	service := NewAlphaViewService(db, factorService)

	view := &models.AlphaView{
		PortfolioID:  1,
		AssetSymbol:  "SPY",
		ViewReturn:   decimal.NewFromFloat(0.05),
		Confidence:   decimal.NewFromFloat(75.0),
		ViewType:     models.ViewTypeAbsolute,
		ViewMethod:   models.ViewMethodFactorTiming,
		GeneratedAt:  time.Now(),
		ValidUntil:   time.Now().AddDate(0, 0, 30),
		Status:       models.ViewStatusActive,
		SourceFactor: "Mkt-RF",
		CreatedAt:    time.Now(),
	}

	_ = service.CreateAlphaView(view)

	retrieved, err := service.GetAlphaView(view.ID)
	if err != nil {
		t.Errorf("GetAlphaView failed: %v", err)
	}

	if retrieved.AssetSymbol != view.AssetSymbol {
		t.Errorf("Expected AssetSymbol %s, got %s", view.AssetSymbol, retrieved.AssetSymbol)
	}

	if !retrieved.ViewReturn.Equal(view.ViewReturn) {
		t.Errorf("Expected ViewReturn %s, got %s", view.ViewReturn.String(), retrieved.ViewReturn.String())
	}
}

func TestAlphaViewService_GetActiveAlphaViews(t *testing.T) {
	db := setupAlphaTestDB(t)
	defer cleanupAlphaTestDB(db)

	factorService := NewFactorDataService(db)
	service := NewAlphaViewService(db, factorService)

	for i := range 3 {
		view := &models.AlphaView{
			PortfolioID: 1,
			AssetSymbol: "SPY",
			ViewReturn:  decimal.NewFromFloat(0.05 * float64(i+1)),
			Confidence:  decimal.NewFromFloat(75.0),
			ViewType:    models.ViewTypeAbsolute,
			ViewMethod:  models.ViewMethodFactorTiming,
			GeneratedAt: time.Now(),
			ValidUntil:  time.Now().AddDate(0, 0, 30),
			Status:      models.ViewStatusActive,
			CreatedAt:   time.Now(),
		}
		_ = service.CreateAlphaView(view)
	}

	views, err := service.GetActiveAlphaViews("SPY")
	if err != nil {
		t.Errorf("GetActiveAlphaViews failed: %v", err)
	}

	if len(views) != 3 {
		t.Errorf("Expected 3 views, got %d", len(views))
	}
}

func TestAlphaViewService_DeactivateView(t *testing.T) {
	db := setupAlphaTestDB(t)
	defer cleanupAlphaTestDB(db)

	factorService := NewFactorDataService(db)
	service := NewAlphaViewService(db, factorService)

	view := &models.AlphaView{
		PortfolioID: 1,
		AssetSymbol: "SPY",
		ViewReturn:  decimal.NewFromFloat(0.05),
		Confidence:  decimal.NewFromFloat(75.0),
		ViewType:    models.ViewTypeAbsolute,
		ViewMethod:  models.ViewMethodFactorTiming,
		GeneratedAt: time.Now(),
		ValidUntil:  time.Now().AddDate(0, 0, 30),
		Status:      models.ViewStatusActive,
		CreatedAt:   time.Now(),
	}

	_ = service.CreateAlphaView(view)

	err := service.DeactivateView(view.ID)
	if err != nil {
		t.Errorf("DeactivateView failed: %v", err)
	}

	retrieved, _ := service.GetAlphaView(view.ID)
	if retrieved.Status != models.ViewStatusExpired {
		t.Errorf("Expected status %s, got %s", models.ViewStatusExpired, retrieved.Status)
	}
}

func TestAlphaViewService_RecordViewPerformance(t *testing.T) {
	db := setupAlphaTestDB(t)
	defer cleanupAlphaTestDB(db)

	factorService := NewFactorDataService(db)
	service := NewAlphaViewService(db, factorService)

	view := &models.AlphaView{
		PortfolioID: 1,
		AssetSymbol: "SPY",
		ViewReturn:  decimal.NewFromFloat(0.05),
		Confidence:  decimal.NewFromFloat(75.0),
		ViewType:    models.ViewTypeAbsolute,
		ViewMethod:  models.ViewMethodFactorTiming,
		GeneratedAt: time.Now(),
		ValidUntil:  time.Now().AddDate(0, 0, 30),
		Status:      models.ViewStatusActive,
		CreatedAt:   time.Now(),
	}
	_ = service.CreateAlphaView(view)

	performance := &models.AlphaViewPerformance{
		ViewID:          view.ID,
		ActualReturn:    decimal.NewFromFloat(0.06),
		PredictionError: decimal.NewFromFloat(0.01),
		IsValidated:     true,
		ValidationDate:  time.Now(),
		IsCorrect:       true,
		CreatedAt:       time.Now(),
	}

	err := service.RecordViewPerformance(performance)
	if err != nil {
		t.Errorf("RecordViewPerformance failed: %v", err)
	}

	if performance.ID == 0 {
		t.Error("Performance ID should not be zero after creation")
	}
}

func TestAlphaViewService_GetViewPerformance(t *testing.T) {
	db := setupAlphaTestDB(t)
	defer cleanupAlphaTestDB(db)

	factorService := NewFactorDataService(db)
	service := NewAlphaViewService(db, factorService)

	view := &models.AlphaView{
		PortfolioID: 1,
		AssetSymbol: "SPY",
		ViewReturn:  decimal.NewFromFloat(0.05),
		Confidence:  decimal.NewFromFloat(75.0),
		ViewType:    models.ViewTypeAbsolute,
		ViewMethod:  models.ViewMethodFactorTiming,
		GeneratedAt: time.Now(),
		ValidUntil:  time.Now().AddDate(0, 0, 30),
		Status:      models.ViewStatusActive,
		CreatedAt:   time.Now(),
	}
	_ = service.CreateAlphaView(view)

	performance := &models.AlphaViewPerformance{
		ViewID:          view.ID,
		ActualReturn:    decimal.NewFromFloat(0.06),
		PredictionError: decimal.NewFromFloat(0.01),
		IsValidated:     true,
		ValidationDate:  time.Now(),
		IsCorrect:       true,
		CreatedAt:       time.Now(),
	}
	_ = service.RecordViewPerformance(performance)

	retrieved, err := service.GetViewPerformance(view.ID)
	if err != nil {
		t.Errorf("GetViewPerformance failed: %v", err)
	}

	if !retrieved.ActualReturn.Equal(performance.ActualReturn) {
		t.Errorf("Expected ActualReturn %s, got %s", performance.ActualReturn.String(), retrieved.ActualReturn.String())
	}
}

func TestBlackLittermanService_CreateConfig(t *testing.T) {
	db := setupAlphaTestDB(t)
	defer cleanupAlphaTestDB(db)

	factorService := NewFactorDataService(db)
	alphaService := NewAlphaViewService(db, factorService)
	service := NewBlackLittermanService(db, alphaService)

	config := &models.BlackLittermanConfig{
		PortfolioID:    1,
		RiskAversion:   decimal.NewFromFloat(2.5),
		PriorType:      models.PriorTypeEqualWeight,
		PriorWeights:   models.JSONMap{"0": 0.25, "1": 0.25, "2": 0.25, "3": 0.25},
		OmegaMethod:    models.OmegaMethodIdzorek,
		OmegaMatrix:    models.JSONMap{"0": models.JSONMap{"0": 0.01, "1": 0}, "1": models.JSONMap{"0": 0, "1": 0.01}},
		IsActive:       true,
		LastCalculated: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := service.CreateConfig(config)
	if err != nil {
		t.Errorf("CreateConfig failed: %v", err)
	}

	if config.ID == 0 {
		t.Error("Config ID should not be zero after creation")
	}
}

func TestBlackLittermanService_CreateConfig_InvalidPriorType(t *testing.T) {
	db := setupAlphaTestDB(t)
	defer cleanupAlphaTestDB(db)

	factorService := NewFactorDataService(db)
	alphaService := NewAlphaViewService(db, factorService)
	service := NewBlackLittermanService(db, alphaService)

	config := &models.BlackLittermanConfig{
		PortfolioID:    1,
		RiskAversion:   decimal.NewFromFloat(2.5),
		PriorType:      models.PriorType("invalid"),
		PriorWeights:   models.JSONMap{"0": 0.25, "1": 0.25, "2": 0.25, "3": 0.25},
		OmegaMethod:    models.OmegaMethodIdzorek,
		OmegaMatrix:    models.JSONMap{"0": models.JSONMap{"0": 0.01, "1": 0}, "1": models.JSONMap{"0": 0, "1": 0.01}},
		IsActive:       true,
		LastCalculated: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := service.CreateConfig(config)
	if err == nil {
		t.Error("Expected error for invalid prior type")
	}
}

func TestBlackLittermanService_CreateConfig_InvalidRiskAversion(t *testing.T) {
	db := setupAlphaTestDB(t)
	defer cleanupAlphaTestDB(db)

	factorService := NewFactorDataService(db)
	alphaService := NewAlphaViewService(db, factorService)
	service := NewBlackLittermanService(db, alphaService)

	config := &models.BlackLittermanConfig{
		PortfolioID:    1,
		RiskAversion:   decimal.NewFromFloat(-1.0),
		PriorType:      models.PriorTypeEqualWeight,
		PriorWeights:   models.JSONMap{"0": 0.25, "1": 0.25, "2": 0.25, "3": 0.25},
		OmegaMethod:    models.OmegaMethodIdzorek,
		OmegaMatrix:    models.JSONMap{"0": models.JSONMap{"0": 0.01, "1": 0}, "1": models.JSONMap{"0": 0, "1": 0.01}},
		IsActive:       true,
		LastCalculated: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := service.CreateConfig(config)
	if err == nil {
		t.Error("Expected error for invalid risk aversion")
	}
}

func TestBlackLittermanService_GetConfig(t *testing.T) {
	db := setupAlphaTestDB(t)
	defer cleanupAlphaTestDB(db)

	factorService := NewFactorDataService(db)
	alphaService := NewAlphaViewService(db, factorService)
	service := NewBlackLittermanService(db, alphaService)

	config := &models.BlackLittermanConfig{
		PortfolioID:    1,
		RiskAversion:   decimal.NewFromFloat(2.5),
		PriorType:      models.PriorTypeEqualWeight,
		PriorWeights:   models.JSONMap{"0": 0.25, "1": 0.25, "2": 0.25, "3": 0.25},
		OmegaMethod:    models.OmegaMethodIdzorek,
		OmegaMatrix:    models.JSONMap{"0": models.JSONMap{"0": 0.01, "1": 0}, "1": models.JSONMap{"0": 0, "1": 0.01}},
		IsActive:       true,
		LastCalculated: time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
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

func TestBlackLittermanService_parseMarketWeights(t *testing.T) {
	db := setupAlphaTestDB(t)
	defer cleanupAlphaTestDB(db)

	factorService := NewFactorDataService(db)
	alphaService := NewAlphaViewService(db, factorService)
	service := NewBlackLittermanService(db, alphaService)

	weightsJSON := models.JSONMap{"0": 0.25, "1": 0.25, "2": 0.25, "3": 0.25}
	weights, err := service.parseMarketWeights(weightsJSON)
	if err != nil {
		t.Errorf("parseMarketWeights failed: %v", err)
	}

	if len(weights) != 4 {
		t.Errorf("Expected 4 weights, got %d", len(weights))
	}

	for i, w := range weights {
		if !w.Equal(decimal.NewFromFloat(0.25)) {
			t.Errorf("Weight %d should be 0.25, got %s", i, w.String())
		}
	}
}

func TestBlackLittermanService_parseCovarianceMatrix(t *testing.T) {
	db := setupAlphaTestDB(t)
	defer cleanupAlphaTestDB(db)

	factorService := NewFactorDataService(db)
	alphaService := NewAlphaViewService(db, factorService)
	service := NewBlackLittermanService(db, alphaService)

	covJSON := models.JSONMap{
		"0": models.JSONMap{"0": 0.04, "1": 0.01},
		"1": models.JSONMap{"0": 0.01, "1": 0.09},
	}
	cov, err := service.parseCovarianceMatrix(covJSON)
	if err != nil {
		t.Errorf("parseCovarianceMatrix failed: %v", err)
	}

	if len(cov) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(cov))
	}

	if len(cov[0]) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(cov[0]))
	}
}

package services

import (
	"errors"
	"strings"
	"testing"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupErrorHandlingTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(
		&models.AlphaView{},
		&models.BlackLittermanConfig{},
		&models.BLPosteriorReturn{},
		&models.Portfolio{},
		&models.PortfolioPosition{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestBlackLittermanService_CalculatePosteriorReturns_ErrorWrapping(t *testing.T) {
	db := setupErrorHandlingTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	factorService := NewFactorDataService(db)
	alphaService := NewAlphaViewService(db, factorService)
	service := NewBlackLittermanService(db, alphaService)

	config := &models.BlackLittermanConfig{
		PortfolioID:  99999,
		RiskAversion: decimal.NewFromFloat(2.5),
		PriorType:    "equal_weight",
		OmegaMethod:  "idzorek",
	}

	err := db.Create(config).Error
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	views := []models.AlphaView{
		{
			AssetSymbol: "INVALID_SYMBOL",
			ViewReturn:  decimal.NewFromFloat(0.05),
			Confidence:  decimal.NewFromFloat(75.0),
			ViewType:    models.ViewTypeAbsolute,
		},
	}

	_, err = service.CalculatePosteriorReturns(config.ID, views)

	if err == nil {
		t.Error("Expected error for invalid portfolio, got nil")
		return
	}

	errMsg := err.Error()

	if !strings.Contains(errMsg, "failed") && !strings.Contains(errMsg, "error") {
		t.Errorf("Error should be wrapped with context message, got: %s", errMsg)
	}
}

func TestBlackLittermanService_CalculatePosteriorReturns_InvalidWeightsJSON(t *testing.T) {
	db := setupErrorHandlingTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	factorService := NewFactorDataService(db)
	alphaService := NewAlphaViewService(db, factorService)
	service := NewBlackLittermanService(db, alphaService)

	portfolio := &models.Portfolio{
		Name: "Test Portfolio",
		Positions: []models.PortfolioPosition{
			{Symbol: "SPY", Weight: decimal.NewFromFloat(0.6)},
			{Symbol: "QQQ", Weight: decimal.NewFromFloat(0.4)},
		},
	}
	if err := db.Create(portfolio).Error; err != nil {
		t.Fatalf("Failed to create portfolio: %v", err)
	}

	config := &models.BlackLittermanConfig{
		PortfolioID:  portfolio.ID,
		RiskAversion: decimal.NewFromFloat(2.5),
		PriorType:    "custom",
		PriorWeights: models.JSONMap{"invalid": "json"},
		OmegaMethod:  "idzorek",
	}
	if err := db.Create(config).Error; err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	views := []models.AlphaView{}

	_, err := service.CalculatePosteriorReturns(config.ID, views)

	if err == nil {
		t.Error("Expected error for invalid weights JSON, got nil")
		return
	}

	if !strings.Contains(err.Error(), "weights") &&
		!strings.Contains(err.Error(), "parse") {
		t.Errorf("Error should mention weights parsing issue, got: %s", err.Error())
	}
}

func TestBlackLittermanService_CalculatePosteriorReturns_NilDatabase(t *testing.T) {
	service := NewBlackLittermanService(nil, nil)

	views := []models.AlphaView{}

	result, err := service.CalculatePosteriorReturns(1, views)

	if err == nil {
		t.Error("Expected error when database is nil, got nil")
	}

	if result != nil {
		t.Error("Expected nil result when database is nil")
	}

	if !errors.Is(err, ErrDatabaseNotInitialized) {
		t.Errorf("Expected ErrDatabaseNotInitialized, got: %v", err)
	}
}

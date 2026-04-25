package services

import (
	"testing"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.FactorData{},
		&models.FactorTimingSignal{},
	)
	require.NoError(t, err)

	return db
}

func cleanupTestDB(t *testing.T, db *gorm.DB) {
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()
}

func TestFactorDataService_CreateFactorData(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	service := NewFactorDataService(db)

	data := &models.FactorData{
		FactorName: "Mkt-RF",
		Date:       time.Now(),
		Value:      decimal.NewFromFloat(0.005),
		DataSource: "Test",
	}

	err := service.CreateFactorData(data)
	assert.NoError(t, err)
	assert.NotZero(t, data.ID)
}

func TestFactorDataService_GetFactorData(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	service := NewFactorDataService(db)

	now := time.Now()
	testData := []models.FactorData{
		{FactorName: "Mkt-RF", Date: now.AddDate(0, 0, -2), Value: decimal.NewFromFloat(0.003), DataSource: "Test"},
		{FactorName: "Mkt-RF", Date: now.AddDate(0, 0, -1), Value: decimal.NewFromFloat(0.004), DataSource: "Test"},
		{FactorName: "Mkt-RF", Date: now, Value: decimal.NewFromFloat(0.005), DataSource: "Test"},
	}

	for _, d := range testData {
		err := service.CreateFactorData(&d)
		assert.NoError(t, err)
	}

	startDate := now.AddDate(0, 0, -2)
	endDate := now

	data, err := service.GetFactorData("Mkt-RF", startDate, endDate)
	assert.NoError(t, err)
	assert.Len(t, data, 3)
}

func TestFactorDataService_GetLatestFactorData(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	service := NewFactorDataService(db)

	now := time.Now()
	testData := []models.FactorData{
		{FactorName: "Mkt-RF", Date: now.AddDate(0, 0, -1), Value: decimal.NewFromFloat(0.003), DataSource: "Test"},
		{FactorName: "Mkt-RF", Date: now, Value: decimal.NewFromFloat(0.005), DataSource: "Test"},
	}

	for _, d := range testData {
		err := service.CreateFactorData(&d)
		assert.NoError(t, err)
	}

	latest, err := service.GetLatestFactorData("Mkt-RF")
	assert.NoError(t, err)
	assert.Equal(t, decimal.NewFromFloat(0.005), latest.Value)
}

func TestFactorDataService_CalculateTimingSignal(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	service := NewFactorDataService(db)

	now := time.Now()
	for i := 0; i < 120; i++ {
		data := &models.FactorData{
			FactorName: "Mkt-RF",
			Date:       now.AddDate(0, 0, -i),
			Value:      decimal.NewFromFloat(float64(i) * 0.001),
			DataSource: "Test",
		}
		err := service.CreateFactorData(data)
		assert.NoError(t, err)
	}

	signal, err := service.CalculateTimingSignal("Mkt-RF", 60)
	assert.NoError(t, err)
	assert.NotNil(t, signal)
	assert.Equal(t, "Mkt-RF", signal.FactorName)
	assert.True(t, signal.SignalStrength.IsValid())
	assert.NotZero(t, signal.SignalScore)
}

func TestFactorDataService_CalculateTimingSignal_InsufficientData(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	service := NewFactorDataService(db)

	now := time.Now()
	for i := 0; i < 10; i++ {
		data := &models.FactorData{
			FactorName: "Mkt-RF",
			Date:       now.AddDate(0, 0, -i),
			Value:      decimal.NewFromFloat(float64(i) * 0.001),
			DataSource: "Test",
		}
		err := service.CreateFactorData(data)
		assert.NoError(t, err)
	}

	_, err := service.CalculateTimingSignal("Mkt-RF", 60)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient data")
}

func TestCalculateMASlope(t *testing.T) {
	values := make([]decimal.Decimal, 61)
	for i := 0; i < 61; i++ {
		values[i] = decimal.NewFromFloat(float64(i))
	}

	slope := calculateMASlope(values, 60)
	assert.True(t, slope.GreaterThan(decimal.Zero))
}

func TestCalculateZScore(t *testing.T) {
	values := []decimal.Decimal{
		decimal.NewFromFloat(1.0),
		decimal.NewFromFloat(2.0),
		decimal.NewFromFloat(3.0),
		decimal.NewFromFloat(4.0),
		decimal.NewFromFloat(5.0),
	}

	zScore := calculateZScore(values)
	assert.True(t, zScore.GreaterThan(decimal.Zero))
}

func TestCalculatePercentile(t *testing.T) {
	values := []decimal.Decimal{
		decimal.NewFromFloat(1.0),
		decimal.NewFromFloat(2.0),
		decimal.NewFromFloat(3.0),
		decimal.NewFromFloat(4.0),
		decimal.NewFromFloat(5.0),
	}

	percentile := calculatePercentile(values)
	assert.True(t, percentile.GreaterThanOrEqual(decimal.Zero))
	assert.True(t, percentile.LessThanOrEqual(decimal.NewFromInt(100)))
}

func TestDetermineSignalStrength(t *testing.T) {
	tests := []struct {
		name     string
		zScore   decimal.Decimal
		expected models.SignalStrength
	}{
		{
			name:     "Strong positive",
			zScore:   decimal.NewFromFloat(2.5),
			expected: models.SignalStrengthStrongPositive,
		},
		{
			name:     "Weak positive",
			zScore:   decimal.NewFromFloat(1.6),
			expected: models.SignalStrengthWeakPositive,
		},
		{
			name:     "Neutral",
			zScore:   decimal.NewFromFloat(0.5),
			expected: models.SignalStrengthNeutral,
		},
		{
			name:     "Weak negative",
			zScore:   decimal.NewFromFloat(-1.6),
			expected: models.SignalStrengthWeakNegative,
		},
		{
			name:     "Strong negative",
			zScore:   decimal.NewFromFloat(-2.5),
			expected: models.SignalStrengthStrongNegative,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineSignalStrength(tt.zScore)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateExpectedReturn(t *testing.T) {
	maSlope := decimal.NewFromFloat(0.001)
	zScore := decimal.NewFromFloat(1.5)

	expectedReturn := calculateExpectedReturn(maSlope, zScore)
	assert.True(t, expectedReturn.GreaterThan(decimal.Zero))
}

func TestCalculateConfidence(t *testing.T) {
	tests := []struct {
		name       string
		zScore     decimal.Decimal
		percentile decimal.Decimal
		minConf    decimal.Decimal
	}{
		{
			name:       "High z-score",
			zScore:     decimal.NewFromFloat(2.5),
			percentile: decimal.NewFromInt(50),
			minConf:    decimal.NewFromInt(80),
		},
		{
			name:       "Medium z-score",
			zScore:     decimal.NewFromFloat(1.6),
			percentile: decimal.NewFromInt(50),
			minConf:    decimal.NewFromInt(60),
		},
		{
			name:       "Low z-score",
			zScore:     decimal.NewFromFloat(0.5),
			percentile: decimal.NewFromInt(50),
			minConf:    decimal.NewFromInt(40),
		},
		{
			name:       "Extreme percentile",
			zScore:     decimal.NewFromFloat(1.6),
			percentile: decimal.NewFromInt(85),
			minConf:    decimal.NewFromInt(70),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := calculateConfidence(tt.zScore, tt.percentile)
			assert.True(t, confidence.GreaterThanOrEqual(tt.minConf))
		})
	}
}

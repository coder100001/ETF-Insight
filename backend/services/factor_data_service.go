package services

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"etf-insight/models"
	"etf-insight/services/factor"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var (
	ErrInvalidFactor = errors.New("invalid factor name")
)

var validFactors = map[string]bool{
	"Mkt-RF": true,
	"SMB":    true,
	"HML":    true,
	"RMW":    true,
	"CMA":    true,
}

type FactorDataService struct {
	db   *gorm.DB
	mu   sync.Mutex
	seed bool
}

func NewFactorDataService(db *gorm.DB) *FactorDataService {
	return &FactorDataService{db: db}
}

func (s *FactorDataService) CreateFactorData(data *models.FactorData) error {
	return s.db.Create(data).Error
}

func (s *FactorDataService) GetFactorData(factorName string, startDate, endDate time.Time) ([]models.FactorData, error) {
	var data []models.FactorData
	err := s.db.Where("factor_name = ? AND date >= ? AND date <= ?", factorName, startDate, endDate).
		Order("date ASC").
		Find(&data).Error
	return data, err
}

func (s *FactorDataService) GetLatestFactorData(factorName string) (*models.FactorData, error) {
	var data models.FactorData
	err := s.db.Where("factor_name = ?", factorName).
		Order("date DESC").
		First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (s *FactorDataService) BatchCreateFactorData(data []models.FactorData) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, d := range data {
			if err := tx.Create(&d).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *FactorDataService) SeedSampleFactorData(days int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.seed {
		return nil
	}

	// Check if real data already exists
	var count int64
	s.db.Model(&models.FactorData{}).Where("data_source = ?", "kenneth_french").Count(&count)
	if count > 0 {
		s.seed = true
		return nil
	}

	// Try to load real data from Kenneth French Data Library
	if err := s.loadRealFactorData(); err == nil {
		s.seed = true
		return nil
	} else {
		// Log warning but fall through to synthetic data
		fmt.Printf("Warning: failed to load real factor data, using synthetic: %v\n", err)
	}

	// Fall back to synthetic data
	return s.generateSyntheticData(days)
}

// loadRealFactorData attempts to load real Fama-French factor data from the Kenneth French Data Library.
func (s *FactorDataService) loadRealFactorData() error {
	endDate := time.Now()
	startDate := endDate.AddDate(-5, 0, 0) // 5 years of data

	rows, err := factor.LoadFactorDataFromFrench(startDate, endDate, "monthly", false)
	if err != nil {
		return fmt.Errorf("failed to download from Kenneth French: %w", err)
	}

	if len(rows) == 0 {
		return fmt.Errorf("no data returned from Kenneth French for date range %s to %s", startDate.Format("2006-01"), endDate.Format("2006-01"))
	}

	// Convert and store each factor
	factors := []string{"Mkt-RF", "SMB", "HML"}
	for _, factorName := range factors {
		factorData := factor.FrenchRowsToFactorData(rows, factorName)
		if len(factorData) == 0 {
			continue
		}

		// Check if this factor already has real data
		var existingCount int64
		s.db.Model(&models.FactorData{}).Where("factor_name = ? AND data_source = ?", factorName, "kenneth_french").Count(&existingCount)
		if existingCount > 0 {
			continue
		}

		if err := s.db.CreateInBatches(factorData, 100).Error; err != nil {
			return fmt.Errorf("failed to store %s data: %w", factorName, err)
		}
	}

	return nil
}

// generateSyntheticData generates synthetic factor data as a fallback.
func (s *FactorDataService) generateSyntheticData(days int) error {
	rand.Seed(time.Now().UnixNano())

	factors := []string{"Mkt-RF", "SMB", "HML", "RMW", "CMA"}
	baseValues := map[string]float64{
		"Mkt-RF": 0.005,
		"SMB":    0.002,
		"HML":    0.003,
		"RMW":    0.0025,
		"CMA":    0.002,
	}
	volatilities := map[string]float64{
		"Mkt-RF": 0.045,
		"SMB":    0.030,
		"HML":    0.030,
		"RMW":    0.025,
		"CMA":    0.025,
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	for _, factorName := range factors {
		var count int64
		s.db.Model(&models.FactorData{}).Where("factor_name = ?", factorName).Count(&count)
		if count > 0 {
			continue
		}

		baseValue := baseValues[factorName]
		volatility := volatilities[factorName]

		var dataToInsert []models.FactorData
		currentDate := startDate
		for currentDate.Before(endDate) || currentDate.Equal(endDate) {
			if currentDate.Weekday() != time.Saturday && currentDate.Weekday() != time.Sunday {
				value := baseValue + randNorm()*volatility
				dataToInsert = append(dataToInsert, models.FactorData{
					FactorName: factorName,
					Date:       currentDate,
					Value:      decimal.NewFromFloat(value),
					DataSource: "sample",
					CreatedAt:  time.Now(),
				})
			}
			currentDate = currentDate.AddDate(0, 0, 1)
		}

		if len(dataToInsert) > 0 {
			if err := s.db.CreateInBatches(dataToInsert, 100).Error; err != nil {
				return err
			}
		}
	}

	s.seed = true
	return nil
}

func (s *FactorDataService) GetFactorCount(factorName string) (int64, error) {
	var count int64
	err := s.db.Model(&models.FactorData{}).Where("factor_name = ?", factorName).Count(&count).Error
	return count, err
}

func randNorm() float64 {
	sum := 0.0
	for range 12 {
		sum += rand.Float64()
	}
	return sum - 6.0
}

func (s *FactorDataService) CalculateTimingSignal(factorName string, lookbackDays int) (*models.FactorTimingSignal, error) {
	if lookbackDays <= 0 {
		return nil, errors.New("lookbackDays must be positive")
	}

	if !validFactors[factorName] {
		return nil, ErrInvalidFactor
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -lookbackDays*2)

	data, err := s.GetFactorData(factorName, startDate, endDate)
	if err != nil {
		return nil, err
	}

	if len(data) < lookbackDays {
		count, err := s.GetFactorCount(factorName)
		if err != nil {
			return nil, fmt.Errorf("failed to check factor count: %w", err)
		}
		if count == 0 {
			if err := s.SeedSampleFactorData(lookbackDays * 3); err != nil {
				return nil, fmt.Errorf("failed to seed sample data: %w", err)
			}
			data, err = s.GetFactorData(factorName, startDate, endDate)
			if err != nil {
				return nil, err
			}
		}

		if len(data) < lookbackDays {
			return nil, fmt.Errorf("insufficient data for calculation: need %d, got %d", lookbackDays, len(data))
		}
	}

	values := make([]decimal.Decimal, len(data))
	for i, d := range data {
		values[i] = d.Value
	}

	maSlope := calculateMASlope(values, 60)
	zScore := calculateZScore(values)
	percentile := calculatePercentile(values)

	signalStrength := determineSignalStrength(zScore)
	signalScore := signalStrength.ToScore()

	expectedReturn := calculateExpectedReturn(maSlope, zScore)
	confidence := calculateConfidence(zScore, percentile)

	signal := &models.FactorTimingSignal{
		FactorName:     factorName,
		SignalDate:     endDate,
		MASlope60:      maSlope,
		ZScore:         zScore,
		Percentile:     percentile,
		SignalStrength: signalStrength,
		SignalScore:    signalScore,
		ExpectedReturn: expectedReturn,
		Confidence:     confidence,
		CreatedAt:      time.Now(),
	}

	return signal, nil
}

func (s *FactorDataService) CreateTimingSignal(signal *models.FactorTimingSignal) error {
	return s.db.Create(signal).Error
}

func (s *FactorDataService) GetTimingSignals(factorName string, startDate, endDate time.Time) ([]models.FactorTimingSignal, error) {
	var signals []models.FactorTimingSignal
	err := s.db.Where("factor_name = ? AND signal_date >= ? AND signal_date <= ?", factorName, startDate, endDate).
		Order("signal_date DESC").
		Find(&signals).Error
	return signals, err
}

func (s *FactorDataService) GetLatestTimingSignal(factorName string) (*models.FactorTimingSignal, error) {
	var signal models.FactorTimingSignal
	err := s.db.Where("factor_name = ?", factorName).
		Order("signal_date DESC").
		First(&signal).Error
	if err != nil {
		return nil, err
	}
	return &signal, nil
}

func calculateMASlope(values []decimal.Decimal, period int) decimal.Decimal {
	if len(values) < period {
		return decimal.Zero
	}

	recent := values[len(values)-period:]
	sum := decimal.Zero
	for _, v := range recent {
		sum = sum.Add(v)
	}
	ma := sum.Div(decimal.NewFromInt(int64(period)))

	if len(values) < period+1 {
		return decimal.Zero
	}

	previous := values[len(values)-period-1 : len(values)-1]
	prevSum := decimal.Zero
	for _, v := range previous {
		prevSum = prevSum.Add(v)
	}
	prevMA := prevSum.Div(decimal.NewFromInt(int64(period)))

	slope := ma.Sub(prevMA)
	return slope
}

func calculateZScore(values []decimal.Decimal) decimal.Decimal {
	if len(values) == 0 {
		return decimal.Zero
	}

	current := values[len(values)-1]

	sum := decimal.Zero
	for _, v := range values {
		sum = sum.Add(v)
	}
	mean := sum.Div(decimal.NewFromInt(int64(len(values))))

	variance := decimal.Zero
	for _, v := range values {
		diff := v.Sub(mean)
		variance = variance.Add(diff.Mul(diff))
	}
	variance = variance.Div(decimal.NewFromInt(int64(len(values))))
	stdDev := decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64()))

	if stdDev.IsZero() {
		return decimal.Zero
	}

	zScore := current.Sub(mean).Div(stdDev)
	return zScore
}

func calculatePercentile(values []decimal.Decimal) decimal.Decimal {
	if len(values) == 0 {
		return decimal.Zero
	}

	current := values[len(values)-1]
	count := 0
	for _, v := range values {
		if v.LessThan(current) {
			count++
		}
	}

	percentile := decimal.NewFromInt(int64(count * 100 / len(values)))
	return percentile
}

func determineSignalStrength(zScore decimal.Decimal) models.SignalStrength {
	absZScore := zScore.Abs()

	switch {
	case absZScore.GreaterThanOrEqual(decimal.NewFromFloat(2.0)):
		if zScore.GreaterThan(decimal.Zero) {
			return models.SignalStrengthStrongPositive
		}
		return models.SignalStrengthStrongNegative
	case absZScore.GreaterThanOrEqual(decimal.NewFromFloat(1.5)):
		if zScore.GreaterThan(decimal.Zero) {
			return models.SignalStrengthWeakPositive
		}
		return models.SignalStrengthWeakNegative
	default:
		return models.SignalStrengthNeutral
	}
}

func calculateExpectedReturn(maSlope, zScore decimal.Decimal) decimal.Decimal {
	return maSlope.Add(zScore.Mul(decimal.NewFromFloat(0.001)))
}

func calculateConfidence(zScore, percentile decimal.Decimal) decimal.Decimal {
	absZScore := zScore.Abs()

	var confidence decimal.Decimal
	switch {
	case absZScore.GreaterThanOrEqual(decimal.NewFromFloat(2.0)):
		confidence = decimal.NewFromFloat(80.0)
	case absZScore.GreaterThanOrEqual(decimal.NewFromFloat(1.5)):
		confidence = decimal.NewFromFloat(60.0)
	default:
		confidence = decimal.NewFromFloat(40.0)
	}

	if percentile.GreaterThan(decimal.NewFromInt(80)) || percentile.LessThan(decimal.NewFromInt(20)) {
		confidence = confidence.Add(decimal.NewFromInt(10))
	}

	if confidence.GreaterThan(decimal.NewFromInt(100)) {
		confidence = decimal.NewFromInt(100)
	}

	return confidence
}

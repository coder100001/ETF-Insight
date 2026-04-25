package services

import (
	"errors"
	"math"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var (
	ErrInvalidFactor = errors.New("invalid factor name")
)

type FactorDataService struct {
	db *gorm.DB
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

func (s *FactorDataService) CalculateTimingSignal(factorName string, lookbackDays int) (*models.FactorTimingSignal, error) {
	if lookbackDays <= 0 {
		return nil, errors.New("lookbackDays must be positive")
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -lookbackDays*2)

	data, err := s.GetFactorData(factorName, startDate, endDate)
	if err != nil {
		return nil, err
	}

	if len(data) < lookbackDays {
		return nil, errors.New("insufficient data for calculation")
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

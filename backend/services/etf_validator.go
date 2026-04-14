package services

import (
	"fmt"
	"regexp"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
)

type ValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
}

type ETFValidator struct {
	maxPrice    decimal.Decimal
	minPrice    decimal.Decimal
	maxVolume   int64
	dateFormats []string
}

func NewETFValidator() *ETFValidator {
	return &ETFValidator{
		maxPrice:    decimal.NewFromFloat(1000000),
		minPrice:    decimal.NewFromFloat(0.0001),
		maxVolume:   1_000_000_000_000,
		dateFormats: []string{"2006-01-02", "2006-01-02T15:04:05Z", "2006-01-02 15:04:05"},
	}
}

func (v *ETFValidator) ValidateSymbol(symbol string) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}

	if symbol == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "symbol cannot be empty")
		return result
	}

	if len(symbol) > 10 {
		result.Valid = false
		result.Errors = append(result.Errors, "symbol length must not exceed 10 characters")
	}

	if !regexp.MustCompile(`^[A-Za-z0-9\-\.]+$`).MatchString(symbol) {
		result.Valid = false
		result.Errors = append(result.Errors, "symbol contains invalid characters")
	}

	return result
}

func (v *ETFValidator) ValidatePrice(price decimal.Decimal, fieldName string) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}

	if price.IsZero() {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s cannot be zero", fieldName))
		return result
	}

	if price.LessThan(v.minPrice) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s is below minimum allowed value: %s", fieldName, v.minPrice.String()))
	}

	if price.GreaterThan(v.maxPrice) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s exceeds maximum allowed value: %s", fieldName, v.maxPrice.String()))
	}

	if price.IsNegative() && fieldName != "change" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("%s is negative: %s", fieldName, price.String()))
	}

	return result
}

func (v *ETFValidator) ValidateVolume(volume int64) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}

	if volume < 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "volume cannot be negative")
		return result
	}

	if volume > v.maxVolume {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("volume exceeds maximum allowed value: %d", v.maxVolume))
	}

	if volume == 0 {
		result.Warnings = append(result.Warnings, "volume is zero, possible trading halt")
	}

	return result
}

func (v *ETFValidator) ValidateDate(date time.Time, fieldName string) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}

	now := time.Now()

	if date.After(now) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("%s cannot be in the future", fieldName))
	}

	tenYearsAgo := now.AddDate(-10, 0, 0)
	if date.Before(tenYearsAgo) {
		result.Warnings = append(result.Warnings, fmt.Sprintf("%s is more than 10 years in the past, data may be stale", fieldName))
	}

	return result
}

func (v *ETFValidator) ValidateETFData(data *models.ETFData) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}

	if data == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "ETF data cannot be nil")
		return result
	}

	symbolResult := v.ValidateSymbol(data.Symbol)
	result.merge(symbolResult)

	priceResult := v.ValidatePrice(data.ClosePrice, "close_price")
	result.merge(priceResult)

	if !data.OpenPrice.IsZero() {
		openResult := v.ValidatePrice(data.OpenPrice, "open_price")
		result.merge(openResult)
	}

	if !data.HighPrice.IsZero() {
		highResult := v.ValidatePrice(data.HighPrice, "high_price")
		result.merge(highResult)

		if data.HighPrice.LessThan(data.LowPrice) {
			result.Valid = false
			result.Errors = append(result.Errors, "high_price cannot be less than low_price")
		}
	}

	if !data.LowPrice.IsZero() {
		lowResult := v.ValidatePrice(data.LowPrice, "low_price")
		result.merge(lowResult)
	}

	volumeResult := v.ValidateVolume(data.Volume)
	result.merge(volumeResult)

	dateResult := v.ValidateDate(data.Date, "date")
	result.merge(dateResult)

	if data.ClosePrice.GreaterThan(decimal.Zero) && data.OpenPrice.GreaterThan(decimal.Zero) {
		dailyChange := data.ClosePrice.Sub(data.OpenPrice).Abs()
		changePercent := dailyChange.Div(data.OpenPrice).Mul(decimal.NewFromInt(100))
		if changePercent.GreaterThan(decimal.NewFromFloat(50)) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("unusual daily change detected: %s%%", changePercent.StringFixed(2)))
		}
	}

	return result
}

func (v *ETFValidator) ValidateETFConfig(cfg *models.ETFConfig) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}

	if cfg == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "ETF config cannot be nil")
		return result
	}

	symbolResult := v.ValidateSymbol(cfg.Symbol)
	result.merge(symbolResult)

	if cfg.Name == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "name cannot be empty")
	}

	if len(cfg.Name) > 200 {
		result.Valid = false
		result.Errors = append(result.Errors, "name length must not exceed 200 characters")
	}

	if cfg.Category != "" {
		validCategories := map[string]bool{
			"US": true, "CN": true, "HK": true, "JP": true, "EU": true,
			" Bond": true, "Tech": true, "Finance": true, "Energy": true,
			"Healthcare": true, "Consumer": true, "Industrial": true,
			"Real Estate": true, "Materials": true, "Utilities": true,
		}
		if !validCategories[cfg.Category] {
			result.Warnings = append(result.Warnings, fmt.Sprintf("unusual category: %s", cfg.Category))
		}
	}

	return result
}

func (v *ETFValidator) ValidateETFDataBatch(dataList []models.ETFData) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}

	if len(dataList) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "data list is empty")
		return result
	}

	if len(dataList) < 2 {
		result.Warnings = append(result.Warnings, "data list has only one record, insufficient for time series analysis")
	}

	symbol := dataList[0].Symbol
	for i, data := range dataList {
		dataResult := v.ValidateETFData(&data)
		if !dataResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("record %d: %v", i, dataResult.Errors))
		}
		result.Warnings = append(result.Warnings, dataResult.Warnings...)

		if data.Symbol != symbol {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("record %d: symbol mismatch, expected %s, got %s", i, symbol, data.Symbol))
		}
	}

	return result
}

func (v *ETFValidator) ValidatePriceRange(prices []models.ETFData) ValidationResult {
	result := ValidationResult{Valid: true, Errors: []string{}, Warnings: []string{}}

	if len(prices) < 2 {
		result.Warnings = append(result.Warnings, "insufficient data for price range validation")
		return result
	}

	sortedPrices := make([]models.ETFData, len(prices))
	copy(sortedPrices, prices)

	highest := sortedPrices[0].HighPrice
	lowest := sortedPrices[0].LowPrice

	for _, p := range sortedPrices {
		if p.HighPrice.GreaterThan(highest) {
			highest = p.HighPrice
		}
		if p.LowPrice.LessThan(lowest) {
			lowest = p.LowPrice
		}
	}

	if lowest.GreaterThan(decimal.Zero) {
		priceRange := highest.Sub(lowest).Div(lowest).Mul(decimal.NewFromInt(100))
		if priceRange.GreaterThan(decimal.NewFromFloat(500)) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("unusual price range: %s%% (highest/lowest)", priceRange.StringFixed(2)))
		}
	}

	return result
}

func (r *ValidationResult) merge(other ValidationResult) {
	if !other.Valid {
		r.Valid = false
	}
	r.Errors = append(r.Errors, other.Errors...)
	r.Warnings = append(r.Warnings, other.Warnings...)
}

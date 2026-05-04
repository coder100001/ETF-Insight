package models

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func ValidateEuropeanOptionRequest(req EuropeanOptionRequest) error {
	if req.Spot.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("spot must be greater than 0, got %s", req.Spot.String())
	}
	if req.Strike.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("strike must be greater than 0, got %s", req.Strike.String())
	}
	if req.Volatility.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("volatility must be greater than 0, got %s", req.Volatility.String())
	}
	if req.TimeToExpiry.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("time_to_expiry must be greater than 0, got %s", req.TimeToExpiry.String())
	}
	if !req.OptionType.IsValid() {
		return fmt.Errorf("invalid option_type: %s, must be 'call' or 'put'", req.OptionType)
	}
	if req.DividendYield.IsNegative() {
		return fmt.Errorf("dividend_yield cannot be negative, got %s", req.DividendYield.String())
	}
	return nil
}

func ValidateAmericanOptionRequest(req AmericanOptionRequest) error {
	if req.Spot.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("spot must be greater than 0, got %s", req.Spot.String())
	}
	if req.Strike.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("strike must be greater than 0, got %s", req.Strike.String())
	}
	if req.Volatility.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("volatility must be greater than 0, got %s", req.Volatility.String())
	}
	if req.TimeToExpiry.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("time_to_expiry must be greater than 0, got %s", req.TimeToExpiry.String())
	}
	if !req.OptionType.IsValid() {
		return fmt.Errorf("invalid option_type: %s, must be 'call' or 'put'", req.OptionType)
	}
	if req.Steps < 10 || req.Steps > 1000 {
		return fmt.Errorf("steps must be between 10 and 1000, got %d", req.Steps)
	}
	if req.DividendYield.IsNegative() {
		return fmt.Errorf("dividend_yield cannot be negative, got %s", req.DividendYield.String())
	}
	return nil
}

func ValidateGreeksRequest(req GreeksRequest) error {
	if req.Spot.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("spot must be greater than 0, got %s", req.Spot.String())
	}
	if req.Strike.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("strike must be greater than 0, got %s", req.Strike.String())
	}
	if req.Volatility.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("volatility must be greater than 0, got %s", req.Volatility.String())
	}
	if req.TimeToExpiry.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("time_to_expiry must be greater than 0, got %s", req.TimeToExpiry.String())
	}
	if !req.OptionType.IsValid() {
		return fmt.Errorf("invalid option_type: %s, must be 'call' or 'put'", req.OptionType)
	}
	return nil
}

func ValidateYieldCurveRequest(req YieldCurveRequest) error {
	if len(req.Tenors) == 0 {
		return fmt.Errorf("tenors cannot be empty")
	}
	if len(req.Rates) == 0 {
		return fmt.Errorf("rates cannot be empty")
	}
	if len(req.Tenors) != len(req.Rates) {
		return fmt.Errorf("tenors and rates must have the same length, got %d and %d", len(req.Tenors), len(req.Rates))
	}
	for i, rate := range req.Rates {
		if rate.IsNegative() {
			return fmt.Errorf("rate at index %d cannot be negative, got %s", i, rate.String())
		}
	}
	return nil
}

func ValidateBondRequest(req BondRequest) error {
	if req.FaceValue.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("face_value must be greater than 0, got %s", req.FaceValue.String())
	}
	if req.CouponRate.IsNegative() {
		return fmt.Errorf("coupon_rate cannot be negative, got %s", req.CouponRate.String())
	}
	if req.Frequency < 1 || req.Frequency > 12 {
		return fmt.Errorf("frequency must be between 1 and 12, got %d", req.Frequency)
	}
	if req.YieldToMaturity.IsNegative() {
		return fmt.Errorf("yield_to_maturity cannot be negative, got %s", req.YieldToMaturity.String())
	}
	if req.Maturity == "" {
		return fmt.Errorf("maturity is required")
	}
	return nil
}

func ValidateVaRRequest(req VaRRequest) error {
	if req.PortfolioValue.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("portfolio_value must be greater than 0, got %s", req.PortfolioValue.String())
	}
	if len(req.Returns) < 2 {
		return fmt.Errorf("returns must have at least 2 elements for VaR calculation, got %d", len(req.Returns))
	}
	if req.Confidence.LessThanOrEqual(decimal.Zero) || req.Confidence.GreaterThanOrEqual(decimal.NewFromInt(1)) {
		return fmt.Errorf("confidence must be between 0 and 1 (exclusive), got %s", req.Confidence.String())
	}
	if req.HoldingPeriod < 1 {
		return fmt.Errorf("holding_period must be at least 1, got %d", req.HoldingPeriod)
	}
	validMethods := map[string]bool{"historical": true, "parametric": true, "monte_carlo": true}
	if !validMethods[req.Method] {
		return fmt.Errorf("invalid method: %s, must be 'historical', 'parametric', or 'monte_carlo'", req.Method)
	}
	return nil
}

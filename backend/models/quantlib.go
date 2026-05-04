package models

import "github.com/shopspring/decimal"

type OptionType string

const (
	OptionTypeCall OptionType = "call"
	OptionTypePut  OptionType = "put"
)

func (o OptionType) IsValid() bool {
	return o == OptionTypeCall || o == OptionTypePut
}

type ExerciseStyle string

const (
	ExerciseStyleEuropean ExerciseStyle = "european"
	ExerciseStyleAmerican ExerciseStyle = "american"
)

func (e ExerciseStyle) IsValid() bool {
	return e == ExerciseStyleEuropean || e == ExerciseStyleAmerican
}

type EuropeanOptionRequest struct {
	Spot          decimal.Decimal `json:"spot" binding:"required"`
	Strike        decimal.Decimal `json:"strike" binding:"required"`
	Rate          decimal.Decimal `json:"rate" binding:"required"`
	Volatility    decimal.Decimal `json:"volatility" binding:"required"`
	TimeToExpiry  decimal.Decimal `json:"time_to_expiry" binding:"required"`
	OptionType    OptionType      `json:"option_type" binding:"required"`
	DividendYield decimal.Decimal `json:"dividend_yield"`
}

type AmericanOptionRequest struct {
	Spot          decimal.Decimal `json:"spot" binding:"required"`
	Strike        decimal.Decimal `json:"strike" binding:"required"`
	Rate          decimal.Decimal `json:"rate" binding:"required"`
	Volatility    decimal.Decimal `json:"volatility" binding:"required"`
	TimeToExpiry  decimal.Decimal `json:"time_to_expiry" binding:"required"`
	OptionType    OptionType      `json:"option_type" binding:"required"`
	Steps         int             `json:"steps" binding:"required"`
	DividendYield decimal.Decimal `json:"dividend_yield"`
}

type OptionResult struct {
	Price decimal.Decimal `json:"price"`
	Delta decimal.Decimal `json:"delta"`
	Gamma decimal.Decimal `json:"gamma"`
	Theta decimal.Decimal `json:"theta"`
	Vega  decimal.Decimal `json:"vega"`
	Rho   decimal.Decimal `json:"rho"`
}

type GreeksRequest struct {
	Spot         decimal.Decimal `json:"spot" binding:"required"`
	Strike       decimal.Decimal `json:"strike" binding:"required"`
	Rate         decimal.Decimal `json:"rate" binding:"required"`
	Volatility   decimal.Decimal `json:"volatility" binding:"required"`
	TimeToExpiry decimal.Decimal `json:"time_to_expiry" binding:"required"`
	OptionType   OptionType      `json:"option_type" binding:"required"`
}

type YieldCurveRequest struct {
	Currency    string            `json:"currency" binding:"required"`
	Calendar    string            `json:"calendar" binding:"required"`
	DayCount    string            `json:"day_count" binding:"required"`
	Tenors      []string          `json:"tenors" binding:"required"`
	Rates       []decimal.Decimal `json:"rates" binding:"required"`
	Compounding string            `json:"compounding"`
	Frequency   string            `json:"frequency"`
}

type YieldCurveResult struct {
	Currency        string            `json:"currency"`
	Tenors          []string          `json:"tenors"`
	Rates           []decimal.Decimal `json:"rates"`
	ZeroRates       []decimal.Decimal `json:"zero_rates"`
	ForwardRates    []decimal.Decimal `json:"forward_rates"`
	DiscountFactors []decimal.Decimal `json:"discount_factors"`
}

type BondRequest struct {
	FaceValue       decimal.Decimal `json:"face_value" binding:"required"`
	CouponRate      decimal.Decimal `json:"coupon_rate" binding:"required"`
	Frequency       int             `json:"frequency" binding:"required"`
	Maturity        string          `json:"maturity" binding:"required"`
	YieldToMaturity decimal.Decimal `json:"yield_to_maturity" binding:"required"`
	SettlementDate  string          `json:"settlement_date"`
	DayCount        string          `json:"day_count"`
}

type BondResult struct {
	DirtyPrice       decimal.Decimal `json:"dirty_price"`
	CleanPrice       decimal.Decimal `json:"clean_price"`
	Duration         decimal.Decimal `json:"duration"`
	ModifiedDuration decimal.Decimal `json:"modified_duration"`
	Convexity        decimal.Decimal `json:"convexity"`
	YieldToMaturity  decimal.Decimal `json:"yield_to_maturity"`
	AccruedInterest  decimal.Decimal `json:"accrued_interest"`
}

type VaRRequest struct {
	PortfolioValue decimal.Decimal   `json:"portfolio_value" binding:"required"`
	Returns        []decimal.Decimal `json:"returns" binding:"required"`
	Confidence     decimal.Decimal   `json:"confidence" binding:"required"`
	HoldingPeriod  int               `json:"holding_period" binding:"required"`
	Method         string            `json:"method" binding:"required"`
}

type VaRResult struct {
	VaR           decimal.Decimal `json:"var"`
	CVaR          decimal.Decimal `json:"cvar"`
	Confidence    decimal.Decimal `json:"confidence"`
	HoldingPeriod int             `json:"holding_period"`
	Method        string          `json:"method"`
}

type QuantLibAPIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

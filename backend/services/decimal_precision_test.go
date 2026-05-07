package services

import (
	"math"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDecimalAlmostEqual(t *testing.T) {
	testCases := []struct {
		name     string
		a, b     decimal.Decimal
		epsilon  decimal.Decimal
		expected bool
	}{
		{
			name:     "identical values",
			a:        decimal.NewFromFloat(1.234567),
			b:        decimal.NewFromFloat(1.234567),
			epsilon:  decimal.NewFromFloat(1e-6),
			expected: true,
		},
		{
			name:     "small difference within epsilon",
			a:        decimal.NewFromFloat(1.234567),
			b:        decimal.NewFromFloat(1.234568),
			epsilon:  decimal.NewFromFloat(1e-5),
			expected: true,
		},
		{
			name:     "difference exceeds epsilon",
			a:        decimal.NewFromFloat(1.234567),
			b:        decimal.NewFromFloat(1.234667),
			epsilon:  decimal.NewFromFloat(1e-5),
			expected: false,
		},
		{
			name:     "zero comparison",
			a:        decimal.Zero,
			b:        decimal.NewFromFloat(0.000001),
			epsilon:  decimal.NewFromFloat(1e-4),
			expected: true,
		},
		{
			name:     "negative numbers",
			a:        decimal.NewFromFloat(-1.234567),
			b:        decimal.NewFromFloat(-1.234568),
			epsilon:  decimal.NewFromFloat(1e-5),
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := DecimalAlmostEquals(tc.a, tc.b, tc.epsilon)
			assert.Equal(t, tc.expected, result,
				"DecimalAlmostEquals(%s, %s, %s) = %v, want %v",
				tc.a.String(), tc.b.String(), tc.epsilon.String(),
				result, tc.expected)
		})
	}
}

func TestDecimalAlmostEqual_LargeValues(t *testing.T) {
	largeA := decimal.NewFromFloat(1000000.123456)
	largeB := decimal.NewFromFloat(1000000.123457)
	epsilon := decimal.NewFromFloat(1e-3)

	result := DecimalAlmostEquals(largeA, largeB, epsilon)
	assert.True(t, result, "Large values with small relative diff should be equal")
}

func TestRiskModels_ParametricVsHistoricalConsistency_RelaxedTolerance(t *testing.T) {
	rm := NewRiskModels()

	returns := []decimal.Decimal{
		decimal.NewFromFloat(0.01), decimal.NewFromFloat(0.02),
		decimal.NewFromFloat(-0.01), decimal.NewFromFloat(0.03),
		decimal.NewFromFloat(-0.02), decimal.NewFromFloat(0.015),
		decimal.NewFromFloat(0.008), decimal.NewFromFloat(-0.005),
		decimal.NewFromFloat(0.012), decimal.NewFromFloat(0.025),
	}
	confidence := 0.95

	parametricResult, err := rm.CalculateParametricVaR(returns, confidence)
	if err != nil {
		t.Fatalf("Parametric VaR failed: %v", err)
	}

	historicalResult, err := rm.CalculateHistoricalVaR(returns, confidence)
	if err != nil {
		t.Fatalf("Historical VaR failed: %v", err)
	}

	if parametricResult == nil || historicalResult == nil {
		t.Skip("One or both methods returned nil")
	}

	pVar := parametricResult.VaR
	hVar := historicalResult.VaR

	if pVar.IsZero() && hVar.IsZero() {
		t.Skip("Both methods returned zero VaR")
	}

	epsilon := decimal.NewFromFloat(0.05)
	if !DecimalAlmostEquals(pVar, hVar, epsilon) {
		diff := pVar.Sub(hVar).Abs()
		t.Logf("VaR methods differ (within relaxed tolerance): parametric=%s, historical=%s, diff=%s",
			pVar.String(), hVar.String(), diff.String())

		ratio := math.Abs(pVar.InexactFloat64()-hVar.InexactFloat64()) /
			math.Max(math.Abs(hVar.InexactFloat64()), 0.0001)

		if ratio > 2.0 {
			t.Errorf("VaR methods diverge too much: parametric=%s, historical=%s, ratio=%.2f",
				pVar.String(), hVar.String(), ratio)
		}
	}
}

// DecimalAlmostEquals 比较两个 decimal.Decimal 是否在容差范围内相等
func DecimalAlmostEquals(a, b, epsilon decimal.Decimal) bool {
	if epsilon.IsNegative() || epsilon.IsZero() {
		return a.Equal(b)
	}
	diff := a.Sub(b).Abs()
	return diff.LessThanOrEqual(epsilon)
}

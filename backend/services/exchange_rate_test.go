package services

import (
	"testing"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func setupExchangeRateTestDB(t *testing.T) {
	// 初始化内存数据库
	err := models.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
}

func TestExchangeRateService_GetRate_SameCurrency(t *testing.T) {
	setupExchangeRateTestDB(t)
	service := NewExchangeRateService()

	rate := service.GetRate("USD", "USD")

	assert.Equal(t, 1.0, rate)
}

func TestExchangeRateService_Convert_SameCurrency(t *testing.T) {
	setupExchangeRateTestDB(t)
	service := NewExchangeRateService()

	amount := decimal.NewFromFloat(100)
	converted := service.Convert(amount, "USD", "USD")

	assert.True(t, amount.Equal(converted))
}

func TestExchangeRateService_Convert_ZeroAmount(t *testing.T) {
	setupExchangeRateTestDB(t)
	service := NewExchangeRateService()

	amount := decimal.Zero
	from := "USD"
	to := "CNY"

	converted := service.Convert(amount, from, to)

	assert.True(t, converted.Equal(decimal.Zero))
}

func TestExchangeRateService_NewExchangeRateService(t *testing.T) {
	service := NewExchangeRateService()

	assert.NotNil(t, service)
	assert.NotNil(t, service.client)
}

func TestExchangeRateService_GetRate_DefaultRates(t *testing.T) {
	setupExchangeRateTestDB(t)
	service := NewExchangeRateService()

	// 测试默认汇率 - 只测试支持的货币对
	testCases := []struct {
		from     string
		to       string
		expected float64
	}{
		{"USD", "CNY", 7.2},
		{"USD", "HKD", 7.8},
		{"CNY", "USD", 0.138889},
		{"CNY", "HKD", 1.083333},
		{"HKD", "USD", 0.128205},
		{"HKD", "CNY", 0.923077},
	}

	for _, tc := range testCases {
		t.Run(tc.from+"_"+tc.to, func(t *testing.T) {
			rate := service.GetRate(tc.from, tc.to)
			// 允许一定的误差
			assert.InDelta(t, tc.expected, rate, 0.01)
		})
	}
}

func TestExchangeRateService_CalculateCrossRate(t *testing.T) {
	setupExchangeRateTestDB(t)
	service := NewExchangeRateService()

	// 测试交叉汇率计算
	tests := []struct {
		name        string
		from        string
		to          string
		shouldBeOne bool
	}{
		{
			name:        "Same currency",
			from:        "USD",
			to:          "USD",
			shouldBeOne: true,
		},
		{
			name:        "Direct rate",
			from:        "USD",
			to:          "CNY",
			shouldBeOne: false,
		},
		{
			name:        "Cross rate through USD",
			from:        "CNY",
			to:          "HKD",
			shouldBeOne: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := service.CalculateCrossRate(tt.from, tt.to)
			if tt.shouldBeOne {
				assert.Equal(t, 1.0, rate)
			} else {
				assert.Greater(t, rate, 0.0)
			}
		})
	}
}

func TestExchangeRateService_GetHistory(t *testing.T) {
	service := NewExchangeRateService()

	history, err := service.GetHistory("USD", "CNY", 30)

	assert.NoError(t, err)
	assert.NotNil(t, history)
}

func TestExchangeRateService_getDefaultRates(t *testing.T) {
	service := NewExchangeRateService()

	rates := service.getDefaultRates()

	assert.NotNil(t, rates)
	assert.Contains(t, rates, "USD")
	assert.Contains(t, rates, "CNY")
	assert.Contains(t, rates, "HKD")

	// 验证USD到CNY的汇率
	assert.Equal(t, 7.2, rates["USD"]["CNY"])
	// 验证CNY到USD的汇率
	assert.InDelta(t, 0.138889, rates["CNY"]["USD"], 0.0001)
}

func TestExchangeRateService_getDefaultRate_Unsupported(t *testing.T) {
	service := NewExchangeRateService()

	// 测试不支持的货币对
	rate := service.getDefaultRate("EUR", "GBP")
	assert.Equal(t, 1.0, rate)
}

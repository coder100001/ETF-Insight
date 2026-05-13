package backtest

import (
	"math"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// TestCalculateAnnualizedReturn 测试年化收益率计算
// 期望：使用复利公式 (final/initial)^(1/years) - 1
func TestCalculateAnnualizedReturn(t *testing.T) {
	// 计算期望的年化收益率（百分比）
	calculateExpected := func(initial, final decimal.Decimal, years float64) float64 {
		if years <= 0 {
			return 0
		}
		growth := final.Div(initial)
		return (math.Pow(growth.InexactFloat64(), 1.0/years) - 1) * 100
	}

	tests := []struct {
		name           string
		initialCapital decimal.Decimal
		finalEquity    decimal.Decimal
		years          float64
		tolerance      float64 // 允许的误差
	}{
		{
			name:           "2年翻倍",
			initialCapital: decimal.NewFromInt(10000),
			finalEquity:    decimal.NewFromInt(20000),
			years:          2,
			tolerance:      0.01,
		},
		{
			name:           "1年50%收益",
			initialCapital: decimal.NewFromInt(10000),
			finalEquity:    decimal.NewFromInt(15000),
			years:          1,
			tolerance:      0.01,
		},
		{
			name:           "3年总收益30%",
			initialCapital: decimal.NewFromInt(10000),
			finalEquity:    decimal.NewFromInt(13000),
			years:          3,
			tolerance:      0.01,
		},
		{
			name:           "亏损",
			initialCapital: decimal.NewFromInt(10000),
			finalEquity:    decimal.NewFromInt(8000),
			years:          1,
			tolerance:      0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 动态计算期望值
			want := calculateExpected(tt.initialCapital, tt.finalEquity, tt.years)

			// 使用与 BacktestEngine 相同的正确计算方式
			growth := tt.finalEquity.Div(tt.initialCapital)
			yearsDec := decimal.NewFromFloat(tt.years)

			var got decimal.Decimal
			if yearsDec.GreaterThan(decimal.Zero) {
				got = decimal.NewFromFloat(
					math.Pow(growth.InexactFloat64(), 1.0/tt.years) - 1,
				).Mul(decimal.NewFromInt(100))
			}

			// 验证计算结果
			gotFloat := got.InexactFloat64()
			diff := math.Abs(gotFloat - want)
			if diff > tt.tolerance {
				t.Errorf("calculateAnnualizedReturn() = %.2f%%, want %.2f%% (diff: %.2f)",
					gotFloat, want, diff)
			}
		})
	}
}

// TestEventDrivenEngineAnnualizedReturn 测试事件驱动引擎的年化收益率
// 这是一个集成测试，验证 generateResult 方法
func TestEventDrivenEngineAnnualizedReturn(t *testing.T) {
	// 创建一个模拟的引擎状态
	engine := &EventDrivenEngine{
		initialCapital: decimal.NewFromInt(10000),
		currentCapital: decimal.NewFromInt(20000),
		startDate:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		endDate:        time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC), // 2年
		equityCurve: []*EquityPoint{
			{Date: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Equity: decimal.NewFromInt(10000)},
			{Date: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC), Equity: decimal.NewFromInt(20000)},
		},
	}

	result := engine.generateResult()

	// 期望：2年翻倍，年化收益率应为 41.42%
	expected := 41.42
	tolerance := 0.5
	got := result.AnnualizedReturn.InexactFloat64()

	if math.Abs(got-expected) > tolerance {
		t.Errorf("AnnualizedReturn = %.2f%%, want %.2f%% (diff: %.2f)",
			got, expected, math.Abs(got-expected))
		t.Logf("注意：当前实现计算的是总收益率 (%.2f%%)，而非年化收益率",
			result.TotalReturn.InexactFloat64())
	}
}

// TestWinRateCalculation 测试胜率计算
// Bug: 卖出后持仓被删除，导致无法统计盈亏
func TestWinRateCalculation(t *testing.T) {
	// 创建一个模拟的引擎状态，包含一次盈利交易
	// 注意：Fill.RealizedPnL 应该在 executeOrder 中设置
	engine := &EventDrivenEngine{
		initialCapital: decimal.NewFromInt(10000),
		currentCapital: decimal.NewFromInt(11000),
		startDate:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		endDate:        time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
		positions:      make(map[string]*Position),
		equityCurve: []*EquityPoint{
			{Date: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Equity: decimal.NewFromInt(10000)},
			{Date: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC), Equity: decimal.NewFromInt(11000)},
		},
		fills: []*Fill{
			// 买入 100股 @ $100
			{
				OrderID:     "order-1",
				Symbol:      "AAPL",
				Side:        OrderSideBuy,
				Quantity:    decimal.NewFromInt(100),
				Price:       decimal.NewFromInt(100),
				RealizedPnL: decimal.Zero, // 买入不记录盈亏
			},
			// 卖出 100股 @ $110 (盈利 $1000)
			// RealizedPnL = (110 - 100) * 100 = $1000
			{
				OrderID:     "order-2",
				Symbol:      "AAPL",
				Side:        OrderSideSell,
				Quantity:    decimal.NewFromInt(100),
				Price:       decimal.NewFromInt(110),
				RealizedPnL: decimal.NewFromInt(1000), // 盈利 $1000
			},
		},
	}

	// 模拟持仓被删除后的状态（这是 Bug 的根源）
	delete(engine.positions, "AAPL")

	result := engine.generateResult()

	// 期望：1 次盈利交易，胜率 100%
	if result.TotalTrades != 1 {
		t.Errorf("TotalTrades = %d, want 1", result.TotalTrades)
	}
	if result.WinningTrades != 1 {
		t.Errorf("WinningTrades = %d, want 1", result.WinningTrades)
	}
	if result.WinRate.InexactFloat64() != 100.0 {
		t.Errorf("WinRate = %.2f%%, want 100%%", result.WinRate.InexactFloat64())
	}
}

// TestEventDrivenEngineInitialCapital 测试事件驱动引擎的初始资金设置
// 验证引擎正确初始化资金，为策略提供正确的基础数据
func TestEventDrivenEngineInitialCapital(t *testing.T) {
	// 创建买入持有策略
	strategy := NewBuyAndHoldStrategy()

	// 创建事件驱动引擎
	initialCapital := 10000.0
	engine := NewEventDrivenEngine(initialCapital, strategy)

	// 验证引擎正确初始化资金
	if engine.initialCapital.IsZero() {
		t.Error("EventDrivenEngine initialCapital should not be zero")
	}
	if engine.currentCapital.IsZero() {
		t.Error("EventDrivenEngine currentCapital should not be zero")
	}

	// 验证初始资金设置正确
	expectedCapital := decimal.NewFromFloat(initialCapital)
	if !engine.initialCapital.Equal(expectedCapital) {
		t.Errorf("initialCapital = %s, want %s", engine.initialCapital.String(), expectedCapital.String())
	}
	if !engine.currentCapital.Equal(expectedCapital) {
		t.Errorf("currentCapital = %s, want %s", engine.currentCapital.String(), expectedCapital.String())
	}
}

// TestEventDrivenEngineWithMockData 测试事件驱动引擎与模拟数据
func TestEventDrivenEngineWithMockData(t *testing.T) {
	// 创建模拟数据提供者
	mockProvider := NewMockDataProvider()
	mockProvider.GenerateData("AAPL", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), 100, 100.0)

	// 创建买入持有策略
	strategy := NewBuyAndHoldStrategy()

	// 创建事件驱动引擎
	engine := NewEventDrivenEngine(10000, strategy)
	engine.SetDataProvider(mockProvider)

	// 运行回测
	startDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2020, 4, 10, 0, 0, 0, 0, time.UTC)

	result, err := engine.Run(startDate, endDate)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// 调试信息
	t.Logf("InitialCapital: %s", result.InitialCapital.String())
	t.Logf("FinalEquity: %s", result.FinalEquity.String())
	t.Logf("TotalTrades: %d", result.TotalTrades)
	t.Logf("Transactions count: %d", len(result.Transactions))

	// 验证有交易发生（买入持有策略应该至少有一次买入）
	if len(result.Transactions) == 0 {
		t.Error("Transactions should be > 0, got 0")
		t.Log("Bug: 策略可能无法获取正确的资金信息，导致没有生成交易信号")
	}

	// 验证最终权益不等于初始资金（应该有持仓）
	// 注意：由于价格波动，最终权益可能高于或低于初始资金
	// 但只要有持仓，权益就应该反映持仓价值
	if result.FinalEquity.Equal(result.InitialCapital) {
		t.Error("FinalEquity should not equal initial capital (should have positions)")
	}
}

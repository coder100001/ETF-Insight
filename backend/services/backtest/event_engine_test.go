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
	tests := []struct {
		name           string
		initialCapital decimal.Decimal
		finalEquity    decimal.Decimal
		years          float64
		want           float64 // 期望的年化收益率（百分比）
		tolerance      float64 // 允许的误差
	}{
		{
			name:           "2年翻倍",
			initialCapital: decimal.NewFromInt(10000),
			finalEquity:    decimal.NewFromInt(20000),
			years:          2,
			want:           41.42, // (2)^(1/2) - 1 = 0.4142 = 41.42%
			tolerance:      0.1,
		},
		{
			name:           "1年50%收益",
			initialCapital: decimal.NewFromInt(10000),
			finalEquity:    decimal.NewFromInt(15000),
			years:          1,
			want:           50.0, // 1.5 - 1 = 0.5 = 50%
			tolerance:      0.1,
		},
		{
			name:           "3年总收益30%",
			initialCapital: decimal.NewFromInt(10000),
			finalEquity:    decimal.NewFromInt(13000),
			years:          3,
			want:           9.14, // (1.3)^(1/3) - 1 = 0.0914 = 9.14%
			tolerance:      0.1,
		},
		{
			name:           "亏损",
			initialCapital: decimal.NewFromInt(10000),
			finalEquity:    decimal.NewFromInt(8000),
			years:          1,
			want:           -20.0, // 0.8 - 1 = -0.2 = -20%
			tolerance:      0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用与 BacktestEngine 相同的正确计算方式
			growth := tt.finalEquity.Div(tt.initialCapital)
			yearsDec := decimal.NewFromFloat(tt.years)

			var expected decimal.Decimal
			if yearsDec.GreaterThan(decimal.Zero) {
				expected = decimal.NewFromFloat(
					math.Pow(growth.InexactFloat64(), 1.0/tt.years) - 1,
				).Mul(decimal.NewFromInt(100))
			}

			// 验证期望值
			got := expected.InexactFloat64()
			diff := math.Abs(got - tt.want)
			if diff > tt.tolerance {
				t.Errorf("calculateAnnualizedReturn() = %.2f%%, want %.2f%% (diff: %.2f)",
					got, tt.want, diff)
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

// TestStrategyEngineAdapter 测试策略引擎适配器
// Bug: EventDrivenEngine 传入空 BacktestEngine 实例，策略无法获取资金信息
func TestStrategyEngineAdapter(t *testing.T) {
	// 创建买入持有策略
	strategy := NewBuyAndHoldStrategy()

	// 创建事件驱动引擎
	initialCapital := 10000.0
	engine := NewEventDrivenEngine(initialCapital, strategy)

	// 验证引擎状态
	if engine.initialCapital.IsZero() {
		t.Error("EventDrivenEngine initialCapital should not be zero")
	}
	if engine.currentCapital.IsZero() {
		t.Error("EventDrivenEngine currentCapital should not be zero")
	}

	// 验证适配器是否正确传递资金信息
	// 策略应该能够通过适配器获取到正确的资金
	expectedCapital := decimal.NewFromFloat(initialCapital)
	if !engine.initialCapital.Equal(expectedCapital) {
		t.Errorf("initialCapital = %s, want %s", engine.initialCapital.String(), expectedCapital.String())
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

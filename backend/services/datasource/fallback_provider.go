package datasource

import (
	"context"
	"math"
	"math/rand/v2"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// MockDataProvider 模拟数据源提供者
// 当所有主要数据源都不可用时使用，返回模拟数据用于开发和测试
type MockDataProvider struct {
	basePrices map[string]float64
}

func NewMockDataProvider() *MockDataProvider {
	provider := &MockDataProvider{
		basePrices: make(map[string]float64),
	}
	provider.loadBasePricesFromDB()
	return provider
}

func (f *MockDataProvider) GetName() string {
	return "mock"
}

func (f *MockDataProvider) IsAvailable(ctx context.Context) bool {
	return true
}

func (f *MockDataProvider) GetRateLimit() int {
	return 1000
}

func (f *MockDataProvider) GetQuote(ctx context.Context, symbol string) (*QuoteData, error) {
	if symbol == "" {
		return nil, ErrInvalidSymbol
	}

	basePrice, ok := f.basePrices[symbol]
	if !ok {
		// 未知 symbol：无基准价则返回错误，不伪造 $100 报价
		return nil, ErrInvalidSymbol
	}

	return f.generateQuote(symbol, basePrice), nil
}

func (f *MockDataProvider) GetQuotes(ctx context.Context, symbols []string) ([]*QuoteData, error) {
	if len(symbols) == 0 {
		return nil, ErrInvalidSymbol
	}

	results := make([]*QuoteData, 0, len(symbols))
	for _, symbol := range symbols {
		quote, err := f.GetQuote(ctx, symbol)
		if err != nil {
			continue
		}
		results = append(results, quote)
	}

	return results, nil
}

// generateQuote 生成更真实的模拟报价数据
// 使用带漂移的几何布朗运动（GBM）风格扰动，日内波动控制在 ±1.5% 以内
// math/rand/v2 顶层函数并发安全，handler 与 scheduler 可同时调用
func (f *MockDataProvider) generateQuote(symbol string, basePrice float64) *QuoteData {
	// 日内波动：使用正态分布近似，标准差约 0.8%，限制在 ±2%
	dayChange := (rand.NormFloat64() * 0.008)
	dayChange = math.Max(-0.02, math.Min(0.02, dayChange))

	closePrice := basePrice * (1 + dayChange)

	// 开盘价围绕前收盘价微幅波动
	openChange := (rand.NormFloat64() * 0.003)
	openPrice := basePrice * (1 + openChange)

	// 最高/最低价基于开盘和收盘的极值加上影线
	highPrice := math.Max(openPrice, closePrice) * (1 + rand.Float64()*0.005)
	lowPrice := math.Min(openPrice, closePrice) * (1 - rand.Float64()*0.005)

	// 成交量：根据价格调整，低价股成交量通常更大
	baseVolume := int64(2000000)
	if basePrice < 50 {
		baseVolume = 5000000
	} else if basePrice > 300 {
		baseVolume = 1000000
	}
	volume := baseVolume + rand.Int64N(baseVolume*5)

	change := closePrice - basePrice
	changePercent := 0.0
	if basePrice > 0 {
		changePercent = (change / basePrice) * 100
	}

	return &QuoteData{
		Symbol:        symbol,
		CurrentPrice:  math.Round(closePrice*100) / 100,
		OpenPrice:     math.Round(openPrice*100) / 100,
		DayHigh:       math.Round(highPrice*100) / 100,
		DayLow:        math.Round(lowPrice*100) / 100,
		PreviousClose: math.Round(basePrice*100) / 100,
		Change:        math.Round(change*100) / 100,
		ChangePercent: math.Round(changePercent*100) / 100,
		Volume:        volume,
		Currency:      "USD",
		Exchange:      "NASDAQ",
		Timestamp:     time.Now(),
		DataSource:    "mock",
	}
}

// GetETFHoldings 获取ETF底层持仓数据
// Mock provider 返回模拟持仓数据，用于开发和测试
func (f *MockDataProvider) GetETFHoldings(ctx context.Context, symbol string, date time.Time) ([]*ETFHoldingData, error) {
	if symbol == "" {
		return nil, ErrInvalidSymbol
	}

	// 返回模拟持仓数据，便于前端开发和测试
	holdings := []*ETFHoldingData{
		{Symbol: "AAPL", Name: "Apple Inc.", Weight: 12.5, Shares: 1000000, MarketValue: 215000000, Date: time.Now()},
		{Symbol: "MSFT", Name: "Microsoft Corp.", Weight: 11.2, Shares: 800000, MarketValue: 304000000, Date: time.Now()},
		{Symbol: "GOOGL", Name: "Alphabet Inc.", Weight: 8.7, Shares: 500000, MarketValue: 82500000, Date: time.Now()},
		{Symbol: "AMZN", Name: "Amazon.com Inc.", Weight: 7.3, Shares: 400000, MarketValue: 76000000, Date: time.Now()},
		{Symbol: "NVDA", Name: "NVIDIA Corp.", Weight: 6.8, Shares: 300000, MarketValue: 36000000, Date: time.Now()},
	}

	return holdings, nil
}

func (f *MockDataProvider) SetBasePrice(symbol string, price float64) {
	f.basePrices[symbol] = price
}

// loadBasePricesFromDB 从数据库加载最新收盘价作为基准价
// 改进：不再限定 data_source = 'finage'，而是取该 symbol 最近的任何有效数据
// 这避免了 mock/generated 数据与 finage 数据之间的价格跳空
func (f *MockDataProvider) loadBasePricesFromDB() {
	defaults := defaultBasePrices()

	// 先尝试从数据库加载所有启用 ETF 的最新价格
	var configs []models.ETFConfig
	if err := models.DB.Where("status = ?", 1).Find(&configs).Error; err != nil {
		f.basePrices = defaults
		return
	}

	for _, cfg := range configs {
		var etfData models.ETFData
		// 查询该 symbol 最新的收盘价，不限数据源
		err := models.DB.Where("symbol = ?", cfg.Symbol).
			Order("date DESC").
			First(&etfData).Error

		if err == nil && etfData.ID > 0 && etfData.ClosePrice.GreaterThan(decimal.Zero) {
			price := etfData.ClosePrice.InexactFloat64()
			// 合理性校验：只拦截"物理不可能"的价格（≤0 或 >10000）
			// 不对比静态锚点——长期上涨/下跌的真实价格不能被误判为异常
			if price > 0 && price <= 10000 {
				f.basePrices[cfg.Symbol] = price
			} else if defaultPrice, ok := defaults[cfg.Symbol]; ok {
				f.basePrices[cfg.Symbol] = defaultPrice
			}
		} else if err == gorm.ErrRecordNotFound {
			// 数据库中完全没有该 symbol 的数据，使用默认价格
			if defaultPrice, ok := defaults[cfg.Symbol]; ok {
				f.basePrices[cfg.Symbol] = defaultPrice
			}
		} else {
			// 查询出错，使用默认价格
			if defaultPrice, ok := defaults[cfg.Symbol]; ok {
				f.basePrices[cfg.Symbol] = defaultPrice
			}
		}
	}

	// 确保所有默认价格都在 map 中（即使数据库没有对应配置）
	for sym, price := range defaults {
		if _, exists := f.basePrices[sym]; !exists {
			f.basePrices[sym] = price
		}
	}
}

// defaultBasePrices 返回所有支持 ETF 的合理基准价格（约 2026 年 8 月近似价格）
// 这些价格仅在无真实 API 数据、无历史数据时使用
func defaultBasePrices() map[string]float64 {
	return map[string]float64{
		// 主流宽基指数
		"QQQ": 650.0, // Invesco QQQ Trust（纳斯达克100）
		"VOO": 655.0, // Vanguard S&P 500
		"VTI": 275.0, // Vanguard Total Stock Market
		"IWM": 220.0, // iShares Russell 2000
		"SPY": 590.0, // SPDR S&P 500
		"DIA": 420.0, // SPDR Dow Jones Industrial Average

		// 股息/收益型
		"SCHD": 31.0,  // Schwab US Dividend Equity
		"JEPQ": 58.0,  // JPMorgan Nasdaq Equity Premium Income
		"JEPI": 58.0,  // JPMorgan Equity Premium Income
		"SPYD": 46.0,  // SPDR Portfolio S&P 500 High Dividend
		"VYM":  130.0, // Vanguard High Dividend Yield
		"DGRO": 60.0,  // iShares Core Dividend Growth
		"HDV":  100.0, // iShares Core High Dividend
		"VIG":  190.0, // Vanguard Dividend Appreciation
		"QYLD": 17.0,  // Global X Nasdaq 100 Covered Call
		"XYLD": 48.0,  // Global X S&P 500 Covered Call
		"PGX":  14.0,  // Invesco Preferred ETF

		// 债券型
		"BND": 72.0, // Vanguard Total Bond Market
		"AGG": 98.0, // iShares Core US Aggregate Bond
		"TLT": 90.0, // iShares 20+ Year Treasury Bond

		// 国际/新兴市场
		"VXUS": 58.0, // Vanguard Total International Stock
		"VEA":  47.0, // Vanguard FTSE Developed Markets
		"VWO":  43.0, // Vanguard FTSE Emerging Markets
		"EFA":  75.0, // iShares MSCI EAFE
		"EEM":  42.0, // iShares MSCI Emerging Markets

		// 其他
		"VNQ": 85.0,  // Vanguard Real Estate
		"GLD": 290.0, // SPDR Gold Shares

		// 个股（用于持仓展示）
		"AAPL":  215.0,
		"MSFT":  380.0,
		"GOOGL": 165.0,
		"AMZN":  185.0,
		"NVDA":  120.0,
	}
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

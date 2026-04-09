package datasource

import (
	"context"
	"math/rand"
	"time"

	"etf-insight/models"
)

type FallbackProvider struct {
	basePrices map[string]float64
	rnd        *rand.Rand
}

func NewFallbackProvider() *FallbackProvider {
	provider := &FallbackProvider{
		basePrices: make(map[string]float64),
		rnd:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	provider.loadBasePricesFromDB()
	return provider
}

func (f *FallbackProvider) GetName() string {
	return "fallback"
}

func (f *FallbackProvider) IsAvailable(ctx context.Context) bool {
	return true
}

func (f *FallbackProvider) GetRateLimit() int {
	return 1000
}

func (f *FallbackProvider) GetQuote(ctx context.Context, symbol string) (*QuoteData, error) {
	if symbol == "" {
		return nil, ErrInvalidSymbol
	}

	basePrice, ok := f.basePrices[symbol]
	if !ok {
		basePrice = 100.0
	}

	return f.generateQuote(symbol, basePrice), nil
}

func (f *FallbackProvider) GetQuotes(ctx context.Context, symbols []string) ([]*QuoteData, error) {
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

func (f *FallbackProvider) generateQuote(symbol string, basePrice float64) *QuoteData {
	previousClose := basePrice

	openChange := (f.rnd.Float64() - 0.5) * 0.02
	openPrice := basePrice * (1 + openChange)

	closeChange := (f.rnd.Float64() - 0.5) * 00.01
	closePrice := basePrice * (1 + closeChange)

	highPrice := max(openPrice, closePrice) * (1 + f.rnd.Float64()*0.005)
	lowPrice := min(openPrice, closePrice) * (1 - f.rnd.Float64()*0.005)

	volume := int64(1000000 + f.rnd.Int63n(49000000))

	change := closePrice - previousClose
	changePercent := 0.0
	if previousClose > 0 {
		changePercent = (change / previousClose) * 100
	}

	return &QuoteData{
		Symbol:        symbol,
		CurrentPrice:  closePrice,
		OpenPrice:     openPrice,
		DayHigh:       highPrice,
		DayLow:        lowPrice,
		PreviousClose: previousClose,
		Change:        change,
		ChangePercent: changePercent,
		Volume:        volume,
		Currency:      "USD",
		Exchange:      "NASDAQ",
		Timestamp:     time.Now(),
		DataSource:    "fallback",
	}
}

func (f *FallbackProvider) SetBasePrice(symbol string, price float64) {
	f.basePrices[symbol] = price
}

func (f *FallbackProvider) loadBasePricesFromDB() {
	var configs []models.ETFConfig
	if err := models.DB.Where("status = ?", 1).Find(&configs).Error; err != nil || len(configs) == 0 {
		f.basePrices = defaultBasePrices()
		return
	}

	for _, cfg := range configs {
		var etfData models.ETFData
		err := models.DB.Where("symbol = ? AND data_source = ?", cfg.Symbol, "finage").
			Order("date DESC").
			First(&etfData).Error

		if err == nil && etfData.ID > 0 {
			f.basePrices[cfg.Symbol] = etfData.ClosePrice.InexactFloat64()
		} else {
			if defaultPrice, ok := defaultBasePrices()[cfg.Symbol]; ok {
				f.basePrices[cfg.Symbol] = defaultPrice
			}
		}
	}
}

func defaultBasePrices() map[string]float64 {
	return map[string]float64{
		"QQQ":   460.0,
		"SCHD":  85.0,
		"VNQ":   85.0,
		"VYM":   130.0,
		"SPYD":  52.0,
		"JEPQ":  58.0,
		"JEPI":  60.0,
		"VTI":   275.0,
		"VXUS":  58.0,
		"BND":   72.0,
		"DGRO":  60.0,
		"HDV":   100.0,
		"VOO":   500.0,
		"VEA":   47.0,
		"VWO":   43.0,
		"PGX":   14.0,
		"QYLD":  17.0,
		"XYLD":  48.0,
		"AGG":   98.0,
		"GLD":   290.0,
		"TLT":   90.0,
		"AAPL":  215.0,
		"MSFT":  380.0,
		"GOOGL": 165.0,
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

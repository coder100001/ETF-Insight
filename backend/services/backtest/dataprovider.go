package backtest

import (
	"fmt"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// DBDataProvider 数据库数据提供者
type DBDataProvider struct {
	db      *gorm.DB
	symbols []string
}

// NewDBDataProvider 创建数据库数据提供者
func NewDBProvider(db *gorm.DB, symbols []string) *DBDataProvider {
	return &DBDataProvider{
		db:      db,
		symbols: symbols,
	}
}

// GetData 获取历史数据
func (p *DBDataProvider) GetData(startDate, endDate time.Time) ([]*Bar, error) {
	var etfData []models.ETFData

	err := p.db.Where("symbol IN ? AND date >= ? AND date <= ?",
		p.symbols, startDate, endDate).
		Order("date ASC").
		Find(&etfData).Error

	if err != nil {
		return nil, fmt.Errorf("查询历史数据失败: %w", err)
	}

	// 获取各ETF的股息率配置
	dividendYields := p.getDividendYields()

	bars := make([]*Bar, len(etfData))
	for i, data := range etfData {
		// 估算每股股息
		// 假设股息按季度支付，每股股息 = 收盘价 * 年化股息率 / 4
		dividend := decimal.Zero
		if yield, ok := dividendYields[data.Symbol]; ok && yield.GreaterThan(decimal.Zero) {
			// 检查是否为股息支付日（简化：假设每季度末支付）
			month := data.Date.Month()
			day := data.Date.Day()
			// 季度末月份的最后几天（3月、6月、9月、12月）
			isDividendMonth := month == 3 || month == 6 || month == 9 || month == 12
			isEndOfMonth := day >= 25 // 假设月末几天支付
			if isDividendMonth && isEndOfMonth {
				dividend = data.ClosePrice.Mul(yield).Div(decimal.NewFromInt(4))
			}
		}

		bars[i] = &Bar{
			Symbol:   data.Symbol,
			Date:     data.Date,
			Open:     data.OpenPrice,
			High:     data.HighPrice,
			Low:      data.LowPrice,
			Close:    data.ClosePrice,
			Volume:   data.Volume,
			Dividend: dividend,
		}
	}

	return bars, nil
}

// getDividendYields 获取各ETF的股息率
func (p *DBDataProvider) getDividendYields() map[string]decimal.Decimal {
	yields := make(map[string]decimal.Decimal)

	var etfConfigs []models.ETFConfig
	if err := p.db.Find(&etfConfigs).Error; err != nil {
		return yields
	}

	for _, config := range etfConfigs {
		// 从ETF配置中获取股息率（如果有）
		// 这里使用预设的股息率数据
		yield := getDividendYieldByCategory(config.Symbol, config.Category)
		if yield.GreaterThan(decimal.Zero) {
			yields[config.Symbol] = yield
		}
	}

	return yields
}

// getDividendYieldByCategory 根据ETF类别获取默认股息率
func getDividendYieldByCategory(symbol, category string) decimal.Decimal {
	// 预设股息率数据（年化）
	dividendYields := map[string]decimal.Decimal{
		// 高股息ETF
		"SCHD": decimal.NewFromFloat(0.035), // Schwab US Dividend Equity
		"VYM":  decimal.NewFromFloat(0.030), // Vanguard High Dividend Yield
		"SPYD": decimal.NewFromFloat(0.040), // SPDR S&P 500 High Dividend
		"HDV":  decimal.NewFromFloat(0.038), // iShares Core High Dividend
		"VIG":  decimal.NewFromFloat(0.020), // Vanguard Dividend Appreciation
		"DGRO": decimal.NewFromFloat(0.025), // iShares Core Dividend Growth
		// 债券ETF
		"BND": decimal.NewFromFloat(0.025), // Vanguard Total Bond Market
		"AGG": decimal.NewFromFloat(0.024), // iShares Core US Aggregate Bond
		"LQD": decimal.NewFromFloat(0.035), // iShares iBoxx Investment Grade
		"HYG": decimal.NewFromFloat(0.050), // iShares iBoxx High Yield
		// REIT ETF
		"VNQ": decimal.NewFromFloat(0.040), // Vanguard Real Estate
		"IYR": decimal.NewFromFloat(0.038), // iShares US Real Estate
		// 其他
		"SPY": decimal.NewFromFloat(0.015), // S&P 500
		"IVV": decimal.NewFromFloat(0.015), // iShares Core S&P 500
		"VOO": decimal.NewFromFloat(0.015), // Vanguard S&P 500
		"QQQ": decimal.NewFromFloat(0.008), // Nasdaq 100
		"IWM": decimal.NewFromFloat(0.012), // Russell 2000
		"EFA": decimal.NewFromFloat(0.022), // iShares MSCI EAFE
		"EEM": decimal.NewFromFloat(0.018), // iShares MSCI Emerging Markets
	}

	if yield, ok := dividendYields[symbol]; ok {
		return yield
	}

	// 根据类别返回默认股息率
	switch category {
	case "股息", "备兑认购":
		return decimal.NewFromFloat(0.035)
	case "债券":
		return decimal.NewFromFloat(0.030)
	case "REIT":
		return decimal.NewFromFloat(0.040)
	default:
		return decimal.Zero
	}
}

// GetSymbols 获取可用标的列表
func (p *DBDataProvider) GetSymbols() []string {
	return p.symbols
}

// MultiSymbolDataProvider 多标的数据提供者
type MultiSymbolDataProvider struct {
	db      *gorm.DB
	symbols []string
	data    map[string][]*Bar // symbol -> bars
}

// NewMultiSymbolDataProvider 创建多标的数据提供者
func NewMultiSymbolDataProvider(db *gorm.DB, symbols []string) *MultiSymbolDataProvider {
	return &MultiSymbolDataProvider{
		db:      db,
		symbols: symbols,
		data:    make(map[string][]*Bar),
	}
}

// LoadData 加载数据
func (p *MultiSymbolDataProvider) LoadData(startDate, endDate time.Time) error {
	// 获取各ETF的股息率配置
	dividendYields := p.getDividendYields()

	for _, symbol := range p.symbols {
		var etfData []models.ETFData

		err := p.db.Where("symbol = ? AND date >= ? AND date <= ?",
			symbol, startDate, endDate).
			Order("date ASC").
			Find(&etfData).Error

		if err != nil {
			return fmt.Errorf("查询 %s 历史数据失败: %w", symbol, err)
		}

		bars := make([]*Bar, len(etfData))
		for i, data := range etfData {
			// 估算每股股息
			dividend := decimal.Zero
			if yield, ok := dividendYields[data.Symbol]; ok && yield.GreaterThan(decimal.Zero) {
				month := data.Date.Month()
				day := data.Date.Day()
				isDividendMonth := month == 3 || month == 6 || month == 9 || month == 12
				isEndOfMonth := day >= 25
				if isDividendMonth && isEndOfMonth {
					dividend = data.ClosePrice.Mul(yield).Div(decimal.NewFromInt(4))
				}
			}

			bars[i] = &Bar{
				Symbol:   data.Symbol,
				Date:     data.Date,
				Open:     data.OpenPrice,
				High:     data.HighPrice,
				Low:      data.LowPrice,
				Close:    data.ClosePrice,
				Volume:   data.Volume,
				Dividend: dividend,
			}
		}

		p.data[symbol] = bars
	}

	return nil
}

// getDividendYields 获取各ETF的股息率
func (p *MultiSymbolDataProvider) getDividendYields() map[string]decimal.Decimal {
	yields := make(map[string]decimal.Decimal)

	var etfConfigs []models.ETFConfig
	if err := p.db.Find(&etfConfigs).Error; err != nil {
		return yields
	}

	for _, config := range etfConfigs {
		yield := getDividendYieldByCategory(config.Symbol, config.Category)
		if yield.GreaterThan(decimal.Zero) {
			yields[config.Symbol] = yield
		}
	}

	return yields
}

// GetData 获取指定标的的数据
func (p *MultiSymbolDataProvider) GetData(symbol string, startDate, endDate time.Time) ([]*Bar, error) {
	bars, ok := p.data[symbol]
	if !ok {
		return nil, fmt.Errorf("未找到标的 %s 的数据", symbol)
	}

	// 过滤日期范围
	filtered := make([]*Bar, 0)
	for _, bar := range bars {
		if (bar.Date.Equal(startDate) || bar.Date.After(startDate)) &&
			(bar.Date.Equal(endDate) || bar.Date.Before(endDate)) {
			filtered = append(filtered, bar)
		}
	}

	return filtered, nil
}

// GetAllData 获取所有数据
func (p *MultiSymbolDataProvider) GetAllData() map[string][]*Bar {
	return p.data
}

// GetSymbols 获取可用标的列表
func (p *MultiSymbolDataProvider) GetSymbols() []string {
	return p.symbols
}

// MockDataProvider 模拟数据提供者 (用于测试)
type MockDataProvider struct {
	data []*Bar
}

// NewMockDataProvider 创建模拟数据提供者
func NewMockDataProvider() *MockDataProvider {
	return &MockDataProvider{
		data: make([]*Bar, 0),
	}
}

// GenerateData 生成模拟数据
func (p *MockDataProvider) GenerateData(symbol string, startDate time.Time, days int, startPrice float64) {
	price := decimal.NewFromFloat(startPrice)
	date := startDate

	for i := range days {
		// 模拟价格波动
		change := decimal.NewFromFloat(0.02 * (float64(i%10) - 5) / 5)
		price = price.Mul(decimal.NewFromInt(1).Add(change))

		open := price.Mul(decimal.NewFromFloat(0.99))
		high := price.Mul(decimal.NewFromFloat(1.01))
		low := price.Mul(decimal.NewFromFloat(0.98))
		close := price

		bar := &Bar{
			Symbol:   symbol,
			Date:     date,
			Open:     open,
			High:     high,
			Low:      low,
			Close:    close,
			Volume:   1000000,
			Dividend: decimal.Zero,
		}

		p.data = append(p.data, bar)
		date = date.AddDate(0, 0, 1)
	}
}

// GetData 获取历史数据
func (p *MockDataProvider) GetData(startDate, endDate time.Time) ([]*Bar, error) {
	filtered := make([]*Bar, 0)
	for _, bar := range p.data {
		if (bar.Date.Equal(startDate) || bar.Date.After(startDate)) &&
			(bar.Date.Equal(endDate) || bar.Date.Before(endDate)) {
			filtered = append(filtered, bar)
		}
	}
	return filtered, nil
}

// GetSymbols 获取可用标的列表
func (p *MockDataProvider) GetSymbols() []string {
	symbols := make(map[string]bool)
	for _, bar := range p.data {
		symbols[bar.Symbol] = true
	}

	result := make([]string, 0, len(symbols))
	for symbol := range symbols {
		result = append(result, symbol)
	}
	return result
}

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

	bars := make([]*Bar, len(etfData))
	for i, data := range etfData {
		bars[i] = &Bar{
			Symbol:   data.Symbol,
			Date:     data.Date,
			Open:     data.OpenPrice,
			High:     data.HighPrice,
			Low:      data.LowPrice,
			Close:    data.ClosePrice,
			Volume:   data.Volume,
			Dividend: decimal.Zero, // 需要从其他数据源获取
		}
	}

	return bars, nil
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
			bars[i] = &Bar{
				Symbol:   data.Symbol,
				Date:     data.Date,
				Open:     data.OpenPrice,
				High:     data.HighPrice,
				Low:      data.LowPrice,
				Close:    data.ClosePrice,
				Volume:   data.Volume,
				Dividend: decimal.Zero,
			}
		}

		p.data[symbol] = bars
	}

	return nil
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

	for i := 0; i < days; i++ {
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

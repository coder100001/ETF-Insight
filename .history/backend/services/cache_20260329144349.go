package services

import (
	"fmt"
	"time"

	"etf-insight/config"
	"etf-insight/utils"

	"github.com/patrickmn/go-cache"
)

// CacheService 缓存服务
type CacheService struct {
	memory *cache.Cache
	cfg    *config.CacheConfig
}

// RealtimeData 实时数据
type RealtimeData struct {
	Symbol           string    `json:"symbol"`
	Name             string    `json:"name"`
	CurrentPrice     float64   `json:"current_price"`
	PreviousClose    float64   `json:"previous_close"`
	OpenPrice        float64   `json:"open_price"`
	DayHigh          float64   `json:"day_high"`
	DayLow           float64   `json:"day_low"`
	Volume           int64     `json:"volume"`
	Change           float64   `json:"change"`
	ChangePercent    float64   `json:"change_percent"`
	MarketCap        int64     `json:"market_cap"`
	DividendYield    float64   `json:"dividend_yield"`
	FiftyTwoWeekHigh float64   `json:"fifty_two_week_high"`
	FiftyTwoWeekLow  float64   `json:"fifty_two_week_low"`
	AverageVolume    int64     `json:"average_volume"`
	Beta             float64   `json:"beta"`
	PERatio          float64   `json:"pe_ratio"`
	Currency         string    `json:"currency"`
	DataSource       string    `json:"data_source"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// MockRealtimeData 模拟实时数据
var MockRealtimeData = map[string]*RealtimeData{
	"QQQ": {
		Symbol:           "QQQ",
		Name:             "Invesco QQQ Trust",
		CurrentPrice:     485.23,
		PreviousClose:    482.15,
		OpenPrice:        483.50,
		DayHigh:          486.95,
		DayLow:           482.10,
		Volume:           25000000,
		Change:           3.08,
		ChangePercent:    0.64,
		MarketCap:        1850000000000,
		DividendYield:    0.55,
		FiftyTwoWeekHigh: 500.12,
		FiftyTwoWeekLow:  360.15,
		AverageVolume:    35000000,
		Beta:             1.05,
		PERatio:          32.5,
		Currency:         "USD",
		DataSource:       "mock",
	},
	"SCHD": {
		Symbol:           "SCHD",
		Name:             "Schwab US Dividend Equity ETF",
		CurrentPrice:     85.45,
		PreviousClose:    85.20,
		OpenPrice:        85.10,
		DayHigh:          85.60,
		DayLow:           84.95,
		Volume:           3500000,
		Change:           0.25,
		ChangePercent:    0.29,
		MarketCap:        32000000000,
		DividendYield:    3.45,
		FiftyTwoWeekHigh: 88.25,
		FiftyTwoWeekLow:  75.30,
		AverageVolume:    4000000,
		Beta:             0.85,
		PERatio:          28.3,
		Currency:         "USD",
		DataSource:       "mock",
	},
	"VNQ": {
		Symbol:           "VNQ",
		Name:             "Vanguard Real Estate ETF",
		CurrentPrice:     115.80,
		PreviousClose:    115.50,
		OpenPrice:        115.30,
		DayHigh:          116.10,
		DayLow:           115.10,
		Volume:           2800000,
		Change:           0.30,
		ChangePercent:    0.26,
		MarketCap:        58000000000,
		DividendYield:    3.85,
		FiftyTwoWeekHigh: 122.40,
		FiftyTwoWeekLow:  105.20,
		AverageVolume:    3200000,
		Beta:             0.75,
		PERatio:          22.1,
		Currency:         "USD",
		DataSource:       "mock",
	},
	"VYM": {
		Symbol:           "VYM",
		Name:             "Vanguard High Dividend Yield ETF",
		CurrentPrice:     82.30,
		PreviousClose:    82.00,
		OpenPrice:        82.10,
		DayHigh:          82.50,
		DayLow:           81.80,
		Volume:           2200000,
		Change:           0.30,
		ChangePercent:    0.37,
		MarketCap:        45000000000,
		DividendYield:    3.65,
		FiftyTwoWeekHigh: 86.50,
		FiftyTwoWeekLow:  74.20,
		AverageVolume:    2800000,
		Beta:             0.80,
		PERatio:          24.5,
		Currency:         "USD",
		DataSource:       "mock",
	},
	"SPYD": {
		Symbol:           "SPYD",
		Name:             "SPDR S&P 500 High Dividend ETF",
		CurrentPrice:     52.60,
		PreviousClose:    52.40,
		OpenPrice:        52.30,
		DayHigh:          52.80,
		DayLow:           52.10,
		Volume:           1800000,
		Change:           0.20,
		ChangePercent:    0.38,
		MarketCap:        12000000000,
		DividendYield:    3.75,
		FiftyTwoWeekHigh: 55.20,
		FiftyTwoWeekLow:  48.50,
		AverageVolume:    2000000,
		Beta:             0.70,
		PERatio:          18.9,
		Currency:         "USD",
		DataSource:       "mock",
	},
	"JEPQ": {
		Symbol:           "JEPQ",
		Name:             "JPMorgan Nasdaq Equity Premium Income ETF",
		CurrentPrice:     58.40,
		PreviousClose:    58.10,
		OpenPrice:        58.20,
		DayHigh:          58.60,
		DayLow:           57.90,
		Volume:           1200000,
		Change:           0.30,
		ChangePercent:    0.52,
		MarketCap:        3500000000,
		DividendYield:    6.25,
		FiftyTwoWeekHigh: 62.80,
		FiftyTwoWeekLow:  52.10,
		AverageVolume:    1500000,
		Beta:             0.90,
		PERatio:          15.2,
		Currency:         "USD",
		DataSource:       "mock",
	},
	"JEPI": {
		Symbol:           "JEPI",
		Name:             "JPMorgan Equity Premium Income ETF",
		CurrentPrice:     42.50,
		PreviousClose:    42.30,
		OpenPrice:        42.40,
		DayHigh:          42.70,
		DayLow:           42.10,
		Volume:           950000,
		Change:           0.20,
		ChangePercent:    0.47,
		MarketCap:        3200000000,
		DividendYield:    6.85,
		FiftyTwoWeekHigh: 45.80,
		FiftyTwoWeekLow:  39.20,
		AverageVolume:    1100000,
		Beta:             0.85,
		PERatio:          14.8,
		Currency:         "USD",
		DataSource:       "mock",
	},
}

// NewCacheService 创建新的缓存服务
func NewCacheService(cacheCfg *config.CacheConfig) *CacheService {
	s := &CacheService{
		memory: cache.New(cacheCfg.RealtimeTTL, 10*time.Minute),
		cfg:    cacheCfg,
	}

	utils.Info("Cache service initialized with memory cache only")
	return s
}

// GetRealtimeData 获取实时数据
func (s *CacheService) GetRealtimeData(symbol string) (*RealtimeData, error) {
	// 从内存缓存获取
	key := fmt.Sprintf("realtime:%s", symbol)
	if data, found := s.memory.Get(key); found {
		if realtimeData, ok := data.(*RealtimeData); ok {
			return realtimeData, nil
		}
	}

	// 如果缓存中没有，使用模拟数据
	if mockData, ok := MockRealtimeData[symbol]; ok {
		return mockData, nil
	}

	return nil, fmt.Errorf("realtime data not found for %s", symbol)
}

// SetRealtimeData 设置实时数据
func (s *CacheService) SetRealtimeData(symbol string, data *RealtimeData) {
	key := fmt.Sprintf("realtime:%s", symbol)
	data.UpdatedAt = time.Now()

	// 写入内存缓存
	s.memory.Set(key, data, s.cfg.RealtimeTTL)
}

// GetHistoricalData 获取历史数据
func (s *CacheService) GetHistoricalData(symbol string, period string) ([]byte, error) {
	key := fmt.Sprintf("historical:%s:%s", symbol, period)

	// 从内存缓存获取
	if data, found := s.memory.Get(key); found {
		if bytes, ok := data.([]byte); ok {
			return bytes, nil
		}
	}

	return nil, fmt.Errorf("historical data not found for %s", symbol)
}

// SetHistoricalData 设置历史数据
func (s *CacheService) SetHistoricalData(symbol string, period string, data []byte) {
	key := fmt.Sprintf("historical:%s:%s", symbol, period)

	// 写入内存缓存
	s.memory.Set(key, data, s.cfg.HistoricalTTL)
}

// GetMetrics 获取指标数据
func (s *CacheService) GetMetrics(symbol string, period string) ([]byte, error) {
	key := fmt.Sprintf("metrics:%s:%s", symbol, period)

	// 从内存缓存获取
	if data, found := s.memory.Get(key); found {
		if bytes, ok := data.([]byte); ok {
			return bytes, nil
		}
	}

	return nil, fmt.Errorf("metrics data not found for %s", symbol)
}

// SetMetrics 设置指标数据
func (s *CacheService) SetMetrics(symbol string, period string, data []byte) {
	key := fmt.Sprintf("metrics:%s:%s", symbol, period)

	// 写入内存缓存
	s.memory.Set(key, data, s.cfg.MetricsTTL)
}

// GetComparison 获取对比数据
func (s *CacheService) GetComparison(symbols []string, period string) ([]byte, error) {
	key := fmt.Sprintf("comparison:%s:%s", symbols[0], period)

	// 从内存缓存获取
	if data, found := s.memory.Get(key); found {
		if bytes, ok := data.([]byte); ok {
			return bytes, nil
		}
	}

	return nil, fmt.Errorf("comparison data not found for %s", symbols[0])
}

// SetComparison 设置对比数据
func (s *CacheService) SetComparison(symbols []string, period string, data []byte) {
	key := fmt.Sprintf("comparison:%s:%s", symbols[0], period)

	// 写入内存缓存
	s.memory.Set(key, data, s.cfg.ComparisonTTL)
}

// GetCacheStats 获取缓存统计信息
func (s *CacheService) GetCacheStats() map[string]interface{} {
	stats := s.memory.Items()
	return map[string]interface{}{
		"cache_size": len(stats),
	}
}

// ClearCache 清除缓存
func (s *CacheService) ClearCache() {
	s.memory.Flush()
}

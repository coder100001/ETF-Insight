package services

import (
	"encoding/json"
	"fmt"
	"time"

	"etf-insight/config"
	"etf-insight/utils"

	"github.com/patrickmn/go-cache"
)

// CacheService 缓存服务
type CacheService struct {
	memory    *cache.Cache
	cfg       *config.CacheConfig
}

// RealtimeData 实时数据
type RealtimeData struct {
	Symbol             string  `json:"symbol"`
	Name               string  `json:"name"`
	CurrentPrice       float64 `json:"current_price"`
	PreviousClose      float64 `json:"previous_close"`
	OpenPrice          float64 `json:"open_price"`
	DayHigh            float64 `json:"day_high"`
	DayLow             float64 `json:"day_low"`
	Volume             int64   `json:"volume"`
	Change             float64 `json:"change"`
	ChangePercent      float64 `json:"change_percent"`
	MarketCap          int64   `json:"market_cap"`
	DividendYield      float64 `json:"dividend_yield"`
	FiftyTwoWeekHigh   float64 `json:"fifty_two_week_high"`
	FiftyTwoWeekLow    float64 `json:"fifty_two_week_low"`
	AverageVolume      int64   `json:"average_volume"`
	Beta               float64 `json:"beta"`
	PERatio            float64 `json:"pe_ratio"`
	Currency           string  `json:"currency"`
	DataSource         string  `json:"data_source"`
	UpdatedAt          time.Time `json:"updated_at"`
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

	return nil, fmt.Errorf("metrics not found for %s", symbol)
}

// SetMetrics 设置指标数据
func (s *CacheService) SetMetrics(symbol string, period string, data []byte) {
	key := fmt.Sprintf("metrics:%s:%s", symbol, period)

	// 写入内存缓存
	s.memory.Set(key, data, s.cfg.MetricsTTL)
}

// GetComparison 获取对比数据
func (s *CacheService) GetComparison(symbols []string, period string) ([]byte, error) {
	key := fmt.Sprintf("comparison:%v:%s", symbols, period)

	// 从内存缓存获取
	if data, found := s.memory.Get(key); found {
		if bytes, ok := data.([]byte); ok {
			return bytes, nil
		}
	}

	return nil, fmt.Errorf("comparison data not found")
}

// SetComparison 设置对比数据
func (s *CacheService) SetComparison(symbols []string, period string, data []byte) {
	key := fmt.Sprintf("comparison:%v:%s", symbols, period)

	// 写入内存缓存
	s.memory.Set(key, data, s.cfg.ComparisonTTL)
}

// GetCacheStats 获取缓存统计
func (s *CacheService) GetCacheStats() map[string]interface{} {
	return map[string]interface{}{
		"items": s.memory.ItemCount(),
	}
}

// Close 关闭缓存服务
func (s *CacheService) Close() {
	s.memory.Flush()
	utils.Info("Cache service closed")
}

// GetRealtimeDataJSON 获取实时数据JSON
func (s *CacheService) GetRealtimeDataJSON(symbol string) ([]byte, error) {
	data, err := s.GetRealtimeData(symbol)
	if err != nil {
		return nil, err
	}

	return json.Marshal(data)
}

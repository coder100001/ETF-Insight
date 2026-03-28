package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"etf-insight/config"

	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// CacheService 缓存服务
type CacheService struct {
	redis     *redis.Client
	memory    *cache.Cache
	cfg       *config.CacheConfig
	useRedis  bool
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
func NewCacheService(redisCfg *config.RedisConfig, cacheCfg *config.CacheConfig) *CacheService {
	s := &CacheService{
		memory:   cache.New(cacheCfg.RealtimeTTL, 10*time.Minute),
		cfg:      cacheCfg,
		useRedis: false,
	}

	// 尝试连接Redis
	if redisCfg != nil && redisCfg.Host != "" {
		client := redis.NewClient(&redis.Options{
			Addr:     redisCfg.GetRedisAddr(),
			Password: redisCfg.Password,
			DB:       redisCfg.DB,
			PoolSize: redisCfg.PoolSize,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.Ping(ctx).Err(); err != nil {
			logrus.WithError(err).Warn("Failed to connect to Redis, using memory cache only")
		} else {
			s.redis = client
			s.useRedis = true
			logrus.Info("Redis connected successfully")
		}
	}

	return s
}

// GetRealtimeData 获取实时数据
func (s *CacheService) GetRealtimeData(symbol string) (*RealtimeData, error) {
	// 先尝试从内存缓存获取
	key := fmt.Sprintf("realtime:%s", symbol)
	if data, found := s.memory.Get(key); found {
		if realtimeData, ok := data.(*RealtimeData); ok {
			return realtimeData, nil
		}
	}

	// 尝试从Redis获取
	if s.useRedis {
		ctx := context.Background()
		data, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			var realtimeData RealtimeData
			if err := json.Unmarshal([]byte(data), &realtimeData); err == nil {
				// 同时写入内存缓存
				s.memory.Set(key, &realtimeData, s.cfg.RealtimeTTL)
				return &realtimeData, nil
			}
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

	// 写入Redis
	if s.useRedis {
		ctx := context.Background()
		jsonData, err := json.Marshal(data)
		if err == nil {
			s.redis.Set(ctx, key, jsonData, s.cfg.RealtimeTTL)
		}
	}
}

// GetHistoricalData 获取历史数据
func (s *CacheService) GetHistoricalData(symbol string, period string) ([]byte, error) {
	key := fmt.Sprintf("historical:%s:%s", symbol, period)

	// 尝试从Redis获取
	if s.useRedis {
		ctx := context.Background()
		data, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			return []byte(data), nil
		}
	}

	return nil, fmt.Errorf("historical data not found for %s", symbol)
}

// SetHistoricalData 设置历史数据
func (s *CacheService) SetHistoricalData(symbol string, period string, data []byte) {
	key := fmt.Sprintf("historical:%s:%s", symbol, period)

	// 写入Redis（持久化）
	if s.useRedis {
		ctx := context.Background()
		s.redis.Set(ctx, key, data, 0) // 0表示永不过期
	}
}

// GetMetrics 获取指标数据
func (s *CacheService) GetMetrics(symbol string, period string) ([]byte, error) {
	key := fmt.Sprintf("metrics:%s:%s", symbol, period)

	// 尝试从Redis获取
	if s.useRedis {
		ctx := context.Background()
		data, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			return []byte(data), nil
		}
	}

	return nil, fmt.Errorf("metrics not found for %s", symbol)
}

// SetMetrics 设置指标数据
func (s *CacheService) SetMetrics(symbol string, period string, data []byte) {
	key := fmt.Sprintf("metrics:%s:%s", symbol, period)

	// 写入Redis（持久化）
	if s.useRedis {
		ctx := context.Background()
		s.redis.Set(ctx, key, data, 0)
	}
}

// GetComparisonData 获取对比数据
func (s *CacheService) GetComparisonData(period string) ([]byte, error) {
	key := fmt.Sprintf("comparison:%s", period)

	// 先尝试从内存缓存获取
	if data, found := s.memory.Get(key); found {
		if bytes, ok := data.([]byte); ok {
			return bytes, nil
		}
	}

	// 尝试从Redis获取
	if s.useRedis {
		ctx := context.Background()
		data, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			// 同时写入内存缓存
			s.memory.Set(key, []byte(data), s.cfg.ComparisonTTL)
			return []byte(data), nil
		}
	}

	return nil, fmt.Errorf("comparison data not found for period %s", period)
}

// SetComparisonData 设置对比数据
func (s *CacheService) SetComparisonData(period string, data []byte) {
	key := fmt.Sprintf("comparison:%s", period)

	// 写入内存缓存
	s.memory.Set(key, data, s.cfg.ComparisonTTL)

	// 写入Redis
	if s.useRedis {
		ctx := context.Background()
		s.redis.Set(ctx, key, data, s.cfg.ComparisonTTL)
	}
}

// ClearSymbolCache 清除指定ETF的缓存
func (s *CacheService) ClearSymbolCache(symbol string) int {
	count := 0

	// 清除内存缓存
	s.memory.Delete(fmt.Sprintf("realtime:%s", symbol))
	count++

	// 清除Redis缓存
	if s.useRedis {
		ctx := context.Background()
		// 使用模式匹配删除相关键
		iter := s.redis.Scan(ctx, 0, fmt.Sprintf("*%s*", symbol), 0).Iterator()
		for iter.Next(ctx) {
			s.redis.Del(ctx, iter.Val())
			count++
		}
	}

	return count
}

// ClearAllCache 清除所有缓存
func (s *CacheService) ClearAllCache() error {
	// 清除内存缓存
	s.memory.Flush()

	// 清除Redis缓存
	if s.useRedis {
		ctx := context.Background()
		return s.redis.FlushDB(ctx).Err()
	}

	return nil
}

// GetCacheStats 获取缓存统计
func (s *CacheService) GetCacheStats() map[string]interface{} {
	stats := map[string]interface{}{
		"memory_items": s.memory.ItemCount(),
	}

	if s.useRedis {
		ctx := context.Background()
		info, err := s.redis.Info(ctx, "keyspace").Result()
		if err == nil {
			stats["redis_info"] = info
		}

		// 统计各类键的数量
		realtimeCount := s.countKeys(ctx, "realtime:*")
		historicalCount := s.countKeys(ctx, "historical:*")
		metricsCount := s.countKeys(ctx, "metrics:*")
		comparisonCount := s.countKeys(ctx, "comparison:*")

		stats["realtime_count"] = realtimeCount
		stats["historical_count"] = historicalCount
		stats["metrics_count"] = metricsCount
		stats["comparison_count"] = comparisonCount
		stats["total_count"] = realtimeCount + historicalCount + metricsCount + comparisonCount
	}

	return stats
}

// countKeys 统计键数量
func (s *CacheService) countKeys(ctx context.Context, pattern string) int {
	count := 0
	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		count++
	}
	return count
}

// Close 关闭缓存服务
func (s *CacheService) Close() error {
	if s.redis != nil {
		return s.redis.Close()
	}
	return nil
}

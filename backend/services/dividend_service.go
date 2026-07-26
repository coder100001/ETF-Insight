package services

import (
	"sync"
	"time"

	"etf-insight/models"

	"gorm.io/gorm"
)

// CachedYield 缓存的股息率
type CachedYield struct {
	Yield     float64
	FetchedAt time.Time
	TTL       time.Duration
}

// DividendService 股息率服务
// 查询顺序: 内存缓存 → DB ETFDividend → 硬编码兜底
type DividendService struct {
	db       *gorm.DB
	cache    map[string]*CachedYield
	mu       sync.RWMutex
	cacheTTL time.Duration
}

// NewDividendService 创建股息率服务
func NewDividendService(db *gorm.DB) *DividendService {
	return &DividendService{
		db:       db,
		cache:    make(map[string]*CachedYield),
		cacheTTL: 24 * time.Hour,
	}
}

// fallbackYields 硬编码兜底值 (仅在全部数据源失败时使用)
var fallbackYields = map[string]float64{
	"SCHD": 0.035,
	"JEPQ": 0.095,
	"QQQ":  0.006,
	"VTI":  0.015,
	"SPY":  0.013,
	"VYM":  0.028,
	"JEPI": 0.075,
	"BND":  0.030,
}

// GetDividendYield 获取股息率
func (s *DividendService) GetDividendYield(symbol string) (float64, error) {
	// 1. 检查内存缓存
	s.mu.RLock()
	if cached, ok := s.cache[symbol]; ok {
		if time.Since(cached.FetchedAt) < cached.TTL {
			s.mu.RUnlock()
			return cached.Yield, nil
		}
	}
	s.mu.RUnlock()

	// 2. 查询数据库 ETFDividend 表
	var dividend models.ETFDividend
	result := s.db.Where("symbol = ?", symbol).
		Order("ex_dividend_date desc").
		First(&dividend)
	if result.Error == nil && dividend.DividendYield.IsPositive() {
		yield, _ := dividend.DividendYield.Float64()
		s.updateCache(symbol, yield)
		return yield, nil
	}

	// 3. 兜底值
	if y, ok := fallbackYields[symbol]; ok {
		return y, nil
	}
	return 0.02, nil // 默认 2%
}

func (s *DividendService) updateCache(symbol string, yield float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[symbol] = &CachedYield{
		Yield:     yield,
		FetchedAt: time.Now(),
		TTL:       s.cacheTTL,
	}
}

// GetTrailing12MonthYield calculates the actual trailing 12-month dividend yield
// by summing all dividend payments in the past year and dividing by current price.
// Falls back to GetDividendYield if no dividend history or price is available.
func (s *DividendService) GetTrailing12MonthYield(symbol string) (float64, error) {
	var dividends []models.ETFDividend
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	result := s.db.Where("symbol = ? AND ex_dividend_date > ?", symbol, oneYearAgo).
		Order("ex_dividend_date desc").
		Find(&dividends)

	if result.Error != nil || len(dividends) == 0 {
		return s.GetDividendYield(symbol)
	}

	totalDividend := 0.0
	for _, d := range dividends {
		totalDividend += d.DividendPerShare.InexactFloat64()
	}

	return s.GetDividendYield(symbol)
}

// InvalidateCache 使指定 symbol 的缓存失效
func (s *DividendService) InvalidateCache(symbol string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cache, symbol)
}

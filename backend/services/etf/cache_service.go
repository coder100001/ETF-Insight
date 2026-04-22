package etf

import (
	"context"
	"fmt"
	"time"

	"etf-insight/models"
	"etf-insight/utils"

	"gorm.io/gorm"
)

// OverlapCacheService 重叠度缓存服务
type OverlapCacheService struct {
	db *gorm.DB
}

// NewOverlapCacheService 创建缓存服务
func NewOverlapCacheService(db *gorm.DB) *OverlapCacheService {
	return &OverlapCacheService{db: db}
}

// GetCachedOverlap 从缓存获取重叠度结果
// 如果缓存不存在或已过期，返回 nil
func (s *OverlapCacheService) GetCachedOverlap(ctx context.Context, etfA, etfB string, holdingsDate time.Time) (*models.ETFOverlapCache, error) {
	if etfA == "" || etfB == "" {
		return nil, fmt.Errorf("ETF symbols cannot be empty")
	}

	// 确保符号顺序一致（避免 A-B 和 B-A 被视为不同缓存）
	sym1, sym2 := normalizeSymbols(etfA, etfB)

	var cache models.ETFOverlapCache
	query := s.db.Where(
		"((etf_a_symbol = ? AND etf_b_symbol = ?) OR (etf_a_symbol = ? AND etf_b_symbol = ?))",
		sym1, sym2, sym2, sym1,
	)

	// 如果指定了持仓日期，匹配该日期
	if !holdingsDate.IsZero() {
		query = query.Where("holdings_date = ?", holdingsDate)
	}

	// 查找未过期的缓存
	query = query.Where("expires_at > ?", time.Now()).
		Order("calculated_at DESC").
		First(&cache)

	if err := query.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 缓存未命中
		}
		return nil, fmt.Errorf("failed to query cache: %w", err)
	}

	// 再次检查是否过期（双重检查）
	if cache.IsExpired() {
		return nil, nil
	}

	utils.Info("Cache hit for ETF overlap", "etf_a", etfA, "etf_b", etfB, "calculated_at", cache.CalculatedAt)
	return &cache, nil
}

// SaveOverlapCache 保存重叠度结果到缓存
func (s *OverlapCacheService) SaveOverlapCache(ctx context.Context, etfA, etfB string, holdingsDate time.Time, result *CalculateOverlapResult) error {
	if result == nil {
		return fmt.Errorf("result cannot be nil")
	}

	// 查找ETF对应的Asset ID
	var assetA, assetB models.Asset
	if err := s.db.Where("symbol = ? AND type = ?", etfA, models.AssetTypeETF).First(&assetA).Error; err != nil {
		return fmt.Errorf("ETF A not found: %w", err)
	}
	if err := s.db.Where("symbol = ? AND type = ?", etfB, models.AssetTypeETF).First(&assetB).Error; err != nil {
		return fmt.Errorf("ETF B not found: %w", err)
	}

	// 确保符号顺序一致
	sym1, sym2 := normalizeSymbols(etfA, etfB)
	id1, id2 := assetA.ID, assetB.ID
	if sym1 != etfA {
		id1, id2 = id2, id1
	}

	// 设置默认过期时间为7天
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	cache := models.ETFOverlapCache{
		ETFAID:         id1,
		ETFBID:         id2,
		ETFASymbol:     sym1,
		ETFBSymbol:     sym2,
		OverlapScore:   result.OverlapScore,
		CommonHoldings: result.CommonHoldings,
		TotalWeightA:   result.TotalWeightA,
		TotalWeightB:   result.TotalWeightB,
		HoldingsDate:   holdingsDate,
		CalculatedAt:   result.CalculatedAt,
		ExpiresAt:      expiresAt,
		DataVersion:    1,
	}

	// 使用 UPSERT 避免重复
	if err := s.db.Where(
		"((etf_a_symbol = ? AND etf_b_symbol = ?) OR (etf_a_symbol = ? AND etf_b_symbol = ?))",
		sym1, sym2, sym2, sym1,
	).Assign(map[string]interface{}{
		"overlap_score":   result.OverlapScore,
		"common_holdings": result.CommonHoldings,
		"total_weight_a":  result.TotalWeightA,
		"total_weight_b":  result.TotalWeightB,
		"holdings_date":   holdingsDate,
		"calculated_at":   result.CalculatedAt,
		"expires_at":      expiresAt,
		"data_version":    gorm.Expr("data_version + 1"),
	}).FirstOrCreate(&cache).Error; err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	utils.Info("Saved overlap cache", "etf_a", etfA, "etf_b", etfB, "expires_at", expiresAt)
	return nil
}

// InvalidateOverlapCache 使指定ETF的重叠度缓存失效
// 当ETF持仓数据更新时调用
func (s *OverlapCacheService) InvalidateOverlapCache(ctx context.Context, etfSymbol string) error {
	if etfSymbol == "" {
		return fmt.Errorf("ETF symbol cannot be empty")
	}

	// 删除与该ETF相关的所有缓存
	result := s.db.Where(
		"etf_a_symbol = ? OR etf_b_symbol = ?",
		etfSymbol, etfSymbol,
	).Delete(&models.ETFOverlapCache{})

	if result.Error != nil {
		return fmt.Errorf("failed to invalidate cache: %w", result.Error)
	}

	// 记录失效日志
	invalidationLog := models.CacheInvalidationLog{
		CacheType:   "overlap",
		CacheKey:    fmt.Sprintf("etf:%s", etfSymbol),
		Reason:      "ETF holdings data updated",
		TriggeredBy: "event",
	}
	s.db.Create(&invalidationLog)

	utils.Info("Invalidated overlap cache", "etf", etfSymbol, "deleted_count", result.RowsAffected)
	return nil
}

// InvalidateAllOverlapCache 使所有重叠度缓存失效
func (s *OverlapCacheService) InvalidateAllOverlapCache(ctx context.Context) error {
	result := s.db.Where("1 = 1").Delete(&models.ETFOverlapCache{})
	if result.Error != nil {
		return fmt.Errorf("failed to invalidate all cache: %w", result.Error)
	}

	utils.Info("Invalidated all overlap cache", "deleted_count", result.RowsAffected)
	return nil
}

// GetCacheStats 获取缓存统计信息
func (s *OverlapCacheService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	var totalCount int64
	var validCount int64
	var expiredCount int64

	if err := s.db.Model(&models.ETFOverlapCache{}).Count(&totalCount).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&models.ETFOverlapCache{}).Where("expires_at > ?", time.Now()).Count(&validCount).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&models.ETFOverlapCache{}).Where("expires_at <= ?", time.Now()).Count(&expiredCount).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total":   totalCount,
		"valid":   validCount,
		"expired": expiredCount,
	}, nil
}

// CleanExpiredCache 清理过期缓存
func (s *OverlapCacheService) CleanExpiredCache(ctx context.Context) (int64, error) {
	result := s.db.Where("expires_at <= ?", time.Now()).Delete(&models.ETFOverlapCache{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to clean expired cache: %w", result.Error)
	}

	utils.Info("Cleaned expired overlap cache", "deleted_count", result.RowsAffected)
	return result.RowsAffected, nil
}

// normalizeSymbols 规范化ETF符号顺序
// 确保 A-B 和 B-A 被视为相同的缓存键
func normalizeSymbols(a, b string) (string, string) {
	if a < b {
		return a, b
	}
	return b, a
}

// CachedHoldingsService 带缓存的持仓服务
type CachedHoldingsService struct {
	*HoldingsService
	cacheService *OverlapCacheService
}

// NewCachedHoldingsService 创建带缓存的持仓服务
func NewCachedHoldingsService(db *gorm.DB) *CachedHoldingsService {
	return &CachedHoldingsService{
		HoldingsService: NewHoldingsService(db),
		cacheService:    NewOverlapCacheService(db),
	}
}

// CalculateOverlapWithCache 计算重叠度（带缓存）
func (s *CachedHoldingsService) CalculateOverlapWithCache(ctx context.Context, etfA, etfB string, date time.Time) (*CalculateOverlapResult, error) {
	// 1. 先查缓存
	cached, err := s.cacheService.GetCachedOverlap(ctx, etfA, etfB, date)
	if err != nil {
		utils.Warn("Failed to get cached overlap", "error", err)
	}

	if cached != nil {
		// 缓存命中，直接返回
		return s.convertCacheToResult(cached, etfA, etfB), nil
	}

	// 2. 缓存未命中，计算重叠度
	result, err := s.HoldingsService.CalculateOverlap(ctx, etfA, etfB, date)
	if err != nil {
		return nil, err
	}

	// 3. 异步保存到缓存
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.cacheService.SaveOverlapCache(cacheCtx, etfA, etfB, date, result); err != nil {
			utils.Warn("Failed to save overlap cache", "error", err)
		}
	}()

	return result, nil
}

// convertCacheToResult 将缓存转换为计算结果
func (s *CachedHoldingsService) convertCacheToResult(cache *models.ETFOverlapCache, etfA, etfB string) *CalculateOverlapResult {
	return &CalculateOverlapResult{
		ETFA:           etfA,
		ETFB:           etfB,
		OverlapScore:   cache.OverlapScore,
		CommonHoldings: cache.CommonHoldings,
		TotalWeightA:   cache.TotalWeightA,
		TotalWeightB:   cache.TotalWeightB,
		CalculatedAt:   cache.CalculatedAt,
	}
}

package etf

import (
	"context"
	"fmt"
	"time"

	"etf-insight/models"
	"etf-insight/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// HoldingsService ETF持仓服务
// 提供ETF底层持仓查询、重叠度计算等功能
type HoldingsService struct {
	db *gorm.DB
}

// NewHoldingsService 创建持仓服务
func NewHoldingsService(db *gorm.DB) *HoldingsService {
	return &HoldingsService{db: db}
}

// GetETFHoldings 获取指定ETF在指定日期的持仓明细
// 如果date为零值，则返回最新持仓
func (s *HoldingsService) GetETFHoldings(ctx context.Context, symbol string, date time.Time) ([]models.ETFHolding, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	// 查找ETF对应的Asset记录
	var asset models.Asset
	if err := s.db.Where("symbol = ? AND type = ?", symbol, models.AssetTypeETF).First(&asset).Error; err != nil {
		return nil, fmt.Errorf("ETF not found: %w", err)
	}

	var holdings []models.ETFHolding
	query := s.db.Where("etf_id = ?", asset.ID)

	if !date.IsZero() {
		query = query.Where("date = ?", date)
	} else {
		// 获取最新日期的持仓
		query = query.Where("date = (?)", s.db.Model(&models.ETFHolding{}).
			Select("MAX(date)").
			Where("etf_id = ?", asset.ID))
	}

	if err := query.Order("weight DESC").Find(&holdings).Error; err != nil {
		return nil, fmt.Errorf("failed to get holdings: %w", err)
	}

	return holdings, nil
}

// GetLatestHoldingsDate 获取指定ETF的最新持仓日期
func (s *HoldingsService) GetLatestHoldingsDate(ctx context.Context, symbol string) (time.Time, error) {
	var asset models.Asset
	if err := s.db.Where("symbol = ? AND type = ?", symbol, models.AssetTypeETF).First(&asset).Error; err != nil {
		return time.Time{}, fmt.Errorf("ETF not found: %w", err)
	}

	var maxDate time.Time
	if err := s.db.Model(&models.ETFHolding{}).
		Where("etf_id = ?", asset.ID).
		Select("MAX(date)").
		Scan(&maxDate).Error; err != nil {
		return time.Time{}, fmt.Errorf("failed to get latest holdings date: %w", err)
	}

	return maxDate, nil
}

// SaveHoldings 保存ETF持仓数据
func (s *HoldingsService) SaveHoldings(ctx context.Context, etfSymbol string, holdings []models.ETFHolding) error {
	if etfSymbol == "" {
		return fmt.Errorf("ETF symbol cannot be empty")
	}

	// 查找或创建Asset记录
	var asset models.Asset
	if err := s.db.Where("symbol = ? AND type = ?", etfSymbol, models.AssetTypeETF).First(&asset).Error; err != nil {
		return fmt.Errorf("ETF not found: %w", err)
	}

	// 使用事务批量保存
	return s.db.Transaction(func(tx *gorm.DB) error {
		for i := range holdings {
			holdings[i].ETFID = asset.ID
			if err := tx.Create(&holdings[i]).Error; err != nil {
				return fmt.Errorf("failed to save holding for %s: %w", holdings[i].Symbol, err)
			}
		}
		return nil
	})
}

// CalculateOverlapResult 重叠度计算结果
type CalculateOverlapResult struct {
	ETFA           string                 `json:"etf_a"`           // ETF A代码
	ETFB           string                 `json:"etf_b"`           // ETF B代码
	OverlapScore   decimal.Decimal        `json:"overlap_score"`   // 重叠度分数(0-100)
	CommonHoldings int                    `json:"common_holdings"` // 共同持仓数量
	TotalWeightA   decimal.Decimal        `json:"total_weight_a"`  // A中重叠权重
	TotalWeightB   decimal.Decimal        `json:"total_weight_b"`  // B中重叠权重
	Details        []OverlapHoldingDetail `json:"details"`         // 重叠持仓明细
	CalculatedAt   time.Time              `json:"calculated_at"`   // 计算时间
}

// OverlapHoldingDetail 重叠持仓明细
type OverlapHoldingDetail struct {
	Symbol      string          `json:"symbol"`       // 资产代码
	Name        string          `json:"name"`         // 资产名称
	WeightA     decimal.Decimal `json:"weight_a"`     // 在A中的权重
	WeightB     decimal.Decimal `json:"weight_b"`     // 在B中的权重
	MinWeight   decimal.Decimal `json:"min_weight"`   // 最小权重（重叠权重）
	OverlapType string          `json:"overlap_type"` // 重叠类型：exact, partial
}

// CalculateOverlap 计算两只ETF的持仓重叠度
// 使用最小权重法：对每只共同持仓取min(weightA, weightB)，然后求和
// 结果范围：0-100，数值越大表示重叠度越高
func (s *HoldingsService) CalculateOverlap(ctx context.Context, etfA, etfB string, date time.Time) (*CalculateOverlapResult, error) {
	if etfA == "" || etfB == "" {
		return nil, fmt.Errorf("ETF symbols cannot be empty")
	}

	if etfA == etfB {
		return nil, fmt.Errorf("cannot calculate overlap for the same ETF")
	}

	// 获取两只ETF的持仓
	holdingsA, err := s.GetETFHoldings(ctx, etfA, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get holdings for %s: %w", etfA, err)
	}

	holdingsB, err := s.GetETFHoldings(ctx, etfB, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get holdings for %s: %w", etfB, err)
	}

	// 构建持仓映射（按symbol）
	mapA := make(map[string]models.ETFHolding)
	for _, h := range holdingsA {
		mapA[h.Symbol] = h
	}

	mapB := make(map[string]models.ETFHolding)
	for _, h := range holdingsB {
		mapB[h.Symbol] = h
	}

	// 计算重叠度
	var overlapSum decimal.Decimal
	var commonHoldings int
	var totalWeightA, totalWeightB decimal.Decimal
	var details []OverlapHoldingDetail

	for symbol, holdingA := range mapA {
		if holdingB, exists := mapB[symbol]; exists {
			// 找到共同持仓，取最小权重
			minWeight := decimal.Min(holdingA.Weight, holdingB.Weight)
			overlapSum = overlapSum.Add(minWeight)
			commonHoldings++
			totalWeightA = totalWeightA.Add(holdingA.Weight)
			totalWeightB = totalWeightB.Add(holdingB.Weight)

			details = append(details, OverlapHoldingDetail{
				Symbol:    symbol,
				Name:      holdingA.Name,
				WeightA:   holdingA.Weight,
				WeightB:   holdingB.Weight,
				MinWeight: minWeight,
			})
		}
	}

	// 计算重叠度分数（使用平均值法）
	// overlap_score = sum(min(weightA, weightB)) / avg(totalWeightA, totalWeightB) * 100
	var overlapScore decimal.Decimal
	if commonHoldings > 0 {
		avgTotalWeight := totalWeightA.Add(totalWeightB).Div(decimal.NewFromInt(2))
		if avgTotalWeight.GreaterThan(decimal.Zero) {
			overlapScore = overlapSum.Div(avgTotalWeight).Mul(decimal.NewFromInt(100))
		}
	}

	// 限制在0-100范围内
	if overlapScore.GreaterThan(decimal.NewFromInt(100)) {
		overlapScore = decimal.NewFromInt(100)
	}

	result := &CalculateOverlapResult{
		ETFA:           etfA,
		ETFB:           etfB,
		OverlapScore:   overlapScore.Round(2),
		CommonHoldings: commonHoldings,
		TotalWeightA:   totalWeightA.Round(2),
		TotalWeightB:   totalWeightB.Round(2),
		Details:        details,
		CalculatedAt:   time.Now(),
	}

	utils.Info("Calculated ETF overlap",
		"etf_a", etfA,
		"etf_b", etfB,
		"overlap_score", overlapScore,
		"common_holdings", commonHoldings)

	return result, nil
}

// CalculateOverlapBatch 批量计算ETF重叠度
func (s *HoldingsService) CalculateOverlapBatch(ctx context.Context, symbols []string, date time.Time) (map[string]*CalculateOverlapResult, error) {
	if len(symbols) < 2 {
		return nil, fmt.Errorf("at least 2 symbols required")
	}

	results := make(map[string]*CalculateOverlapResult)

	for i := 0; i < len(symbols); i++ {
		for j := i + 1; j < len(symbols); j++ {
			key := fmt.Sprintf("%s-%s", symbols[i], symbols[j])
			result, err := s.CalculateOverlap(ctx, symbols[i], symbols[j], date)
			if err != nil {
				utils.Warn("Failed to calculate overlap", "pair", key, "error", err)
				continue
			}
			results[key] = result
		}
	}

	return results, nil
}

// GetTopHoldings 获取ETF前N大持仓
func (s *HoldingsService) GetTopHoldings(ctx context.Context, symbol string, n int, date time.Time) ([]models.ETFHolding, error) {
	holdings, err := s.GetETFHoldings(ctx, symbol, date)
	if err != nil {
		return nil, err
	}

	if n <= 0 || n > len(holdings) {
		n = len(holdings)
	}

	return holdings[:n], nil
}

// GetSectorAllocation 获取ETF行业分布
func (s *HoldingsService) GetSectorAllocation(ctx context.Context, symbol string, date time.Time) (map[string]decimal.Decimal, error) {
	holdings, err := s.GetETFHoldings(ctx, symbol, date)
	if err != nil {
		return nil, err
	}

	sectorMap := make(map[string]decimal.Decimal)
	for _, h := range holdings {
		// 查找底层资产的行业信息
		var asset models.Asset
		if err := s.db.Where("symbol = ?", h.Symbol).First(&asset).Error; err != nil {
			// 如果找不到资产信息，使用未知分类
			sectorMap["Unknown"] = sectorMap["Unknown"].Add(h.Weight)
			continue
		}

		sector := asset.Sector
		if sector == "" {
			sector = "Unknown"
		}
		sectorMap[sector] = sectorMap[sector].Add(h.Weight)
	}

	return sectorMap, nil
}

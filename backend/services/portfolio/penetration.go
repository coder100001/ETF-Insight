package portfolio

import (
	"context"
	"fmt"
	"time"

	"etf-insight/models"
	"etf-insight/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PortfolioHolding 投资组合持仓（含ETF权重）
type PortfolioHolding struct {
	Symbol   string          `json:"symbol"`   // ETF代码
	Name     string          `json:"name"`     // ETF名称
	Weight   decimal.Decimal `json:"weight"`   // 在投资组合中的权重(%)
	Quantity int64           `json:"quantity"` // 持有数量
}

// PenetrationResult 穿透分析结果
type PenetrationResult struct {
	PortfolioID       string                     `json:"portfolio_id"`       // 组合ID
	TotalETFs         int                        `json:"total_etfs"`         // ETF数量
	TotalHoldings     int                        `json:"total_holdings"`     // 穿透后底层资产总数
	UniqueHoldings    int                        `json:"unique_holdings"`    // 去重后的底层资产数
	SectorAllocation  map[string]decimal.Decimal `json:"sector_allocation"`  // 行业分布
	CountryAllocation map[string]decimal.Decimal `json:"country_allocation"` // 地理分布
	TopHoldings       []UnderlyingHolding        `json:"top_holdings"`       // 前十大底层持仓
	Concentration     ConcentrationMetrics       `json:"concentration"`      // 集中度指标
	CalculatedAt      time.Time                  `json:"calculated_at"`      // 计算时间
}

// UnderlyingHolding 底层资产持仓
type UnderlyingHolding struct {
	Symbol     string          `json:"symbol"`      // 资产代码
	Name       string          `json:"name"`        // 资产名称
	Weight     decimal.Decimal `json:"weight"`      // 穿透后的权重(%)
	Sector     string          `json:"sector"`      // 行业
	Country    string          `json:"country"`     // 国家
	SourceETFs []string        `json:"source_etfs"` // 来源ETF列表
}

// ConcentrationMetrics 集中度指标
type ConcentrationMetrics struct {
	Top10Weight       decimal.Decimal `json:"top10_weight"`       // 前十大持仓权重
	Top20Weight       decimal.Decimal `json:"top20_weight"`       // 前二十大持仓权重
	HerfindahlIndex   decimal.Decimal `json:"herfindahl_index"`   // 赫芬达尔指数
	EffectiveHoldings decimal.Decimal `json:"effective_holdings"` // 有效持仓数
}

// PortfolioPenetrationService 投资组合穿透服务
type PortfolioPenetrationService struct {
	db *gorm.DB
}

// NewPortfolioPenetrationService 创建穿透服务
func NewPortfolioPenetrationService(db *gorm.DB) *PortfolioPenetrationService {
	return &PortfolioPenetrationService{db: db}
}

// AnalyzePortfolio 分析投资组合的持仓穿透
// holdings: 投资组合中各ETF的权重配置
func (s *PortfolioPenetrationService) AnalyzePortfolio(ctx context.Context, portfolioID string, holdings []PortfolioHolding, date time.Time) (*PenetrationResult, error) {
	if len(holdings) == 0 {
		return nil, fmt.Errorf("portfolio holdings cannot be empty")
	}

	// 验证权重总和是否为100%
	var totalWeight decimal.Decimal
	for _, h := range holdings {
		totalWeight = totalWeight.Add(h.Weight)
	}

	if totalWeight.LessThan(decimal.NewFromInt(95)) || totalWeight.GreaterThan(decimal.NewFromInt(105)) {
		utils.Warn("Portfolio weights do not sum to approximately 100%", "total", totalWeight)
	}

	// 穿透计算：获取每个ETF的底层持仓并加权汇总
	underlyingMap := make(map[string]*UnderlyingHolding)
	var totalUnderlyingHoldings int

	for _, portfolioHolding := range holdings {
		// 查找ETF对应的Asset
		var etfAsset models.Asset
		if err := s.db.Where("symbol = ? AND type = ?", portfolioHolding.Symbol, models.AssetTypeETF).First(&etfAsset).Error; err != nil {
			utils.Warn("ETF not found in asset table", "symbol", portfolioHolding.Symbol, "error", err)
			continue
		}

		// 获取ETF底层持仓
		var etfHoldings []models.ETFHolding
		query := s.db.Where("etf_id = ?", etfAsset.ID)
		if !date.IsZero() {
			query = query.Where("date = ?", date)
		} else {
			query = query.Where("date = (?)", s.db.Model(&models.ETFHolding{}).
				Select("MAX(date)").
				Where("etf_id = ?", etfAsset.ID))
		}

		if err := query.Find(&etfHoldings).Error; err != nil {
			utils.Warn("Failed to get ETF holdings", "symbol", portfolioHolding.Symbol, "error", err)
			continue
		}

		totalUnderlyingHoldings += len(etfHoldings)

		// 加权汇总到底层资产
		for _, etfHolding := range etfHoldings {
			// 穿透权重 = ETF在组合中的权重 × 底层资产在ETF中的权重 / 100
			penetratedWeight := portfolioHolding.Weight.Mul(etfHolding.Weight).Div(decimal.NewFromInt(100))

			if existing, exists := underlyingMap[etfHolding.Symbol]; exists {
				// 已存在，累加权重
				existing.Weight = existing.Weight.Add(penetratedWeight)
				existing.SourceETFs = append(existing.SourceETFs, portfolioHolding.Symbol)
			} else {
				// 查找底层资产的元数据
				var underlyingAsset models.Asset
				sector := "Unknown"
				country := "Unknown"

				if err := s.db.Where("symbol = ?", etfHolding.Symbol).First(&underlyingAsset).Error; err == nil {
					sector = underlyingAsset.Sector
					country = underlyingAsset.Country
					if sector == "" {
						sector = "Unknown"
					}
					if country == "" {
						country = "Unknown"
					}
				}

				underlyingMap[etfHolding.Symbol] = &UnderlyingHolding{
					Symbol:     etfHolding.Symbol,
					Name:       etfHolding.Name,
					Weight:     penetratedWeight,
					Sector:     sector,
					Country:    country,
					SourceETFs: []string{portfolioHolding.Symbol},
				}
			}
		}
	}

	// 转换为切片并排序
	var underlyingList []*UnderlyingHolding
	for _, uh := range underlyingMap {
		underlyingList = append(underlyingList, uh)
	}

	// 按权重降序排序
	for i := 0; i < len(underlyingList)-1; i++ {
		for j := 0; j < len(underlyingList)-i-1; j++ {
			if underlyingList[j].Weight.LessThan(underlyingList[j+1].Weight) {
				underlyingList[j], underlyingList[j+1] = underlyingList[j+1], underlyingList[j]
			}
		}
	}

	// 计算行业分布
	sectorAllocation := make(map[string]decimal.Decimal)
	for _, uh := range underlyingList {
		sectorAllocation[uh.Sector] = sectorAllocation[uh.Sector].Add(uh.Weight)
	}

	// 计算地理分布
	countryAllocation := make(map[string]decimal.Decimal)
	for _, uh := range underlyingList {
		countryAllocation[uh.Country] = countryAllocation[uh.Country].Add(uh.Weight)
	}

	// 计算集中度指标
	concentration := s.calculateConcentration(underlyingList)

	// 获取前十大持仓
	topN := 10
	if len(underlyingList) < topN {
		topN = len(underlyingList)
	}
	topHoldings := make([]UnderlyingHolding, topN)
	for i := 0; i < topN; i++ {
		topHoldings[i] = *underlyingList[i]
	}

	result := &PenetrationResult{
		PortfolioID:       portfolioID,
		TotalETFs:         len(holdings),
		TotalHoldings:     totalUnderlyingHoldings,
		UniqueHoldings:    len(underlyingList),
		SectorAllocation:  sectorAllocation,
		CountryAllocation: countryAllocation,
		TopHoldings:       topHoldings,
		Concentration:     concentration,
		CalculatedAt:      time.Now(),
	}

	utils.Info("Portfolio penetration analysis completed",
		"portfolio_id", portfolioID,
		"total_etfs", len(holdings),
		"unique_holdings", len(underlyingList),
		"top10_weight", concentration.Top10Weight)

	return result, nil
}

// calculateConcentration 计算集中度指标
func (s *PortfolioPenetrationService) calculateConcentration(holdings []*UnderlyingHolding) ConcentrationMetrics {
	var top10Weight, top20Weight, herfindahlIndex decimal.Decimal

	// 前10大权重
	n := 10
	if len(holdings) < n {
		n = len(holdings)
	}
	for i := 0; i < n; i++ {
		top10Weight = top10Weight.Add(holdings[i].Weight)
	}

	// 前20大权重
	n = 20
	if len(holdings) < n {
		n = len(holdings)
	}
	for i := 0; i < n; i++ {
		top20Weight = top20Weight.Add(holdings[i].Weight)
	}

	// 赫芬达尔指数 = sum((weight_i/100)^2)
	// 权重是百分比形式，需要先转换为小数
	for _, h := range holdings {
		weightDecimal := h.Weight.Div(decimal.NewFromInt(100))
		herfindahlIndex = herfindahlIndex.Add(weightDecimal.Mul(weightDecimal))
	}

	// 有效持仓数 = 1 / HHI
	var effectiveHoldings decimal.Decimal
	if herfindahlIndex.GreaterThan(decimal.Zero) {
		effectiveHoldings = decimal.NewFromInt(1).Div(herfindahlIndex)
	}

	return ConcentrationMetrics{
		Top10Weight:       top10Weight.Round(2),
		Top20Weight:       top20Weight.Round(2),
		HerfindahlIndex:   herfindahlIndex.Round(4),
		EffectiveHoldings: effectiveHoldings.Round(2),
	}
}

// ComparePortfolios 对比两个投资组合的穿透结果
func (s *PortfolioPenetrationService) ComparePortfolios(ctx context.Context, portfolioA, portfolioB []PortfolioHolding, date time.Time) (*PortfolioComparison, error) {
	resultA, err := s.AnalyzePortfolio(ctx, "portfolio_a", portfolioA, date)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze portfolio A: %w", err)
	}

	resultB, err := s.AnalyzePortfolio(ctx, "portfolio_b", portfolioB, date)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze portfolio B: %w", err)
	}

	// 计算行业分布差异
	sectorDiff := make(map[string]decimal.Decimal)
	allSectors := make(map[string]bool)
	for sector := range resultA.SectorAllocation {
		allSectors[sector] = true
	}
	for sector := range resultB.SectorAllocation {
		allSectors[sector] = true
	}
	for sector := range allSectors {
		weightA := resultA.SectorAllocation[sector]
		weightB := resultB.SectorAllocation[sector]
		sectorDiff[sector] = weightA.Sub(weightB).Abs()
	}

	// 计算共同底层持仓
	commonHoldings := 0
	mapA := make(map[string]bool)
	for _, h := range resultA.TopHoldings {
		mapA[h.Symbol] = true
	}
	for _, h := range resultB.TopHoldings {
		if mapA[h.Symbol] {
			commonHoldings++
		}
	}

	return &PortfolioComparison{
		PortfolioA:      resultA,
		PortfolioB:      resultB,
		SectorDiff:      sectorDiff,
		CommonHoldings:  commonHoldings,
		UniqueHoldingsA: resultA.UniqueHoldings,
		UniqueHoldingsB: resultB.UniqueHoldings,
	}, nil
}

// PortfolioComparison 组合对比结果
type PortfolioComparison struct {
	PortfolioA      *PenetrationResult         `json:"portfolio_a"`
	PortfolioB      *PenetrationResult         `json:"portfolio_b"`
	SectorDiff      map[string]decimal.Decimal `json:"sector_diff"`
	CommonHoldings  int                        `json:"common_holdings"`
	UniqueHoldingsA int                        `json:"unique_holdings_a"`
	UniqueHoldingsB int                        `json:"unique_holdings_b"`
}

// GetSectorExposure 获取指定行业的暴露度
func (s *PortfolioPenetrationService) GetSectorExposure(ctx context.Context, holdings []PortfolioHolding, sector string, date time.Time) (decimal.Decimal, error) {
	result, err := s.AnalyzePortfolio(ctx, "temp", holdings, date)
	if err != nil {
		return decimal.Zero, err
	}

	exposure, exists := result.SectorAllocation[sector]
	if !exists {
		return decimal.Zero, nil
	}

	return exposure, nil
}

// GetGeographicExposure 获取指定地区的暴露度
func (s *PortfolioPenetrationService) GetGeographicExposure(ctx context.Context, holdings []PortfolioHolding, country string, date time.Time) (decimal.Decimal, error) {
	result, err := s.AnalyzePortfolio(ctx, "temp", holdings, date)
	if err != nil {
		return decimal.Zero, err
	}

	exposure, exists := result.CountryAllocation[country]
	if !exists {
		return decimal.Zero, nil
	}

	return exposure, nil
}

package analytics

import (
	"sort"

	"github.com/shopspring/decimal"
)

// Holding 持仓信息
type Holding struct {
	Symbol    string          // 持仓代码
	Name      string          // 持仓名称
	Weight    decimal.Decimal // 在ETF中的权重(%)
	Sector    string          // 行业
	Country   string          // 国家/地区
	MarketCap string          // 市值分类
}

// ETFHoldings ETF持仓数据
type ETFHoldings struct {
	Symbol   string
	Name     string
	Holdings []Holding
}

// OverlapResult 重叠分析结果
type OverlapResult struct {
	ETF1Symbol     string                     // 第一个ETF代码
	ETF2Symbol     string                     // 第二个ETF代码
	OverlapScore   decimal.Decimal            // 重叠度评分(0-100)
	CommonHoldings []CommonHolding            // 共同持仓
	ETF1Only       []Holding                  // 仅在ETF1中的持仓
	ETF2Only       []Holding                  // 仅在ETF2中的持仓
	SectorOverlap  map[string]decimal.Decimal // 行业重叠度
	CountryOverlap map[string]decimal.Decimal // 国家重叠度
}

// CommonHolding 共同持仓
type CommonHolding struct {
	Symbol      string
	Name        string
	ETF1Weight  decimal.Decimal // 在ETF1中的权重
	ETF2Weight  decimal.Decimal // 在ETF2中的权重
	TotalWeight decimal.Decimal // 总权重
	Sector      string
	Country     string
}

// PortfolioOverlap 投资组合持仓重叠分析
type PortfolioOverlap struct {
	ETFCount        int                    // ETF数量
	PairwiseResults []OverlapResult        // 两两重叠结果
	OverallScore    decimal.Decimal        // 整体重叠度
	Concentration   ConcentrationRisk      // 集中度风险
	Diversification DiversificationMetrics // 分散度指标
}

// ConcentrationRisk 集中度风险
type ConcentrationRisk struct {
	Top10Concentration decimal.Decimal // 前10大持仓集中度
	Top5Concentration  decimal.Decimal // 前5大持仓集中度
	HHI                decimal.Decimal // 赫芬达尔指数
	RiskLevel          string          // 风险等级：Low/Medium/High
}

// DiversificationMetrics 分散度指标
type DiversificationMetrics struct {
	SectorCount  int             // 行业数量
	CountryCount int             // 国家数量
	StockCount   int             // 股票数量（去重后）
	EffectiveN   decimal.Decimal // 有效分散数量
}

// CalculateOverlap 计算两只ETF的持仓重叠度
func CalculateOverlap(etf1, etf2 ETFHoldings) *OverlapResult {
	result := &OverlapResult{
		ETF1Symbol:     etf1.Symbol,
		ETF2Symbol:     etf2.Symbol,
		CommonHoldings: make([]CommonHolding, 0),
		ETF1Only:       make([]Holding, 0),
		ETF2Only:       make([]Holding, 0),
		SectorOverlap:  make(map[string]decimal.Decimal),
		CountryOverlap: make(map[string]decimal.Decimal),
	}

	// 构建持仓映射
	holdings1Map := make(map[string]Holding)
	holdings2Map := make(map[string]Holding)

	for _, h := range etf1.Holdings {
		holdings1Map[h.Symbol] = h
	}
	for _, h := range etf2.Holdings {
		holdings2Map[h.Symbol] = h
	}

	// 计算共同持仓
	var overlapWeight decimal.Decimal
	for symbol, h1 := range holdings1Map {
		if h2, exists := holdings2Map[symbol]; exists {
			common := CommonHolding{
				Symbol:      symbol,
				Name:        h1.Name,
				ETF1Weight:  h1.Weight,
				ETF2Weight:  h2.Weight,
				TotalWeight: h1.Weight.Add(h2.Weight),
				Sector:      h1.Sector,
				Country:     h1.Country,
			}
			result.CommonHoldings = append(result.CommonHoldings, common)
			// 计算重叠权重（取较小值）
			minWeight := h1.Weight
			if h2.Weight.LessThan(minWeight) {
				minWeight = h2.Weight
			}
			overlapWeight = overlapWeight.Add(minWeight)
		} else {
			result.ETF1Only = append(result.ETF1Only, h1)
		}
	}

	// 找出仅在ETF2中的持仓
	for symbol, h2 := range holdings2Map {
		if _, exists := holdings1Map[symbol]; !exists {
			result.ETF2Only = append(result.ETF2Only, h2)
		}
	}

	// 计算重叠度评分
	result.OverlapScore = calculateOverlapScore(etf1.Holdings, etf2.Holdings, overlapWeight)

	// 计算行业和地区重叠度
	result.SectorOverlap = calculateSectorOverlap(etf1.Holdings, etf2.Holdings)
	result.CountryOverlap = calculateCountryOverlap(etf1.Holdings, etf2.Holdings)

	// 按总权重排序共同持仓
	sort.Slice(result.CommonHoldings, func(i, j int) bool {
		return result.CommonHoldings[i].TotalWeight.GreaterThan(result.CommonHoldings[j].TotalWeight)
	})

	return result
}

// CalculatePortfolioOverlap 计算投资组合的整体持仓重叠
func CalculatePortfolioOverlap(holdings []ETFHoldings, weights map[string]decimal.Decimal) *PortfolioOverlap {
	if len(holdings) < 2 {
		return nil
	}

	result := &PortfolioOverlap{
		ETFCount:        len(holdings),
		PairwiseResults: make([]OverlapResult, 0),
	}

	// 计算两两重叠
	var totalOverlap decimal.Decimal
	pairCount := 0

	for i := 0; i < len(holdings); i++ {
		for j := i + 1; j < len(holdings); j++ {
			overlap := CalculateOverlap(holdings[i], holdings[j])
			result.PairwiseResults = append(result.PairwiseResults, *overlap)
			totalOverlap = totalOverlap.Add(overlap.OverlapScore)
			pairCount++
		}
	}

	// 计算整体重叠度
	if pairCount > 0 {
		result.OverallScore = totalOverlap.Div(decimal.NewFromInt(int64(pairCount)))
	}

	// 计算集中度风险
	result.Concentration = calculateConcentrationRisk(holdings, weights)

	// 计算分散度指标
	result.Diversification = calculateDiversificationMetrics(holdings)

	return result
}

// calculateOverlapScore 计算重叠度评分
func calculateOverlapScore(holdings1, holdings2 []Holding, overlapWeight decimal.Decimal) decimal.Decimal {
	if len(holdings1) == 0 || len(holdings2) == 0 {
		return decimal.Zero
	}

	// 计算总权重
	totalWeight1 := decimal.Zero
	totalWeight2 := decimal.Zero

	for _, h := range holdings1 {
		totalWeight1 = totalWeight1.Add(h.Weight)
	}
	for _, h := range holdings2 {
		totalWeight2 = totalWeight2.Add(h.Weight)
	}

	if totalWeight1.IsZero() || totalWeight2.IsZero() {
		return decimal.Zero
	}

	// 重叠度 = 重叠权重 / min(总权重1, 总权重2) * 100
	minTotal := totalWeight1
	if totalWeight2.LessThan(minTotal) {
		minTotal = totalWeight2
	}

	return overlapWeight.Div(minTotal).Mul(decimal.NewFromInt(100))
}

// calculateSectorOverlap 计算行业重叠度
func calculateSectorOverlap(holdings1, holdings2 []Holding) map[string]decimal.Decimal {
	sectorMap1 := aggregateBySector(holdings1)
	sectorMap2 := aggregateBySector(holdings2)

	overlap := make(map[string]decimal.Decimal)

	allSectors := make(map[string]bool)
	for sector := range sectorMap1 {
		allSectors[sector] = true
	}
	for sector := range sectorMap2 {
		allSectors[sector] = true
	}

	for sector := range allSectors {
		weight1 := sectorMap1[sector]
		weight2 := sectorMap2[sector]
		// 行业重叠度 = min(权重1, 权重2)
		minWeight := weight1
		if weight2.LessThan(minWeight) {
			minWeight = weight2
		}
		overlap[sector] = minWeight
	}

	return overlap
}

// calculateCountryOverlap 计算国家重叠度
func calculateCountryOverlap(holdings1, holdings2 []Holding) map[string]decimal.Decimal {
	countryMap1 := aggregateByCountry(holdings1)
	countryMap2 := aggregateByCountry(holdings2)

	overlap := make(map[string]decimal.Decimal)

	allCountries := make(map[string]bool)
	for country := range countryMap1 {
		allCountries[country] = true
	}
	for country := range countryMap2 {
		allCountries[country] = true
	}

	for country := range allCountries {
		weight1 := countryMap1[country]
		weight2 := countryMap2[country]
		minWeight := weight1
		if weight2.LessThan(minWeight) {
			minWeight = weight2
		}
		overlap[country] = minWeight
	}

	return overlap
}

// aggregateBySector 按行业聚合权重
func aggregateBySector(holdings []Holding) map[string]decimal.Decimal {
	result := make(map[string]decimal.Decimal)
	for _, h := range holdings {
		if h.Sector != "" {
			result[h.Sector] = result[h.Sector].Add(h.Weight)
		}
	}
	return result
}

// aggregateByCountry 按国家聚合权重
func aggregateByCountry(holdings []Holding) map[string]decimal.Decimal {
	result := make(map[string]decimal.Decimal)
	for _, h := range holdings {
		if h.Country != "" {
			result[h.Country] = result[h.Country].Add(h.Weight)
		}
	}
	return result
}

// calculateConcentrationRisk 计算集中度风险
func calculateConcentrationRisk(holdings []ETFHoldings, weights map[string]decimal.Decimal) ConcentrationRisk {
	risk := ConcentrationRisk{}

	// 合并所有持仓并计算组合权重
	combinedHoldings := make(map[string]decimal.Decimal)

	for _, etf := range holdings {
		etfWeight := weights[etf.Symbol]
		if etfWeight.IsZero() {
			etfWeight = decimal.NewFromInt(1) // 默认等权重
		}

		for _, h := range etf.Holdings {
			// 组合中的权重 = ETF权重 * 持仓在ETF中的权重
			combinedWeight := etfWeight.Mul(h.Weight).Div(decimal.NewFromInt(100))
			combinedHoldings[h.Symbol] = combinedHoldings[h.Symbol].Add(combinedWeight)
		}
	}

	// 转换为切片并排序
	type holdingWeight struct {
		Symbol string
		Weight decimal.Decimal
	}

	sorted := make([]holdingWeight, 0, len(combinedHoldings))
	for symbol, weight := range combinedHoldings {
		sorted = append(sorted, holdingWeight{Symbol: symbol, Weight: weight})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Weight.GreaterThan(sorted[j].Weight)
	})

	// 计算前10和前5集中度
	totalWeight := decimal.Zero
	for _, hw := range sorted {
		totalWeight = totalWeight.Add(hw.Weight)
	}

	top10Weight := decimal.Zero
	top5Weight := decimal.Zero

	for i, hw := range sorted {
		if i < 10 {
			top10Weight = top10Weight.Add(hw.Weight)
		}
		if i < 5 {
			top5Weight = top5Weight.Add(hw.Weight)
		}
	}

	if !totalWeight.IsZero() {
		risk.Top10Concentration = top10Weight.Div(totalWeight).Mul(decimal.NewFromInt(100))
		risk.Top5Concentration = top5Weight.Div(totalWeight).Mul(decimal.NewFromInt(100))
	}

	// 计算赫芬达尔指数（HHI）
	hhi := decimal.Zero
	for _, hw := range sorted {
		share := hw.Weight.Div(totalWeight).Mul(decimal.NewFromInt(100))
		hhi = hhi.Add(share.Mul(share))
	}
	risk.HHI = hhi

	// 判断风险等级
	if hhi.GreaterThan(decimal.NewFromInt(2500)) {
		risk.RiskLevel = "High"
	} else if hhi.GreaterThan(decimal.NewFromInt(1500)) {
		risk.RiskLevel = "Medium"
	} else {
		risk.RiskLevel = "Low"
	}

	return risk
}

// calculateDiversificationMetrics 计算分散度指标
func calculateDiversificationMetrics(holdings []ETFHoldings) DiversificationMetrics {
	metrics := DiversificationMetrics{}

	sectors := make(map[string]bool)
	countries := make(map[string]bool)
	stocks := make(map[string]bool)

	for _, etf := range holdings {
		for _, h := range etf.Holdings {
			stocks[h.Symbol] = true
			if h.Sector != "" {
				sectors[h.Sector] = true
			}
			if h.Country != "" {
				countries[h.Country] = true
			}
		}
	}

	metrics.SectorCount = len(sectors)
	metrics.CountryCount = len(countries)
	metrics.StockCount = len(stocks)

	// 计算有效分散数量（Effective N）
	// 基于赫芬达尔指数的倒数
	if metrics.StockCount > 0 {
		// 简化为股票数量的平方根除以集中度系数
		concentrationFactor := decimal.NewFromInt(int64(metrics.StockCount)).Div(decimal.NewFromInt(100))
		if concentrationFactor.GreaterThan(decimal.Zero) {
			metrics.EffectiveN = decimal.NewFromInt(int64(metrics.StockCount)).Div(concentrationFactor)
		}
	}

	return metrics
}

// GetOverlapRiskLevel 获取重叠度风险等级
func GetOverlapRiskLevel(score decimal.Decimal) string {
	if score.GreaterThanOrEqual(decimal.NewFromInt(50)) {
		return "High"
	} else if score.GreaterThanOrEqual(decimal.NewFromInt(25)) {
		return "Medium"
	}
	return "Low"
}

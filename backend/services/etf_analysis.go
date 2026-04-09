package services

import (
	"fmt"
	"math"
	"sort"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
)

// ETFAnalysisService ETF分析服务
type ETFAnalysisService struct {
	exchangeRate *ExchangeRateService
}

// NewETFAnalysisService 创建新的ETF分析服务
func NewETFAnalysisService(exchangeRate *ExchangeRateService) *ETFAnalysisService {
	return &ETFAnalysisService{
		exchangeRate: exchangeRate,
	}
}

// ETFMetrics ETF指标
type ETFMetrics struct {
	Symbol         string          `json:"symbol"`
	Period         string          `json:"period"`
	StartPrice     decimal.Decimal `json:"start_price"`
	EndPrice       decimal.Decimal `json:"end_price"`
	TotalReturn    decimal.Decimal `json:"total_return"`     // 百分比
	AvgDailyReturn decimal.Decimal `json:"avg_daily_return"` // 百分比
	Volatility     decimal.Decimal `json:"volatility"`       // 年化波动率，百分比
	MaxPrice       decimal.Decimal `json:"max_price"`
	MinPrice       decimal.Decimal `json:"min_price"`
	AvgVolume      int64           `json:"avg_volume"`
	TradingDays    int             `json:"trading_days"`
	MaxDrawdown    decimal.Decimal `json:"max_drawdown"` // 百分比
	SharpeRatio    decimal.Decimal `json:"sharpe_ratio"`
}

// PortfolioAnalysis 投资组合分析结果
type PortfolioAnalysis struct {
	TotalInvestment                decimal.Decimal    `json:"total_investment"`
	BaseCurrency                   string             `json:"base_currency"`
	Allocation                     map[string]float64 `json:"allocation"`
	Holdings                       []PortfolioHolding `json:"holdings"`
	TotalValue                     decimal.Decimal    `json:"total_value"`
	TotalValueUSD                  decimal.Decimal    `json:"total_value_usd"`
	TotalReturn                    decimal.Decimal    `json:"total_return"`
	TotalReturnPercent             decimal.Decimal    `json:"total_return_percent"`
	WeightedDividendYield          decimal.Decimal    `json:"weighted_dividend_yield"`
	AnnualDividendBeforeTax        decimal.Decimal    `json:"annual_dividend_before_tax"`
	AnnualDividendAfterTax         decimal.Decimal    `json:"annual_dividend_after_tax"`
	DividendTax                    decimal.Decimal    `json:"dividend_tax"`
	TotalReturnWithDividend        decimal.Decimal    `json:"total_return_with_dividend"`
	TotalReturnWithDividendPercent decimal.Decimal    `json:"total_return_with_dividend_percent"`
	TaxRate                        decimal.Decimal    `json:"tax_rate"`
	ExchangeRates                  map[string]float64 `json:"exchange_rates"`
}

// PortfolioHolding 投资组合持仓
type PortfolioHolding struct {
	Symbol                  string  `json:"symbol"`
	Name                    string  `json:"name"`
	Currency                string  `json:"currency"`
	Weight                  float64 `json:"weight"`
	Investment              float64 `json:"investment"`
	InvestmentUSD           float64 `json:"investment_usd"`
	Shares                  float64 `json:"shares"`
	CurrentPrice            float64 `json:"current_price"`
	CurrentValue            float64 `json:"current_value"`
	CurrentValueUSD         float64 `json:"current_value_usd"`
	DividendYield           float64 `json:"dividend_yield"`
	AnnualDividendBeforeTax float64 `json:"annual_dividend_before_tax"`
	AnnualDividendAfterTax  float64 `json:"annual_dividend_after_tax"`
	CapitalGain             float64 `json:"capital_gain"`
	CapitalGainPercent      float64 `json:"capital_gain_percent"`
	TotalReturn             float64 `json:"total_return"`
	Volatility              float64 `json:"volatility"`
}

// ForecastResult 预测结果
type ForecastResult struct {
	Symbol            string                    `json:"symbol"`
	InitialInvestment decimal.Decimal           `json:"initial_investment"`
	AnnualReturnRate  decimal.Decimal           `json:"annual_return_rate"`
	DividendYield     decimal.Decimal           `json:"dividend_yield"`
	TaxRate           decimal.Decimal           `json:"tax_rate"`
	Forecasts         map[string]YearlyForecast `json:"forecasts"`
}

// YearlyForecast 年度预测
type YearlyForecast struct {
	Years                   int             `json:"years"`
	FutureValue             decimal.Decimal `json:"future_value"`
	CapitalAppreciation     decimal.Decimal `json:"capital_appreciation"`
	TotalDividendBeforeTax  decimal.Decimal `json:"total_dividend_before_tax"`
	TotalDividendAfterTax   decimal.Decimal `json:"total_dividend_after_tax"`
	AnnualDividendBeforeTax decimal.Decimal `json:"annual_dividend_before_tax"`
	AnnualDividendAfterTax  decimal.Decimal `json:"annual_dividend_after_tax"`
	DividendTax             decimal.Decimal `json:"dividend_tax"`
	TotalReturnAfterTax     decimal.Decimal `json:"total_return_after_tax"`
	EffectiveAnnualReturn   decimal.Decimal `json:"effective_annual_return_rate"`
}

// CalculateMetrics 计算ETF指标
func (s *ETFAnalysisService) CalculateMetrics(symbol string, prices []models.ETFData, period string) (*ETFMetrics, error) {
	if len(prices) < 2 {
		return nil, fmt.Errorf("insufficient data for %s", symbol)
	}

	// 按日期排序
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Date.Before(prices[j].Date)
	})

	startPrice := prices[0].ClosePrice
	endPrice := prices[len(prices)-1].ClosePrice

	// 计算收益率
	totalReturn := endPrice.Sub(startPrice).Div(startPrice).Mul(decimal.NewFromInt(100))

	// 计算日收益率和波动率
	var dailyReturns []decimal.Decimal
	var maxPrice = startPrice
	var minPrice = startPrice
	var totalVolume int64

	for i := 1; i < len(prices); i++ {
		prevClose := prices[i-1].ClosePrice
		currClose := prices[i].ClosePrice

		if prevClose.IsPositive() {
			dailyReturn := currClose.Sub(prevClose).Div(prevClose)
			dailyReturns = append(dailyReturns, dailyReturn)
		}

		if currClose.GreaterThan(maxPrice) {
			maxPrice = currClose
		}
		if currClose.LessThan(minPrice) {
			minPrice = currClose
		}

		totalVolume += prices[i].Volume
	}

	// 计算平均日收益率
	avgDailyReturn := decimal.Zero
	if len(dailyReturns) > 0 {
		sum := decimal.Zero
		for _, r := range dailyReturns {
			sum = sum.Add(r)
		}
		avgDailyReturn = sum.Div(decimal.NewFromInt(int64(len(dailyReturns)))).Mul(decimal.NewFromInt(100))
	}

	// 计算年化波动率
	volatility := decimal.Zero
	if len(dailyReturns) > 1 {
		mean := decimal.Zero
		for _, r := range dailyReturns {
			mean = mean.Add(r)
		}
		mean = mean.Div(decimal.NewFromInt(int64(len(dailyReturns))))

		variance := decimal.Zero
		for _, r := range dailyReturns {
			diff := r.Sub(mean)
			variance = variance.Add(diff.Mul(diff))
		}
		variance = variance.Div(decimal.NewFromInt(int64(len(dailyReturns) - 1)))

		// 年化波动率 = 日波动率 * sqrt(252)
		volatility = decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64() * 252)).Mul(decimal.NewFromInt(100))
	}

	// 计算最大回撤
	maxDrawdown := calculateMaxDrawdown(prices)

	// 计算夏普比率 (假设无风险利率4%)
	sharpeRatio := decimal.Zero
	if volatility.IsPositive() {
		excessReturn := avgDailyReturn.Mul(decimal.NewFromInt(252)).Sub(decimal.NewFromInt(4))
		sharpeRatio = excessReturn.Div(volatility)
	}

	avgVolume := int64(0)
	if len(prices) > 0 {
		avgVolume = totalVolume / int64(len(prices))
	}

	return &ETFMetrics{
		Symbol:         symbol,
		Period:         period,
		StartPrice:     startPrice,
		EndPrice:       endPrice,
		TotalReturn:    totalReturn,
		AvgDailyReturn: avgDailyReturn,
		Volatility:     volatility,
		MaxPrice:       maxPrice,
		MinPrice:       minPrice,
		AvgVolume:      avgVolume,
		TradingDays:    len(prices),
		MaxDrawdown:    maxDrawdown,
		SharpeRatio:    sharpeRatio,
	}, nil
}

// AnalyzePortfolio 分析投资组合
func (s *ETFAnalysisService) AnalyzePortfolio(allocation map[string]float64, totalInvestment decimal.Decimal, taxRate decimal.Decimal) (*PortfolioAnalysis, error) {
	if taxRate.IsZero() {
		taxRate = decimal.NewFromFloat(0.10) // 默认10%税率
	}

	result := &PortfolioAnalysis{
		TotalInvestment: totalInvestment,
		BaseCurrency:    "USD",
		Allocation:      allocation,
		Holdings:        []PortfolioHolding{},
		ExchangeRates:   make(map[string]float64),
		TaxRate:         taxRate,
	}

	var totalValueUSD decimal.Decimal
	var totalAnnualDividendBeforeTax decimal.Decimal
	var totalAnnualDividendAfterTax decimal.Decimal
	var totalCapitalGain decimal.Decimal

	for symbol, weight := range allocation {
		weightDecimal := decimal.NewFromFloat(weight).Div(decimal.NewFromInt(100))
		if weightDecimal.IsZero() {
			continue
		}

		// 从数据库获取最新价格
		var etfData models.ETFData
		var currentPrice decimal.Decimal
		var totalReturnPercent decimal.Decimal
		var volatility decimal.Decimal

		if err := models.DB.Where("symbol = ?", symbol).Order("date DESC").First(&etfData).Error; err == nil {
			currentPrice = etfData.ClosePrice
		} else {
			currentPrice = decimal.NewFromFloat(100.0)
		}

		// 从历史数据计算收益率
		var prices []models.ETFData
		if err := models.DB.Where("symbol = ?", symbol).Order("date DESC").Limit(252).Find(&prices).Error; err == nil && len(prices) > 1 {
			metrics, _ := s.CalculateMetrics(symbol, prices, "1y")
			if metrics != nil {
				totalReturnPercent = metrics.TotalReturn
				volatility = metrics.Volatility
			}
		}

		// 从数据库获取 ETF 配置信息
		var etfConfig models.ETFConfig
		etfName := symbol + " ETF"
		if err := models.DB.Where("symbol = ?", symbol).First(&etfConfig).Error; err == nil {
			etfName = etfConfig.Name
		}

		// 计算投资金额
		investmentUSD := totalInvestment.Mul(weightDecimal)

		// 根据 ETF 类型设置合理的默认股息率
		defaultDividendYield := getDividendYieldByCategory(symbol, etfConfig.Category)

		var shares decimal.Decimal
		if currentPrice.IsPositive() {
			shares = investmentUSD.Div(currentPrice)
		}

		// 计算当前价值 = 股数 × 当前价格
		currentValueUSD := shares.Mul(currentPrice)
		if currentValueUSD.IsZero() {
			currentValueUSD = investmentUSD
		}

		// 计算资本利得 = 当前价值 - 初始投资
		capitalGain := currentValueUSD.Sub(investmentUSD)
		capitalGainPercent := decimal.Zero
		if investmentUSD.IsPositive() {
			capitalGainPercent = capitalGain.Div(investmentUSD).Mul(decimal.NewFromInt(100))
		}

		// 如果有历史收益率，使用历史收益率计算资本利得
		if !totalReturnPercent.IsZero() {
			capitalGainPercent = totalReturnPercent
			capitalGain = investmentUSD.Mul(totalReturnPercent.Div(decimal.NewFromInt(100)))
			currentValueUSD = investmentUSD.Add(capitalGain)
		}

		// 计算股息（使用默认股息率）
		annualDividendBeforeTax := investmentUSD.Mul(defaultDividendYield)
		annualDividendAfterTax := annualDividendBeforeTax.Mul(decimal.NewFromInt(1).Sub(taxRate))

		// 计算总收益 = 资本利得 + 税后股息
		totalReturn := capitalGain.Add(annualDividendAfterTax)

		holding := PortfolioHolding{
			Symbol:                  symbol,
			Name:                    etfName,
			Currency:                "USD",
			Weight:                  weightDecimal.Mul(decimal.NewFromInt(100)).InexactFloat64(),
			Investment:              investmentUSD.InexactFloat64(),
			InvestmentUSD:           investmentUSD.InexactFloat64(),
			Shares:                  shares.InexactFloat64(),
			CurrentPrice:            currentPrice.InexactFloat64(),
			CurrentValue:            currentValueUSD.InexactFloat64(),
			CurrentValueUSD:         currentValueUSD.InexactFloat64(),
			DividendYield:           defaultDividendYield.InexactFloat64(),
			AnnualDividendBeforeTax: annualDividendBeforeTax.InexactFloat64(),
			AnnualDividendAfterTax:  annualDividendAfterTax.InexactFloat64(),
			CapitalGain:             capitalGain.InexactFloat64(),
			CapitalGainPercent:      capitalGainPercent.InexactFloat64(),
			TotalReturn:             totalReturn.InexactFloat64(),
			Volatility:              volatility.InexactFloat64(),
		}

		result.Holdings = append(result.Holdings, holding)
		totalValueUSD = totalValueUSD.Add(currentValueUSD)
		totalAnnualDividendBeforeTax = totalAnnualDividendBeforeTax.Add(annualDividendBeforeTax)
		totalAnnualDividendAfterTax = totalAnnualDividendAfterTax.Add(annualDividendAfterTax)
		totalCapitalGain = totalCapitalGain.Add(capitalGain)

		// 计算加权股息率
		result.WeightedDividendYield = result.WeightedDividendYield.Add(
			defaultDividendYield.Mul(weightDecimal),
		)
	}

	result.TotalValue = totalValueUSD
	result.TotalValueUSD = totalValueUSD
	result.TotalReturn = totalCapitalGain
	result.TotalReturnPercent = totalCapitalGain.Div(totalInvestment).Mul(decimal.NewFromInt(100))
	result.AnnualDividendBeforeTax = totalAnnualDividendBeforeTax
	result.AnnualDividendAfterTax = totalAnnualDividendAfterTax
	result.DividendTax = totalAnnualDividendBeforeTax.Sub(totalAnnualDividendAfterTax)
	result.TotalReturnWithDividend = totalCapitalGain.Add(totalAnnualDividendAfterTax)
	result.TotalReturnWithDividendPercent = result.TotalReturnWithDividend.Div(totalInvestment).Mul(decimal.NewFromInt(100))

	return result, nil
}

// getDividendYieldByCategory 根据 ETF 类型返回合理的默认股息率
func getDividendYieldByCategory(symbol string, category string) decimal.Decimal {
	// 高股息 ETF
	if symbol == "SCHD" || symbol == "VYM" || symbol == "SPYD" || symbol == "HDV" || symbol == "DGRO" {
		return decimal.NewFromFloat(0.035) // 3.5%
	}
	// 覆盖收益型 ETF
	if symbol == "JEPI" || symbol == "JEPQ" || symbol == "QYLD" || symbol == "XYLD" {
		return decimal.NewFromFloat(0.07) // 7%
	}
	// 债券 ETF
	if symbol == "BND" || symbol == "AGG" || symbol == "TLT" || symbol == "BND" {
		return decimal.NewFromFloat(0.04) // 4%
	}
	// 房地产 ETF
	if symbol == "VNQ" {
		return decimal.NewFromFloat(0.04) // 4%
	}
	// 黄金 ETF
	if symbol == "GLD" {
		return decimal.NewFromFloat(0.00) // 0%
	}
	// 宽基指数 ETF
	if symbol == "QQQ" || symbol == "VOO" || symbol == "VTI" || symbol == "SPY" {
		return decimal.NewFromFloat(0.015) // 1.5%
	}
	// 国际市场 ETF
	if symbol == "VEA" || symbol == "VWO" || symbol == "VXUS" {
		return decimal.NewFromFloat(0.03) // 3%
	}
	// 默认
	return decimal.NewFromFloat(0.02) // 2%
}

// ForecastETFGrowth 预测ETF增长
func (s *ETFAnalysisService) ForecastETFGrowth(symbol string, initialInvestment decimal.Decimal, annualReturnRate *decimal.Decimal, taxRate decimal.Decimal) (*ForecastResult, error) {
	if taxRate.IsZero() {
		taxRate = decimal.NewFromFloat(0.10)
	}

	// 获取实时数据（移除缓存后需要重新设计）
	// realtimeData, err := s.cacheService.GetRealtimeData(symbol)
	// if err != nil {
	//	return nil, err
	// }
	// dividendYield := decimal.NewFromFloat(realtimeData.DividendYield).Div(decimal.NewFromInt(100))

	// 默认股息率
	dividendYield := decimal.NewFromFloat(0.03) // 3% 默认股息率
	// TODO: 从数据库获取真实股息率数据

	// 如果没有提供年化收益率，使用默认值
	if annualReturnRate == nil {
		// 获取历史数据计算
		var prices []models.ETFData
		if err := models.DB.Where("symbol = ?", symbol).Order("date DESC").Limit(252).Find(&prices).Error; err == nil && len(prices) > 0 {
			metrics, _ := s.CalculateMetrics(symbol, prices, "1y")
			if metrics != nil {
				annualReturnRate = &metrics.TotalReturn
			}
		}

		if annualReturnRate == nil {
			defaultRate := decimal.NewFromFloat(0.08)
			annualReturnRate = &defaultRate
		} else {
			// 转换为小数
			rate := annualReturnRate.Div(decimal.NewFromInt(100))
			annualReturnRate = &rate
		}
	}

	result := &ForecastResult{
		Symbol:            symbol,
		InitialInvestment: initialInvestment,
		AnnualReturnRate:  annualReturnRate.Mul(decimal.NewFromInt(100)),
		DividendYield:     dividendYield.Mul(decimal.NewFromInt(100)),
		TaxRate:           taxRate.Mul(decimal.NewFromInt(100)),
		Forecasts:         make(map[string]YearlyForecast),
	}

	// 计算3年、5年、10年预测
	years := []int{3, 5, 10}
	for _, year := range years {
		forecast := s.calculateYearlyForecast(
			initialInvestment,
			*annualReturnRate,
			dividendYield,
			taxRate,
			year,
		)
		result.Forecasts[fmt.Sprintf("%d", year)] = forecast
	}

	return result, nil
}

// calculateYearlyForecast 计算年度预测
func (s *ETFAnalysisService) calculateYearlyForecast(
	initialInvestment decimal.Decimal,
	annualReturnRate decimal.Decimal,
	dividendYield decimal.Decimal,
	taxRate decimal.Decimal,
	years int,
) YearlyForecast {
	// 未来价值（复利增长）
	futureValue := initialInvestment.Mul(
		decimal.NewFromFloat(1).Add(annualReturnRate).Pow(decimal.NewFromInt(int64(years))),
	)

	capitalAppreciation := futureValue.Sub(initialInvestment)

	// 计算累计股息
	var totalDividendBeforeTax decimal.Decimal
	var totalDividendAfterTax decimal.Decimal

	for year := 1; year <= years; year++ {
		yearValue := initialInvestment.Mul(
			decimal.NewFromFloat(1).Add(annualReturnRate).Pow(decimal.NewFromInt(int64(year))),
		)
		annualDividend := yearValue.Mul(dividendYield)
		totalDividendBeforeTax = totalDividendBeforeTax.Add(annualDividend)
		totalDividendAfterTax = totalDividendAfterTax.Add(
			annualDividend.Mul(decimal.NewFromInt(1).Sub(taxRate)),
		)
	}

	dividendTax := totalDividendBeforeTax.Sub(totalDividendAfterTax)
	totalReturnAfterTax := capitalAppreciation.Add(totalDividendAfterTax)

	// 有效年化收益率
	totalValue := initialInvestment.Add(totalReturnAfterTax)
	effectiveAnnualReturn := decimal.NewFromFloat(
		math.Pow(totalValue.Div(initialInvestment).InexactFloat64(), 1.0/float64(years)) - 1,
	).Mul(decimal.NewFromInt(100))

	// 第一年股息
	firstYearValue := initialInvestment.Mul(decimal.NewFromFloat(1).Add(annualReturnRate))
	firstYearDividendBeforeTax := firstYearValue.Mul(dividendYield)
	firstYearDividendAfterTax := firstYearDividendBeforeTax.Mul(decimal.NewFromInt(1).Sub(taxRate))

	return YearlyForecast{
		Years:                   years,
		FutureValue:             futureValue,
		CapitalAppreciation:     capitalAppreciation,
		TotalDividendBeforeTax:  totalDividendBeforeTax,
		TotalDividendAfterTax:   totalDividendAfterTax,
		AnnualDividendBeforeTax: firstYearDividendBeforeTax,
		AnnualDividendAfterTax:  firstYearDividendAfterTax,
		DividendTax:             dividendTax,
		TotalReturnAfterTax:     totalReturnAfterTax,
		EffectiveAnnualReturn:   effectiveAnnualReturn,
	}
}

// calculateMaxDrawdown 计算最大回撤
func calculateMaxDrawdown(prices []models.ETFData) decimal.Decimal {
	if len(prices) == 0 {
		return decimal.Zero
	}

	maxDrawdown := decimal.Zero
	peak := prices[0].ClosePrice

	for _, price := range prices {
		if price.ClosePrice.GreaterThan(peak) {
			peak = price.ClosePrice
		}

		if peak.IsPositive() {
			drawdown := peak.Sub(price.ClosePrice).Div(peak).Mul(decimal.NewFromInt(100))
			if drawdown.GreaterThan(maxDrawdown) {
				maxDrawdown = drawdown
			}
		}
	}

	return maxDrawdown.Neg() // 返回负值表示回撤
}

// getETFCurrency 获取ETF计价货币
func (s *ETFAnalysisService) getETFCurrency(symbol string) string {
	var cfg models.ETFConfig
	if err := models.DB.Where("symbol = ?", symbol).First(&cfg).Error; err == nil {
		return cfg.Currency
	}
	return "USD"
}

// convertToUSD 转换为美元
func (s *ETFAnalysisService) convertToUSD(amount decimal.Decimal, currency string) decimal.Decimal {
	if currency == "USD" {
		return amount
	}

	rate := s.exchangeRate.GetRate(currency, "USD")
	return amount.Mul(decimal.NewFromFloat(rate))
}

// GetComparisonData 获取ETF对比数据
func (s *ETFAnalysisService) GetComparisonData(symbols []string, period string) ([]map[string]interface{}, error) {
	// 批量获取ETF配置
	var etfConfigs []models.ETFConfig
	models.DB.Where("symbol IN ?", symbols).Find(&etfConfigs)

	// 构建配置映射
	configMap := make(map[string]models.ETFConfig)
	for _, cfg := range etfConfigs {
		configMap[cfg.Symbol] = cfg
	}

	// 批量获取每个ETF的最新行情数据
	type LatestETFData struct {
		Symbol     string
		OpenPrice  decimal.Decimal
		ClosePrice decimal.Decimal
		HighPrice  decimal.Decimal
		LowPrice   decimal.Decimal
		Volume     int64
		Date       time.Time
	}

	var latestData []LatestETFData
	for _, symbol := range symbols {
		var data LatestETFData
		models.DB.Model(&models.ETFData{}).
			Select("symbol, open_price, close_price, high_price, low_price, volume, date").
			Where("symbol = ?", symbol).
			Order("date DESC").
			First(&data)
		if data.Symbol != "" {
			latestData = append(latestData, data)
		}
	}

	// 构建最新数据映射
	latestDataMap := make(map[string]LatestETFData)
	for _, data := range latestData {
		latestDataMap[data.Symbol] = data
	}

	// 批量获取历史数据
	var allPrices []models.ETFData
	models.DB.Where("symbol IN ?", symbols).Order("symbol, date DESC").Find(&allPrices)

	// 构建历史数据映射
	pricesMap := make(map[string][]models.ETFData)
	for _, price := range allPrices {
		pricesMap[price.Symbol] = append(pricesMap[price.Symbol], price)
	}

	// 构建结果
	var results []map[string]interface{}
	for _, symbol := range symbols {
		// 获取ETF配置
		cfg, ok := configMap[symbol]
		if !ok {
			continue
		}

		// 获取最新行情数据
		data, dataOk := latestDataMap[symbol]
		if !dataOk {
			continue
		}

		// 计算涨跌幅 - 基于前一日收盘价，而非当日OpenPrice
		change := decimal.Zero
		changePercent := decimal.Zero

		// 从数据库获取前一日数据
		var prevData models.ETFData
		if err := models.DB.Where("symbol = ? AND date < ?", symbol, data.Date).
			Order("date DESC").First(&prevData).Error; err == nil && prevData.ID > 0 {
			previousClose := prevData.ClosePrice
			change = data.ClosePrice.Sub(previousClose)
			if previousClose.GreaterThan(decimal.Zero) {
				changePercent = change.Div(previousClose).Mul(decimal.NewFromInt(100))
			}
		} else {
			// 没有前日数据，使用当日开盘价作为近似
			change = data.ClosePrice.Sub(data.OpenPrice)
			if data.OpenPrice.GreaterThan(decimal.Zero) {
				changePercent = change.Div(data.OpenPrice).Mul(decimal.NewFromInt(100))
			}
		}

		// 获取历史数据
		prices := pricesMap[symbol]
		if len(prices) > 252 {
			prices = prices[:252] // 限制数量
		}

		// 计算指标
		var metrics *ETFMetrics
		if len(prices) > 0 {
			metrics, _ = s.CalculateMetrics(symbol, prices, period)
		}

		// 构建结果
		result := map[string]interface{}{
			"symbol":         symbol,
			"name":           cfg.Name,
			"current_price":  data.ClosePrice.InexactFloat64(),
			"change":         change.InexactFloat64(),
			"change_percent": changePercent.InexactFloat64(),
			"volume":         data.Volume,
			"strategy":       cfg.Strategy,
			"focus":          cfg.Focus,
			"expense_ratio":  cfg.ExpenseRatio,
		}

		if metrics != nil {
			result["total_return"] = metrics.TotalReturn
			result["volatility"] = metrics.Volatility
			result["max_drawdown"] = metrics.MaxDrawdown
			result["sharpe_ratio"] = metrics.SharpeRatio
		}

		results = append(results, result)
	}

	return results, nil
}

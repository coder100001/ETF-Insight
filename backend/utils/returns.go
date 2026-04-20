package utils

import (
	"etf-insight/models"

	"github.com/shopspring/decimal"
)

// CalculateReturnsFromETFData 从ETF历史数据计算收益率序列
// 按时间正序计算: returns[i] = (price[i] - price[i-1]) / price[i-1]
func CalculateReturnsFromETFData(etfData []models.ETFData) []decimal.Decimal {
	if len(etfData) < 2 {
		return []decimal.Decimal{}
	}

	// 确保数据按时间正序排列(从早到晚)
	// 如果数据是倒序的，需要反转
	sortedData := make([]models.ETFData, len(etfData))
	copy(sortedData, etfData)

	// 检查是否需要反转(如果第一个日期晚于最后一个日期)
	if len(sortedData) > 1 && sortedData[0].Date.After(sortedData[len(sortedData)-1].Date) {
		// 数据是倒序的，需要反转
		for i, j := 0, len(sortedData)-1; i < j; i, j = i+1, j-1 {
			sortedData[i], sortedData[j] = sortedData[j], sortedData[i]
		}
	}

	returns := make([]decimal.Decimal, 0, len(sortedData)-1)
	for i := 1; i < len(sortedData); i++ {
		currentPrice := sortedData[i].ClosePrice
		previousPrice := sortedData[i-1].ClosePrice
		if previousPrice.GreaterThan(decimal.Zero) {
			ret := currentPrice.Sub(previousPrice).Div(previousPrice)
			returns = append(returns, ret)
		}
	}

	return returns
}

// CalculateReturnsFromPrices 从价格序列计算收益率
// prices 应该按时间正序排列(从早到晚)
func CalculateReturnsFromPrices(prices []decimal.Decimal) []decimal.Decimal {
	if len(prices) < 2 {
		return []decimal.Decimal{}
	}

	returns := make([]decimal.Decimal, 0, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		currentPrice := prices[i]
		previousPrice := prices[i-1]
		if previousPrice.GreaterThan(decimal.Zero) {
			ret := currentPrice.Sub(previousPrice).Div(previousPrice)
			returns = append(returns, ret)
		}
	}

	return returns
}

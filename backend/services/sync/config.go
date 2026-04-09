package sync

import (
	"etf-insight/services/datasource"
)

// ETFPortfolio 预定义的ETF投资组合
type ETFPortfolio struct {
	Name        string
	Description string
	ETFs        []datasource.ETFInfo
}

// GetDefaultETFList 获取默认ETF列表
func GetDefaultETFList() []datasource.ETFInfo {
	return []datasource.ETFInfo{
		{
			Symbol:       "QQQ",
			Name:         "Invesco QQQ Trust",
			Category:     "大盘股",
			Description:  "追踪纳斯达克100指数",
			Provider:     "Invesco",
			ExpenseRatio: 0.0020,
		},
		{
			Symbol:       "SCHD",
			Name:         "Schwab US Dividend Equity ETF",
			Category:     "股息",
			Description:  "美国股息股票ETF",
			Provider:     "Schwab",
			ExpenseRatio: 0.0006,
		},
		{
			Symbol:       "VNQ",
			Name:         "Vanguard Real Estate ETF",
			Category:     "REITs",
			Description:  "房地产投资信托ETF",
			Provider:     "Vanguard",
			ExpenseRatio: 0.0012,
		},
		{
			Symbol:       "VYM",
			Name:         "Vanguard High Dividend Yield ETF",
			Category:     "股息",
			Description:  "高股息收益ETF",
			Provider:     "Vanguard",
			ExpenseRatio: 0.0006,
		},
		{
			Symbol:       "SPYD",
			Name:         "SPDR Portfolio S&P 500 High Dividend ETF",
			Category:     "股息",
			Description:  "S&P 500高股息ETF",
			Provider:     "SPDR",
			ExpenseRatio: 0.0035,
		},
		{
			Symbol:       "JEPQ",
			Name:         "JPMorgan Nasdaq Equity Premium Income ETF",
			Category:     "备兑认购",
			Description:  "纳斯达克备兑认购收入ETF",
			Provider:     "JPMorgan",
			ExpenseRatio: 0.0035,
		},
		{
			Symbol:       "JEPI",
			Name:         "JPMorgan Equity Premium Income ETF",
			Category:     "备兑认购",
			Description:  "股票备兑认购收入ETF",
			Provider:     "JPMorgan",
			ExpenseRatio: 0.0035,
		},
		{
			Symbol:       "VTI",
			Name:         "Vanguard Total Stock Market ETF",
			Category:     "全市场",
			Description:  "全美股票市场ETF",
			Provider:     "Vanguard",
			ExpenseRatio: 0.0003,
		},
		{
			Symbol:       "VOO",
			Name:         "Vanguard S&P 500 ETF",
			Category:     "大盘股",
			Description:  "S&P 500指数ETF",
			Provider:     "Vanguard",
			ExpenseRatio: 0.0003,
		},
		{
			Symbol:       "VEA",
			Name:         "Vanguard FTSE Developed Markets ETF",
			Category:     "国际",
			Description:  "发达市场ETF",
			Provider:     "Vanguard",
			ExpenseRatio: 0.0005,
		},
		{
			Symbol:       "VWO",
			Name:         "Vanguard FTSE Emerging Markets ETF",
			Category:     "新兴市场",
			Description:  "新兴市场ETF",
			Provider:     "Vanguard",
			ExpenseRatio: 0.0010,
		},
		{
			Symbol:       "BND",
			Name:         "Vanguard Total Bond Market ETF",
			Category:     "债券",
			Description:  "全美债券市场ETF",
			Provider:     "Vanguard",
			ExpenseRatio: 0.0003,
		},
		{
			Symbol:       "AGG",
			Name:         "iShares Core U.S. Aggregate Bond ETF",
			Category:     "债券",
			Description:  "美国综合债券ETF",
			Provider:     "iShares",
			ExpenseRatio: 0.0003,
		},
		{
			Symbol:       "GLD",
			Name:         "SPDR Gold Shares",
			Category:     "商品",
			Description:  "黄金ETF",
			Provider:     "SPDR",
			ExpenseRatio: 0.0040,
		},
		{
			Symbol:       "TLT",
			Name:         "iShares 20+ Year Treasury Bond ETF",
			Category:     "国债",
			Description:  "20年以上美国国债ETF",
			Provider:     "iShares",
			ExpenseRatio: 0.0015,
		},
	}
}

// GetDividendPortfolio 获取股息投资组合
func GetDividendPortfolio() ETFPortfolio {
	allETFs := GetDefaultETFList()
	var dividendETFs []datasource.ETFInfo

	for _, etf := range allETFs {
		if etf.Category == "股息" || etf.Category == "备兑认购" {
			dividendETFs = append(dividendETFs, etf)
		}
	}

	return ETFPortfolio{
		Name:        "股息收益组合",
		Description: "专注于高股息收益的ETF组合",
		ETFs:        dividendETFs,
	}
}

// GetGrowthPortfolio 获取成长投资组合
func GetGrowthPortfolio() ETFPortfolio {
	allETFs := GetDefaultETFList()
	var growthETFs []datasource.ETFInfo

	for _, etf := range allETFs {
		if etf.Category == "大盘股" || etf.Category == "全市场" {
			growthETFs = append(growthETFs, etf)
		}
	}

	return ETFPortfolio{
		Name:        "成长型组合",
		Description: "专注于成长型股票的ETF组合",
		ETFs:        growthETFs,
	}
}

// GetGlobalPortfolio 获取全球配置组合
func GetGlobalPortfolio() ETFPortfolio {
	allETFs := GetDefaultETFList()
	var globalETFs []datasource.ETFInfo

	for _, etf := range allETFs {
		if etf.Category == "国际" || etf.Category == "新兴市场" {
			globalETFs = append(globalETFs, etf)
		}
	}

	return ETFPortfolio{
		Name:        "全球配置组合",
		Description: "国际多元化ETF组合",
		ETFs:        globalETFs,
	}
}

package etf

import (
	"fmt"
	"net/http"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// UniversalETFService 通用ETF服务
type UniversalETFService struct {
	db         *gorm.DB
	httpClient *http.Client
}

// NewUniversalETFService 创建通用ETF服务
func NewUniversalETFService(db *gorm.DB) *UniversalETFService {
	return &UniversalETFService{
		db: db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ETFDataSource ETF数据源接口
type ETFDataSource interface {
	GetETFList() ([]models.UniversalETF, error)
	GetETFQuote(symbol string) (*models.UniversalETF, error)
	GetHistoricalData(symbol string, startDate, endDate time.Time) ([]models.ETFHistoricalData, error)
}

// InitializeDefaultETFs 初始化默认ETF数据
func (s *UniversalETFService) InitializeDefaultETFs() error {
	// 美股ETF
	usETFs := []models.UniversalETF{
		// 宽基指数
		{
			Symbol:       "VTI",
			Name:         "Vanguard Total Stock Market ETF",
			FullName:     "Vanguard Total Stock Market ETF",
			AssetClass:   models.AssetClassEquity,
			Region:       models.RegionUS,
			ETFType:      models.ETFTypeIndex,
			Exchange:     "NYSE",
			Currency:     "USD",
			Benchmark:    "CRSP US Total Market Index",
			Provider:     "Vanguard",
			ExpenseRatio: decimal.NewFromFloat(0.0003),
			Status:       1,
			DataSource:   "Yahoo Finance",
		},
		{
			Symbol:       "VOO",
			Name:         "Vanguard S&P 500 ETF",
			FullName:     "Vanguard S&P 500 ETF",
			AssetClass:   models.AssetClassEquity,
			Region:       models.RegionUS,
			ETFType:      models.ETFTypeIndex,
			Exchange:     "NYSE",
			Currency:     "USD",
			Benchmark:    "S&P 500 Index",
			Provider:     "Vanguard",
			ExpenseRatio: decimal.NewFromFloat(0.0003),
			Status:       1,
			DataSource:   "Yahoo Finance",
		},
		{
			Symbol:       "QQQ",
			Name:         "Invesco QQQ Trust",
			FullName:     "Invesco QQQ Trust Series 1",
			AssetClass:   models.AssetClassEquity,
			Region:       models.RegionUS,
			ETFType:      models.ETFTypeIndex,
			Exchange:     "NASDAQ",
			Currency:     "USD",
			Benchmark:    "NASDAQ-100 Index",
			Provider:     "Invesco",
			ExpenseRatio: decimal.NewFromFloat(0.0020),
			Status:       1,
			DataSource:   "Yahoo Finance",
		},
		// 债券ETF
		{
			Symbol:       "AGG",
			Name:         "iShares Core U.S. Aggregate Bond ETF",
			FullName:     "iShares Core U.S. Aggregate Bond ETF",
			AssetClass:   models.AssetClassBond,
			Region:       models.RegionUS,
			ETFType:      models.ETFTypeIndex,
			Exchange:     "NYSE",
			Currency:     "USD",
			Benchmark:    "Bloomberg U.S. Aggregate Bond Index",
			Provider:     "BlackRock",
			ExpenseRatio: decimal.NewFromFloat(0.0003),
			Status:       1,
			DataSource:   "Yahoo Finance",
		},
		{
			Symbol:       "TLT",
			Name:         "iShares 20+ Year Treasury Bond ETF",
			FullName:     "iShares 20+ Year Treasury Bond ETF",
			AssetClass:   models.AssetClassBond,
			Region:       models.RegionUS,
			ETFType:      models.ETFTypeIndex,
			Exchange:     "NASDAQ",
			Currency:     "USD",
			Benchmark:    "ICE U.S. Treasury 20+ Year Bond Index",
			Provider:     "BlackRock",
			ExpenseRatio: decimal.NewFromFloat(0.0015),
			Status:       1,
			DataSource:   "Yahoo Finance",
		},
		// 商品ETF
		{
			Symbol:       "GLD",
			Name:         "SPDR Gold Shares",
			FullName:     "SPDR Gold Shares",
			AssetClass:   models.AssetClassCommodity,
			Region:       models.RegionGlobal,
			ETFType:      models.ETFTypeIndex,
			Exchange:     "NYSE",
			Currency:     "USD",
			Benchmark:    "Gold Price",
			Provider:     "State Street",
			ExpenseRatio: decimal.NewFromFloat(0.0040),
			Status:       1,
			DataSource:   "Yahoo Finance",
		},
		{
			Symbol:       "USO",
			Name:         "United States Oil Fund",
			FullName:     "United States Oil Fund, LP",
			AssetClass:   models.AssetClassCommodity,
			Region:       models.RegionGlobal,
			ETFType:      models.ETFTypeIndex,
			Exchange:     "NYSE",
			Currency:     "USD",
			Benchmark:    "WTI Crude Oil",
			Provider:     "US Commodity Funds",
			ExpenseRatio: decimal.NewFromFloat(0.0079),
			Status:       1,
			DataSource:   "Yahoo Finance",
		},
		// REIT ETF
		{
			Symbol:       "VNQ",
			Name:         "Vanguard Real Estate ETF",
			FullName:     "Vanguard Real Estate ETF",
			AssetClass:   models.AssetClassREIT,
			Region:       models.RegionUS,
			ETFType:      models.ETFTypeIndex,
			Exchange:     "NYSE",
			Currency:     "USD",
			Benchmark:    "MSCI US Investable Market Real Estate 25/50 Index",
			Provider:     "Vanguard",
			ExpenseRatio: decimal.NewFromFloat(0.0012),
			Status:       1,
			DataSource:   "Yahoo Finance",
		},
		// 国际ETF
		{
			Symbol:       "VEA",
			Name:         "Vanguard Developed Markets ETF",
			FullName:     "Vanguard FTSE Developed Markets ETF",
			AssetClass:   models.AssetClassEquity,
			Region:       models.RegionGlobal,
			ETFType:      models.ETFTypeIndex,
			Exchange:     "NYSE",
			Currency:     "USD",
			Benchmark:    "FTSE Developed All Cap ex US Index",
			Provider:     "Vanguard",
			ExpenseRatio: decimal.NewFromFloat(0.0005),
			Status:       1,
			DataSource:   "Yahoo Finance",
		},
		{
			Symbol:       "VWO",
			Name:         "Vanguard Emerging Markets ETF",
			FullName:     "Vanguard FTSE Emerging Markets ETF",
			AssetClass:   models.AssetClassEquity,
			Region:       models.RegionEmerging,
			ETFType:      models.ETFTypeIndex,
			Exchange:     "NYSE",
			Currency:     "USD",
			Benchmark:    "FTSE Emerging Markets All Cap China A Inclusion Index",
			Provider:     "Vanguard",
			ExpenseRatio: decimal.NewFromFloat(0.0010),
			Status:       1,
			DataSource:   "Yahoo Finance",
		},
		// 行业ETF
		{
			Symbol:       "XLK",
			Name:         "Technology Select Sector SPDR",
			FullName:     "Technology Select Sector SPDR Fund",
			AssetClass:   models.AssetClassEquity,
			Region:       models.RegionUS,
			ETFType:      models.ETFTypeSector,
			Sector:       "Technology",
			Exchange:     "NYSE",
			Currency:     "USD",
			Benchmark:    "Technology Select Sector Index",
			Provider:     "State Street",
			ExpenseRatio: decimal.NewFromFloat(0.0009),
			Status:       1,
			DataSource:   "Yahoo Finance",
		},
		{
			Symbol:       "XLF",
			Name:         "Financial Select Sector SPDR",
			FullName:     "Financial Select Sector SPDR Fund",
			AssetClass:   models.AssetClassEquity,
			Region:       models.RegionUS,
			ETFType:      models.ETFTypeSector,
			Sector:       "Financials",
			Exchange:     "NYSE",
			Currency:     "USD",
			Benchmark:    "Financial Select Sector Index",
			Provider:     "State Street",
			ExpenseRatio: decimal.NewFromFloat(0.0009),
			Status:       1,
			DataSource:   "Yahoo Finance",
		},
		// 因子ETF
		{
			Symbol:        "SCHD",
			Name:          "Schwab US Dividend Equity ETF",
			FullName:      "Schwab U.S. Dividend Equity ETF",
			AssetClass:    models.AssetClassEquity,
			Region:        models.RegionUS,
			ETFType:       models.ETFTypeFactor,
			Exchange:      "NYSE",
			Currency:      "USD",
			Benchmark:     "Dow Jones U.S. Dividend 100 Index",
			Provider:      "Charles Schwab",
			ExpenseRatio:  decimal.NewFromFloat(0.0006),
			DividendYield: decimal.NewFromFloat(0.035),
			Status:        1,
			DataSource:    "Yahoo Finance",
		},
		{
			Symbol:       "USMV",
			Name:         "iShares MSCI USA Min Vol Factor ETF",
			FullName:     "iShares MSCI USA Min Vol Factor ETF",
			AssetClass:   models.AssetClassEquity,
			Region:       models.RegionUS,
			ETFType:      models.ETFTypeFactor,
			Exchange:     "NYSE",
			Currency:     "USD",
			Benchmark:    "MSCI USA Minimum Volatility Index",
			Provider:     "BlackRock",
			ExpenseRatio: decimal.NewFromFloat(0.0015),
			Status:       1,
			DataSource:   "Yahoo Finance",
		},
	}

	// A股ETF
	chinaETFs := []models.UniversalETF{
		{
			Symbol:       "510300",
			Name:         "华泰柏瑞沪深300ETF",
			FullName:     "华泰柏瑞沪深300交易型开放式指数证券投资基金",
			AssetClass:   models.AssetClassEquity,
			Region:       models.RegionChina,
			ETFType:      models.ETFTypeIndex,
			Exchange:     "SSE",
			Currency:     "CNY",
			Benchmark:    "沪深300指数",
			Provider:     "华泰柏瑞基金",
			ExpenseRatio: decimal.NewFromFloat(0.005),
			Status:       1,
			DataSource:   "AKShare",
		},
		{
			Symbol:       "510050",
			Name:         "华夏上证50ETF",
			FullName:     "华夏上证50交易型开放式指数证券投资基金",
			AssetClass:   models.AssetClassEquity,
			Region:       models.RegionChina,
			ETFType:      models.ETFTypeIndex,
			Exchange:     "SSE",
			Currency:     "CNY",
			Benchmark:    "上证50指数",
			Provider:     "华夏基金",
			ExpenseRatio: decimal.NewFromFloat(0.005),
			Status:       1,
			DataSource:   "AKShare",
		},
		{
			Symbol:            "515080",
			Name:              "中证红利ETF",
			FullName:          "招商中证红利交易型开放式指数证券投资基金",
			AssetClass:        models.AssetClassEquity,
			Region:            models.RegionChina,
			ETFType:           models.ETFTypeFactor,
			Exchange:          "SSE",
			Currency:          "CNY",
			Benchmark:         "中证红利指数",
			Provider:          "招商基金",
			ExpenseRatio:      decimal.NewFromFloat(0.003),
			DividendFrequency: "季分",
			Status:            1,
			DataSource:        "AKShare",
		},
		{
			Symbol:            "510880",
			Name:              "红利ETF",
			FullName:          "华泰柏瑞上证红利交易型开放式指数证券投资基金",
			AssetClass:        models.AssetClassEquity,
			Region:            models.RegionChina,
			ETFType:           models.ETFTypeFactor,
			Exchange:          "SSE",
			Currency:          "CNY",
			Benchmark:         "上证红利指数",
			Provider:          "华泰柏瑞基金",
			ExpenseRatio:      decimal.NewFromFloat(0.006),
			DividendFrequency: "年分",
			Status:            1,
			DataSource:        "AKShare",
		},
	}

	// 合并所有ETF
	allETFs := append(usETFs, chinaETFs...)

	// 保存到数据库
	for _, etf := range allETFs {
		var existing models.UniversalETF
		result := s.db.Where("symbol = ?", etf.Symbol).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := s.db.Create(&etf).Error; err != nil {
				fmt.Printf("创建ETF失败 %s: %v\n", etf.Symbol, err)
			}
		} else if result.Error == nil {
			// 更新现有记录
			if err := s.db.Model(&existing).Updates(etf).Error; err != nil {
				fmt.Printf("更新ETF失败 %s: %v\n", etf.Symbol, err)
			}
		}
	}

	return nil
}

// GetETFBySymbol 根据代码获取ETF
func (s *UniversalETFService) GetETFBySymbol(symbol string) (*models.UniversalETF, error) {
	var etf models.UniversalETF
	result := s.db.Where("symbol = ? AND status = ?", symbol, 1).First(&etf)
	if result.Error != nil {
		return nil, result.Error
	}
	return &etf, nil
}

// GetETFsByAssetClass 根据资产类别获取ETF
func (s *UniversalETFService) GetETFsByAssetClass(assetClass models.AssetClass) ([]models.UniversalETF, error) {
	var etfs []models.UniversalETF
	result := s.db.Where("asset_class = ? AND status = ?", assetClass, 1).Find(&etfs)
	return etfs, result.Error
}

// GetETFsByRegion 根据地区获取ETF
func (s *UniversalETFService) GetETFsByRegion(region models.Region) ([]models.UniversalETF, error) {
	var etfs []models.UniversalETF
	result := s.db.Where("region = ? AND status = ?", region, 1).Find(&etfs)
	return etfs, result.Error
}

// GetETFsByType 根据类型获取ETF
func (s *UniversalETFService) GetETFsByType(etfType models.ETFType) ([]models.UniversalETF, error) {
	var etfs []models.UniversalETF
	result := s.db.Where("etf_type = ? AND status = ?", etfType, 1).Find(&etfs)
	return etfs, result.Error
}

// SearchETFs 搜索ETF
func (s *UniversalETFService) SearchETFs(keyword string) ([]models.UniversalETF, error) {
	var etfs []models.UniversalETF
	result := s.db.Where(
		"(symbol LIKE ? OR name LIKE ? OR full_name LIKE ?) AND status = ?",
		"%"+keyword+"%",
		"%"+keyword+"%",
		"%"+keyword+"%",
		1,
	).Find(&etfs)
	return etfs, result.Error
}

// GetAllETFs 获取所有ETF
func (s *UniversalETFService) GetAllETFs() ([]models.UniversalETF, error) {
	var etfs []models.UniversalETF
	result := s.db.Where("status = ?", 1).Find(&etfs)
	return etfs, result.Error
}

// GetETFsByFilter 根据多条件筛选ETF
func (s *UniversalETFService) GetETFsByFilter(filter ETFFilter) ([]models.UniversalETF, error) {
	query := s.db.Where("status = ?", 1)

	if filter.AssetClass != "" {
		query = query.Where("asset_class = ?", filter.AssetClass)
	}
	if filter.Region != "" {
		query = query.Where("region = ?", filter.Region)
	}
	if filter.ETFType != "" {
		query = query.Where("etf_type = ?", filter.ETFType)
	}
	if filter.Sector != "" {
		query = query.Where("sector = ?", filter.Sector)
	}
	if filter.Provider != "" {
		query = query.Where("provider = ?", filter.Provider)
	}
	if filter.Currency != "" {
		query = query.Where("currency = ?", filter.Currency)
	}
	if filter.MinExpenseRatio.GreaterThan(decimal.Zero) {
		query = query.Where("expense_ratio >= ?", filter.MinExpenseRatio)
	}
	if filter.MaxExpenseRatio.GreaterThan(decimal.Zero) {
		query = query.Where("expense_ratio <= ?", filter.MaxExpenseRatio)
	}

	var etfs []models.UniversalETF
	result := query.Find(&etfs)
	return etfs, result.Error
}

// ETFFilter ETF筛选条件
type ETFFilter struct {
	AssetClass      models.AssetClass
	Region          models.Region
	ETFType         models.ETFType
	Sector          string
	Provider        string
	Currency        string
	MinExpenseRatio decimal.Decimal
	MaxExpenseRatio decimal.Decimal
}

// GetAssetClassDistribution 获取资产类别分布
func (s *UniversalETFService) GetAssetClassDistribution() (map[string]int, error) {
	var results []struct {
		AssetClass string
		Count      int
	}

	result := s.db.Model(&models.UniversalETF{}).
		Select("asset_class, COUNT(*) as count").
		Where("status = ?", 1).
		Group("asset_class").
		Scan(&results)

	if result.Error != nil {
		return nil, result.Error
	}

	distribution := make(map[string]int)
	for _, r := range results {
		distribution[r.AssetClass] = r.Count
	}

	return distribution, nil
}

// GetRegionDistribution 获取地区分布
func (s *UniversalETFService) GetRegionDistribution() (map[string]int, error) {
	var results []struct {
		Region string
		Count  int
	}

	result := s.db.Model(&models.UniversalETF{}).
		Select("region, COUNT(*) as count").
		Where("status = ?", 1).
		Group("region").
		Scan(&results)

	if result.Error != nil {
		return nil, result.Error
	}

	distribution := make(map[string]int)
	for _, r := range results {
		distribution[r.Region] = r.Count
	}

	return distribution, nil
}

// CompareETFs 对比多个ETF
func (s *UniversalETFService) CompareETFs(symbols []string) ([]models.UniversalETF, error) {
	var etfs []models.UniversalETF
	result := s.db.Where("symbol IN ? AND status = ?", symbols, 1).Find(&etfs)
	return etfs, result.Error
}

// GetPortfolioAllocation 获取组合配置建议
func (s *UniversalETFService) GetPortfolioAllocation(strategy string) (map[string]float64, error) {
	// 基于不同策略的配置建议
	allocations := map[string]map[string]float64{
		"conservative": {
			"AGG": 40.0, // 债券
			"VTI": 30.0, // 美股
			"VNQ": 10.0, // REIT
			"GLD": 10.0, // 黄金
			"VEA": 10.0, // 国际股票
		},
		"balanced": {
			"VTI": 35.0,
			"VEA": 20.0,
			"VWO": 10.0,
			"AGG": 20.0,
			"VNQ": 10.0,
			"GLD": 5.0,
		},
		"aggressive": {
			"VTI": 40.0,
			"QQQ": 20.0,
			"VEA": 15.0,
			"VWO": 15.0,
			"VNQ": 10.0,
		},
		"dividend": {
			"SCHD": 40.0,
			"VTI":  20.0,
			"VNQ":  20.0,
			"AGG":  20.0,
		},
	}

	if allocation, ok := allocations[strategy]; ok {
		return allocation, nil
	}

	return nil, fmt.Errorf("未知的策略: %s", strategy)
}

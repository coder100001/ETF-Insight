package ashare

import (
	"etf-insight/constants"
	"fmt"
	"time"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ETFDataService A股ETF数据服务
type ETFDataService struct {
	db              *gorm.DB
	akshareProvider *AKShareProvider
	tushareProvider *TuShareProvider
	useAKShare      bool
	useTuShare      bool
}

// NewETFDataService 创建ETF数据服务
func NewETFDataService(db *gorm.DB) *ETFDataService {
	return &ETFDataService{
		db:              db,
		akshareProvider: NewAKShareProvider(""),
		useAKShare:      false, // 默认不启用，需要配置Python服务
		useTuShare:      false,
	}
}

// EnableAKShare 启用AKShare数据源
func (s *ETFDataService) EnableAKShare(baseURL string) {
	s.akshareProvider = NewAKShareProvider(baseURL)
	s.useAKShare = true
}

// EnableTuShare 启用TuShare数据源
func (s *ETFDataService) EnableTuShare(apiKey string) {
	s.tushareProvider = NewTuShareProvider(apiKey, "")
	s.useTuShare = true
}

// SyncETFList 同步ETF列表（只同步核心组合ETF）
func (s *ETFDataService) SyncETFList() error {
	if !s.useAKShare && !s.useTuShare {
		return fmt.Errorf("未启用任何数据源")
	}

	var etfInfos []ETFInfo
	var err error

	if s.useAKShare {
		etfInfos, err = s.akshareProvider.GetETFList()
		if err != nil {
			return fmt.Errorf("从AKShare获取ETF列表失败: %w", err)
		}
	}

	coreSet := make(map[string]bool, len(constants.CoreETFSymbols))
	for _, sym := range constants.CoreETFSymbols {
		coreSet[sym] = true
	}

	for _, info := range etfInfos {
		if !coreSet[info.Symbol] {
			continue
		}

		etf := &models.AShareDividendETF{
			Symbol:        info.Symbol,
			Name:          info.Name,
			Benchmark:     info.Benchmark,
			ManagementFee: info.ManagementFee,
			Exchange:      info.Exchange,
			Status:        1,
		}

		var existing models.AShareDividendETF
		result := s.db.Where("symbol = ?", info.Symbol).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := s.db.Create(etf).Error; err != nil {
				fmt.Printf("创建ETF记录失败 %s: %v\n", info.Symbol, err)
			}
		} else if result.Error == nil {
			if err := s.db.Model(&existing).Updates(etf).Error; err != nil {
				fmt.Printf("更新ETF记录失败 %s: %v\n", info.Symbol, err)
			}
		}
	}

	return nil
}

// SyncETFPrices 同步ETF价格
func (s *ETFDataService) SyncETFPrices(symbols []string) error {
	if !s.useAKShare {
		return fmt.Errorf("AKShare数据源未启用")
	}

	quotes, err := s.akshareProvider.GetETFQuotes(symbols)
	if err != nil {
		return fmt.Errorf("获取ETF行情失败: %w", err)
	}

	for symbol, quote := range quotes {
		var etf models.AShareDividendETF
		result := s.db.Where("symbol = ?", symbol).First(&etf)

		if result.Error != nil {
			continue
		}

		// 更新价格信息
		updates := map[string]any{
			"current_price":    quote.CurrentPrice,
			"previous_close":   quote.PreviousClose,
			"price_change":     quote.PriceChange,
			"price_change_pct": quote.PriceChangePct,
			"volume":           quote.Volume,
			"turnover":         quote.Turnover,
			"price_updated_at": quote.UpdateTime,
		}

		if err := s.db.Model(&etf).Updates(updates).Error; err != nil {
			fmt.Printf("更新ETF价格失败 %s: %v\n", symbol, err)
		}
	}

	return nil
}

// GetETFPrice 获取单个ETF价格
func (s *ETFDataService) GetETFPrice(symbol string) (*models.AShareDividendETF, error) {
	var etf models.AShareDividendETF
	result := s.db.Where("symbol = ?", symbol).First(&etf)
	if result.Error != nil {
		return nil, result.Error
	}

	// 如果启用了AKShare且价格过期，尝试刷新
	if s.useAKShare && time.Since(etf.PriceUpdatedAt) > 5*time.Minute {
		quote, err := s.akshareProvider.GetETFQuote(symbol)
		if err == nil {
			etf.CurrentPrice = quote.CurrentPrice
			etf.PreviousClose = quote.PreviousClose
			etf.PriceChange = quote.PriceChange
			etf.PriceChangePct = quote.PriceChangePct
			etf.Volume = quote.Volume
			etf.Turnover = quote.Turnover
			etf.PriceUpdatedAt = quote.UpdateTime

			// 更新数据库
			s.db.Model(&etf).Updates(map[string]any{
				"current_price":    quote.CurrentPrice,
				"previous_close":   quote.PreviousClose,
				"price_change":     quote.PriceChange,
				"price_change_pct": quote.PriceChangePct,
				"volume":           quote.Volume,
				"turnover":         quote.Turnover,
				"price_updated_at": quote.UpdateTime,
			})
		}
	}

	return &etf, nil
}

// GetAllETFPrices 获取所有ETF价格
func (s *ETFDataService) GetAllETFPrices() ([]models.AShareDividendETF, error) {
	var etfs []models.AShareDividendETF
	result := s.db.Where("status = ?", 1).Find(&etfs)
	if result.Error != nil {
		return nil, result.Error
	}

	// 如果启用了AKShare，批量刷新价格
	if s.useAKShare && len(etfs) > 0 {
		symbols := make([]string, 0, len(etfs))
		for _, etf := range etfs {
			symbols = append(symbols, etf.Symbol)
		}

		// 分批获取，每批50个
		batchSize := 50
		for i := 0; i < len(symbols); i += batchSize {
			end := min(i+batchSize, len(symbols))
			batch := symbols[i:end]
			s.SyncETFPrices(batch)
		}

		// 重新查询
		s.db.Where("status = ?", 1).Find(&etfs)
	}

	return etfs, nil
}

// GetCoreETFPrices 获取核心ETF价格（白名单过滤）
func (s *ETFDataService) GetCoreETFPrices() ([]models.AShareDividendETF, error) {
	var etfs []models.AShareDividendETF
	result := s.db.Where("symbol IN ? AND status = ?", constants.CoreETFSymbols, 1).Find(&etfs)
	if result.Error != nil {
		return nil, result.Error
	}

	if s.useAKShare && len(etfs) > 0 {
		s.SyncETFPrices(constants.CoreETFSymbols)
		s.db.Where("symbol IN ? AND status = ?", constants.CoreETFSymbols, 1).Find(&etfs)
	}

	return etfs, nil
}

// NeedsPriceRefresh 检查核心ETF是否需要刷新价格（检查所有核心ETF）
// 检查条件：从未刷新过 或 超过1小时未刷新
func (s *ETFDataService) NeedsPriceRefresh() bool {
	var etfs []models.AShareDividendETF
	result := s.db.Where("symbol IN ? AND status = ?", constants.CoreETFSymbols, 1).Find(&etfs)
	if result.Error != nil || len(etfs) == 0 {
		return true
	}
	// 只要有一个ETF需要刷新就返回true
	maxAge := 1 * time.Hour
	for _, etf := range etfs {
		if etf.PriceUpdatedAt.IsZero() || time.Since(etf.PriceUpdatedAt) > maxAge {
			return true
		}
	}
	return false
}

// GetHistoricalPrices 获取历史价格
func (s *ETFDataService) GetHistoricalPrices(symbol string, startDate, endDate time.Time) ([]ETFHistoricalData, error) {
	if !s.useAKShare {
		return nil, fmt.Errorf("AKShare数据源未启用")
	}

	return s.akshareProvider.GetHistoricalData(symbol, startDate, endDate)
}

// CalculateDividendYield 计算股息率
func (s *ETFDataService) CalculateDividendYield(symbol string) (decimal.Decimal, error) {
	if !s.useAKShare {
		return decimal.Zero, fmt.Errorf("AKShare数据源未启用")
	}

	dividends, err := s.akshareProvider.GetDividendHistory(symbol)
	if err != nil {
		return decimal.Zero, err
	}

	if len(dividends) == 0 {
		return decimal.Zero, nil
	}

	// 获取最新价格
	quote, err := s.akshareProvider.GetETFQuote(symbol)
	if err != nil {
		return decimal.Zero, err
	}

	// 计算年化股息率（基于最近一年的分红）
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	var totalDividend decimal.Decimal
	for _, div := range dividends {
		if div.ExDividendDate.After(oneYearAgo) {
			totalDividend = totalDividend.Add(div.DividendPerShare)
		}
	}

	if quote.CurrentPrice.IsZero() {
		return decimal.Zero, nil
	}

	yield := totalDividend.Div(quote.CurrentPrice).Mul(decimal.NewFromInt(100))
	return yield, nil
}

// RefreshAllData 刷新所有数据
func (s *ETFDataService) RefreshAllData() error {
	// 同步ETF列表
	if err := s.SyncETFList(); err != nil {
		fmt.Printf("同步ETF列表失败: %v\n", err)
	}

	// 同步价格
	var etfs []models.AShareDividendETF
	s.db.Where("status = ?", 1).Pluck("symbol", &etfs)

	symbols := make([]string, len(etfs))
	for i, etf := range etfs {
		symbols[i] = etf.Symbol
	}

	if err := s.SyncETFPrices(symbols); err != nil {
		return fmt.Errorf("同步价格失败: %w", err)
	}

	return nil
}

// GetETFByFrequency 按分红频率筛选ETF
func (s *ETFDataService) GetETFByFrequency(frequency models.DividendFrequency) ([]models.AShareDividendETF, error) {
	var etfs []models.AShareDividendETF
	result := s.db.Where("dividend_frequency = ? AND status = ?", frequency, 1).Find(&etfs)
	return etfs, result.Error
}

// SearchETFs 搜索ETF
func (s *ETFDataService) SearchETFs(keyword string) ([]models.AShareDividendETF, error) {
	var etfs []models.AShareDividendETF
	result := s.db.Where(
		"(symbol LIKE ? OR name LIKE ?) AND status = ?",
		"%"+keyword+"%",
		"%"+keyword+"%",
		1,
	).Find(&etfs)
	return etfs, result.Error
}

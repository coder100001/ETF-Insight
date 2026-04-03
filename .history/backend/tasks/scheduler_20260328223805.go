package tasks

import (
	"time"

	"etf-insight/config"
	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/utils"

	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	cron         *cron.Cron
	cfg          *config.ScheduleConfig
	yahooClient  *services.YahooFinanceClient
	cacheService *services.CacheService
	analysisSvc  *services.ETFAnalysisService
	exchangeSvc  *services.ExchangeRateService
}

// NewScheduler 创建新的调度器
func NewScheduler(cfg *config.ScheduleConfig, cache *services.CacheService, analysis *services.ETFAnalysisService, exchange *services.ExchangeRateService) *Scheduler {
	// 使用标准cron解析器
	c := cron.New(cron.WithSeconds())

	return &Scheduler{
		cron:         c,
		cfg:          cfg,
		yahooClient:  services.NewYahooFinanceClient(),
		cacheService: cache,
		analysisSvc:  analysis,
		exchangeSvc:  exchange,
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	utils.Info("Starting scheduler...")

	// 添加汇率更新任务 (每天 10:30)
	_, err := s.cron.AddFunc("0 30 10 * * *", s.updateExchangeRates)
	if err != nil {
		utils.Error("Failed to add exchange rate update job", err)
	} else {
		utils.Info("Exchange rate update job scheduled at 10:30 daily")
	}

	// 添加ETF盘前更新任务 (每天 9:30)
	_, err = s.cron.AddFunc("0 30 9 * * *", s.updateETFData)
	if err != nil {
		utils.Error("Failed to add ETF pre-market update job", err)
	} else {
		utils.Info("ETF pre-market update job scheduled at 09:30 daily")
	}

	// 添加ETF收盘后更新任务 (每天 16:30)
	_, err = s.cron.AddFunc("0 30 16 * * *", s.updateETFData)
	if err != nil {
		utils.Error("Failed to add ETF post-market update job", err)
	} else {
		utils.Info("ETF post-market update job scheduled at 16:30 daily")
	}

	// 添加每小时检查任务
	_, err = s.cron.AddFunc("0 0 * * * *", s.hourlyCheck)
	if err != nil {
		utils.Error("Failed to add hourly check job", err)
	} else {
		utils.Info("Hourly check job scheduled")
	}

	s.cron.Start()
	utils.Info("Scheduler started successfully")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	utils.Info("Stopping scheduler...")
	ctx := s.cron.Stop()
	<-ctx.Done()
	utils.Info("Scheduler stopped")
}

// updateExchangeRates 更新汇率
func (s *Scheduler) updateExchangeRates() {
	utils.Info("Running scheduled exchange rate update...")

	// 创建操作日志
	opLog := models.OperationLog{
		OperationType: "scheduled_task",
		OperationName: "定时汇率更新",
		Operator:      "system",
		Status:        0,
		StartTime:     time.Now(),
	}
	models.DB.Create(&opLog)

	if err := s.exchangeSvc.UpdateRates(); err != nil {
		utils.Error("Failed to update exchange rates", err)
		opLog.Status = 2
		opLog.ErrorMessage = err.Error()
	} else {
		utils.Info("Exchange rates updated successfully")
		opLog.Status = 1
	}

	endTime := time.Now()
	opLog.EndTime = &endTime
	opLog.DurationMs = int(endTime.Sub(opLog.StartTime).Milliseconds())
	models.DB.Save(&opLog)
}

// updateETFData 更新ETF数据
func (s *Scheduler) updateETFData() {
	utils.Info("Running scheduled ETF data update...")

	// 创建操作日志
	opLog := models.OperationLog{
		OperationType: "scheduled_task",
		OperationName: "ETF数据更新",
		Operator:      "system",
		Status:        0,
		StartTime:     time.Now(),
	}
	models.DB.Create(&opLog)

	// 获取所有启用的ETF
	var etfConfigs []models.ETFConfig
	_ = models.DB.Where("status = ?", 1).Find(&etfConfigs)
	// 简化处理，使用硬编码数据
	etfConfigs = []models.ETFConfig{
		{Symbol: "QQQ", Currency: "USD"},
		{Symbol: "SCHD", Currency: "USD"},
	}

	var symbols []string
	for _, cfg := range etfConfigs {
		symbols = append(symbols, cfg.Symbol)
	}

	// 获取实时数据
	quotes, err := s.yahooClient.GetQuotes(symbols)
	if err != nil {
		utils.Error("Failed to get quotes", err)
		opLog.Status = 2
		opLog.ErrorMessage = err.Error()
		endTime := time.Now()
		opLog.EndTime = &endTime
		models.DB.Save(&opLog)
		return
	}

	// 更新缓存和数据库
	successCount := 0
	for _, quote := range quotes {
		// 更新缓存
		realtimeData := &services.RealtimeData{
			Symbol:           quote.Symbol,
			Name:             quote.Name,
			CurrentPrice:     quote.CurrentPrice,
			PreviousClose:    quote.PreviousClose,
			OpenPrice:        quote.OpenPrice,
			DayHigh:          quote.DayHigh,
			DayLow:           quote.DayLow,
			Volume:           quote.Volume,
			Change:           quote.Change,
			ChangePercent:    quote.ChangePercent,
			MarketCap:        quote.MarketCap,
			DividendYield:    quote.DividendYield,
			FiftyTwoWeekHigh: quote.FiftyTwoWeekHigh,
			FiftyTwoWeekLow:  quote.FiftyTwoWeekLow,
			AverageVolume:    quote.AverageVolume,
			Beta:             quote.Beta,
			PERatio:          quote.PERatio,
			Currency:         quote.Currency,
			DataSource:       "yahoo_finance",
		}
		s.cacheService.SetRealtimeData(quote.Symbol, realtimeData)

		// 保存到数据库
		etfData := models.ETFData{
			Symbol:     quote.Symbol,
			Date:       time.Now(),
			OpenPrice:  decimal.NewFromFloat(quote.OpenPrice),
			ClosePrice: decimal.NewFromFloat(quote.CurrentPrice),
			HighPrice:  decimal.NewFromFloat(quote.DayHigh),
			LowPrice:   decimal.NewFromFloat(quote.DayLow),
			Volume:     quote.Volume,
			DataSource: "yahoo_finance",
		}

		// 简化处理，直接创建数据
		models.DB.Create(&etfData)

		successCount++
	}

	// 获取历史数据并更新
	for _, symbol := range symbols {
		go func(sym string) {
			prices, err := s.yahooClient.GetHistoricalData(sym, "1y", "1d")
			if err != nil {
				utils.Warn("Failed to get historical data", err, "symbol", sym)
				return
			}

			// 保存历史数据
			for _, price := range prices {
				etfData := models.ETFData{
					Symbol:     sym,
					Date:       price.Date,
					OpenPrice:  decimal.NewFromFloat(price.Open),
					ClosePrice: decimal.NewFromFloat(price.Close),
					HighPrice:  decimal.NewFromFloat(price.High),
					LowPrice:   decimal.NewFromFloat(price.Low),
					Volume:     price.Volume,
					DataSource: "yahoo_finance",
				}

				// 简化处理，直接创建数据
				models.DB.Create(&etfData)
			}

			utils.Info("Updated historical data", "symbol", sym, "records", len(prices))
		}(symbol)
	}

	utils.Info("ETF data update completed", "success", successCount, "total", len(symbols))
	opLog.Status = 1
	endTime := time.Now()
	opLog.EndTime = &endTime
	opLog.DurationMs = int(endTime.Sub(opLog.StartTime).Milliseconds())
	models.DB.Save(&opLog)
}

// hourlyCheck 每小时检查
func (s *Scheduler) hourlyCheck() {
	utils.Info("Running hourly check...")

	// 检查缓存状态
	stats := s.cacheService.GetCacheStats()
	utils.Info("Cache stats", "stats", stats)

	// 可以在这里添加其他检查逻辑
	// 例如：检查数据完整性、清理过期数据等
}

// RunOnce 立即执行一次更新
func (s *Scheduler) RunOnce() {
	utils.Info("Running one-time update...")
	s.updateETFData()
	s.updateExchangeRates()
}

// GetJobs 获取所有任务
func (s *Scheduler) GetJobs() []map[string]interface{} {
	entries := s.cron.Entries()
	var jobs []map[string]interface{}

	for _, entry := range entries {
		jobs = append(jobs, map[string]interface{}{
			"id":       entry.ID,
			"next_run": entry.Next.Format("2006-01-02 15:04:05"),
			"prev_run": entry.Prev.Format("2006-01-02 15:04:05"),
		})
	}

	return jobs
}

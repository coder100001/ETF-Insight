package tasks

import (
	"context"
	"time"

	"etf-insight/config"
	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/services/datasource"
	"etf-insight/utils"

	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"gorm.io/gorm/clause"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	cron        *cron.Cron
	cfg         *config.ScheduleConfig
	analysisSvc *services.ETFAnalysisService
	exchangeSvc *services.ExchangeRateService
	provider    datasource.DataSourceProvider
}

// NewScheduler 创建新的调度器
func NewScheduler(cfg *config.ScheduleConfig, analysis *services.ETFAnalysisService, exchange *services.ExchangeRateService, provider datasource.DataSourceProvider) *Scheduler {
	c := cron.New(cron.WithSeconds())

	return &Scheduler{
		cron:        c,
		cfg:         cfg,
		analysisSvc: analysis,
		exchangeSvc: exchange,
		provider:    provider,
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

	opLog := models.OperationLog{
		OperationType: "scheduled_task",
		OperationName: "定时汇率更新",
		Operator:      "system",
		Status:        0,
		StartTime:     time.Now(),
	}
	models.DB.Create(&opLog)

	if s.exchangeSvc != nil {
		if err := s.exchangeSvc.UpdateRates(); err != nil {
			utils.Error("Failed to update exchange rates", err)
			opLog.Status = 2
			opLog.ErrorMessage = err.Error()
		} else {
			utils.Info("Exchange rates updated successfully")
			opLog.Status = 1
		}
	}

	endTime := time.Now()
	opLog.EndTime = &endTime
	opLog.DurationMs = int(endTime.Sub(opLog.StartTime).Milliseconds())
	models.DB.Save(&opLog)
}

// updateETFData 更新ETF数据 - 使用Finage聚合API获取完整OHLCV数据并入库
func (s *Scheduler) updateETFData() {
	utils.Info("Running scheduled ETF data update...")

	opLog := models.OperationLog{
		OperationType: "scheduled_task",
		OperationName: "ETF数据更新",
		Operator:      "system",
		Status:        0,
		StartTime:     time.Now(),
	}
	models.DB.Create(&opLog)

	// 获取所有启用的ETF配置（从数据库读取，不硬编码）
	var etfConfigs []models.ETFConfig
	if err := models.DB.Where("status = ?", 1).Find(&etfConfigs).Error; err != nil || len(etfConfigs) == 0 {
		utils.Warn("No enabled ETF configs found", "error", err)
		opLog.Status = 2
		opLog.ErrorMessage = "no enabled ETF configs"
		endTime := time.Now()
		opLog.EndTime = &endTime
		models.DB.Save(&opLog)
		return
	}

	symbols := make([]string, 0, len(etfConfigs))
	for _, cfg := range etfConfigs {
		symbols = append(symbols, cfg.Symbol)
	}

	// 使用数据源获取完整OHLCV数据
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if s.provider == nil || !s.provider.IsAvailable(ctx) {
		providerName := "nil"
		if s.provider != nil {
			providerName = s.provider.GetName()
		}
		utils.Warn("Data source not available", "provider", providerName)
		opLog.Status = 2
		opLog.ErrorMessage = "data source not available: " + providerName
		endTime := time.Now()
		opLog.EndTime = &endTime
		models.DB.Save(&opLog)
		return
	}

	// 逐个获取ETF数据（使用聚合API获取完整OHLCV）
	quotes, err := s.provider.GetQuotes(ctx, symbols)
	if err != nil {
		utils.Error("Failed to get quotes from data source", err, "provider", s.provider.GetName())
		opLog.Status = 2
		opLog.ErrorMessage = err.Error()
		endTime := time.Now()
		opLog.EndTime = &endTime
		models.DB.Save(&opLog)
		return
	}

	// 入库完整OHLCV数据
	successCount := 0
	failCount := 0
	for _, quote := range quotes {
		// 检查OHLCV数据是否完整，不完整的跳过不入库
		if quote.OpenPrice == 0 && quote.DayHigh == 0 && quote.DayLow == 0 && quote.Volume == 0 {
			utils.Warn("Skipping incomplete OHLCV data",
				"symbol", quote.Symbol,
				"dataSource", quote.DataSource,
				"currentPrice", quote.CurrentPrice)
			failCount++
			continue
		}

		// 确定日期：使用聚合API返回的Timestamp
		date := quote.Timestamp
		if date.IsZero() {
			date = time.Now().Truncate(24 * time.Hour)
		}

		// Upsert 数据 (symbol + date 联合唯一)
		etfData := models.ETFData{
			Symbol:     quote.Symbol,
			Date:       date,
			OpenPrice:  decimal.NewFromFloat(quote.OpenPrice),
			ClosePrice: decimal.NewFromFloat(quote.CurrentPrice),
			HighPrice:  decimal.NewFromFloat(quote.DayHigh),
			LowPrice:   decimal.NewFromFloat(quote.DayLow),
			Volume:     quote.Volume,
			DataSource: quote.DataSource,
		}

		result := models.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "symbol"}, {Name: "date"}},
			DoUpdates: clause.AssignmentColumns([]string{"open_price", "close_price", "high_price", "low_price", "volume", "data_source"}),
		}).Create(&etfData)

		if result.Error != nil {
			utils.Warn("Failed to upsert ETF data", result.Error, "symbol", quote.Symbol)
			failCount++
			continue
		}

		successCount++
		utils.Info("Updated ETF data",
			"symbol", quote.Symbol,
			"open", quote.OpenPrice,
			"high", quote.DayHigh,
			"low", quote.DayLow,
			"close", quote.CurrentPrice,
			"volume", quote.Volume,
			"source", quote.DataSource)

		// 缓存已移除，不再更新缓存中的实时数据
	}

	utils.Info("ETF data update completed", "success", successCount, "fail", failCount, "total", len(symbols), "provider", s.provider.GetName())

	opLog.Status = 1
	if failCount > 0 && successCount == 0 {
		opLog.Status = 2
	} else if failCount > 0 {
		opLog.Status = 3
	}
	endTime := time.Now()
	opLog.EndTime = &endTime
	opLog.DurationMs = int(endTime.Sub(opLog.StartTime).Milliseconds())
	models.DB.Save(&opLog)
}

// hourlyCheck 每小时检查
func (s *Scheduler) hourlyCheck() {
	utils.Info("Running hourly check...")

	// 缓存已移除：if s.cacheService != nil {
	//	stats := s.cacheService.GetCacheStats()
	//	utils.Info("Cache stats", "stats", stats)
	// }
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

// Package sync 提供ETF数据同步服务
package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"etf-insight/models"
	"etf-insight/services/datasource"
	"etf-insight/utils"

	"github.com/shopspring/decimal"
)

// SyncService 同步服务
type SyncService struct {
	provider  datasource.DataSourceProvider
	factory   *datasource.ProviderFactory
	operation string
}

// SyncResult 同步结果
type SyncResult struct {
	TotalCount   int
	SuccessCount int
	FailCount    int
	UpdatedCount int
	Errors       []error
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	DataSource   string
	Details      []SyncDetail
}

// SyncDetail 单个ETF同步详情
type SyncDetail struct {
	Symbol     string
	Success    bool
	Error      string
	Price      float64
	ChangePct  float64
	DataSource string
}

// NewSyncService 创建同步服务
func NewSyncService(provider datasource.DataSourceProvider) *SyncService {
	factory := datasource.NewProviderFactory()
	factory.Register(provider.GetName(), provider)
	factory.Register("fallback", datasource.NewFallbackProvider())

	return &SyncService{
		provider:  provider,
		factory:   factory,
		operation: "ETF_SYNC",
	}
}

// NewSyncServiceWithFactory 使用工厂创建同步服务
// 自动选择可用的数据源
func NewSyncServiceWithFactory(factory *datasource.ProviderFactory) (*SyncService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	provider, err := factory.GetDefault(ctx)
	if err != nil {
		return nil, err
	}

	return &SyncService{
		provider:  provider,
		factory:   factory,
		operation: "ETF_SYNC",
	}, nil
}

// SyncETFs 同步ETF列表
func (s *SyncService) SyncETFs(ctx context.Context, etfList []datasource.ETFInfo) (*SyncResult, error) {
	result := &SyncResult{
		StartTime:  time.Now(),
		TotalCount: len(etfList),
		DataSource: s.provider.GetName(),
		Details:    make([]SyncDetail, 0, len(etfList)),
	}

	// 记录操作日志
	logID := s.startOperationLog(len(etfList))

	// 检查数据源可用性
	utils.Info("Checking data source availability", "provider", s.provider.GetName())
	if !s.provider.IsAvailable(ctx) {
		utils.Warn("Primary data source not available, switching to fallback", "provider", s.provider.GetName())
		// 尝试切换到后备数据源
		fallback, ok := s.factory.Get("fallback")
		if ok {
			s.provider = fallback
			result.DataSource = fallback.GetName()
			utils.Info("Switched to fallback data source")
		} else {
			s.failOperationLog(logID, "no data source available")
			return nil, datasource.ErrNoAvailableProvider
		}
	} else {
		utils.Info("Primary data source is available", "provider", s.provider.GetName())
	}

	// 提取symbols
	symbols := make([]string, len(etfList))
	etfMap := make(map[string]datasource.ETFInfo)
	for i, etf := range etfList {
		symbols[i] = etf.Symbol
		etfMap[etf.Symbol] = etf
	}

	// 批量获取报价
	quotes, err := s.provider.GetQuotes(ctx, symbols)
	if err != nil {
		s.failOperationLog(logID, err.Error())
		return nil, err
	}

	// 同步每个ETF
	for _, quote := range quotes {
		detail := SyncDetail{
			Symbol:     quote.Symbol,
			DataSource: quote.DataSource,
		}

		etf := etfMap[quote.Symbol]

		// 更新ETF配置
		if err := s.updateETFConfig(etf); err != nil {
			detail.Success = false
			detail.Error = fmt.Sprintf("config update failed: %v", err)
			result.FailCount++
			result.Errors = append(result.Errors, err)
			result.Details = append(result.Details, detail)
			continue
		}

		// 更新ETF数据
		if err := s.updateETFData(quote); err != nil {
			detail.Success = false
			detail.Error = fmt.Sprintf("data update failed: %v", err)
			result.FailCount++
			result.Errors = append(result.Errors, err)
			result.Details = append(result.Details, detail)
			continue
		}

		detail.Success = true
		detail.Price = quote.CurrentPrice
		detail.ChangePct = quote.ChangePercent
		result.SuccessCount++
		result.UpdatedCount++
		result.Details = append(result.Details, detail)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// 完成操作日志
	s.completeOperationLog(logID, result)

	return result, nil
}

// SyncSingleETF 同步单个ETF
func (s *SyncService) SyncSingleETF(ctx context.Context, etf datasource.ETFInfo) (*SyncDetail, error) {
	quote, err := s.provider.GetQuote(ctx, etf.Symbol)
	if err != nil {
		return nil, err
	}

	// 更新配置和数据
	if err := s.updateETFConfig(etf); err != nil {
		return nil, fmt.Errorf("config update failed: %w", err)
	}

	if err := s.updateETFData(quote); err != nil {
		return nil, fmt.Errorf("data update failed: %w", err)
	}

	return &SyncDetail{
		Symbol:     etf.Symbol,
		Success:    true,
		Price:      quote.CurrentPrice,
		ChangePct:  quote.ChangePercent,
		DataSource: quote.DataSource,
	}, nil
}

// GetProvider 获取当前使用的数据源
func (s *SyncService) GetProvider() datasource.DataSourceProvider {
	return s.provider
}

// SwitchProvider 切换数据源
func (s *SyncService) SwitchProvider(name string) error {
	provider, ok := s.factory.Get(name)
	if !ok {
		return fmt.Errorf("provider %s not found", name)
	}

	s.provider = provider
	return nil
}

// updateETFConfig 更新ETF配置
func (s *SyncService) updateETFConfig(etf datasource.ETFInfo) error {
	var config models.ETFConfig
	result := models.DB.Where("symbol = ?", etf.Symbol).First(&config)

	if result.Error != nil {
		// 创建新配置
		config = models.ETFConfig{
			Symbol:          etf.Symbol,
			Name:            etf.Name,
			Description:     etf.Description,
			Category:        etf.Category,
			Currency:        "USD",
			Exchange:        "NASDAQ",
			Provider:        etf.Provider,
			ExpenseRatio:    decimal.NewFromFloat(etf.ExpenseRatio),
			Status:          1,
			AutoUpdate:      true,
			UpdateFrequency: "每日",
			DataSource:      "Finnhub",
		}
		return models.DB.Create(&config).Error
	}

	// 更新现有配置
	return models.DB.Model(&config).Updates(map[string]interface{}{
		"name":          etf.Name,
		"description":   etf.Description,
		"category":      etf.Category,
		"provider":      etf.Provider,
		"expense_ratio": decimal.NewFromFloat(etf.ExpenseRatio),
		"data_source":   "Finnhub",
		"updated_at":    time.Now(),
	}).Error
}

// updateETFData 更新ETF数据
func (s *SyncService) updateETFData(quote *datasource.QuoteData) error {
	date := time.Now().Truncate(24 * time.Hour)

	// 检查数据来源是否为不完整的 last API
	// 如果 OHLCV 全为0，说明只有 ask/bid 数据，不应写入数据库
	if quote.OpenPrice == 0 && quote.DayHigh == 0 && quote.DayLow == 0 && quote.Volume == 0 {
		utils.Warn("Skipping data with incomplete OHLCV",
			"symbol", quote.Symbol,
			"dataSource", quote.DataSource,
			"currentPrice", quote.CurrentPrice)
		return fmt.Errorf("incomplete OHLCV data for %s (source: %s), only currentPrice=%.2f available",
			quote.Symbol, quote.DataSource, quote.CurrentPrice)
	}

	var existing models.ETFData
	result := models.DB.Where("symbol = ? AND date = ?", quote.Symbol, date).First(&existing)

	if result.Error == nil {
		// 记录已存在，更新
		existing.OpenPrice = decimal.NewFromFloat(quote.OpenPrice)
		existing.ClosePrice = decimal.NewFromFloat(quote.CurrentPrice)
		existing.HighPrice = decimal.NewFromFloat(quote.DayHigh)
		existing.LowPrice = decimal.NewFromFloat(quote.DayLow)
		existing.Volume = quote.Volume
		existing.DataSource = quote.DataSource
		return models.DB.Save(&existing).Error
	}

	// 创建新记录
	data := &models.ETFData{
		Symbol:     quote.Symbol,
		Date:       date,
		OpenPrice:  decimal.NewFromFloat(quote.OpenPrice),
		ClosePrice: decimal.NewFromFloat(quote.CurrentPrice),
		HighPrice:  decimal.NewFromFloat(quote.DayHigh),
		LowPrice:   decimal.NewFromFloat(quote.DayLow),
		Volume:     quote.Volume,
		DataSource: quote.DataSource,
		CreatedAt:  time.Now(),
	}
	return models.DB.Create(data).Error
}

// startOperationLog 开始记录操作日志
func (s *SyncService) startOperationLog(totalCount int) uint {
	log := models.OperationLog{
		OperationType: "SYNC",
		OperationName: s.operation,
		Operator:      "system",
		Status:        0, // 进行中
		StartTime:     time.Now(),
		Details:       fmt.Sprintf("开始同步 %d 只ETF", totalCount),
	}
	models.DB.Create(&log)
	return log.ID
}

// failOperationLog 记录失败日志
func (s *SyncService) failOperationLog(id uint, errorMsg string) {
	now := time.Now()
	models.DB.Model(&models.OperationLog{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        2, // 失败
		"error_message": errorMsg,
		"end_time":      &now,
	})
}

// completeOperationLog 完成操作日志
func (s *SyncService) completeOperationLog(id uint, result *SyncResult) {
	now := time.Now()
	details, _ := json.Marshal(map[string]interface{}{
		"total":      result.TotalCount,
		"success":    result.SuccessCount,
		"fail":       result.FailCount,
		"updated":    result.UpdatedCount,
		"duration":   result.Duration.String(),
		"dataSource": result.DataSource,
	})

	status := 1 // 成功
	if result.FailCount > 0 {
		if result.SuccessCount == 0 {
			status = 2 // 全部失败
		} else {
			status = 3 // 部分成功
		}
	}

	var errorMsg string
	if len(result.Errors) > 0 {
		errorMsg = result.Errors[0].Error()
	}

	durationMs := int(result.Duration.Milliseconds())

	models.DB.Model(&models.OperationLog{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        status,
		"end_time":      &now,
		"duration_ms":   durationMs,
		"details":       string(details),
		"error_message": errorMsg,
	})
}

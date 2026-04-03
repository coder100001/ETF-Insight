package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"etf-insight/models"
	"etf-insight/utils"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ExchangeRateSyncService 汇率同步服务
type ExchangeRateSyncService struct {
	apiClient      *http.Client
	syncMutex      sync.Mutex
	isSyncing      bool
	dataSource     string
	maxRetries     int
	retryDelay     time.Duration
	requestTimeout time.Duration
}

// NewExchangeRateSyncService 创建汇率同步服务
func NewExchangeRateSyncService() *ExchangeRateSyncService {
	return &ExchangeRateSyncService{
		apiClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		dataSource:     "exchangerate-api",
		maxRetries:     3,
		retryDelay:     2 * time.Second,
		requestTimeout: 30 * time.Second,
	}
}

// ExchangeRateAPIResponse 汇率API响应
type ExchangeRateAPIResponse struct {
	Result  string             `json:"result"`
	Base    string             `json:"base_code"`
	Rates   map[string]float64 `json:"conversion_rates"`
	Date    string             `json:"time_last_update_utc"`
	NextDate string            `json:"time_next_update_utc"`
}

// SyncOptions 同步选项
type SyncOptions struct {
	SyncType     string   // full:全量, incremental:增量
	CurrencyPairs []models.CurrencyPair
	ForceUpdate  bool     // 强制更新，即使数据未过期
}

// SyncResult 同步结果
type SyncResult struct {
	BatchID      string
	TotalCount   int
	SuccessCount int
	FailedCount  int
	SkippedCount int
	Duration     time.Duration
	Errors       []error
}

// FullSync 全量同步
func (s *ExchangeRateSyncService) FullSync(ctx context.Context) (*SyncResult, error) {
	return s.Sync(ctx, &SyncOptions{
		SyncType: "full",
	})
}

// IncrementalSync 增量同步
func (s *ExchangeRateSyncService) IncrementalSync(ctx context.Context) (*SyncResult, error) {
	return s.Sync(ctx, &SyncOptions{
		SyncType: "incremental",
	})
}

// Sync 执行同步
func (s *ExchangeRateSyncService) Sync(ctx context.Context, opts *SyncOptions) (*SyncResult, error) {
	s.syncMutex.Lock()
	if s.isSyncing {
		s.syncMutex.Unlock()
		return nil, fmt.Errorf("sync is already in progress")
	}
	s.isSyncing = true
	s.syncMutex.Unlock()

	defer func() {
		s.syncMutex.Lock()
		s.isSyncing = false
		s.syncMutex.Unlock()
	}()

	startTime := time.Now()
	batchID := uuid.New().String()

	// 创建同步日志
	syncLog := &models.ExchangeRateSyncLog{
		BatchID:       batchID,
		SyncType:      opts.SyncType,
		DataSource:    s.dataSource,
		Status:        "running",
		StartedAt:     startTime,
		MaxRetryCount: s.maxRetries,
	}
	models.DB.Create(syncLog)

	// 获取需要同步的货币对
	var currencyPairs []models.CurrencyPair
	if len(opts.CurrencyPairs) > 0 {
		currencyPairs = opts.CurrencyPairs
	} else {
		if err := models.DB.Where("is_active = ?", 1).Order("priority desc").Find(&currencyPairs).Error; err != nil {
			s.failSync(syncLog, err)
			return nil, err
		}
	}

	result := &SyncResult{
		BatchID: batchID,
		TotalCount: len(currencyPairs),
	}

	// 使用工作池进行并发同步
	workerCount := 5
	if len(currencyPairs) < workerCount {
		workerCount = len(currencyPairs)
	}

	type syncJob struct {
		pair models.CurrencyPair
		index int
	}

	jobChan := make(chan syncJob, len(currencyPairs))
	resultChan := make(chan struct {
		success bool
		skipped bool
		err     error
	}, len(currencyPairs))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChan {
				success, skipped, err := s.syncSinglePair(ctx, job.pair, batchID, opts.ForceUpdate)
				resultChan <- struct {
					success bool
					skipped bool
					err     error
				}{success, skipped, err}
			}
		}()
	}

	// 发送任务
	for i, pair := range currencyPairs {
		jobChan <- syncJob{pair: pair, index: i}
	}
	close(jobChan)

	// 等待所有任务完成
	wg.Wait()
	close(resultChan)

	// 收集结果
	for res := range resultChan {
		if res.success {
			result.SuccessCount++
		} else if res.skipped {
			result.SkippedCount++
		} else {
			result.FailedCount++
			if res.err != nil {
				result.Errors = append(result.Errors, res.err)
			}
		}
	}

	result.Duration = time.Since(startTime)

	// 更新同步日志
	s.completeSync(syncLog, result)

	return result, nil
}

// syncSinglePair 同步单个货币对
func (s *ExchangeRateSyncService) syncSinglePair(ctx context.Context, pair models.CurrencyPair, batchID string, forceUpdate bool) (bool, bool, error) {
	// 检查是否需要更新（增量同步时）
	if !forceUpdate {
		var existingRate models.ExchangeRate
		if err := models.DB.Where("from_currency = ? AND to_currency = ? AND valid_status = ?",
			pair.FromCurrency, pair.ToCurrency, 1).
			Order("updated_at desc").First(&existingRate).Error; err == nil {
			// 如果数据在1小时内更新过，则跳过
			if time.Since(existingRate.UpdatedAt) < time.Hour {
				return false, true, nil
			}
		}
	}

	// 获取汇率数据（带重试）
	var rate float64
	var err error
	for i := 0; i < s.maxRetries; i++ {
		rate, err = s.fetchRateFromAPI(pair.FromCurrency, pair.ToCurrency)
		if err == nil {
			break
		}
		if i < s.maxRetries-1 {
			time.Sleep(s.retryDelay * time.Duration(i+1))
		}
	}

	if err != nil {
		// 记录同步明细 - 失败
		s.createSyncDetail(pair, batchID, decimal.Zero, decimal.Zero, decimal.Zero, "failed", err.Error())
		return false, false, err
	}

	// 验证数据
	if !s.validateRate(rate) {
		err := fmt.Errorf("invalid rate value: %f", rate)
		s.createSyncDetail(pair, batchID, decimal.Zero, decimal.Zero, decimal.Zero, "failed", err.Error())
		return false, false, err
	}

	// 获取旧汇率用于计算变化
	var oldRate decimal.Decimal
	var existingRate models.ExchangeRate
	if err := models.DB.Where("from_currency = ? AND to_currency = ? AND valid_status = ?",
		pair.FromCurrency, pair.ToCurrency, 1).
		First(&existingRate).Error; err == nil {
		oldRate = existingRate.Rate
	}

	newRate := decimal.NewFromFloat(rate)
	changePercent := decimal.Zero
	if !oldRate.IsZero() {
		changePercent = newRate.Sub(oldRate).Div(oldRate).Mul(decimal.NewFromInt(100))
	}

	// 保存汇率数据
	exchangeRate := &models.ExchangeRate{
		FromCurrency:  pair.FromCurrency,
		ToCurrency:    pair.ToCurrency,
		Rate:          newRate,
		PreviousRate:  oldRate,
		ChangePercent: changePercent,
		DataSource:    s.dataSource,
		SourceType:    "api",
		ValidStatus:   1,
		SyncBatchID:   batchID,
		SyncedAt:      &[]time.Time{time.Now()}[0],
		ExpiresAt:     &[]time.Time{time.Now().Add(24 * time.Hour)}[0],
	}

	// 使用 Upsert 操作
	if err := models.DB.Where("from_currency = ? AND to_currency = ?", pair.FromCurrency, pair.ToCurrency).
		Assign(exchangeRate).
		FirstOrCreate(exchangeRate).Error; err != nil {
		s.createSyncDetail(pair, batchID, oldRate, newRate, changePercent, "failed", err.Error())
		return false, false, err
	}

	// 记录同步明细 - 成功
	s.createSyncDetail(pair, batchID, oldRate, newRate, changePercent, "success", "")

	return true, false, nil
}

// fetchRateFromAPI 从API获取汇率
func (s *ExchangeRateSyncService) fetchRateFromAPI(fromCurrency, toCurrency string) (float64, error) {
	// 使用 exchangerate-api.com 的免费API
	url := fmt.Sprintf("https://api.exchangerate-api.com/v4/latest/%s", fromCurrency)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := s.apiClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var apiResp ExchangeRateAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return 0, err
	}

	if apiResp.Result != "success" {
		return 0, fmt.Errorf("API returned error result")
	}

	rate, ok := apiResp.Rates[toCurrency]
	if !ok {
		return 0, fmt.Errorf("currency %s not found in rates", toCurrency)
	}

	return rate, nil
}

// validateRate 验证汇率值
func (s *ExchangeRateSyncService) validateRate(rate float64) bool {
	if rate <= 0 {
		return false
	}
	// 检查是否为异常值（比如汇率突然变化超过50%）
	if rate > 1000000 || rate < 0.000001 {
		return false
	}
	return true
}

// createSyncDetail 创建同步明细记录
func (s *ExchangeRateSyncService) createSyncDetail(pair models.CurrencyPair, batchID string, oldRate, newRate, changePercent decimal.Decimal, status, errorMsg string) {
	// 先获取同步日志ID
	var syncLog models.ExchangeRateSyncLog
	if err := models.DB.Where("batch_id = ?", batchID).First(&syncLog).Error; err != nil {
		utils.Error("Failed to find sync log", err)
		return
	}

	detail := &models.ExchangeRateSyncDetail{
		SyncLogID:     syncLog.ID,
		FromCurrency:  pair.FromCurrency,
		ToCurrency:    pair.ToCurrency,
		OldRate:       oldRate,
		NewRate:       newRate,
		ChangePercent: changePercent,
		Status:        status,
		ErrorMessage:  errorMsg,
	}

	if err := models.DB.Create(detail).Error; err != nil {
		utils.Error("Failed to create sync detail", err)
	}
}

// failSync 标记同步失败
func (s *ExchangeRateSyncService) failSync(syncLog *models.ExchangeRateSyncLog, err error) {
	now := time.Now()
	syncLog.Status = "failed"
	syncLog.ErrorMessage = err.Error()
	syncLog.CompletedAt = &now
	syncLog.DurationMs = time.Since(syncLog.StartedAt).Milliseconds()
	models.DB.Save(syncLog)
}

// completeSync 完成同步
func (s *ExchangeRateSyncService) completeSync(syncLog *models.ExchangeRateSyncLog, result *SyncResult) {
	now := time.Now()
	syncLog.CompletedAt = &now
	syncLog.DurationMs = result.Duration.Milliseconds()
	syncLog.TotalCount = result.TotalCount
	syncLog.SuccessCount = result.SuccessCount
	syncLog.FailedCount = result.FailedCount
	syncLog.SkippedCount = result.SkippedCount

	if result.FailedCount == 0 {
		syncLog.Status = "success"
	} else if result.SuccessCount == 0 {
		syncLog.Status = "failed"
		syncLog.ErrorMessage = fmt.Sprintf("All %d pairs failed", result.FailedCount)
	} else {
		syncLog.Status = "partial"
		syncLog.ErrorMessage = fmt.Sprintf("%d failed, %d succeeded", result.FailedCount, result.SuccessCount)
	}

	models.DB.Save(syncLog)
}

// GetLatestSyncLog 获取最新的同步日志
func (s *ExchangeRateSyncService) GetLatestSyncLog() (*models.ExchangeRateSyncLog, error) {
	var log models.ExchangeRateSyncLog
	if err := models.DB.Order("started_at desc").First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// GetSyncLogs 获取同步日志列表
func (s *ExchangeRateSyncService) GetSyncLogs(limit int) ([]models.ExchangeRateSyncLog, error) {
	var logs []models.ExchangeRateSyncLog
	if err := models.DB.Order("started_at desc").Limit(limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

package sync

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"etf-insight/models"
	"etf-insight/services/exchange_rate/datasource"
	"etf-insight/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm/clause"
)

// Synchronizer 数据同步器
// 负责从数据源获取汇率数据并同步到数据库
type Synchronizer struct {
	manager *datasource.DataSourceManager
}

// 默认货币对列表（当数据库为空时使用）
var defaultCurrencyPairs = []models.CurrencyPair{
	{FromCurrency: "USD", ToCurrency: "CNY", IsActive: 1},
	{FromCurrency: "USD", ToCurrency: "EUR", IsActive: 1},
	{FromCurrency: "USD", ToCurrency: "GBP", IsActive: 1},
	{FromCurrency: "USD", ToCurrency: "JPY", IsActive: 1},
	{FromCurrency: "USD", ToCurrency: "HKD", IsActive: 1},
	{FromCurrency: "EUR", ToCurrency: "CNY", IsActive: 1},
	{FromCurrency: "GBP", ToCurrency: "CNY", IsActive: 1},
	{FromCurrency: "JPY", ToCurrency: "CNY", IsActive: 1},
	{FromCurrency: "CNY", ToCurrency: "USD", IsActive: 1},
	{FromCurrency: "HKD", ToCurrency: "USD", IsActive: 1},
}

// NewSynchronizer 创建数据同步器
func NewSynchronizer(manager *datasource.DataSourceManager) *Synchronizer {
	return &Synchronizer{
		manager: manager,
	}
}

// SyncResult 同步结果
type SyncResult struct {
	BatchID      string       // 批次ID
	SyncType     string       // 同步类型: full / incremental
	DataSource   string       // 使用的数据源
	Status       string       // success / failed / partial
	TotalCount   int          // 总数
	SuccessCount int          // 成功数
	FailedCount  int          // 失败数
	DurationMs   int64        // 耗时(毫秒)
	Details      []SyncDetail // 同步详情
	Errors       []string     // 错误信息
}

// SyncDetail 同步详情
type SyncDetail struct {
	FromCurrency string          // 源货币
	ToCurrency   string          // 目标货币
	OldRate      decimal.Decimal // 旧汇率
	NewRate      decimal.Decimal // 新汇率
	Change       float64         // 变动百分比
	Status       string          // success / failed / skipped
}

// ============================================================================
// 高性能批处理同步 (BatchSync)
// ============================================================================

// BatchSyncConfig 批处理配置
type BatchSyncConfig struct {
	// 并发控制
	MaxParallelFetches int // 最大并行获取数（默认：4）
	MaxParallelCalcs   int // 最大并行计算数（默认：CPU核心数）
	DBBatchSize        int // 数据库批量写入大小（默认：100）

	// 超时控制
	FetchTimeout time.Duration // 获取超时（默认：30秒）
	CalcTimeout  time.Duration // 计算超时（默认：10秒）
	TotalTimeout time.Duration // 总超时（默认：60秒）

	// 错误处理
	ContinueOnError bool // 遇到错误继续处理（默认：true）
	MaxRetries      int  // 最大重试次数（默认：2）
}

// DefaultBatchSyncConfig 返回默认配置
func DefaultBatchSyncConfig() *BatchSyncConfig {
	return &BatchSyncConfig{
		MaxParallelFetches: 4,
		MaxParallelCalcs:   runtime.NumCPU(),
		DBBatchSize:        100,
		FetchTimeout:       30 * time.Second,
		CalcTimeout:        10 * time.Second,
		TotalTimeout:       60 * time.Second,
		ContinueOnError:    true,
		MaxRetries:         2,
	}
}

// BatchSyncResult 批处理同步结果
type BatchSyncResult struct {
	BatchID      string             // 批次ID
	TotalPairs   int                // 总货币对数
	SuccessCount int32              // 成功数（原子操作）
	FailedCount  int32              // 失败数（原子操作）
	SkippedCount int32              // 跳过数（原子操作）
	Duration     time.Duration      // 总耗时
	DataSource   string             // 使用的数据源
	Status       string             // success / partial / failed
	Errors       []string           // 错误列表
	RateLimitHit bool               // 是否触发速率限制
	Details      []*BatchSyncDetail // 详细信息
}

// BatchSyncDetail 批处理详情
type BatchSyncDetail struct {
	Pair          string          // 货币对
	Rate          decimal.Decimal // 汇率
	OldRate       decimal.Decimal // 旧汇率
	ChangePercent float64         // 变动百分比
	Status        string          // success / failed / skipped
	Error         string          // 错误信息
}

// SyncAllRates 同步所有汇率（高性能批处理版）
func (s *Synchronizer) SyncAllRates(ctx context.Context) (*SyncResult, error) {
	return s.BatchSync(ctx, DefaultBatchSyncConfig())
}

// BatchSync 高性能批处理同步
// 特性：
// 1. 并行获取多个基准货币的汇率
// 2. 并行计算交叉汇率
// 3. 批量并发写入数据库
// 4. 完善的错误处理和重试机制
func (s *Synchronizer) BatchSync(ctx context.Context, config *BatchSyncConfig) (*SyncResult, error) {
	startTime := time.Now()
	batchID := fmt.Sprintf("BATCH_%s", startTime.Format("20060102150405"))

	// 使用超时上下文
	ctx, cancel := context.WithTimeout(ctx, config.TotalTimeout)
	defer cancel()

	utils.Info("开始高性能批处理汇率同步",
		"batch_id", batchID,
		"parallel_fetches", config.MaxParallelFetches,
		"parallel_calcs", config.MaxParallelCalcs)

	// 获取所有活跃的货币对
	currencyPairs, err := s.getActiveCurrencyPairs()
	if err != nil {
		return nil, fmt.Errorf("获取货币对失败: %w", err)
	}

	// 第一阶段：并行获取多个基准货币的汇率
	baseRates, err := s.parallelFetchBaseRates(ctx, config)
	if err != nil {
		utils.Warn("获取基准汇率失败，尝试单基准获取", "error", err)
		// 降级：使用单基准获取
		baseRates, err = s.fetchSingleBaseRates(ctx)
		if err != nil {
			return nil, fmt.Errorf("获取基准汇率失败: %w", err)
		}
	}

	// 第二阶段：并行计算所有货币对的汇率
	calcResults := s.parallelCalculateRates(currencyPairs, baseRates, config)

	// 第三阶段：并发批量写入数据库
	dbErrors := s.parallelBatchSave(ctx, calcResults, config)

	// 统计结果
	result := &SyncResult{
		BatchID:      batchID,
		SyncType:     "batch",
		DataSource:   s.manager.GetCurrentProvider().GetName(),
		Status:       "success",
		TotalCount:   len(currencyPairs),
		SuccessCount: 0,
		FailedCount:  0,
		DurationMs:   time.Since(startTime).Milliseconds(),
		Details:      make([]SyncDetail, 0),
		Errors:       make([]string, 0),
	}

	for _, r := range calcResults {
		detail := SyncDetail{
			FromCurrency: r.FromCurrency,
			ToCurrency:   r.ToCurrency,
			OldRate:      r.OldRate,
			NewRate:      r.NewRate,
			Change:       r.ChangePercent,
			Status:       r.Status,
		}
		result.Details = append(result.Details, detail)

		if r.Status == "success" {
			result.SuccessCount++
		} else if r.Status != "skipped" {
			result.FailedCount++
			if r.Error != "" {
				result.Errors = append(result.Errors, r.FromCurrency+"/"+r.ToCurrency+": "+r.Error)
			}
		}
	}

	// 添加数据库错误
	for _, err := range dbErrors {
		result.Errors = append(result.Errors, err)
		result.FailedCount++
	}

	// 判断最终状态
	if result.FailedCount > 0 && result.SuccessCount > 0 {
		result.Status = "partial"
	} else if result.FailedCount == result.TotalCount {
		result.Status = "failed"
	}

	// 保存同步日志
	s.saveSyncLog(result)

	// 更新Fallback缓存
	s.updateFallbackCache(ctx, baseRates)

	// 尝试恢复到主数据源
	s.manager.RestoreToPrimary(ctx)

	utils.Info("批处理汇率同步完成",
		"batch_id", batchID,
		"status", result.Status,
		"total", result.TotalCount,
		"success", result.SuccessCount,
		"failed", result.FailedCount,
		"duration_ms", result.DurationMs)

	return result, nil
}

// calcResult 计算结果
type calcResult struct {
	FromCurrency  string
	ToCurrency    string
	NewRate       decimal.Decimal
	OldRate       decimal.Decimal
	ChangePercent float64
	Status        string
	Error         string
}

// ---------------------------------------------------------------------------
// 第一阶段：并行获取多个基准货币的汇率
// ---------------------------------------------------------------------------

// baseRateResult 基准货币汇率结果
type baseRateResult struct {
	BaseCurrency string                     // 基准货币
	Rates        map[string]decimal.Decimal // 汇率数据
	DataSource   string                     // 数据源
	Success      bool                       // 是否成功
	Error        string                     // 错误信息
}

// parallelFetchBaseRates 并行获取多个基准货币的汇率
func (s *Synchronizer) parallelFetchBaseRates(ctx context.Context, config *BatchSyncConfig) (map[string]map[string]decimal.Decimal, error) {
	// 基准货币列表（并行请求）
	baseCurrencies := []string{"USD", "EUR", "CNY"}

	// 创建结果通道
	results := make(chan *baseRateResult, len(baseCurrencies))
	errors := make([]string, 0)

	// 使用信号量控制并发数
	sem := make(chan struct{}, config.MaxParallelFetches)
	var wg sync.WaitGroup

	for _, base := range baseCurrencies {
		wg.Add(1)
		go func(baseCurrency string) {
			defer wg.Done()

			// 获取信号量
			sem <- struct{}{}
			defer func() { <-sem }()

			// 创建带超时的上下文
			fetchCtx, cancel := context.WithTimeout(ctx, config.FetchTimeout)
			defer cancel()

			// 获取基准汇率
			result := s.fetchBaseRateWithRetry(fetchCtx, baseCurrency, config.MaxRetries)

			select {
			case results <- result:
			case <-fetchCtx.Done():
				results <- &baseRateResult{
					BaseCurrency: baseCurrency,
					Success:      false,
					Error:        "超时",
				}
			}
		}(base)
	}

	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	baseRates := make(map[string]map[string]decimal.Decimal)
	for result := range results {
		if result.Success && len(result.Rates) > 0 {
			baseRates[result.BaseCurrency] = result.Rates
			utils.Info("基准汇率获取成功",
				"base", result.BaseCurrency,
				"currencies", len(result.Rates),
				"source", result.DataSource)
		} else {
			errors = append(errors, fmt.Sprintf("%s: %s", result.BaseCurrency, result.Error))
		}
	}

	if len(baseRates) == 0 {
		return nil, fmt.Errorf("所有基准货币获取失败: %v", errors)
	}

	return baseRates, nil
}

// fetchBaseRateWithRetry 带重试的基准汇率获取
func (s *Synchronizer) fetchBaseRateWithRetry(ctx context.Context, baseCurrency string, maxRetries int) *baseRateResult {
	result := &baseRateResult{
		BaseCurrency: baseCurrency,
		Rates:        make(map[string]decimal.Decimal),
	}

	for i := 0; i <= maxRetries; i++ {
		batchResult, err := s.manager.GetRates(ctx, baseCurrency)
		if err == nil && batchResult.Success {
			result.Rates = batchResult.Data
			result.DataSource = batchResult.DataSource
			result.Success = true
			return result
		}

		if i < maxRetries {
			// 重试前短暂等待
			time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
		} else {
			result.Error = fmt.Sprintf("%v", err)
		}
	}

	return result
}

// fetchSingleBaseRates 单基准获取（降级方案）
func (s *Synchronizer) fetchSingleBaseRates(ctx context.Context) (map[string]map[string]decimal.Decimal, error) {
	batchResult, err := s.manager.GetRates(ctx, "USD")
	if err != nil {
		return nil, err
	}

	if !batchResult.Success {
		return nil, fmt.Errorf("获取失败")
	}

	return map[string]map[string]decimal.Decimal{
		"USD": batchResult.Data,
	}, nil
}

// ---------------------------------------------------------------------------
// 第二阶段：并行计算所有货币对的汇率
// ---------------------------------------------------------------------------

// parallelCalculateRates 并行计算所有货币对的汇率
func (s *Synchronizer) parallelCalculateRates(pairs []models.CurrencyPair, baseRates map[string]map[string]decimal.Decimal, config *BatchSyncConfig) []*calcResult {
	results := make([]*calcResult, len(pairs))
	var wg sync.WaitGroup

	// 使用信号量控制并发数
	sem := make(chan struct{}, config.MaxParallelCalcs)

	for i, pair := range pairs {
		wg.Add(1)
		go func(idx int, p models.CurrencyPair) {
			defer wg.Done()

			// 获取信号量
			sem <- struct{}{}
			defer func() { <-sem }()

			result := s.calculateCrossRate(p.FromCurrency, p.ToCurrency, baseRates)
			results[idx] = result
		}(i, pair)
	}

	wg.Wait()
	return results
}

// calculateCrossRate 计算交叉汇率
func (s *Synchronizer) calculateCrossRate(fromCurrency, toCurrency string, baseRates map[string]map[string]decimal.Decimal) *calcResult {
	result := &calcResult{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		Status:       "failed",
	}

	// 尝试从多个基准货币计算
	for baseCurrency, rates := range baseRates {
		rate := s.calculateFromBase(fromCurrency, toCurrency, baseCurrency, rates)
		if !rate.IsZero() {
			result.NewRate = rate
			result.Status = "success"
			return result
		}
	}

	result.Error = "无法计算汇率（所有基准货币都不可用）"
	return result
}

// calculateFromBase 从指定基准货币计算汇率
func (s *Synchronizer) calculateFromBase(fromCurrency, toCurrency, baseCurrency string, rates map[string]decimal.Decimal) decimal.Decimal {
	// 同货币直接返回1
	if fromCurrency == toCurrency {
		return decimal.NewFromFloat(1.0)
	}

	// 获取基准汇率的辅助函数
	getRate := func(currency string) (decimal.Decimal, bool) {
		if currency == baseCurrency {
			return decimal.NewFromFloat(1.0), true
		}
		rate, ok := rates[currency]
		return rate, ok
	}

	// 1. 直接 BASE/X
	if fromCurrency == baseCurrency {
		if rate, ok := getRate(toCurrency); ok {
			return rate
		}
	}

	// 2. 直接 X/BASE = 1 / (BASE/X)
	if toCurrency == baseCurrency {
		if rate, ok := getRate(fromCurrency); ok && rate.GreaterThan(decimal.Zero) {
			return decimal.NewFromFloat(1.0).Div(rate)
		}
	}

	// 3. 交叉汇率: X/Y = (BASE/Y) / (BASE/X)
	fromRate, fromOK := getRate(fromCurrency)
	toRate, toOK := getRate(toCurrency)

	if fromOK && toOK && fromRate.GreaterThan(decimal.Zero) {
		return toRate.Div(fromRate)
	}

	return decimal.Zero
}

// ---------------------------------------------------------------------------
// 第三阶段：并发批量写入数据库
// ---------------------------------------------------------------------------

// parallelBatchSave 并发批量保存到数据库
func (s *Synchronizer) parallelBatchSave(ctx context.Context, results []*calcResult, config *BatchSyncConfig) []string {
	errors := make([]string, 0)
	var mu sync.Mutex

	// 先获取所有旧汇率（并行）
	oldRates := s.parallelFetchOldRates(results)

	// 更新计算结果中的旧汇率和变动百分比
	for _, r := range results {
		if oldRate, ok := oldRates[r.FromCurrency+"_"+r.ToCurrency]; ok {
			r.OldRate = oldRate
			if !oldRate.IsZero() {
				r.ChangePercent, _ = r.NewRate.Sub(oldRate).Div(oldRate).Float64()
				r.ChangePercent *= 100
			}
		}
	}

	// 批量写入（分批处理）
	batchSize := config.DBBatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	var wg sync.WaitGroup
	errorChan := make(chan string, len(results))

	// 并行写入不同批次
	for i := 0; i < len(results); i += batchSize {
		end := min(i+batchSize, len(results))
		batch := results[i:end]

		wg.Add(1)
		go func(batch []*calcResult) {
			defer wg.Done()

			if err := s.batchUpsert(batch); err != nil {
				errorChan <- fmt.Sprintf("批次%d-%d写入失败: %v", i, end, err)
			}
		}(batch)
	}

	// 等待所有写入完成
	wg.Wait()
	close(errorChan)

	// 收集错误
	for err := range errorChan {
		mu.Lock()
		errors = append(errors, err)
		mu.Unlock()
	}

	return errors
}

// parallelFetchOldRates 并行获取旧汇率
func (s *Synchronizer) parallelFetchOldRates(results []*calcResult) map[string]decimal.Decimal {
	oldRates := make(map[string]decimal.Decimal)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 并发获取
	for _, r := range results {
		wg.Add(1)
		go func(from, to string) {
			defer wg.Done()

			oldRate := s.getExistingRate(from, to)

			mu.Lock()
			oldRates[from+"_"+to] = oldRate
			mu.Unlock()
		}(r.FromCurrency, r.ToCurrency)
	}

	wg.Wait()
	return oldRates
}

// batchUpsert 批量Upsert汇率数据
func (s *Synchronizer) batchUpsert(results []*calcResult) error {
	if len(results) == 0 {
		return nil
	}

	// 过滤出需要写入的数据
	rates := make([]models.ExchangeRate, 0, len(results))
	for _, r := range results {
		if r.Status == "success" && !r.NewRate.IsZero() {
			rates = append(rates, models.ExchangeRate{
				FromCurrency: r.FromCurrency,
				ToCurrency:   r.ToCurrency,
				Rate:         r.NewRate,
				DataSource:   s.manager.GetCurrentProvider().GetName(),
				SourceType:   "api",
				ValidStatus:  1,
			})
		}
	}

	if len(rates) == 0 {
		return nil
	}

	// 使用GORM的批量Upsert
	return models.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "from_currency"}, {Name: "to_currency"}},
		DoUpdates: clause.AssignmentColumns([]string{"rate", "data_source", "source_type", "valid_status", "updated_at"}),
	}).CreateInBatches(rates, 50).Error
}

// ---------------------------------------------------------------------------
// 旧版方法（保留兼容性）
// ---------------------------------------------------------------------------

// SyncSingleRate 同步单个货币对汇率
func (s *Synchronizer) SyncSingleRate(ctx context.Context, fromCurrency, toCurrency string) (*SyncDetail, error) {
	utils.Info("同步单个汇率",
		"from", fromCurrency,
		"to", toCurrency)

	// 获取新汇率
	newRate, err := s.manager.GetRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		return &SyncDetail{
			FromCurrency: fromCurrency,
			ToCurrency:   toCurrency,
			Status:       "failed",
		}, err
	}

	// 获取旧汇率
	oldRate := s.getExistingRate(fromCurrency, toCurrency)

	// 计算变动
	var change float64
	if !oldRate.IsZero() {
		change = newRate.Sub(oldRate).Div(oldRate).InexactFloat64() * 100
	}

	// 保存到数据库
	if err := s.saveRate(fromCurrency, toCurrency, newRate, s.manager.GetCurrentProvider().GetName()); err != nil {
		return &SyncDetail{
			FromCurrency: fromCurrency,
			ToCurrency:   toCurrency,
			OldRate:      oldRate,
			NewRate:      newRate,
			Change:       change,
			Status:       "failed",
		}, err
	}

	utils.Info("汇率同步成功",
		"from", fromCurrency,
		"to", toCurrency,
		"old_rate", oldRate.String(),
		"new_rate", newRate.String(),
		"change", fmt.Sprintf("%.4f%%", change))

	return &SyncDetail{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		OldRate:      oldRate,
		NewRate:      newRate,
		Change:       change,
		Status:       "success",
	}, nil
}

// SyncFromPrimary 从主数据源同步数据（恢复后同步）
func (s *Synchronizer) SyncFromPrimary(ctx context.Context) (*SyncResult, error) {
	utils.Info("开始从主数据源同步数据（恢复后同步）")

	// 尝试恢复到主数据源
	if !s.manager.RestoreToPrimary(ctx) {
		return nil, fmt.Errorf("主数据源不可用，无法执行恢复同步")
	}

	// 执行全量同步
	return s.SyncAllRates(ctx)
}

// ConsistencyCheck 数据一致性检查
func (s *Synchronizer) ConsistencyCheck(ctx context.Context) (*ConsistencyReport, error) {
	utils.Info("开始数据一致性检查")

	report := &ConsistencyReport{
		CheckTime: time.Now(),
		Items:     make([]ConsistencyItem, 0),
	}

	// 获取所有活跃的货币对
	pairs, err := s.getActiveCurrencyPairs()
	if err != nil {
		return nil, err
	}

	for _, pair := range pairs {
		item := s.checkConsistency(ctx, pair)
		report.Items = append(report.Items, item)
		if item.IsConsistent {
			report.ConsistentCount++
		} else {
			report.InconsistentCount++
		}
	}

	report.TotalCount = len(report.Items)

	utils.Info("数据一致性检查完成",
		"total", report.TotalCount,
		"consistent", report.ConsistentCount,
		"inconsistent", report.InconsistentCount)

	return report, nil
}

// ConsistencyReport 一致性报告
type ConsistencyReport struct {
	CheckTime         time.Time
	TotalCount        int
	ConsistentCount   int
	InconsistentCount int
	Items             []ConsistencyItem
}

// ConsistencyItem 一致性检查项
type ConsistencyItem struct {
	FromCurrency string
	ToCurrency   string
	DBRate       decimal.Decimal
	SourceRate   decimal.Decimal
	DiffPercent  float64
	IsConsistent bool
	DataSource   string
}

// ---------------------------------------------------------------------------
// 内部方法
// ---------------------------------------------------------------------------

// getExistingRate 获取数据库中已有的汇率
func (s *Synchronizer) getExistingRate(fromCurrency, toCurrency string) decimal.Decimal {
	var rate models.ExchangeRate
	result := models.DB.Where(
		"from_currency = ? AND to_currency = ?",
		fromCurrency, toCurrency,
	).Order("updated_at DESC").First(&rate)

	if result.Error == nil {
		return rate.Rate
	}
	return decimal.Zero
}

// saveRate 保存汇率到数据库
func (s *Synchronizer) saveRate(fromCurrency, toCurrency string, rate decimal.Decimal, dataSource string) error {
	exchangeRate := models.ExchangeRate{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		Rate:         rate,
		DataSource:   dataSource,
		SourceType:   "api",
		ValidStatus:  1,
	}

	result := models.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "from_currency"}, {Name: "to_currency"}},
		DoUpdates: clause.AssignmentColumns([]string{"rate", "data_source", "source_type", "valid_status"}),
	}).Create(&exchangeRate)

	return result.Error
}

// getActiveCurrencyPairs 获取所有活跃的货币对
// 如果数据库中没有配置，返回默认货币对列表
func (s *Synchronizer) getActiveCurrencyPairs() ([]models.CurrencyPair, error) {
	var pairs []models.CurrencyPair
	result := models.DB.Where("is_active = ?", 1).Find(&pairs)
	if result.Error != nil {
		return nil, result.Error
	}

	// 如果数据库为空，返回默认货币对列表
	if len(pairs) == 0 {
		utils.Info("数据库中没有配置的货币对，使用默认货币对列表")
		return defaultCurrencyPairs, nil
	}

	return pairs, nil
}

// saveSyncLog 保存同步日志
func (s *Synchronizer) saveSyncLog(result *SyncResult) {
	syncLog := models.ExchangeRateSyncLog{
		BatchID:      result.BatchID,
		SyncType:     result.SyncType,
		DataSource:   result.DataSource,
		Status:       result.Status,
		TotalCount:   result.TotalCount,
		SuccessCount: result.SuccessCount,
		FailedCount:  result.FailedCount,
		DurationMs:   result.DurationMs,
	}

	if err := models.DB.Create(&syncLog).Error; err != nil {
		utils.Error("保存同步日志失败", err)
	}

	// 保存同步详情
	for _, detail := range result.Details {
		syncDetail := models.ExchangeRateSyncDetail{
			SyncLogID:     syncLog.ID,
			FromCurrency:  detail.FromCurrency,
			ToCurrency:    detail.ToCurrency,
			OldRate:       detail.OldRate,
			NewRate:       detail.NewRate,
			ChangePercent: decimal.NewFromFloat(detail.Change),
			Status:        detail.Status,
		}
		models.DB.Create(&syncDetail)
	}
}

// updateFallbackCache 更新Fallback缓存
func (s *Synchronizer) updateFallbackCache(ctx context.Context, baseRates map[string]map[string]decimal.Decimal) {
	// 获取Fallback提供者
	providers := s.manager.GetBackupProviders()
	for _, p := range providers {
		if p.GetName() == "fallback" {
			if fallback, ok := p.(*datasource.FallbackProvider); ok {
				for baseCurrency, rates := range baseRates {
					fallback.BatchUpdateCache(baseCurrency, rates, "batch_sync")
				}
			}
		}
	}
}

// checkConsistency 检查单个货币对的一致性
func (s *Synchronizer) checkConsistency(ctx context.Context, pair models.CurrencyPair) ConsistencyItem {
	item := ConsistencyItem{
		FromCurrency: pair.FromCurrency,
		ToCurrency:   pair.ToCurrency,
	}

	// 获取数据库中的汇率
	item.DBRate = s.getExistingRate(pair.FromCurrency, pair.ToCurrency)

	// 获取数据源中的汇率
	sourceRate, err := s.manager.GetRate(ctx, pair.FromCurrency, pair.ToCurrency)
	if err != nil {
		item.IsConsistent = false
		item.DataSource = s.manager.GetCurrentProvider().GetName()
		return item
	}
	item.SourceRate = sourceRate
	item.DataSource = s.manager.GetCurrentProvider().GetName()

	// 计算差异百分比
	if !item.DBRate.IsZero() {
		item.DiffPercent, _ = sourceRate.Sub(item.DBRate).Div(item.DBRate).Float64()
		item.DiffPercent *= 100
	}

	// 判断一致性（差异不超过0.5%认为一致）
	item.IsConsistent = item.DBRate.IsZero() || (item.DiffPercent >= -0.5 && item.DiffPercent <= 0.5)

	return item
}

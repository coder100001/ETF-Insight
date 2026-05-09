package tasks

import (
	"context"
	"time"

	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/utils"
)

// ASharePriceScheduler A股ETF价格更新调度器
type ASharePriceScheduler struct {
	priceService *services.ASharePriceService
	ticker       *time.Ticker
	stopCh       chan struct{}
	isRunning    bool
}

// NewASharePriceScheduler 创建价格更新调度器
func NewASharePriceScheduler() *ASharePriceScheduler {
	return &ASharePriceScheduler{
		priceService: services.NewASharePriceService(models.DB),
		stopCh:       make(chan struct{}),
		isRunning:    false,
	}
}

// Start 启动定时任务
// interval: 更新间隔（默认5分钟）
func (s *ASharePriceScheduler) Start(interval time.Duration) {
	if s.isRunning {
		utils.Warn("价格更新调度器已在运行")
		return
	}

	if interval <= 0 {
		interval = 5 * time.Minute
	}

	s.ticker = time.NewTicker(interval)
	s.isRunning = true

	utils.Info("启动A股ETF价格更新调度器", "interval", interval)

	// 立即执行一次更新
	go s.updatePrices()

	// 定时更新
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.updatePrices()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop 停止定时任务
func (s *ASharePriceScheduler) Stop() {
	if !s.isRunning {
		return
	}

	s.ticker.Stop()
	close(s.stopCh)
	s.isRunning = false

	utils.Info("停止A股ETF价格更新调度器")
}

// IsRunning 检查是否正在运行
func (s *ASharePriceScheduler) IsRunning() bool {
	return s.isRunning
}

// updatePrices 执行价格更新
func (s *ASharePriceScheduler) updatePrices() {
	utils.Info("开始定时更新A股ETF价格")

	// 检查是否在交易时间（可选）
	if !s.isTradingTime() {
		utils.Info("当前非交易时间，跳过价格更新")
		return
	}

	if err := s.priceService.UpdateAllETFPrices(); err != nil {
		utils.Error("定时更新价格失败", err)
		return
	}

	utils.Info("定时更新A股ETF价格完成")
}

// isTradingTime 检查当前是否在A股交易时间
// 交易时间：周一至周五 9:30-11:30, 13:00-15:00
func (s *ASharePriceScheduler) isTradingTime() bool {
	now := time.Now()
	weekday := now.Weekday()

	// 周末不交易
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	hour := now.Hour()
	minute := now.Minute()
	timeVal := hour*100 + minute

	// 上午交易时间：9:30-11:30
	// 下午交易时间：13:00-15:00
	isMorningSession := timeVal >= 930 && timeVal <= 1130
	isAfternoonSession := timeVal >= 1300 && timeVal <= 1500

	return isMorningSession || isAfternoonSession
}

// UpdateNow 立即执行一次价格更新
func (s *ASharePriceScheduler) UpdateNow() error {
	utils.Info("手动触发A股ETF价格更新")
	return s.priceService.UpdateAllETFPrices()
}

// ==================== 全局调度器管理 ====================

var (
	globalASharePriceScheduler *ASharePriceScheduler
)

// InitASharePriceScheduler 初始化全局价格更新调度器
func InitASharePriceScheduler(interval time.Duration) {
	if globalASharePriceScheduler != nil {
		globalASharePriceScheduler.Stop()
	}

	globalASharePriceScheduler = NewASharePriceScheduler()
	globalASharePriceScheduler.Start(interval)
}

// StopASharePriceScheduler 停止全局价格更新调度器
func StopASharePriceScheduler() {
	if globalASharePriceScheduler != nil {
		globalASharePriceScheduler.Stop()
		globalASharePriceScheduler = nil
	}
}

// GetASharePriceScheduler 获取全局价格更新调度器
func GetASharePriceScheduler() *ASharePriceScheduler {
	return globalASharePriceScheduler
}

// UpdateASharePricesNow 立即更新A股ETF价格
func UpdateASharePricesNow() error {
	if globalASharePriceScheduler != nil {
		return globalASharePriceScheduler.UpdateNow()
	}

	// 如果没有全局调度器，创建临时服务执行更新
	service := services.NewASharePriceService(models.DB)
	return service.UpdateAllETFPrices()
}

// ASharePriceJob A股价格更新任务（用于统一的任务调度框架）
type ASharePriceJob struct{}

// Name 任务名称
func (j *ASharePriceJob) Name() string {
	return "ashare_price_update"
}

// Description 任务描述
func (j *ASharePriceJob) Description() string {
	return "更新A股红利ETF实时价格"
}

// Execute 执行任务
func (j *ASharePriceJob) Execute(ctx context.Context) error {
	return UpdateASharePricesNow()
}

// Schedule 任务调度规则（每5分钟）
func (j *ASharePriceJob) Schedule() string {
	return "*/5 * * * *" // cron格式：每5分钟
}

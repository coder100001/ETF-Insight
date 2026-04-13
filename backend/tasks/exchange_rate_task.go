package tasks

import (
	"time"

	exchangerate "etf-insight/services/exchange_rate"
	"etf-insight/services/exchange_rate/datasource"
	"etf-insight/utils"

	"github.com/robfig/cron/v3"
)

type ExchangeRateTask struct {
	cron        *cron.Cron
	exchangeSvc *exchangerate.ExchangeRateService
	isRunning   bool
}

func NewExchangeRateTask(config *datasource.DataSourceConfig) *ExchangeRateTask {
	return &ExchangeRateTask{
		cron:        cron.New(cron.WithSeconds()),
		exchangeSvc: exchangerate.NewExchangeRateService(config),
	}
}

func (t *ExchangeRateTask) Start() {
	if t.isRunning {
		return
	}

	// 每5分钟同步一次
	_, err := t.cron.AddFunc("0 */5 * * * *", func() {
		t.runFrequentSync()
	})
	if err != nil {
		utils.Error("Failed to add frequent sync cron job", err)
		return
	}

	// 每天10:30全量同步
	_, err = t.cron.AddFunc("0 30 10 * * *", func() {
		t.runDailySync()
	})
	if err != nil {
		utils.Error("Failed to add daily sync cron job", err)
		return
	}

	// 每小时尝试恢复主数据源
	_, err = t.cron.AddFunc("0 0 * * * *", func() {
		t.tryRestorePrimary()
	})
	if err != nil {
		utils.Error("Failed to add restore primary cron job", err)
		return
	}

	t.cron.Start()
	t.isRunning = true
	utils.Info("汇率同步任务已启动",
		"frequent_interval", "5m",
		"daily_sync", "10:30",
		"restore_check", "1h")
}

func (t *ExchangeRateTask) Stop() {
	if !t.isRunning {
		return
	}
	t.cron.Stop()
	t.isRunning = false
	utils.Info("汇率同步任务已停止")
}

func (t *ExchangeRateTask) runFrequentSync() {
	utils.Info("开始高频汇率同步（5分钟间隔）")

	start := time.Now()
	err := t.exchangeSvc.UpdateRates()
	duration := time.Since(start)

	if err != nil {
		utils.Error("高频同步失败", err, "duration_ms", duration.Milliseconds())
		return
	}

	utils.Info("高频同步完成", "duration_ms", duration.Milliseconds())
}

func (t *ExchangeRateTask) runDailySync() {
	utils.Info("开始每日汇率同步")

	start := time.Now()
	err := t.exchangeSvc.UpdateRates()
	duration := time.Since(start)

	if err != nil {
		utils.Error("每日同步失败", err, "duration_ms", duration.Milliseconds())
		return
	}

	utils.Info("每日同步完成", "duration_ms", duration.Milliseconds())
}

func (t *ExchangeRateTask) tryRestorePrimary() {
	if t.exchangeSvc.RestoreToPrimary() {
		utils.Info("已恢复到主数据源")
		// 恢复后执行一次同步
		if err := t.exchangeSvc.SyncFromPrimary(); err != nil {
			utils.Warn("主数据源同步失败", "error", err.Error())
		}
	}
}

func (t *ExchangeRateTask) TriggerManualSync() error {
	return t.exchangeSvc.UpdateRates()
}

func (t *ExchangeRateTask) IsRunning() bool {
	return t.isRunning
}

// GetService 获取汇率服务实例
func (t *ExchangeRateTask) GetService() *exchangerate.ExchangeRateService {
	return t.exchangeSvc
}

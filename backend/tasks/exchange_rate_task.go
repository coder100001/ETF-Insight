package tasks

import (
	"time"

	"etf-insight/services"
	"etf-insight/utils"

	"github.com/robfig/cron/v3"
)

type ExchangeRateTask struct {
	cron        *cron.Cron
	exchangeSvc *services.ExchangeRateService
	isRunning   bool
}

func NewExchangeRateTask() *ExchangeRateTask {
	return &ExchangeRateTask{
		cron:        cron.New(cron.WithSeconds()),
		exchangeSvc: services.NewExchangeRateService(),
	}
}

func (t *ExchangeRateTask) Start() {
	if t.isRunning {
		return
	}

	_, err := t.cron.AddFunc("0 */5 * * * *", func() {
		t.runFrequentSync()
	})
	if err != nil {
		utils.Error("Failed to add frequent sync cron job", err)
		return
	}

	_, err = t.cron.AddFunc("0 30 10 * * *", func() {
		t.runDailySync()
	})
	if err != nil {
		utils.Error("Failed to add daily sync cron job", err)
		return
	}

	t.cron.Start()
	t.isRunning = true
	utils.Info("Exchange rate sync task started", map[string]interface{}{
		"frequent_interval": "5m",
		"daily_sync":        "10:30",
	})
}

func (t *ExchangeRateTask) Stop() {
	if !t.isRunning {
		return
	}
	t.cron.Stop()
	t.isRunning = false
	utils.Info("Exchange rate sync task stopped", nil)
}

func (t *ExchangeRateTask) runFrequentSync() {
	utils.Info("Starting frequent exchange rate sync (5min interval)", nil)

	start := time.Now()
	err := t.exchangeSvc.UpdateRates()
	duration := time.Since(start)

	if err != nil {
		utils.Error("Frequent sync failed", err)
		return
	}

	utils.Info("Frequent sync completed", map[string]interface{}{
		"duration_ms": duration.Milliseconds(),
	})
}

func (t *ExchangeRateTask) runDailySync() {
	utils.Info("Starting daily exchange rate sync", nil)

	start := time.Now()
	err := t.exchangeSvc.UpdateRates()
	duration := time.Since(start)

	if err != nil {
		utils.Error("Daily sync failed", err)
		return
	}

	utils.Info("Daily sync completed", map[string]interface{}{
		"duration_ms": duration.Milliseconds(),
	})
}

func (t *ExchangeRateTask) TriggerManualSync() error {
	return t.exchangeSvc.UpdateRates()
}

func (t *ExchangeRateTask) IsRunning() bool {
	return t.isRunning
}

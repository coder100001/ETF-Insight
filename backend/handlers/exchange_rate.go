package handlers

import (
	"net/http"
	"strconv"
	"time"

	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/tasks"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// ExchangeRateHandler 汇率处理器
type ExchangeRateHandler struct {
	syncService *services.ExchangeRateSyncService
	syncTask    *tasks.ExchangeRateTask
}

// NewExchangeRateHandler 创建汇率处理器
func NewExchangeRateHandler() *ExchangeRateHandler {
	return &ExchangeRateHandler{
		syncService: services.NewExchangeRateSyncService(),
		syncTask:    tasks.NewExchangeRateTask(),
	}
}

// GetExchangeRates 获取汇率列表
func (h *ExchangeRateHandler) GetExchangeRates(c *gin.Context) {
	var rates []models.ExchangeRate
	
	// 只获取有效的汇率数据
	query := models.DB.Where("valid_status = ?", 1)
	
	// 支持按货币对筛选
	fromCurrency := c.Query("from")
	if fromCurrency != "" {
		query = query.Where("from_currency = ?", fromCurrency)
	}
	
	toCurrency := c.Query("to")
	if toCurrency != "" {
		query = query.Where("to_currency = ?", toCurrency)
	}
	
	// 支持按数据源筛选
	dataSource := c.Query("source")
	if dataSource != "" {
		query = query.Where("data_source = ?", dataSource)
	}
	
	// 排序和限制
	if err := query.Order("priority desc, updated_at desc").Find(&rates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch exchange rates",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rates,
		"count":   len(rates),
	})
}

// GetExchangeRate 获取单个汇率
func (h *ExchangeRateHandler) GetExchangeRate(c *gin.Context) {
	fromCurrency := c.Param("from")
	toCurrency := c.Param("to")
	
	if fromCurrency == "" || toCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "From and to currency are required",
		})
		return
	}
	
	var rate models.ExchangeRate
	if err := models.DB.Where("from_currency = ? AND to_currency = ? AND valid_status = ?",
		fromCurrency, toCurrency, 1).First(&rate).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Exchange rate not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rate,
	})
}

// ConvertCurrency 货币转换
func (h *ExchangeRateHandler) ConvertCurrency(c *gin.Context) {
	var req struct {
		From   string  `json:"from" binding:"required"`
		To     string  `json:"to" binding:"required"`
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	var rate models.ExchangeRate
	if err := models.DB.Where("from_currency = ? AND to_currency = ? AND valid_status = ?",
		req.From, req.To, 1).First(&rate).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Exchange rate not found",
		})
		return
	}
	
	// 检查数据是否过期
	if rate.ExpiresAt != nil && rate.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Exchange rate data has expired",
		})
		return
	}
	
	convertedAmount := rate.Rate.Mul(decimal.NewFromFloat(req.Amount))
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"from":             req.From,
			"to":               req.To,
			"amount":           req.Amount,
			"rate":             rate.Rate,
			"converted_amount": convertedAmount,
			"updated_at":       rate.UpdatedAt,
		},
	})
}

// TriggerSync 触发手动同步
func (h *ExchangeRateHandler) TriggerSync(c *gin.Context) {
	var req struct {
		SyncType string `json:"sync_type" binding:"required,oneof=full incremental"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	// 异步执行同步
	go func() {
		result, err := h.syncTask.TriggerManualSync(req.SyncType)
		if err != nil {
			// 记录错误日志
			return
		}
		// 记录成功日志
		_ = result
	}()
	
	c.JSON(http.StatusAccepted, gin.H{
		"success": true,
		"message": "Sync task started",
		"sync_type": req.SyncType,
	})
}

// GetSyncLogs 获取同步日志
func (h *ExchangeRateHandler) GetSyncLogs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	
	logs, err := h.syncService.GetSyncLogs(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch sync logs",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"count":   len(logs),
	})
}

// GetSyncLogDetails 获取同步日志详情
func (h *ExchangeRateHandler) GetSyncLogDetails(c *gin.Context) {
	logID := c.Param("id")
	if logID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Log ID is required",
		})
		return
	}
	
	var log models.ExchangeRateSyncLog
	if err := models.DB.First(&log, logID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Sync log not found",
		})
		return
	}
	
	var details []models.ExchangeRateSyncDetail
	models.DB.Where("sync_log_id = ?", log.ID).Find(&details)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"log":     log,
			"details": details,
		},
	})
}

// GetCurrencyPairs 获取货币对列表
func (h *ExchangeRateHandler) GetCurrencyPairs(c *gin.Context) {
	var pairs []models.CurrencyPair
	
	query := models.DB
	
	// 支持筛选启用状态
	isActive := c.Query("active")
	if isActive != "" {
		query = query.Where("is_active = ?", isActive)
	}
	
	if err := query.Order("priority desc").Find(&pairs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch currency pairs",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    pairs,
		"count":   len(pairs),
	})
}

// CreateCurrencyPair 创建货币对
func (h *ExchangeRateHandler) CreateCurrencyPair(c *gin.Context) {
	var pair models.CurrencyPair
	if err := c.ShouldBindJSON(&pair); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	// 检查是否已存在
	var existing models.CurrencyPair
	if err := models.DB.Where("from_currency = ? AND to_currency = ?", pair.FromCurrency, pair.ToCurrency).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Currency pair already exists",
		})
		return
	}
	
	if err := models.DB.Create(&pair).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create currency pair",
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    pair,
	})
}

// UpdateCurrencyPair 更新货币对
func (h *ExchangeRateHandler) UpdateCurrencyPair(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ID is required",
		})
		return
	}
	
	var pair models.CurrencyPair
	if err := models.DB.First(&pair, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Currency pair not found",
		})
		return
	}
	
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	if err := models.DB.Model(&pair).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update currency pair",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    pair,
	})
}

// DeleteCurrencyPair 删除货币对
func (h *ExchangeRateHandler) DeleteCurrencyPair(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ID is required",
		})
		return
	}
	
	var pair models.CurrencyPair
	if err := models.DB.First(&pair, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Currency pair not found",
		})
		return
	}
	
	if err := models.DB.Delete(&pair).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete currency pair",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Currency pair deleted",
	})
}

// GetSyncStatus 获取同步状态
func (h *ExchangeRateHandler) GetSyncStatus(c *gin.Context) {
	latestLog, err := h.syncService.GetLatestSyncLog()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"is_syncing":  false,
				"last_sync":   nil,
				"task_status": "not_started",
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"is_syncing":   h.syncTask.IsRunning(),
			"last_sync":    latestLog,
			"task_status":  latestLog.Status,
		},
	})
}

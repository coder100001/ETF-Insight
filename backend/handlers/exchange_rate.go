package handlers

import (
	"net/http"
	"strconv"
	"time"

	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/tasks"

	"github.com/gin-gonic/gin"
)

type ExchangeRateHandler struct {
	exchangeSvc *services.ExchangeRateService
	syncTask    *tasks.ExchangeRateTask
}

func NewExchangeRateHandler() *ExchangeRateHandler {
	return &ExchangeRateHandler{
		exchangeSvc: services.NewExchangeRateService(),
		syncTask:    tasks.NewExchangeRateTask(),
	}
}

type ExchangeRateResponse struct {
	ID            uint       `json:"id"`
	FromCurrency  string     `json:"from_currency"`
	ToCurrency    string     `json:"to_currency"`
	Rate          float64    `json:"rate"`
	PreviousRate  float64    `json:"previous_rate"`
	ChangePercent float64    `json:"change_percent"`
	DataSource    string     `json:"data_source"`
	SourceType    string     `json:"source_type"`
	ValidStatus   int        `json:"valid_status"`
	Priority      int        `json:"priority"`
	SyncedAt      *time.Time `json:"synced_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func convertToResponse(rate *models.ExchangeRate) *ExchangeRateResponse {
	return &ExchangeRateResponse{
		ID:            rate.ID,
		FromCurrency:  rate.FromCurrency,
		ToCurrency:    rate.ToCurrency,
		Rate:          rate.Rate.InexactFloat64(),
		PreviousRate:  rate.PreviousRate.InexactFloat64(),
		ChangePercent: rate.ChangePercent.InexactFloat64(),
		DataSource:    rate.DataSource,
		SourceType:    rate.SourceType,
		ValidStatus:   rate.ValidStatus,
		Priority:      rate.Priority,
		SyncedAt:      rate.SyncedAt,
		ExpiresAt:     rate.ExpiresAt,
		CreatedAt:     rate.CreatedAt,
		UpdatedAt:     rate.UpdatedAt,
	}
}

func (h *ExchangeRateHandler) GetExchangeRates(c *gin.Context) {
	var rates []models.ExchangeRate

	query := models.DB.Where("valid_status = ?", 1)

	fromCurrency := c.Query("from")
	if fromCurrency != "" {
		query = query.Where("from_currency = ?", fromCurrency)
	}

	toCurrency := c.Query("to")
	if toCurrency != "" {
		query = query.Where("to_currency = ?", toCurrency)
	}

	dataSource := c.Query("source")
	if dataSource != "" {
		query = query.Where("data_source = ?", dataSource)
	}

	if err := query.Order("priority desc, updated_at desc").Find(&rates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch exchange rates",
		})
		return
	}

	responses := make([]*ExchangeRateResponse, len(rates))
	for i := range rates {
		responses[i] = convertToResponse(&rates[i])
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
		"count":   len(responses),
	})
}

func (h *ExchangeRateHandler) GetExchangeRate(c *gin.Context) {
	fromCurrency := c.Param("from")
	toCurrency := c.Param("to")

	if fromCurrency == "" || toCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Currency pair is required",
		})
		return
	}

	var rate models.ExchangeRate
	result := models.DB.Where(
		"from_currency = ? AND to_currency = ?",
		fromCurrency, toCurrency,
	).Order("updated_at DESC").First(&rate)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Exchange rate not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    convertToResponse(&rate),
	})
}

func (h *ExchangeRateHandler) TriggerSync(c *gin.Context) {
	go func() {
		_ = h.syncTask.TriggerManualSync()
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"success": true,
		"message": "Sync task started",
	})
}

func (h *ExchangeRateHandler) GetCurrencyPairs(c *gin.Context) {
	var pairs []models.CurrencyPair

	query := models.DB

	isActive := c.Query("active")
	if isActive != "" {
		query = query.Where("is_active = ?", isActive)
	}

	if err := query.Order("display_order ASC").Find(&pairs).Error; err != nil {
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

func (h *ExchangeRateHandler) ConvertCurrency(c *gin.Context) {
	var req struct {
		Amount       float64 `json:"amount" binding:"required"`
		FromCurrency string  `json:"from_currency" binding:"required"`
		ToCurrency   string  `json:"to_currency" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	rate := h.exchangeSvc.GetRate(req.FromCurrency, req.ToCurrency)
	converted := req.Amount * rate

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"from_currency":    req.FromCurrency,
			"to_currency":      req.ToCurrency,
			"original_amount":  req.Amount,
			"converted_amount": converted,
			"rate":             rate,
		},
	})
}

func (h *ExchangeRateHandler) GetSupportedCurrencies(c *gin.Context) {
	var currencies []string
	models.DB.Model(&models.ExchangeRate{}).
		Distinct("from_currency").
		Pluck("from_currency", &currencies)

	if currencies == nil {
		currencies = []string{"USD", "CNY", "HKD"}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"currencies": currencies,
	})
}

func (h *ExchangeRateHandler) GetExchangeRatesSummary(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var rates []models.ExchangeRate
	models.DB.Where("valid_status = ?", 1).
		Order("updated_at DESC").
		Limit(limit).
		Find(&rates)

	responses := make([]*ExchangeRateResponse, len(rates))
	for i := range rates {
		responses[i] = convertToResponse(&rates[i])
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
		"count":   len(responses),
	})
}

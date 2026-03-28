package handlers

import (
	"net/http"
	"strconv"

	"etf-insight/services"

	"github.com/gin-gonic/gin"
)

// ExchangeRateHandler 汇率处理器
type ExchangeRateHandler struct {
	service *services.ExchangeRateService
}

// NewExchangeRateHandler 创建新的汇率处理器
func NewExchangeRateHandler(service *services.ExchangeRateService) *ExchangeRateHandler {
	return &ExchangeRateHandler{
		service: service,
	}
}

// GetRates 获取汇率列表
func (h *ExchangeRateHandler) GetRates(c *gin.Context) {
	// 返回常用汇率对
	currencyPairs := []struct {
		From string
		To   string
	}{
		{"USD", "CNY"},
		{"USD", "HKD"},
		{"CNY", "HKD"},
		{"CNY", "USD"},
		{"HKD", "USD"},
		{"HKD", "CNY"},
	}

	var rates []map[string]interface{}
	for _, pair := range currencyPairs {
		rate := h.service.GetRate(pair.From, pair.To)
		rates = append(rates, map[string]interface{}{
			"from_currency": pair.From,
			"to_currency":   pair.To,
			"rate":          rate,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rates,
	})
}

// GetHistory 获取汇率历史
func (h *ExchangeRateHandler) GetHistory(c *gin.Context) {
	fromCurrency := c.DefaultQuery("from", "USD")
	toCurrency := c.DefaultQuery("to", "CNY")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	history, err := h.service.GetHistory(fromCurrency, toCurrency, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"from_currency": fromCurrency,
			"to_currency":   toCurrency,
			"data":          history,
		},
	})
}

// Convert 货币转换
func (h *ExchangeRateHandler) Convert(c *gin.Context) {
	fromCurrency := c.DefaultQuery("from", "USD")
	toCurrency := c.DefaultQuery("to", "CNY")
	amount, err := strconv.ParseFloat(c.DefaultQuery("amount", "0"), 64)
	if err != nil || amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid amount",
		})
		return
	}

	rate := h.service.GetRate(fromCurrency, toCurrency)
	result := amount * rate

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"from_currency": fromCurrency,
			"to_currency":   toCurrency,
			"amount":        amount,
			"rate":          rate,
			"result":        result,
		},
	})
}

// UpdateRates 手动更新汇率
func (h *ExchangeRateHandler) UpdateRates(c *gin.Context) {
	if err := h.service.UpdateRates(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "exchange rates updated",
	})
}

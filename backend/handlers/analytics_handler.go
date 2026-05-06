package handlers

import (
	"net/http"

	"etf-insight/services/analytics"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	client *analytics.Client
}

func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{
		client: analytics.NewClient(),
	}
}

func (h *AnalyticsHandler) Health(c *gin.Context) {
	result, err := h.client.Health()
	if err != nil {
		utils.Error("Analytics service health check failed", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Analytics service unavailable",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *AnalyticsHandler) OptimizePortfolio(c *gin.Context) {
	var req struct {
		Symbols  []string `json:"symbols" binding:"required"`
		Strategy string   `json:"strategy" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.OptimizePortfolio(req.Symbols, req.Strategy)
	if err != nil {
		utils.Error("Portfolio optimization failed", err)
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *AnalyticsHandler) CalculateVaR(c *gin.Context) {
	var req struct {
		Returns    []float64 `json:"returns" binding:"required"`
		Confidence float64   `json:"confidence" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.CalculateVaR(req.Returns, req.Confidence)
	if err != nil {
		utils.Error("VaR calculation failed", err)
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *AnalyticsHandler) CalculateCAPM(c *gin.Context) {
	var req struct {
		RiskFreeRate float64 `json:"risk_free_rate" binding:"required"`
		MarketReturn float64 `json:"market_return" binding:"required"`
		Beta         float64 `json:"beta" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.CalculateCAPM(req.RiskFreeRate, req.MarketReturn, req.Beta)
	if err != nil {
		utils.Error("CAPM calculation failed", err)
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

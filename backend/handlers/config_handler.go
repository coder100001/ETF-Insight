package handlers

import (
	"net/http"

	"etf-insight/config"

	"github.com/gin-gonic/gin"
)

// FinancialConfigResponse 金融配置响应
type FinancialConfigResponse struct {
	RiskFreeRate    float64 `json:"risk_free_rate"`
	TradingDaysYear int     `json:"trading_days_year"`
	DefaultCurrency string  `json:"default_currency"`
}

// GetFinancialConfig 获取金融配置
func GetFinancialConfig(c *gin.Context) {
	cfg := config.GetFinancialConfig()
	c.JSON(http.StatusOK, gin.H{"success": true, "data": FinancialConfigResponse{
		RiskFreeRate:    cfg.RiskFreeRate,
		TradingDaysYear: cfg.TradingDaysYear,
		DefaultCurrency: cfg.DefaultCurrency,
	}})
}

// UpdateFinancialConfigRequest 更新金融配置请求
type UpdateFinancialConfigRequest struct {
	RiskFreeRate    *float64 `json:"risk_free_rate,omitempty"`
	TradingDaysYear *int     `json:"trading_days_year,omitempty"`
}

// UpdateFinancialConfig 更新金融配置
func UpdateFinancialConfig(c *gin.Context) {
	var req UpdateFinancialConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if req.RiskFreeRate != nil {
		if *req.RiskFreeRate < -0.05 || *req.RiskFreeRate > 0.50 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "risk_free_rate must be between -5% and 50%"})
			return
		}
		config.SetRiskFreeRate(*req.RiskFreeRate)
	}
	if req.TradingDaysYear != nil {
		if *req.TradingDaysYear < 1 || *req.TradingDaysYear > 366 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "trading_days_year must be between 1 and 366"})
			return
		}
		config.SetTradingDaysYear(*req.TradingDaysYear)
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "updated"})
}

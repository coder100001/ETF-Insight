package handlers

import (
	"net/http"
	"strconv"

	"etf-insight/services"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type PortfolioHandler struct {
	analysisService *services.ETFAnalysisService
}

func NewPortfolioHandler(analysisService *services.ETFAnalysisService) *PortfolioHandler {
	return &PortfolioHandler{
		analysisService: analysisService,
	}
}

type PortfolioRequest struct {
	Allocation      map[string]float64 `json:"allocation"`
	TotalInvestment float64            `json:"total_investment"`
	TaxRate         float64            `json:"tax_rate"`
}

func (h *PortfolioHandler) AnalyzePortfolio(c *gin.Context) {
	var req PortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	if req.TotalInvestment == 0 {
		req.TotalInvestment = 10000
	}
	if req.TaxRate == 0 {
		req.TaxRate = 0.10
	}

	result, err := h.analysisService.AnalyzePortfolio(
		req.Allocation,
		decimal.NewFromFloat(req.TotalInvestment),
		decimal.NewFromFloat(req.TaxRate),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"total_value":                        result.TotalValue.InexactFloat64(),
		"total_return":                       result.TotalReturn.InexactFloat64(),
		"total_return_pct":                   result.TotalReturnPercent.InexactFloat64(),
		"annual_dividend":                    result.AnnualDividendAfterTax.InexactFloat64(),
		"dividend_yield":                     result.WeightedDividendYield.InexactFloat64(),
		"tax_rate":                           result.TaxRate.InexactFloat64() * 100,
		"after_tax_return":                   result.TotalReturnWithDividend.InexactFloat64(),
		"holdings":                           result.Holdings,
		"total_investment":                   result.TotalInvestment.InexactFloat64(),
		"annual_dividend_before_tax":         result.AnnualDividendBeforeTax.InexactFloat64(),
		"dividend_tax":                       result.DividendTax.InexactFloat64(),
		"total_return_with_dividend":         result.TotalReturnWithDividend.InexactFloat64(),
		"total_return_with_dividend_percent": result.TotalReturnWithDividendPercent.InexactFloat64(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

func (h *PortfolioHandler) GetPortfolioConfigs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": []map[string]interface{}{
			{
				"id":               1,
				"name":             "保守型组合",
				"description":      "低风险稳健配置",
				"allocation":       map[string]float64{"SCHD": 60, "VNQ": 20, "VYM": 20},
				"total_investment": 50000,
				"status":           1,
			},
			{
				"id":               2,
				"name":             "成长型组合",
				"description":      "高风险高收益配置",
				"allocation":       map[string]float64{"QQQ": 70, "SCHD": 30},
				"total_investment": 100000,
				"status":           1,
			},
		},
	})
}

func (h *PortfolioHandler) GetPortfolioConfig(c *gin.Context) {
	id := c.Param("id")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"id":               id,
			"name":             "示例组合",
			"description":      "示例描述",
			"allocation":       map[string]float64{"SCHD": 50, "SPYD": 50},
			"total_investment": 10000,
			"status":           1,
		},
	})
}

func (h *PortfolioHandler) CreatePortfolioConfig(c *gin.Context) {
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	config["id"] = 3
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    config,
	})
}

func (h *PortfolioHandler) UpdatePortfolioConfig(c *gin.Context) {
	id := c.Param("id")

	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	config["id"] = id
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

func (h *PortfolioHandler) DeletePortfolioConfig(c *gin.Context) {
	id := c.Param("id")
	_ = id

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Config deleted",
	})
}

func (h *PortfolioHandler) TogglePortfolioConfigStatus(c *gin.Context) {
	id := c.Param("id")
	idInt, _ := strconv.Atoi(id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"id":     idInt,
			"status": 1,
		},
	})
}

func (h *PortfolioHandler) AnalyzePortfolioConfig(c *gin.Context) {
	id := c.Param("id")
	_ = id

	var req struct {
		TaxRate float64 `json:"tax_rate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.TaxRate = 0.10
	}

	allocation := map[string]float64{"SCHD": 50, "SPYD": 50}
	result, err := h.analysisService.AnalyzePortfolio(
		allocation,
		decimal.NewFromFloat(10000),
		decimal.NewFromFloat(req.TaxRate),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

package handlers

import (
	"net/http"

	"etf-insight/services"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type PortfolioOptimizerHandler struct {
	optimizer *services.PortfolioOptimizer
}

func NewPortfolioOptimizerHandler(optimizer *services.PortfolioOptimizer) *PortfolioOptimizerHandler {
	return &PortfolioOptimizerHandler{
		optimizer: optimizer,
	}
}

type OptimizePortfolioRequest struct {
	Symbols           []string `json:"symbols" binding:"required,min=2,max=20"`
	OptimizationType  string   `json:"optimization_type"`
	RiskFreeRate      float64  `json:"risk_free_rate"`
	TargetReturn      float64  `json:"target_return"`
	MaxWeightPerAsset float64  `json:"max_weight_per_asset"`
	MinWeightPerAsset float64  `json:"min_weight_per_asset"`
}

type FrontierRequest struct {
	Symbols      []string `json:"symbols" binding:"required,min=2,max=20"`
	RiskFreeRate float64  `json:"risk_free_rate"`
	NumPoints    int      `json:"num_points"`
}

func (h *PortfolioOptimizerHandler) OptimizePortfolio(c *gin.Context) {
	var req OptimizePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if len(req.Symbols) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "At least 2 symbols required",
		})
		return
	}

	if len(req.Symbols) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Maximum 20 symbols allowed",
		})
		return
	}

	optType := services.OptimizationTypeMaxSharpe
	switch req.OptimizationType {
	case "max_sharpe":
		optType = services.OptimizationTypeMaxSharpe
	case "min_volatility":
		optType = services.OptimizationTypeMinVolatility
	case "equal_weight":
		optType = services.OptimizationTypeEqualWeight
	}

	riskFreeRate := decimal.NewFromFloat(req.RiskFreeRate)
	if riskFreeRate.IsZero() {
		riskFreeRate = decimal.NewFromFloat(0.04)
	}

	maxWeight := decimal.NewFromFloat(req.MaxWeightPerAsset)
	if maxWeight.IsZero() {
		maxWeight = decimal.NewFromFloat(0.4)
	}

	minWeight := decimal.NewFromFloat(req.MinWeightPerAsset)
	if minWeight.IsZero() {
		minWeight = decimal.NewFromFloat(0.05)
	}

	optReq := services.PortfolioOptimizationRequest{
		Symbols:          req.Symbols,
		OptimizationType: optType,
		RiskFreeRate:     riskFreeRate,
		Constraints: services.OptimizationConstraints{
			MaxWeightPerAsset: maxWeight,
			MinWeightPerAsset: minWeight,
			AllowShort:        false,
		},
	}

	result, err := h.optimizer.Optimize(optReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Optimization failed: " + err.Error(),
		})
		return
	}

	weights := make(map[string]float64)
	for symbol, weight := range result.Weights {
		weights[symbol] = weight.InexactFloat64()
	}

	response := gin.H{
		"success": true,
		"data": gin.H{
			"weights":             weights,
			"expected_return":     result.ExpectedReturn.InexactFloat64(),
			"expected_volatility": result.ExpectedVolatility.InexactFloat64(),
			"sharpe_ratio":        result.SharpeRatio.InexactFloat64(),
			"optimization_type":   result.OptimizationType,
			"risk_free_rate":      result.RiskFreeRate.InexactFloat64(),
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *PortfolioOptimizerHandler) GetEfficientFrontier(c *gin.Context) {
	var req FrontierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	riskFreeRate := decimal.NewFromFloat(req.RiskFreeRate)
	if riskFreeRate.IsZero() {
		riskFreeRate = decimal.NewFromFloat(0.04)
	}

	optReq := services.PortfolioOptimizationRequest{
		Symbols:      req.Symbols,
		RiskFreeRate: riskFreeRate,
		Constraints: services.OptimizationConstraints{
			MaxWeightPerAsset: decimal.NewFromFloat(0.4),
			MinWeightPerAsset: decimal.NewFromFloat(0.05),
			AllowShort:        false,
		},
	}

	frontier, err := h.optimizer.GetEfficientFrontier(optReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate efficient frontier: " + err.Error(),
		})
		return
	}

	frontierData := make([]map[string]any, 0, len(frontier))
	for _, point := range frontier {
		weights := make(map[string]float64)
		for symbol, weight := range point.Weights {
			weights[symbol] = weight.InexactFloat64()
		}
		frontierData = append(frontierData, map[string]any{
			"expected_return": point.ExpectedReturn.InexactFloat64(),
			"volatility":      point.Volatility.InexactFloat64(),
			"weights":         weights,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"frontier":       frontierData,
			"risk_free_rate": riskFreeRate.InexactFloat64(),
		},
	})
}

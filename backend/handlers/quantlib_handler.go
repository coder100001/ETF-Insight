package handlers

import (
	"net/http"

	"etf-insight/models"
	"etf-insight/services/quantlib"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
)

type QuantLibHandler struct {
	client *quantlib.Client
}

func NewQuantLibHandler() *QuantLibHandler {
	return &QuantLibHandler{
		client: quantlib.NewClient(),
	}
}

func (h *QuantLibHandler) PriceEuropeanOption(c *gin.Context) {
	var req models.EuropeanOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request format"})
		return
	}

	if err := models.ValidateEuropeanOptionRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.PriceEuropeanOption(req)
	if err != nil {
		utils.Error("Failed to price European option", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Quantitative analysis service temporarily unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *QuantLibHandler) PriceAmericanOption(c *gin.Context) {
	var req models.AmericanOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request format"})
		return
	}

	if err := models.ValidateAmericanOptionRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.PriceAmericanOption(req)
	if err != nil {
		utils.Error("Failed to price American option", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Quantitative analysis service temporarily unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *QuantLibHandler) CalculateGreeks(c *gin.Context) {
	var req models.GreeksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request format"})
		return
	}

	if err := models.ValidateGreeksRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.CalculateGreeks(req)
	if err != nil {
		utils.Error("Failed to calculate Greeks", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Quantitative analysis service temporarily unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *QuantLibHandler) BuildYieldCurve(c *gin.Context) {
	var req models.YieldCurveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request format"})
		return
	}

	if err := models.ValidateYieldCurveRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.BuildYieldCurve(req)
	if err != nil {
		utils.Error("Failed to build yield curve", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Quantitative analysis service temporarily unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *QuantLibHandler) PriceBond(c *gin.Context) {
	var req models.BondRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request format"})
		return
	}

	if err := models.ValidateBondRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.PriceBond(req)
	if err != nil {
		utils.Error("Failed to price bond", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Quantitative analysis service temporarily unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *QuantLibHandler) CalculateVaR(c *gin.Context) {
	var req models.VaRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request format"})
		return
	}

	if err := models.ValidateVaRRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.client.CalculateVaR(req)
	if err != nil {
		utils.Error("Failed to calculate VaR", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Quantitative analysis service temporarily unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *QuantLibHandler) GetReferenceData(c *gin.Context) {
	dataType := c.Param("type")

	validTypes := map[string]bool{
		"currencies":            true,
		"frequencies":           true,
		"calendars":             true,
		"day-count-conventions": true,
	}

	if !validTypes[dataType] {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid reference data type"})
		return
	}

	var result interface{}
	var err error

	switch dataType {
	case "currencies":
		result, err = h.client.GetSupportedCurrencies()
	case "frequencies":
		result, err = h.client.GetFrequencies()
	case "calendars":
		result, err = h.client.GetCalendars()
	case "day-count-conventions":
		result, err = h.client.GetDayCountConventions()
	}

	if err != nil {
		utils.Error("Failed to get reference data", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Reference data service temporarily unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

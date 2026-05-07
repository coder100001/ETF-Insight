package handlers

import (
	"net/http"
	"time"

	"etf-insight/services"

	"github.com/gin-gonic/gin"
)

type FactorTimingHandler struct {
	service *services.FactorDataService
}

func NewFactorTimingHandler(service *services.FactorDataService) *FactorTimingHandler {
	return &FactorTimingHandler{service: service}
}

func (h *FactorTimingHandler) CalculateFactorTiming(c *gin.Context) {
	var req struct {
		FactorName   string `json:"factor_name" binding:"required"`
		LookbackDays int    `json:"lookback_days"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.LookbackDays <= 0 {
		req.LookbackDays = 252
	}

	signal, err := h.service.CalculateTimingSignal(req.FactorName, req.LookbackDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": signal})
}

func (h *FactorTimingHandler) GetFactorTimingHistory(c *gin.Context) {
	factorName := c.Param("factor_name")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if factorName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "factor_name is required"})
		return
	}

	var signals []any

	if startDate != "" && endDate != "" {
		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid start_date format, use YYYY-MM-DD"})
			return
		}
		end, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid end_date format, use YYYY-MM-DD"})
			return
		}
		result, err := h.service.GetTimingSignals(factorName, start, end)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
		for _, s := range result {
			signals = append(signals, s)
		}
	} else {
		latest, err := h.service.GetLatestTimingSignal(factorName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
		if latest != nil {
			signals = append(signals, latest)
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": signals})
}

func (h *FactorTimingHandler) GetLatestSignal(c *gin.Context) {
	factorName := c.Param("factor_name")
	if factorName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "factor_name is required"})
		return
	}

	signal, err := h.service.GetLatestTimingSignal(factorName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if signal == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "no timing signal found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": signal})
}

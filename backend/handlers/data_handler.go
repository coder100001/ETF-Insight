package handlers

import (
	"net/http"

	"etf-insight/services/data"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
)

type DataHandler struct {
	client *data.Client
}

func NewDataHandler() *DataHandler {
	return &DataHandler{
		client: data.NewClient(),
	}
}

func (h *DataHandler) Health(c *gin.Context) {
	result, err := h.client.Health()
	if err != nil {
		utils.Error("Data service health check failed", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Data service unavailable",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *DataHandler) FredSeries(c *gin.Context) {
	seriesID := c.Param("series_id")
	params := map[string]string{}
	if v := c.Query("start_date"); v != "" {
		params["start_date"] = v
	}
	if v := c.Query("end_date"); v != "" {
		params["end_date"] = v
	}

	result, err := h.client.FredSeries(seriesID, params)
	if err != nil {
		utils.Error("FRED request failed", err)
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *DataHandler) YFinanceQuote(c *gin.Context) {
	symbol := c.Param("symbol")
	result, err := h.client.YFinanceQuote(symbol)
	if err != nil {
		utils.Error("Yahoo Finance request failed", err)
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *DataHandler) AkShareStockSpot(c *gin.Context) {
	result, err := h.client.AkShareStockSpot()
	if err != nil {
		utils.Error("AkShare request failed", err)
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

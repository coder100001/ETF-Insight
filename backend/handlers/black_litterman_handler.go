package handlers

import (
	"net/http"
	"strconv"

	"etf-insight/models"
	"etf-insight/services"

	"github.com/gin-gonic/gin"
)

type BlackLittermanHandler struct {
	service *services.BlackLittermanService
}

func NewBlackLittermanHandler(service *services.BlackLittermanService) *BlackLittermanHandler {
	return &BlackLittermanHandler{service: service}
}

func (h *BlackLittermanHandler) CreateConfig(c *gin.Context) {
	var config models.BlackLittermanConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if err := h.service.CreateConfig(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": config})
}

func (h *BlackLittermanHandler) GetConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid config id"})
		return
	}

	config, err := h.service.GetConfig(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "config not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": config})
}

func (h *BlackLittermanHandler) UpdateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid config id"})
		return
	}

	var config models.BlackLittermanConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	config.ID = uint(id)
	if err := h.service.UpdateConfig(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": config})
}

func (h *BlackLittermanHandler) CalculatePosterior(c *gin.Context) {
	var req struct {
		ConfigID uint   `json:"config_id" binding:"required"`
		ViewIDs  []uint `json:"view_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	result, err := h.service.CalculatePosteriorReturnsByIDs(req.ConfigID, req.ViewIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "no posterior results computed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *BlackLittermanHandler) GetPosteriorResults(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid config id"})
		return
	}

	result, err := h.service.GetPosteriorReturns(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "no posterior results computed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

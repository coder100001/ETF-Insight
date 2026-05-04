package handlers

import (
	"net/http"
	"strconv"

	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
)

type AlphaViewHandler struct {
	service *services.AlphaViewService
}

func NewAlphaViewHandler(service *services.AlphaViewService) *AlphaViewHandler {
	return &AlphaViewHandler{service: service}
}

func (h *AlphaViewHandler) CreateView(c *gin.Context) {
	var view models.AlphaView
	if err := c.ShouldBindJSON(&view); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request body", "error": "bad request"})
		return
	}

	if err := h.service.CreateAlphaView(&view); err != nil {
		utils.Error("Failed to create alpha view", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": view})
}

func (h *AlphaViewHandler) GetActiveViews(c *gin.Context) {
	assetSymbol := c.Query("asset_symbol")

	activeViews, err := h.service.GetActiveAlphaViews(assetSymbol)
	if err != nil {
		utils.Error("Failed to get active views", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": activeViews})
}

func (h *AlphaViewHandler) GetView(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid view id"})
		return
	}

	view, err := h.service.GetAlphaView(uint(id))
	if err != nil {
		utils.Error("Failed to get alpha view", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "internal server error"})
		return
	}

	if view == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "view not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": view})
}

func (h *AlphaViewHandler) GenerateFromFactor(c *gin.Context) {
	var req struct {
		FactorName  string `json:"factor_name" binding:"required"`
		AssetSymbol string `json:"asset_symbol" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request body", "error": "bad request"})
		return
	}

	view, err := h.service.GenerateViewFromFactorTiming(req.FactorName, req.AssetSymbol)
	if err != nil {
		utils.Error("Failed to generate view from factor", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": view})
}

func (h *AlphaViewHandler) UpdateView(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid view id"})
		return
	}

	var view models.AlphaView
	if err := c.ShouldBindJSON(&view); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request body", "error": "bad request"})
		return
	}

	view.ID = uint(id)
	if err := h.service.UpdateAlphaView(&view); err != nil {
		utils.Error("Failed to update alpha view", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": view})
}

func (h *AlphaViewHandler) DeactivateView(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid view id"})
		return
	}

	if err := h.service.DeactivateView(uint(id)); err != nil {
		utils.Error("Failed to deactivate view", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "view deactivated"})
}

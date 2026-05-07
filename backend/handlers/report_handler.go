package handlers

import (
	"net/http"
	"strconv"

	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService *services.ReportService
}

func NewReportHandler(reportService *services.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

// ===== 请求结构体 =====

type GenerateReportRequest struct {
	TemplateID uint           `json:"template_id" binding:"required"`
	Title      string         `json:"title" binding:"required"`
	Format     string         `json:"format" binding:"required"`
	Data       map[string]any `json:"data"`
}

type CreateTemplateRequest struct {
	Name        string                   `json:"name" binding:"required"`
	Description string                   `json:"description"`
	Category    string                   `json:"category" binding:"required"`
	Config      map[string]any           `json:"config"`
	Sections    []models.ReportSection   `json:"sections"`
	Parameters  []models.ReportParameter `json:"parameters"`
}

type UpdateTemplateRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Config      map[string]any `json:"config"`
}

// ===== 模板相关接口 =====

// GetTemplates 获取所有报告模板
func (h *ReportHandler) GetTemplates(c *gin.Context) {
	templates, err := h.reportService.GetTemplates()
	if err != nil {
		utils.Error("Failed to get templates", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get templates",
			"error":   "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    templates,
	})
}

// GetDefaultTemplates 获取默认报告模板
func (h *ReportHandler) GetDefaultTemplates(c *gin.Context) {
	templates, err := h.reportService.GetDefaultTemplates()
	if err != nil {
		utils.Error("Failed to get default templates", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get default templates",
			"error":   "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    templates,
	})
}

// GetTemplate 获取单个报告模板
func (h *ReportHandler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid template ID",
		})
		return
	}

	template, err := h.reportService.GetTemplate(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Template not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    template,
	})
}

// CreateTemplate 创建报告模板
func (h *ReportHandler) CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	template := &models.ReportTemplate{
		Name:        req.Name,
		Description: req.Description,
		Category:    models.ReportCategory(req.Category),
		Sections:    req.Sections,
	}

	if err := h.reportService.CreateTemplate(template); err != nil {
		utils.Error("Failed to create template", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create template",
			"error":   "internal server error",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    template,
		"message": "Template created successfully",
	})
}

// UpdateTemplate 更新报告模板
func (h *ReportHandler) UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid template ID",
		})
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	template, err := h.reportService.GetTemplate(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Template not found",
		})
		return
	}

	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Description != "" {
		template.Description = req.Description
	}

	if err := h.reportService.UpdateTemplate(template); err != nil {
		utils.Error("Failed to update template", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update template",
			"error":   "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    template,
		"message": "Template updated successfully",
	})
}

// DeleteTemplate 删除报告模板
func (h *ReportHandler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid template ID",
		})
		return
	}

	if err := h.reportService.DeleteTemplate(uint(id)); err != nil {
		utils.Error("Failed to delete template", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to delete template",
			"error":   "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Template deleted successfully",
	})
}

// ===== 报告相关接口 =====

// GenerateReport 生成报告
func (h *ReportHandler) GenerateReport(c *gin.Context) {
	var req GenerateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	report, err := h.reportService.GenerateReport(
		req.TemplateID,
		req.Title,
		models.ReportFormat(req.Format),
		req.Data,
	)
	if err != nil {
		utils.Error("Failed to generate report", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to generate report",
			"error":   "internal server error",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"success": true,
		"data":    report,
		"message": "Report generation started",
	})
}

// GetReport 获取报告详情
func (h *ReportHandler) GetReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid report ID",
		})
		return
	}

	report, err := h.reportService.GetReport(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Report not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    report,
	})
}

// GetReports 获取报告列表（分页）
func (h *ReportHandler) GetReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	reports, total, err := h.reportService.GetReports(page, pageSize)
	if err != nil {
		utils.Error("Failed to get reports", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get reports",
			"error":   "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reports,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// DeleteReport 删除报告
func (h *ReportHandler) DeleteReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid report ID",
		})
		return
	}

	if err := h.reportService.DeleteReport(uint(id)); err != nil {
		utils.Error("Failed to delete report", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to delete report",
			"error":   "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Report deleted successfully",
	})
}

// DownloadReport 下载报告
func (h *ReportHandler) DownloadReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid report ID",
		})
		return
	}

	report, err := h.reportService.GetReport(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Report not found",
		})
		return
	}

	if report.Status != models.ReportStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Report not ready yet",
		})
		return
	}

	// TODO: 实际的文件下载逻辑
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"download_url": report.FilePath,
			"file_name":    report.Title,
			"file_size":    report.FileSize,
		},
	})
}

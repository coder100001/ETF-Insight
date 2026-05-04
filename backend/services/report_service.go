package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"etf-insight/models"
	"etf-insight/utils"

	"gorm.io/gorm"
)

var (
	ErrTemplateNotFound = errors.New("report template not found")
	ErrReportNotFound   = errors.New("generated report not found")
	ErrInvalidFormat    = errors.New("invalid report format")
	ErrGenerationFailed = errors.New("report generation failed")
	ErrReportInProgress = errors.New("report generation already in progress")
)

type ReportService struct {
	db *gorm.DB
}

func NewReportService(db *gorm.DB) *ReportService {
	return &ReportService{
		db: db,
	}
}

// ===== 模板相关方法 =====

func (s *ReportService) GetTemplates() ([]models.ReportTemplate, error) {
	var templates []models.ReportTemplate
	err := s.db.Preload("Sections").Find(&templates).Error
	return templates, err
}

func (s *ReportService) GetDefaultTemplates() ([]models.ReportTemplate, error) {
	var templates []models.ReportTemplate
	err := s.db.Preload("Sections").Preload("Parameters").Where("is_default = ?", true).Find(&templates).Error
	return templates, err
}

func (s *ReportService) GetTemplate(id uint) (*models.ReportTemplate, error) {
	var template models.ReportTemplate
	err := s.db.Preload("Sections").Preload("Parameters").First(&template, id).Error
	if err != nil {
		return nil, ErrTemplateNotFound
	}
	return &template, nil
}

func (s *ReportService) CreateTemplate(template *models.ReportTemplate) error {
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	return s.db.Create(template).Error
}

func (s *ReportService) UpdateTemplate(template *models.ReportTemplate) error {
	template.UpdatedAt = time.Now()
	return s.db.Save(template).Error
}

func (s *ReportService) DeleteTemplate(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("template_id = ?", id).Delete(&models.ReportSection{}).Error; err != nil {
			return err
		}
		if err := tx.Where("template_id = ?", id).Delete(&models.ReportParameter{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.ReportTemplate{}, id).Error
	})
}

// ===== 报告生成相关方法 =====

func (s *ReportService) CreateReport(report *models.GeneratedReport) error {
	report.Status = models.ReportStatusPending
	report.CreatedAt = time.Now()
	return s.db.Create(report).Error
}

func (s *ReportService) GetReport(id uint) (*models.GeneratedReport, error) {
	var report models.GeneratedReport
	err := s.db.Preload("Template").Preload("Template.Sections").First(&report, id).Error
	if err != nil {
		return nil, ErrReportNotFound
	}
	return &report, nil
}

func (s *ReportService) GetReports(page, pageSize int) ([]models.GeneratedReport, int64, error) {
	var reports []models.GeneratedReport
	var total int64

	offset := (page - 1) * pageSize

	err := s.db.Model(&models.GeneratedReport{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = s.db.Preload("Template").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&reports).Error

	return reports, total, err
}

func (s *ReportService) UpdateReportStatus(id uint, status models.ReportStatus, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	if status == models.ReportStatusCompleted || status == models.ReportStatusFailed {
		now := time.Now()
		updates["completed_at"] = &now
	}

	return s.db.Model(&models.GeneratedReport{}).Where("id = ?", id).Updates(updates).Error
}

func (s *ReportService) UpdateReportFile(id uint, filePath string, fileSize int64) error {
	return s.db.Model(&models.GeneratedReport{}).Where("id = ?", id).Updates(map[string]interface{}{
		"file_path": filePath,
		"file_size": fileSize,
	}).Error
}

func (s *ReportService) DeleteReport(id uint) error {
	return s.db.Delete(&models.GeneratedReport{}, id).Error
}

// ===== 报告生成核心方法 =====

func (s *ReportService) GenerateReport(templateID uint, title string, format models.ReportFormat, data map[string]interface{}) (*models.GeneratedReport, error) {
	// 1. 获取模板
	template, err := s.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// 2. 验证格式
	if format != models.ReportFormatHTML && format != models.ReportFormatPDF && format != models.ReportFormatExcel {
		return nil, ErrInvalidFormat
	}

	// 3. 创建报告记录
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report data: %w", err)
	}
	report := &models.GeneratedReport{
		TemplateID: templateID,
		Title:      title,
		Format:     format,
		Status:     models.ReportStatusPending,
		Data:       string(dataJSON),
	}

	if err := s.CreateReport(report); err != nil {
		return nil, err
	}

	// 4. 异步生成报告（这里先同步执行，后面可以改为异步）
	go s.asyncGenerateReport(report.ID, template, format, data)

	return report, nil
}

func (s *ReportService) asyncGenerateReport(reportID uint, template *models.ReportTemplate, format models.ReportFormat, data map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("panic during report generation: %v", r)
			utils.Error("Report generation panic", fmt.Errorf("%s", errMsg))
			s.UpdateReportStatus(reportID, models.ReportStatusFailed, errMsg)
		}
	}()

	// 更新状态为生成中
	s.UpdateReportStatus(reportID, models.ReportStatusGenerating, "")

	// 模拟生成过程
	time.Sleep(2 * time.Second)

	// TODO: 实际的报告生成逻辑
	filePath := fmt.Sprintf("/tmp/report_%d.%s", reportID, format)
	fileSize := int64(1024) // 模拟文件大小

	// 更新状态为完成
	s.UpdateReportFile(reportID, filePath, fileSize)
	s.UpdateReportStatus(reportID, models.ReportStatusCompleted, "")
}

// ===== 参数相关方法 =====

func (s *ReportService) GetParameters(templateID uint) ([]models.ReportParameter, error) {
	var params []models.ReportParameter
	err := s.db.Where("template_id = ?", templateID).Order("order ASC").Find(&params).Error
	return params, err
}

func (s *ReportService) CreateParameter(param *models.ReportParameter) error {
	param.CreatedAt = time.Now()
	param.UpdatedAt = time.Now()
	return s.db.Create(param).Error
}

func (s *ReportService) UpdateParameter(param *models.ReportParameter) error {
	param.UpdatedAt = time.Now()
	return s.db.Save(param).Error
}

func (s *ReportService) DeleteParameter(id uint) error {
	return s.db.Delete(&models.ReportParameter{}, id).Error
}

// ===== 章节相关方法 =====

func (s *ReportService) GetSections(templateID uint) ([]models.ReportSection, error) {
	var sections []models.ReportSection
	err := s.db.Where("template_id = ?", templateID).Order("order ASC").Find(&sections).Error
	return sections, err
}

func (s *ReportService) CreateSection(section *models.ReportSection) error {
	section.CreatedAt = time.Now()
	return s.db.Create(section).Error
}

func (s *ReportService) UpdateSection(section *models.ReportSection) error {
	return s.db.Save(section).Error
}

func (s *ReportService) DeleteSection(id uint) error {
	return s.db.Delete(&models.ReportSection{}, id).Error
}

// ===== 初始化默认数据 =====

func (s *ReportService) InitDefaultTemplates() error {
	// 检查是否已有默认模板
	var count int64
	s.db.Model(&models.ReportTemplate{}).Where("is_default = ?", true).Count(&count)
	if count > 0 {
		return nil // 已初始化
	}

	// 默认模板数据
	defaultTemplates := []models.ReportTemplate{
		{
			Name:        "投资组合分析报告",
			Description: "全面的投资组合分析报告，包含收益、风险、配置等指标",
			Category:    models.ReportCategoryPortfolio,
			IsDefault:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "风险分析报告",
			Description: "深度风险分析报告，包含VaR、回撤、相关性分析等",
			Category:    models.ReportCategoryRisk,
			IsDefault:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "ETF对比报告",
			Description: "多只ETF对比分析报告，包含持仓重叠、费率、表现对比",
			Category:    models.ReportCategoryETFComparison,
			IsDefault:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "市场分析周报",
			Description: "每周市场分析报告，包含指数表现、ETF资金流向等",
			Category:    models.ReportCategoryMarket,
			IsDefault:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// 创建模板
	for _, template := range defaultTemplates {
		if err := s.db.Create(&template).Error; err != nil {
			return err
		}
	}

	return nil
}

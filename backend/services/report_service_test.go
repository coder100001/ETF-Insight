package services

import (
	"testing"
	"time"

	"etf-insight/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupReportTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(
		&models.ReportTemplate{},
		&models.ReportSection{},
		&models.ReportParameter{},
		&models.GeneratedReport{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func cleanupReportTestDB(db *gorm.DB) {
	sqlDB, _ := db.DB()
	sqlDB.Close()
}

func TestReportService_CreateTemplate(t *testing.T) {
	db := setupReportTestDB(t)
	defer cleanupReportTestDB(db)

	service := NewReportService(db)

	template := &models.ReportTemplate{
		Name:        "Test Template",
		Description: "A test template",
		Category:    models.ReportCategoryPortfolio,
		IsDefault:   false,
	}

	err := service.CreateTemplate(template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	if template.ID == 0 {
		t.Error("Template ID should be set after creation")
	}
}

func TestReportService_GetTemplate(t *testing.T) {
	db := setupReportTestDB(t)
	defer cleanupReportTestDB(db)

	service := NewReportService(db)

	template := &models.ReportTemplate{
		Name:        "Test Template",
		Description: "A test template",
		Category:    models.ReportCategoryPortfolio,
		IsDefault:   true,
	}
	service.CreateTemplate(template)

	section := &models.ReportSection{
		TemplateID: template.ID,
		Title:      "Section 1",
		Type:       "text",
		Order:      1,
	}
	service.CreateSection(section)

	param := &models.ReportParameter{
		TemplateID: template.ID,
		Name:       "portfolio_id",
		Label:      "Portfolio",
		Type:       "select",
		Required:   true,
	}
	service.CreateParameter(param)

	result, err := service.GetTemplate(template.ID)
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}

	if result.Name != template.Name {
		t.Errorf("Expected name %s, got %s", template.Name, result.Name)
	}

	if len(result.Sections) != 1 {
		t.Errorf("Expected 1 section, got %d", len(result.Sections))
	}

	if len(result.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(result.Parameters))
	}
}

func TestReportService_GetTemplate_NotFound(t *testing.T) {
	db := setupReportTestDB(t)
	defer cleanupReportTestDB(db)

	service := NewReportService(db)

	_, err := service.GetTemplate(999)
	if err != ErrTemplateNotFound {
		t.Errorf("Expected ErrTemplateNotFound, got %v", err)
	}
}

func TestReportService_GetDefaultTemplates(t *testing.T) {
	db := setupReportTestDB(t)
	defer cleanupReportTestDB(db)

	service := NewReportService(db)

	template1 := &models.ReportTemplate{
		Name:      "Default Template 1",
		Category:  models.ReportCategoryPortfolio,
		IsDefault: true,
	}
	template2 := &models.ReportTemplate{
		Name:      "Non-Default Template",
		Category:  models.ReportCategoryRisk,
		IsDefault: false,
	}
	template3 := &models.ReportTemplate{
		Name:      "Default Template 2",
		Category:  models.ReportCategoryMarket,
		IsDefault: true,
	}

	service.CreateTemplate(template1)
	service.CreateTemplate(template2)
	service.CreateTemplate(template3)

	templates, err := service.GetDefaultTemplates()
	if err != nil {
		t.Fatalf("Failed to get default templates: %v", err)
	}

	if len(templates) != 2 {
		t.Errorf("Expected 2 default templates, got %d", len(templates))
	}
}

func TestReportService_DeleteTemplate(t *testing.T) {
	db := setupReportTestDB(t)
	defer cleanupReportTestDB(db)

	service := NewReportService(db)

	template := &models.ReportTemplate{
		Name:     "Template to Delete",
		Category: models.ReportCategoryPortfolio,
	}
	service.CreateTemplate(template)

	section := &models.ReportSection{
		TemplateID: template.ID,
		Title:      "Section",
		Type:       "text",
	}
	service.CreateSection(section)

	param := &models.ReportParameter{
		TemplateID: template.ID,
		Name:       "param1",
		Label:      "Param",
		Type:       "string",
	}
	service.CreateParameter(param)

	err := service.DeleteTemplate(template.ID)
	if err != nil {
		t.Fatalf("Failed to delete template: %v", err)
	}

	_, err = service.GetTemplate(template.ID)
	if err != ErrTemplateNotFound {
		t.Errorf("Expected ErrTemplateNotFound after deletion, got %v", err)
	}

	var sectionCount int64
	db.Model(&models.ReportSection{}).Count(&sectionCount)
	if sectionCount != 0 {
		t.Errorf("Expected 0 sections after deletion, got %d", sectionCount)
	}

	var paramCount int64
	db.Model(&models.ReportParameter{}).Count(&paramCount)
	if paramCount != 0 {
		t.Errorf("Expected 0 parameters after deletion, got %d", paramCount)
	}
}

func TestReportService_GenerateReport(t *testing.T) {
	db := setupReportTestDB(t)
	defer cleanupReportTestDB(db)

	service := NewReportService(db)

	template := &models.ReportTemplate{
		Name:     "Test Template",
		Category: models.ReportCategoryPortfolio,
	}
	service.CreateTemplate(template)

	data := map[string]any{
		"portfolio_id": 1,
		"start_date":   "2024-01-01",
	}

	report, err := service.GenerateReport(template.ID, "Test Report", models.ReportFormatHTML, data)
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	if report.ID == 0 {
		t.Error("Report ID should be set after creation")
	}

	if report.Status != models.ReportStatusPending {
		t.Errorf("Expected status pending, got %s", report.Status)
	}
}

func TestReportService_GenerateReport_InvalidFormat(t *testing.T) {
	db := setupReportTestDB(t)
	defer cleanupReportTestDB(db)

	service := NewReportService(db)

	template := &models.ReportTemplate{
		Name:     "Test Template",
		Category: models.ReportCategoryPortfolio,
	}
	service.CreateTemplate(template)

	_, err := service.GenerateReport(template.ID, "Test Report", models.ReportFormat("invalid"), nil)
	if err != ErrInvalidFormat {
		t.Errorf("Expected ErrInvalidFormat, got %v", err)
	}
}

func TestReportService_GetReports_Pagination(t *testing.T) {
	db := setupReportTestDB(t)
	defer cleanupReportTestDB(db)

	service := NewReportService(db)

	template := &models.ReportTemplate{
		Name:     "Test Template",
		Category: models.ReportCategoryPortfolio,
	}
	service.CreateTemplate(template)

	for range 25 {
		report := &models.GeneratedReport{
			TemplateID: template.ID,
			Title:      "Report",
			Format:     models.ReportFormatHTML,
			Status:     models.ReportStatusCompleted,
		}
		service.CreateReport(report)
	}

	reports, total, err := service.GetReports(1, 10)
	if err != nil {
		t.Fatalf("Failed to get reports: %v", err)
	}

	if total != 25 {
		t.Errorf("Expected total 25, got %d", total)
	}

	if len(reports) != 10 {
		t.Errorf("Expected 10 reports on page 1, got %d", len(reports))
	}

	reports2, _, err := service.GetReports(3, 10)
	if err != nil {
		t.Fatalf("Failed to get reports page 3: %v", err)
	}

	if len(reports2) != 5 {
		t.Errorf("Expected 5 reports on page 3, got %d", len(reports2))
	}
}

func TestReportService_UpdateReportStatus(t *testing.T) {
	db := setupReportTestDB(t)
	defer cleanupReportTestDB(db)

	service := NewReportService(db)

	template := &models.ReportTemplate{
		Name:     "Test Template",
		Category: models.ReportCategoryPortfolio,
	}
	service.CreateTemplate(template)

	report := &models.GeneratedReport{
		TemplateID: template.ID,
		Title:      "Test Report",
		Format:     models.ReportFormatHTML,
		Status:     models.ReportStatusPending,
	}
	service.CreateReport(report)

	err := service.UpdateReportStatus(report.ID, models.ReportStatusCompleted, "")
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	updated, _ := service.GetReport(report.ID)
	if updated.Status != models.ReportStatusCompleted {
		t.Errorf("Expected status completed, got %s", updated.Status)
	}

	if updated.CompletedAt == nil {
		t.Error("CompletedAt should be set for completed status")
	}
}

func TestReportService_AsyncGenerateReport_PanicRecovery(t *testing.T) {
	db := setupReportTestDB(t)
	defer cleanupReportTestDB(db)

	service := NewReportService(db)

	template := &models.ReportTemplate{
		Name:     "Test Template",
		Category: models.ReportCategoryPortfolio,
	}
	service.CreateTemplate(template)

	report := &models.GeneratedReport{
		TemplateID: template.ID,
		Title:      "Panic Test",
		Format:     models.ReportFormatHTML,
		Status:     models.ReportStatusPending,
	}
	service.CreateReport(report)

	service.asyncGenerateReport(report.ID, template, models.ReportFormatHTML, nil)

	time.Sleep(100 * time.Millisecond)

	updated, err := service.GetReport(report.ID)
	if err != nil {
		t.Fatalf("Failed to get report: %v", err)
	}

	if updated.Status != models.ReportStatusCompleted {
		t.Errorf("Expected status completed, got %s", updated.Status)
	}
}

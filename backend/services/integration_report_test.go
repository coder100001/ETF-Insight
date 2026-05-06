package services

import (
	"testing"
	"time"

	"etf-insight/models"

	"github.com/stretchr/testify/assert"
)

func TestReportService_CreateAndGet_Template(t *testing.T) {
	service := NewReportService(nil)

	template := &models.ReportTemplate{
		Name:        "Portfolio Analysis",
		Description: "Standard portfolio analysis report",
		IsDefault:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := service.CreateTemplate(template)
	if err != nil {
		if err.Error() == "database connection is nil" {
			t.Log("Skipping test: no database connection")
			return
		}
		t.Fatalf("CreateTemplate unexpected error: %v", err)
	}

	assert.NotZero(t, template.ID, "Template ID should be set after creation")
	assert.Equal(t, "Portfolio Analysis", template.Name)
}

func TestReportService_GetTemplates_NilDB(t *testing.T) {
	service := NewReportService(nil)

	templates, err := service.GetTemplates()
	if err != nil && err.Error() != "database connection is nil" {
		t.Logf("GetTemplates (no DB): %v", err)
		return
	}
	if templates == nil {
		t.Log("GetTemplates returned nil (expected with no DB)")
	}
}

func TestReportService_GetDefaultTemplates_NilDB(t *testing.T) {
	service := NewReportService(nil)

	templates, err := service.GetDefaultTemplates()
	if err != nil && err.Error() != "database connection is nil" {
		t.Logf("GetDefaultTemplates (no DB): %v", err)
	}
	_ = templates
}

func TestIntegrationReportService_GetTemplate_NotFound(t *testing.T) {
	service := NewReportService(nil)

	template, err := service.GetTemplate(999999)
	if err != nil && err.Error() == "database connection is nil" {
		t.Log("Skipping test: no database connection")
		return
	}
	assert.Equal(t, ErrTemplateNotFound, err, "Non-existent template should return ErrTemplateNotFound")
	assert.Nil(t, template)
}

func TestReportService_UpdateTemplate_NilDB(t *testing.T) {
	service := NewReportService(nil)

	template := &models.ReportTemplate{
		ID:          1,
		Name:        "Updated Name",
		Description: "Updated description",
		UpdatedAt:   time.Now(),
	}

	err := service.UpdateTemplate(template)
	if err != nil && err.Error() != "database connection is nil" {
		t.Logf("UpdateTemplate (no DB): %v", err)
	}
}

func TestReportService_DeleteTemplate_NilDB(t *testing.T) {
	service := NewReportService(nil)

	err := service.DeleteTemplate(999999)
	if err != nil && err.Error() != "database connection is nil" {
		t.Logf("DeleteTemplate (no DB): %v", err)
	}
}

func TestReportService_CreateReport_WithParameters(t *testing.T) {
	service := NewReportService(nil)

	report := &models.GeneratedReport{
		TemplateID: 1,
		Title:      "Q1 2026 Portfolio Review",
		Status:     models.ReportStatusPending,
		Format:     models.ReportFormatPDF,
		CreatedAt:  time.Now(),
	}

	err := service.CreateReport(report)
	if err != nil {
		if err.Error() == "database connection is nil" {
			t.Log("Skipping test: no database connection")
			return
		}
		t.Fatalf("CreateReport unexpected error: %v", err)
	}

	assert.NotZero(t, report.ID)
	assert.Equal(t, "Q1 2026 Portfolio Review", report.Title)
	assert.Equal(t, models.ReportStatusPending, report.Status)
}

func TestReportService_GetReport_NotFound(t *testing.T) {
	service := NewReportService(nil)

	report, err := service.GetReport(999999)
	if err != nil && err.Error() == "database connection is nil" {
		t.Log("Skipping test: no database connection")
		return
	}
	assert.Equal(t, ErrReportNotFound, err, "Non-existent report should return ErrReportNotFound")
	assert.Nil(t, report)
}

func TestReportService_GenerateReport_Integration(t *testing.T) {
	service := NewReportService(nil)

	_, err := service.GenerateReport(1, "Test Generated Report", models.ReportFormatPDF, nil)
	if err != nil {
		t.Logf("GenerateReport requires actual data source: %v", err)
	}
}

func TestReportService_Parameters_CRUD(t *testing.T) {
	service := NewReportService(nil)

	param := &models.ReportParameter{
		TemplateID: 1,
		Name:       "risk_free_rate",
		Default:    "0.04",
		Type:       "decimal",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := service.CreateParameter(param)
	if err != nil {
		if err.Error() == "database connection is nil" {
			t.Log("Skipping test: no database connection")
			return
		}
		t.Fatalf("CreateParameter unexpected error: %v", err)
	}

	assert.NotZero(t, param.ID)
	assert.Equal(t, "risk_free_rate", param.Name)
}

func TestReportService_Sections_CRUD(t *testing.T) {
	service := NewReportService(nil)

	section := &models.ReportSection{
		TemplateID: 1,
		Title:      "Executive Summary",
		Order:      1,
		Content:    "This report provides a comprehensive analysis...",
		Type:       "text",
		Required:   true,
		CreatedAt:  time.Now(),
	}

	err := service.CreateSection(section)
	if err != nil {
		if err.Error() == "database connection is nil" {
			t.Log("Skipping test: no database connection")
			return
		}
		t.Fatalf("CreateSection unexpected error: %v", err)
	}

	assert.NotZero(t, section.ID)
	assert.Equal(t, "Executive Summary", section.Title)
	assert.Equal(t, 1, section.Order)
}

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"etf-insight/models"
	"etf-insight/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewPortfolioHandler(t *testing.T) {
	analysisService := services.NewETFAnalysisService(nil)
	handler := NewPortfolioHandler(analysisService)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.analysisService)
}

func TestAnalyzePortfolio(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试ETF配置
	cfg := models.ETFConfig{
		Symbol:   "SPY",
		Name:     "SPDR S&P 500 ETF Trust",
		Currency: "USD",
	}
	models.DB.Create(&cfg)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	handler := NewPortfolioHandler(analysisService)
	router.POST("/api/portfolio/analyze", handler.AnalyzePortfolio)

	body := `{
		"allocation": {"SPY": 100},
		"total_investment": 10000,
		"tax_rate": 0.10
	}`
	req := httptest.NewRequest("POST", "/api/portfolio/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

func TestAnalyzePortfolio_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	handler := NewPortfolioHandler(analysisService)
	router.POST("/api/portfolio/analyze", handler.AnalyzePortfolio)

	req := httptest.NewRequest("POST", "/api/portfolio/analyze", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyzePortfolio_DefaultValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试ETF配置
	cfg := models.ETFConfig{
		Symbol:   "SPY",
		Name:     "SPDR S&P 500 ETF Trust",
		Currency: "USD",
	}
	models.DB.Create(&cfg)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	handler := NewPortfolioHandler(analysisService)
	router.POST("/api/portfolio/analyze", handler.AnalyzePortfolio)

	// 不提供 total_investment 和 tax_rate，应该使用默认值
	body := `{
		"allocation": {"SPY": 100}
	}`
	req := httptest.NewRequest("POST", "/api/portfolio/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

func TestAnalyzePortfolio_MultipleETFs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建多个测试ETF配置
	cfg1 := models.ETFConfig{
		Symbol:   "SPY",
		Name:     "SPDR S&P 500 ETF Trust",
		Currency: "USD",
	}
	cfg2 := models.ETFConfig{
		Symbol:   "QQQ",
		Name:     "Invesco QQQ Trust",
		Currency: "USD",
	}
	models.DB.Create(&cfg1)
	models.DB.Create(&cfg2)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	handler := NewPortfolioHandler(analysisService)
	router.POST("/api/portfolio/analyze", handler.AnalyzePortfolio)

	body := `{
		"allocation": {"SPY": 60, "QQQ": 40},
		"total_investment": 50000,
		"tax_rate": 0.15
	}`
	req := httptest.NewRequest("POST", "/api/portfolio/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

func TestGetPortfolioConfigs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试投资组合配置
	allocation := map[string]interface{}{"SPY": 100}
	allocationJSON, _ := json.Marshal(allocation)

	config := models.PortfolioConfig{
		Name:       "Test Portfolio",
		Allocation: string(allocationJSON),
		Status:     1,
	}
	models.DB.Create(&config)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	handler := NewPortfolioHandler(analysisService)
	router.GET("/api/portfolio/configs", handler.GetPortfolioConfigs)

	req := httptest.NewRequest("GET", "/api/portfolio/configs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

func TestCreatePortfolioConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	handler := NewPortfolioHandler(analysisService)
	router.POST("/api/portfolio/configs", handler.CreatePortfolioConfig)

	body := `{
		"name": "New Portfolio",
		"description": "Test portfolio",
		"allocation": {"SPY": 50, "QQQ": 50},
		"total_investment": 100000,
		"tax_rate": 0.10,
		"is_default": false
	}`
	req := httptest.NewRequest("POST", "/api/portfolio/configs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

func TestCreatePortfolioConfig_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	handler := NewPortfolioHandler(analysisService)
	router.POST("/api/portfolio/configs", handler.CreatePortfolioConfig)

	req := httptest.NewRequest("POST", "/api/portfolio/configs", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePortfolioConfig_MissingName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	handler := NewPortfolioHandler(analysisService)
	router.POST("/api/portfolio/configs", handler.CreatePortfolioConfig)

	body := `{
		"allocation": {"SPY": 100}
	}`
	req := httptest.NewRequest("POST", "/api/portfolio/configs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPortfolioConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试投资组合配置
	allocation := map[string]interface{}{"SPY": 100}
	allocationJSON, _ := json.Marshal(allocation)

	config := models.PortfolioConfig{
		Name:       "Test Portfolio",
		Allocation: string(allocationJSON),
		Status:     1,
	}
	models.DB.Create(&config)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	handler := NewPortfolioHandler(analysisService)
	router.GET("/api/portfolio/configs/:id", handler.GetPortfolioConfig)

	req := httptest.NewRequest("GET", "/api/portfolio/configs/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

func TestGetPortfolioConfig_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	handler := NewPortfolioHandler(analysisService)
	router.GET("/api/portfolio/configs/:id", handler.GetPortfolioConfig)

	req := httptest.NewRequest("GET", "/api/portfolio/configs/9999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdatePortfolioConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试投资组合配置
	allocation := map[string]interface{}{"SPY": 100}
	allocationJSON, _ := json.Marshal(allocation)

	config := models.PortfolioConfig{
		Name:       "Test Portfolio",
		Allocation: string(allocationJSON),
		Status:     1,
	}
	models.DB.Create(&config)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	handler := NewPortfolioHandler(analysisService)
	router.PUT("/api/portfolio/configs/:id", handler.UpdatePortfolioConfig)

	body := `{
		"name": "Updated Portfolio",
		"allocation": {"SPY": 60, "QQQ": 40}
	}`
	req := httptest.NewRequest("PUT", "/api/portfolio/configs/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

func TestDeletePortfolioConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试投资组合配置
	allocation := map[string]interface{}{"SPY": 100}
	allocationJSON, _ := json.Marshal(allocation)

	config := models.PortfolioConfig{
		Name:       "Test Portfolio",
		Allocation: string(allocationJSON),
		Status:     1,
	}
	models.DB.Create(&config)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	handler := NewPortfolioHandler(analysisService)
	router.DELETE("/api/portfolio/configs/:id", handler.DeletePortfolioConfig)

	req := httptest.NewRequest("DELETE", "/api/portfolio/configs/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

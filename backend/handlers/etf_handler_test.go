package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"etf-insight/models"
	"etf-insight/services"
	"etf-insight/services/datasource"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// MockDataSourceProvider 模拟数据源提供者
type MockDataSourceProvider struct{}

func (m *MockDataSourceProvider) GetName() string {
	return "mock"
}

func (m *MockDataSourceProvider) GetQuote(ctx context.Context, symbol string) (*datasource.QuoteData, error) {
	return &datasource.QuoteData{
		Symbol:        symbol,
		CurrentPrice:  100.0,
		PreviousClose: 99.0,
		Change:        1.0,
		ChangePercent: 1.01,
		Timestamp:     time.Now(),
	}, nil
}

func (m *MockDataSourceProvider) GetQuotes(ctx context.Context, symbols []string) ([]*datasource.QuoteData, error) {
	var quotes []*datasource.QuoteData
	for _, symbol := range symbols {
		quotes = append(quotes, &datasource.QuoteData{
			Symbol:        symbol,
			CurrentPrice:  100.0,
			PreviousClose: 99.0,
			Change:        1.0,
			ChangePercent: 1.01,
			Timestamp:     time.Now(),
		})
	}
	return quotes, nil
}

func (m *MockDataSourceProvider) IsAvailable(ctx context.Context) bool {
	return true
}

func (m *MockDataSourceProvider) GetRateLimit() int {
	return 100
}

func (m *MockDataSourceProvider) GetETFHoldings(ctx context.Context, symbol string, date time.Time) ([]*datasource.ETFHoldingData, error) {
	return []*datasource.ETFHoldingData{}, nil
}

func setupTestDB(t *testing.T) {
	// 初始化内存数据库
	err := models.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
}

func TestGetETFList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.GET("/api/etf/list", handler.GetETFList)

	req := httptest.NewRequest("GET", "/api/etf/list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestGetETFList_WithPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.GET("/api/etf/list", handler.GetETFList)

	req := httptest.NewRequest("GET", "/api/etf/list?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestGetETFList_InvalidPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.GET("/api/etf/list", handler.GetETFList)

	req := httptest.NewRequest("GET", "/api/etf/list?page=-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 无效页面会使用默认值，返回200
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetETFList_InvalidPageSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.GET("/api/etf/list", handler.GetETFList)

	req := httptest.NewRequest("GET", "/api/etf/list?pageSize=1000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回400错误或自动调整为最大值
	assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest}, w.Code)
}

func TestGetETFList_WithSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.GET("/api/etf/list", handler.GetETFList)

	req := httptest.NewRequest("GET", "/api/etf/list?search=SPY", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetETFList_WithSort(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.GET("/api/etf/list", handler.GetETFList)

	req := httptest.NewRequest("GET", "/api/etf/list?sortBy=name&sortOrder=desc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewETFHandler(t *testing.T) {
	analysisService := services.NewETFAnalysisService(nil)
	provider := &MockDataSourceProvider{}

	handler := NewETFHandler(analysisService, provider)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.analysisService)
	assert.NotNil(t, handler.provider)
}

func TestGetETFRealtime(t *testing.T) {
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
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.GET("/api/etf/realtime/:symbol", handler.GetETFRealtime)

	req := httptest.NewRequest("GET", "/api/etf/realtime/SPY", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 可能返回200或404，取决于数据是否存在
	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
}

func TestGetETFRealtime_InvalidSymbol(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.GET("/api/etf/realtime/:symbol", handler.GetETFRealtime)

	req := httptest.NewRequest("GET", "/api/etf/realtime/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 空symbol应该返回404
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetETFComparison(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试ETF配置
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
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.POST("/api/etf/comparison", handler.GetETFComparison)

	body := `{"symbols": ["SPY", "QQQ"], "period": "1y"}`
	req := httptest.NewRequest("POST", "/api/etf/comparison", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestGetETFComparison_NoETFs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 不创建任何ETF配置
	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.GET("/api/etf/comparison", handler.GetETFComparison)

	req := httptest.NewRequest("GET", "/api/etf/comparison?period=1y", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 没有ETF时应该返回200，但数据为空
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

func TestGetETFHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试ETF配置和价格数据
	cfg := models.ETFConfig{
		Symbol:   "SPY",
		Name:     "SPDR S&P 500 ETF Trust",
		Currency: "USD",
	}
	models.DB.Create(&cfg)

	// 创建历史价格数据
	for i := 0; i < 30; i++ {
		price := models.ETFData{
			Symbol:     "SPY",
			Date:       time.Now().AddDate(0, 0, -i),
			ClosePrice: decimal.NewFromFloat(100.0 + float64(i)*0.5),
			OpenPrice:  decimal.NewFromFloat(99.0 + float64(i)*0.5),
			HighPrice:  decimal.NewFromFloat(101.0 + float64(i)*0.5),
			LowPrice:   decimal.NewFromFloat(98.0 + float64(i)*0.5),
			Volume:     1000000,
		}
		models.DB.Create(&price)
	}

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.GET("/api/etf/history/:symbol", handler.GetETFHistory)

	req := httptest.NewRequest("GET", "/api/etf/history/SPY?period=1mo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestGetETFHistory_DefaultPeriod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试ETF配置和价格数据
	cfg := models.ETFConfig{
		Symbol:   "SPY",
		Name:     "SPDR S&P 500 ETF Trust",
		Currency: "USD",
	}
	models.DB.Create(&cfg)

	// 创建历史价格数据
	for i := 0; i < 30; i++ {
		price := models.ETFData{
			Symbol:     "SPY",
			Date:       time.Now().AddDate(0, 0, -i),
			ClosePrice: decimal.NewFromFloat(100.0 + float64(i)*0.5),
			OpenPrice:  decimal.NewFromFloat(99.0 + float64(i)*0.5),
			HighPrice:  decimal.NewFromFloat(101.0 + float64(i)*0.5),
			LowPrice:   decimal.NewFromFloat(98.0 + float64(i)*0.5),
			Volume:     1000000,
		}
		models.DB.Create(&price)
	}

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.GET("/api/etf/history/:symbol", handler.GetETFHistory)

	// 不提供period参数，应该使用默认值
	req := httptest.NewRequest("GET", "/api/etf/history/SPY", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回200，使用默认period
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestGetETFMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试ETF配置和价格数据
	cfg := models.ETFConfig{
		Symbol:   "SPY",
		Name:     "SPDR S&P 500 ETF Trust",
		Currency: "USD",
	}
	models.DB.Create(&cfg)

	// 创建历史价格数据用于计算指标
	for i := 0; i < 60; i++ {
		price := models.ETFData{
			Symbol:     "SPY",
			Date:       time.Now().AddDate(0, 0, -i),
			ClosePrice: decimal.NewFromFloat(100.0 + float64(i%10)*2),
			Volume:     1000000,
		}
		models.DB.Create(&price)
	}

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.GET("/api/etf/metrics/:symbol", handler.GetETFMetrics)

	req := httptest.NewRequest("GET", "/api/etf/metrics/SPY?period=3mo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestGetETFForecast(t *testing.T) {
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
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.GET("/api/etf/forecast/:symbol", handler.GetETFForecast)

	req := httptest.NewRequest("GET", "/api/etf/forecast/SPY?initial_investment=100000&tax_rate=0.10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestGetETFForecast_InvalidInvestment(t *testing.T) {
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
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.GET("/api/etf/forecast/:symbol", handler.GetETFForecast)

	// 无效的投资金额应该使用默认值
	req := httptest.NewRequest("GET", "/api/etf/forecast/SPY?initial_investment=invalid&tax_rate=0.10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回200，因为会回退到默认值
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetETFForecast_MissingSymbol(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.GET("/api/etf/forecast/:symbol", handler.GetETFForecast)

	req := httptest.NewRequest("GET", "/api/etf/forecast/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 空symbol应该返回404
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateRealtimeData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试ETF配置
	cfg := models.ETFConfig{
		Symbol:   "SPY",
		Name:     "SPDR S&P 500 ETF Trust",
		Currency: "USD",
		Status:   1, // 启用状态
	}
	models.DB.Create(&cfg)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.POST("/api/etf/realtime/update", handler.UpdateRealtimeData)

	req := httptest.NewRequest("POST", "/api/etf/realtime/update", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestUpdateRealtimeData_NoEnabledETFs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试ETF配置，但状态为禁用
	cfg := models.ETFConfig{
		Symbol:   "SPY",
		Name:     "SPDR S&P 500 ETF Trust",
		Currency: "USD",
		Status:   0, // 禁用状态
	}
	models.DB.Create(&cfg)

	router := gin.New()
	analysisService := services.NewETFAnalysisService(nil)
	provider := &MockDataSourceProvider{}
	handler := NewETFHandler(analysisService, provider)
	router.POST("/api/etf/realtime/update", handler.UpdateRealtimeData)

	req := httptest.NewRequest("POST", "/api/etf/realtime/update", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

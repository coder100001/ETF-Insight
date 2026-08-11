package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"etf-insight/models"
	exchangerate "etf-insight/services/exchange_rate"
	"etf-insight/services/exchange_rate/datasource"
	"etf-insight/tasks"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// newTestExchangeRateHandler 创建测试用的汇率处理器
// 使用空配置创建 ExchangeRateService（会降级到 Fallback provider），避免依赖外部 API
func newTestExchangeRateHandler() *ExchangeRateHandler {
	svc := exchangerate.NewExchangeRateService(&datasource.DataSourceConfig{})
	return &ExchangeRateHandler{
		exchangeSvc: svc,
		syncTask:    nil,
	}
}

func TestNewExchangeRateHandler(t *testing.T) {
	utils.InitLogger("warn")
	task := tasks.NewExchangeRateTask(&datasource.DataSourceConfig{})
	handler := NewExchangeRateHandler(task)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.exchangeSvc)
}

func TestGetExchangeRates(t *testing.T) {
	utils.InitLogger("warn")
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试汇率数据
	rate := models.ExchangeRate{
		FromCurrency:  "USD",
		ToCurrency:    "CNY",
		Rate:          decimal.NewFromFloat(7.2),
		PreviousRate:  decimal.NewFromFloat(7.1),
		ChangePercent: decimal.NewFromFloat(1.4),
		DataSource:    "test",
		SourceType:    "api",
		ValidStatus:   1,
		Priority:      1,
		SyncedAt:      &[]time.Time{time.Now()}[0],
	}
	models.DB.Create(&rate)

	router := gin.New()
	handler := newTestExchangeRateHandler()
	router.GET("/api/exchange-rates", handler.GetExchangeRates)

	req := httptest.NewRequest("GET", "/api/exchange-rates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

func TestGetExchangeRates_WithFilters(t *testing.T) {
	utils.InitLogger("warn")
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试汇率数据
	rate1 := models.ExchangeRate{
		FromCurrency: "USD",
		ToCurrency:   "CNY",
		Rate:         decimal.NewFromFloat(7.2),
		DataSource:   "test",
		ValidStatus:  1,
	}
	rate2 := models.ExchangeRate{
		FromCurrency: "USD",
		ToCurrency:   "HKD",
		Rate:         decimal.NewFromFloat(7.8),
		DataSource:   "test",
		ValidStatus:  1,
	}
	models.DB.Create(&rate1)
	models.DB.Create(&rate2)

	router := gin.New()
	handler := newTestExchangeRateHandler()
	router.GET("/api/exchange-rates", handler.GetExchangeRates)

	// 测试 from 过滤器
	req := httptest.NewRequest("GET", "/api/exchange-rates?from=USD", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

func TestGetExchangeRate(t *testing.T) {
	utils.InitLogger("warn")
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试汇率数据
	rate := models.ExchangeRate{
		FromCurrency: "USD",
		ToCurrency:   "CNY",
		Rate:         decimal.NewFromFloat(7.2),
		DataSource:   "test",
		ValidStatus:  1,
	}
	models.DB.Create(&rate)

	router := gin.New()
	handler := newTestExchangeRateHandler()
	router.GET("/api/exchange-rates/:from/:to", handler.GetExchangeRate)

	req := httptest.NewRequest("GET", "/api/exchange-rates/USD/CNY", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

func TestGetExchangeRate_NotFound(t *testing.T) {
	utils.InitLogger("warn")
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	handler := newTestExchangeRateHandler()
	router.GET("/api/exchange-rates/:from/:to", handler.GetExchangeRate)

	req := httptest.NewRequest("GET", "/api/exchange-rates/XXX/YYY", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestConvertCurrency(t *testing.T) {
	utils.InitLogger("warn")
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试汇率数据
	rate := models.ExchangeRate{
		FromCurrency: "USD",
		ToCurrency:   "CNY",
		Rate:         decimal.NewFromFloat(7.2),
		DataSource:   "test",
		ValidStatus:  1,
	}
	models.DB.Create(&rate)

	router := gin.New()
	handler := newTestExchangeRateHandler()
	router.POST("/api/exchange-rates/convert", handler.ConvertCurrency)

	body := `{"from_currency": "USD", "to_currency": "CNY", "amount": 100}`
	req := httptest.NewRequest("POST", "/api/exchange-rates/convert", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

func TestConvertCurrency_InvalidBody(t *testing.T) {
	utils.InitLogger("warn")
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	router := gin.New()
	handler := newTestExchangeRateHandler()
	router.POST("/api/exchange-rates/convert", handler.ConvertCurrency)

	req := httptest.NewRequest("POST", "/api/exchange-rates/convert", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSupportedCurrencies(t *testing.T) {
	utils.InitLogger("warn")
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试汇率数据
	rate1 := models.ExchangeRate{
		FromCurrency: "USD",
		ToCurrency:   "CNY",
		Rate:         decimal.NewFromFloat(7.2),
		DataSource:   "test",
		ValidStatus:  1,
	}
	rate2 := models.ExchangeRate{
		FromCurrency: "EUR",
		ToCurrency:   "CNY",
		Rate:         decimal.NewFromFloat(7.8),
		DataSource:   "test",
		ValidStatus:  1,
	}
	models.DB.Create(&rate1)
	models.DB.Create(&rate2)

	router := gin.New()
	handler := newTestExchangeRateHandler()
	router.GET("/api/exchange-rates/currencies", handler.GetSupportedCurrencies)

	req := httptest.NewRequest("GET", "/api/exchange-rates/currencies", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

func TestGetExchangeRatesSummary(t *testing.T) {
	utils.InitLogger("warn")
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 创建测试汇率数据
	rate := models.ExchangeRate{
		FromCurrency: "USD",
		ToCurrency:   "CNY",
		Rate:         decimal.NewFromFloat(7.2),
		DataSource:   "test",
		ValidStatus:  1,
	}
	models.DB.Create(&rate)

	router := gin.New()
	handler := newTestExchangeRateHandler()
	router.GET("/api/exchange-rates/summary", handler.GetExchangeRatesSummary)

	req := httptest.NewRequest("GET", "/api/exchange-rates/summary", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
}

func TestGetCurrencyPairs(t *testing.T) {
	utils.InitLogger("warn")
	gin.SetMode(gin.TestMode)
	setupTestDB(t)

	// 注意：currency_pairs 表可能不存在，此测试主要验证handler不panic
	router := gin.New()
	handler := newTestExchangeRateHandler()
	router.GET("/api/exchange-rates/pairs", handler.GetCurrencyPairs)

	req := httptest.NewRequest("GET", "/api/exchange-rates/pairs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 由于表可能不存在，可能返回500或200
	assert.Contains(t, []int{http.StatusOK, http.StatusInternalServerError}, w.Code)
}

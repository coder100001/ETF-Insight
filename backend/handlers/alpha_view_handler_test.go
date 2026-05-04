package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"etf-insight/models"
	"etf-insight/services"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAlphaViewHandlerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(
		&models.AlphaView{},
		&models.AlphaViewPerformance{},
		&models.FactorTimingSignal{},
		&models.FactorData{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func setupAlphaViewRouter(db *gorm.DB) (*gin.Engine, *AlphaViewHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	factorService := services.NewFactorDataService(db)
	alphaViewService := services.NewAlphaViewService(db, factorService)
	handler := NewAlphaViewHandler(alphaViewService)

	return router, handler
}

func TestAlphaViewHandler_CreateView(t *testing.T) {
	db := setupAlphaViewHandlerTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	router, handler := setupAlphaViewRouter(db)

	router.POST("/api/alpha-views", handler.CreateView)

	view := models.AlphaView{
		PortfolioID:  1,
		AssetSymbol:  "SPY",
		ViewReturn:   decimal.NewFromFloat(0.05),
		Confidence:   decimal.NewFromFloat(75.0),
		ViewType:     models.ViewTypeAbsolute,
		ViewMethod:   models.ViewMethodFactorTiming,
		GeneratedAt:  time.Now(),
		ValidUntil:   time.Now().AddDate(0, 0, 30),
		Status:       models.ViewStatusActive,
		SourceFactor: "Mkt-RF",
	}
	body, _ := json.Marshal(view)

	req, _ := http.NewRequest("POST", "/api/alpha-views", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestAlphaViewHandler_CreateView_InvalidBody(t *testing.T) {
	db := setupAlphaViewHandlerTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	router, handler := setupAlphaViewRouter(db)

	router.POST("/api/alpha-views", handler.CreateView)

	req, _ := http.NewRequest("POST", "/api/alpha-views", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAlphaViewHandler_GetActiveViews(t *testing.T) {
	db := setupAlphaViewHandlerTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	router, handler := setupAlphaViewRouter(db)

	// Create test view
	view := &models.AlphaView{
		PortfolioID:  1,
		AssetSymbol:  "SPY",
		ViewReturn:   decimal.NewFromFloat(0.05),
		Confidence:   decimal.NewFromFloat(75.0),
		ViewType:     models.ViewTypeAbsolute,
		ViewMethod:   models.ViewMethodFactorTiming,
		GeneratedAt:  time.Now(),
		ValidUntil:   time.Now().AddDate(0, 0, 30),
		Status:       models.ViewStatusActive,
		SourceFactor: "Mkt-RF",
		CreatedAt:    time.Now(),
	}
	db.Create(view)

	router.GET("/api/alpha-views/active", handler.GetActiveViews)

	req, _ := http.NewRequest("GET", "/api/alpha-views/active?asset_symbol=SPY", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if !response["success"].(bool) {
		t.Error("Expected success to be true")
	}
}

func TestAlphaViewHandler_GetView(t *testing.T) {
	db := setupAlphaViewHandlerTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	router, handler := setupAlphaViewRouter(db)

	// Create test view
	view := &models.AlphaView{
		PortfolioID:  1,
		AssetSymbol:  "SPY",
		ViewReturn:   decimal.NewFromFloat(0.05),
		Confidence:   decimal.NewFromFloat(75.0),
		ViewType:     models.ViewTypeAbsolute,
		ViewMethod:   models.ViewMethodFactorTiming,
		GeneratedAt:  time.Now(),
		ValidUntil:   time.Now().AddDate(0, 0, 30),
		Status:       models.ViewStatusActive,
		SourceFactor: "Mkt-RF",
		CreatedAt:    time.Now(),
	}
	db.Create(view)

	router.GET("/api/alpha-views/:id", handler.GetView)

	req, _ := http.NewRequest("GET", "/api/alpha-views/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAlphaViewHandler_GetView_NotFound(t *testing.T) {
	db := setupAlphaViewHandlerTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	router, handler := setupAlphaViewRouter(db)

	router.GET("/api/alpha-views/:id", handler.GetView)

	req, _ := http.NewRequest("GET", "/api/alpha-views/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// The service returns an error for non-existent views, which the handler treats as 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestAlphaViewHandler_GetView_InvalidID(t *testing.T) {
	db := setupAlphaViewHandlerTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	router, handler := setupAlphaViewRouter(db)

	router.GET("/api/alpha-views/:id", handler.GetView)

	req, _ := http.NewRequest("GET", "/api/alpha-views/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAlphaViewHandler_UpdateView(t *testing.T) {
	db := setupAlphaViewHandlerTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	router, handler := setupAlphaViewRouter(db)

	// Create test view
	view := &models.AlphaView{
		PortfolioID:  1,
		AssetSymbol:  "SPY",
		ViewReturn:   decimal.NewFromFloat(0.05),
		Confidence:   decimal.NewFromFloat(75.0),
		ViewType:     models.ViewTypeAbsolute,
		ViewMethod:   models.ViewMethodFactorTiming,
		GeneratedAt:  time.Now(),
		ValidUntil:   time.Now().AddDate(0, 0, 30),
		Status:       models.ViewStatusActive,
		SourceFactor: "Mkt-RF",
		CreatedAt:    time.Now(),
	}
	db.Create(view)

	router.PUT("/api/alpha-views/:id", handler.UpdateView)

	updatedView := models.AlphaView{
		PortfolioID:  1,
		AssetSymbol:  "SPY",
		ViewReturn:   decimal.NewFromFloat(0.06),
		Confidence:   decimal.NewFromFloat(80.0),
		ViewType:     models.ViewTypeAbsolute,
		ViewMethod:   models.ViewMethodFactorTiming,
		GeneratedAt:  time.Now(),
		ValidUntil:   time.Now().AddDate(0, 0, 30),
		Status:       models.ViewStatusActive,
		SourceFactor: "Mkt-RF",
	}
	body, _ := json.Marshal(updatedView)

	req, _ := http.NewRequest("PUT", "/api/alpha-views/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAlphaViewHandler_DeactivateView(t *testing.T) {
	db := setupAlphaViewHandlerTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	router, handler := setupAlphaViewRouter(db)

	// Create test view
	view := &models.AlphaView{
		PortfolioID:  1,
		AssetSymbol:  "SPY",
		ViewReturn:   decimal.NewFromFloat(0.05),
		Confidence:   decimal.NewFromFloat(75.0),
		ViewType:     models.ViewTypeAbsolute,
		ViewMethod:   models.ViewMethodFactorTiming,
		GeneratedAt:  time.Now(),
		ValidUntil:   time.Now().AddDate(0, 0, 30),
		Status:       models.ViewStatusActive,
		SourceFactor: "Mkt-RF",
		CreatedAt:    time.Now(),
	}
	db.Create(view)

	router.DELETE("/api/alpha-views/:id", handler.DeactivateView)

	req, _ := http.NewRequest("DELETE", "/api/alpha-views/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if !response["success"].(bool) {
		t.Error("Expected success to be true")
	}

	if response["message"] != "view deactivated" {
		t.Errorf("Expected message 'view deactivated', got '%s'", response["message"])
	}
}

func TestAlphaViewHandler_DeactivateView_InvalidID(t *testing.T) {
	db := setupAlphaViewHandlerTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	router, handler := setupAlphaViewRouter(db)

	router.DELETE("/api/alpha-views/:id", handler.DeactivateView)

	req, _ := http.NewRequest("DELETE", "/api/alpha-views/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

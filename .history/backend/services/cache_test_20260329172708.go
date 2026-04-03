package services

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestNewCacheService(t *testing.T) {
	cs := NewCacheService()
	if cs == nil {
		t.Fatal("NewCacheService returned nil")
	}

	if cs.cache == nil {
		t.Error("Cache should not be nil")
	}
}

func TestCacheService_SetAndGet(t *testing.T) {
	cs := NewCacheService()

	key := "test_key"
	value := map[string]interface{}{"price": 100.0}

	cs.Set(key, value, 5*time.Minute)

	result, found := cs.Get(key)
	if !found {
		t.Error("Key should be found in cache")
	}

	if result == nil {
		t.Error("Value should not be nil")
	}
}

func TestCacheService_GetNotFound(t *testing.T) {
	cs := NewCacheService()

	_, found := cs.Get("nonexistent_key")
	if found {
		t.Error("Non-existent key should not be found")
	}
}

func TestCacheService_Delete(t *testing.T) {
	cs := NewCacheService()

	key := "test_key"
	value := "test_value"

	cs.Set(key, value, 5*time.Minute)
	cs.Delete(key)

	_, found := cs.Get(key)
	if found {
		t.Error("Key should be deleted")
	}
}

func TestCacheService_Clear(t *testing.T) {
	cs := NewCacheService()

	cs.Set("key1", "value1", 5*time.Minute)
	cs.Set("key2", "value2", 5*time.Minute)

	cs.Clear()

	_, found1 := cs.Get("key1")
	_, found2 := cs.Get("key2")

	if found1 || found2 {
		t.Error("Cache should be cleared")
	}
}

func TestCacheService_GetETFData(t *testing.T) {
	cs := NewCacheService()

	data := &ETFRealtimeData{
		Symbol:        "QQQ",
		Price:         decimal.NewFromFloat(350.0),
		Change:        decimal.NewFromFloat(5.0),
		ChangePercent: decimal.NewFromFloat(1.5),
	}

	cs.Set("etf:realtime:QQQ", data, 5*time.Minute)

	result, found := cs.GetETFData("QQQ")
	if !found {
		t.Error("ETF data should be found")
	}

	if result.Symbol != "QQQ" {
		t.Errorf("Expected symbol QQQ, got %s", result.Symbol)
	}
}

func TestCacheService_SetETFData(t *testing.T) {
	cs := NewCacheService()

	data := &ETFRealtimeData{
		Symbol: "SPY",
		Price:  decimal.NewFromFloat(450.0),
	}

	cs.SetETFData("SPY", data, 5*time.Minute)

	result, found := cs.GetETFData("SPY")
	if !found {
		t.Error("ETF data should be found after SetETFData")
	}

	if result.Price.Cmp(data.Price) != 0 {
		t.Errorf("Price mismatch")
	}
}

func TestNewETFAnalysisService(t *testing.T) {
	cs := NewCacheService()
	eas := NewETFAnalysisService(cs)

	if eas == nil {
		t.Fatal("NewETFAnalysisService returned nil")
	}

	if eas.cache == nil {
		t.Error("Cache should not be nil")
	}
}

func TestETFAnalysisService_CalculateMetrics(t *testing.T) {
	cs := NewCacheService()
	eas := NewETFAnalysisService(cs)

	holdings := []Holding{
		{
			Symbol:    "QQQ",
			Shares:    100,
			CostBasis: decimal.NewFromFloat(300.0),
		},
		{
			Symbol:    "SPY",
			Shares:    50,
			CostBasis: decimal.NewFromFloat(400.0),
		},
	}

	metrics := eas.CalculateMetrics(holdings)

	if metrics == nil {
		t.Fatal("CalculateMetrics returned nil")
	}

	if metrics.TotalShares != 150 {
		t.Errorf("Expected total shares 150, got %d", metrics.TotalShares)
	}
}

func TestETFAnalysisService_CalculateTax(t *testing.T) {
	cs := NewCacheService()
	eas := NewETFAnalysisService(cs)

	dividendYield := decimal.NewFromFloat(0.03)
	investment := decimal.NewFromFloat(10000)
	taxRate := decimal.NewFromFloat(0.15)

	tax := eas.CalculateTax(dividendYield, investment, taxRate)

	expected := dividendYield.Mul(investment).Mul(taxRate)
	if !tax.Equals(expected) {
		t.Errorf("Tax calculation mismatch")
	}
}

func TestCacheService_GetWithExpiration(t *testing.T) {
	cs := NewCacheService()

	key := "test_key"
	value := "test_value"

	cs.Set(key, value, 5*time.Minute)

	result, expiration, found := cs.GetWithExpiration(key)
	if !found {
		t.Error("Key should be found")
	}

	if result == nil {
		t.Error("Value should not be nil")
	}

	if expiration.IsZero() {
		t.Error("Expiration should not be zero")
	}
}

func TestCacheService_ItemCount(t *testing.T) {
	cs := NewCacheService()

	initialCount := cs.ItemCount()

	cs.Set("key1", "value1", 5*time.Minute)
	cs.Set("key2", "value2", 5*time.Minute)

	newCount := cs.ItemCount()
	if newCount != initialCount+2 {
		t.Errorf("Expected %d items, got %d", initialCount+2, newCount)
	}
}

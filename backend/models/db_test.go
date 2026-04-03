package models

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestInitDB(t *testing.T) {
	err := InitDB("host=localhost port=5432 user=postgres password=postgres dbname=etf_insight sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: PostgreSQL not available")
	}

	if DB == nil {
		t.Error("DB should not be nil after InitDB")
	}
}

func TestCreateETFConfig(t *testing.T) {
	err := InitDB("host=localhost port=5432 user=postgres password=postgres dbname=etf_insight sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: PostgreSQL not available")
	}

	etf := &ETFConfig{
		Symbol:       "TEST",
		Name:         "Test ETF",
		ExpenseRatio: decimal.NewFromFloat(0.001),
		Status:       1,
	}

	result := DB.Create(etf)
	if result.Error != nil {
		t.Errorf("Create failed: %v", result.Error)
	}

	// 清理
	DB.Where("symbol = ?", "TEST").Delete(&ETFConfig{})
}

func TestInitDefaultData(t *testing.T) {
	err := InitDB("host=localhost port=5432 user=postgres password=postgres dbname=etf_insight sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: PostgreSQL not available")
	}

	err = InitDefaultData()
	if err != nil {
		t.Errorf("InitDefaultData failed: %v", err)
	}
}

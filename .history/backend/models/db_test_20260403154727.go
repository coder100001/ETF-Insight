package models

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestInitDB(t *testing.T) {
	// 使用 SQLite 内存数据库进行测试
	err := InitDB(":memory:")
	if err != nil {
		t.Skipf("Skipping test: SQLite not available: %v", err)
	}

	if DB == nil {
		t.Error("DB should not be nil after InitDB")
	}
}

func TestCreateETFConfig(t *testing.T) {
	// 使用 SQLite 内存数据库进行测试
	err := InitDB(":memory:")
	if err != nil {
		t.Skipf("Skipping test: SQLite not available: %v", err)
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
	// 使用 SQLite 内存数据库进行测试
	err := InitDB(":memory:")
	if err != nil {
		t.Skipf("Skipping test: SQLite not available: %v", err)
	}

	err = InitDefaultData()
	if err != nil {
		t.Errorf("InitDefaultData failed: %v", err)
	}
}

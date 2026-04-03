package models

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestInitDB(t *testing.T) {
	err := InitDB()
	if err != nil {
		t.Errorf("InitDB failed: %v", err)
	}

	if DB == nil {
		t.Error("DB should not be nil after InitDB")
	}
}

func TestMockDB_Create(t *testing.T) {
	InitDB()

	etf := &ETFConfig{
		Symbol:       "TEST",
		Name:         "Test ETF",
		ExpenseRatio: decimal.NewFromFloat(0.001),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	DB.Create(etf)

	if DB.err != nil {
		t.Errorf("Create failed: %v", DB.err)
	}
}

func TestMockDB_CreateETFData(t *testing.T) {
	InitDB()

	data := &ETFData{
		Symbol:    "TEST",
		Price:     decimal.NewFromFloat(100.0),
		Timestamp: time.Now(),
	}

	DB.Create(data)

	if DB.err != nil {
		t.Errorf("Create ETFData failed: %v", DB.err)
	}
}

func TestMockDB_CreateOperationLog(t *testing.T) {
	InitDB()

	log := &OperationLog{
		Operation: "test",
		Message:   "test message",
		CreatedAt: time.Now(),
	}

	DB.Create(log)

	if DB.err != nil {
		t.Errorf("Create OperationLog failed: %v", DB.err)
	}

	if log.ID == 0 {
		t.Error("OperationLog ID should be set")
	}
}

func TestMockDB_ChainMethods(t *testing.T) {
	InitDB()

	result := DB.Where("symbol = ?", "TEST").
		First(&ETFConfig{}).
		Find(&[]ETFConfig{}).
		Order("symbol").
		Limit(10).
		Model(&ETFConfig{}).
		Updates(map[string]interface{}{"status": 1}).
		Save(&ETFConfig{})

	if result == nil {
		t.Error("Chain methods should return DB")
	}
}

func TestInitDefaultData(t *testing.T) {
	InitDB()

	err := InitDefaultData()
	if err != nil {
		t.Errorf("InitDefaultData failed: %v", err)
	}

	if len(DB.etfConfigs) == 0 {
		t.Error("Default ETFs should be created")
	}
}

func TestAutoMigrate(t *testing.T) {
	err := AutoMigrate()
	if err != nil {
		t.Errorf("AutoMigrate failed: %v", err)
	}
}

package models

import (
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// DB 模拟数据库
type MockDB struct {
	mu         sync.RWMutex
	etfConfigs map[string]ETFConfig
	etfData    []ETFData
	opLogs     []OperationLog
	err        error
}

var DB *MockDB

// InitDB 初始化数据库
func InitDB() error {
	DB = &MockDB{
		etfConfigs: make(map[string]ETFConfig),
		etfData:    make([]ETFData, 0),
		opLogs:     make([]OperationLog, 0),
	}
	return nil
}

// AutoMigrate 自动迁移
func AutoMigrate() error {
	return nil
}

// Create 创建记录
func (db *MockDB) Create(value interface{}) *MockDB {
	db.mu.Lock()
	defer db.mu.Unlock()

	switch v := value.(type) {
	case *ETFConfig:
		db.etfConfigs[v.Symbol] = *v
	case *ETFData:
		db.etfData = append(db.etfData, *v)
	case *OperationLog:
		v.ID = uint(len(db.opLogs) + 1)
		db.opLogs = append(db.opLogs, *v)
	case *ExchangeRate:
		// 简化处理，不存储汇率
	}
	return db
}

// Error 返回错误
func (db *MockDB) Error() error {
	return db.err
}

// Where 查询条件
func (db *MockDB) Where(query string, args ...interface{}) *MockDB {
	return db
}

// First 获取第一条记录
func (db *MockDB) First(dest interface{}, conds ...interface{}) *MockDB {
	db.mu.RLock()
	defer db.mu.RUnlock()

	switch v := dest.(type) {
	case *ETFConfig:
		if len(conds) > 0 {
			// 简化处理，根据ID查找（支持int和uint）
			var id uint
			switch val := conds[0].(type) {
			case uint:
				id = val
			case int:
				id = uint(val)
			default:
				// 默认返回第一个
				for _, config := range db.etfConfigs {
					*v = config
					return db
				}
				return db
			}
			for _, config := range db.etfConfigs {
				if config.ID == id {
					*v = config
					return db
				}
			}
		}
		// 默认返回第一个
		for _, config := range db.etfConfigs {
			*v = config
			return db
		}
	}
	return db
}

// Find 查询多条记录
func (db *MockDB) Find(dest interface{}, conds ...interface{}) *MockDB {
	db.mu.RLock()
	defer db.mu.RUnlock()

	switch v := dest.(type) {
	case *[]ETFConfig:
		for _, config := range db.etfConfigs {
			*v = append(*v, config)
		}
	}
	return db
}

// Order 排序
func (db *MockDB) Order(value string) *MockDB {
	return db
}

// Limit 限制数量
func (db *MockDB) Limit(limit int) *MockDB {
	return db
}

// Model 指定模型
func (db *MockDB) Model(value interface{}) *MockDB {
	return db
}

// Updates 更新记录
func (db *MockDB) Updates(values interface{}) *MockDB {
	return db
}

// Save 保存记录
func (db *MockDB) Save(value interface{}) *MockDB {
	db.mu.Lock()
	defer db.mu.Unlock()

	switch v := value.(type) {
	case *ETFConfig:
		db.etfConfigs[v.Symbol] = *v
	}
	return db
}

// Delete 删除记录
func (db *MockDB) Delete(value interface{}, conds ...interface{}) *MockDB {
	db.mu.Lock()
	defer db.mu.Unlock()

	// 简化处理，实际应该根据 conds 删除
	return db
}

// InitDefaultData 初始化默认数据
func InitDefaultData() error {
	defaultETFs := []ETFConfig{
		{
			Symbol:       "QQQ",
			Name:         "Invesco QQQ Trust",
			Description:  "追踪纳斯达克100指数",
			Strategy:     "大盘成长",
			Focus:        "科技",
			ExpenseRatio: decimal.NewFromFloat(0.0020),
			Currency:     "USD",
			Exchange:     "NASDAQ",
			Category:     "大盘股",
			Provider:     "Invesco",
			Status:       1,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Symbol:       "SCHD",
			Name:         "Schwab US Dividend Equity ETF",
			Description:  "美国股息股票ETF",
			Strategy:     "股息价值",
			Focus:        "高股息",
			ExpenseRatio: decimal.NewFromFloat(0.0006),
			Currency:     "USD",
			Exchange:     "NYSE",
			Category:     "股息",
			Provider:     "Charles Schwab",
			Status:       1,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, etf := range defaultETFs {
		DB.Create(&etf)
	}

	return nil
}

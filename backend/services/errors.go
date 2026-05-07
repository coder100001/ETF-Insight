package services

import "errors"

// ErrDatabaseNotInitialized 数据库连接未初始化
// 统一在 services 包中定义，供所有服务使用
var ErrDatabaseNotInitialized = errors.New("database connection is nil")

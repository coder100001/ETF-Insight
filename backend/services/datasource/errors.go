package datasource

import "errors"

// 数据源相关错误
var (
	// ErrNoAvailableProvider 没有可用的数据源
	ErrNoAvailableProvider = errors.New("no available data provider")

	// ErrRateLimitExceeded 超出速率限制
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrInvalidSymbol 无效的股票代码
	ErrInvalidSymbol = errors.New("invalid stock symbol")

	// ErrAPINotAvailable API服务不可用
	ErrAPINotAvailable = errors.New("API service not available")

	// ErrInvalidResponse 无效的API响应
	ErrInvalidResponse = errors.New("invalid API response")

	// ErrTimeout 请求超时
	ErrTimeout = errors.New("request timeout")

	// ErrNetwork 网络错误
	ErrNetwork = errors.New("network error")
)

// DataSourceError 数据源错误
type DataSourceError struct {
	Provider string
	Op       string
	Err      error
	Symbol   string
	Status   int
}

func (e *DataSourceError) Error() string {
	if e.Symbol != "" {
		return e.Provider + ":" + e.Op + " [" + e.Symbol + "]: " + e.Err.Error()
	}
	return e.Provider + ":" + e.Op + ": " + e.Err.Error()
}

func (e *DataSourceError) Unwrap() error {
	return e.Err
}

// IsDataSourceError 检查是否为数据源错误
func IsDataSourceError(err error) bool {
	var dsErr *DataSourceError
	return errors.As(err, &dsErr)
}

package config

import (
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	ETF      ETFConfig      `yaml:"etf"`
	Schedule ScheduleConfig `yaml:"schedule"`
	Log      LogConfig      `yaml:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `yaml:"driver"` // mysql, sqlite
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
	DSN      string `yaml:"dsn"` // 直接指定DSN
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// ETFConfig ETF相关配置
type ETFConfig struct {
	DefaultSymbols []string          `yaml:"default_symbols"`
	DataFetch      DataFetchConfig   `yaml:"data_fetch"`
	Cache          CacheConfig       `yaml:"cache"`
}

// DataFetchConfig 数据获取配置
type DataFetchConfig struct {
	RetryTimes       int           `yaml:"retry_times"`
	RetryDelay       time.Duration `yaml:"retry_delay"`
	RequestTimeout   time.Duration `yaml:"request_timeout"`
	RateLimitDelay   time.Duration `yaml:"rate_limit_delay"`
	MaxWorkers       int           `yaml:"max_workers"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	RealtimeTTL    time.Duration `yaml:"realtime_ttl"`
	HistoricalTTL  time.Duration `yaml:"historical_ttl"`
	MetricsTTL     time.Duration `yaml:"metrics_ttl"`
	ComparisonTTL  time.Duration `yaml:"comparison_ttl"`
}

// ScheduleConfig 定时任务配置
type ScheduleConfig struct {
	DailyUpdateTime    string `yaml:"daily_update_time"`
	MarketCloseUpdate  string `yaml:"market_close_update"`
	Timezone           string `yaml:"timezone"`
	ExchangeRateTime   string `yaml:"exchange_rate_time"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "sqlite"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 3306),
			Username: getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_NAME", "etf_insight"),
			Charset:  "utf8mb4",
			DSN:      getEnv("DB_DSN", "etf_insight.db"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			PoolSize: 50,
		},
		ETF: ETFConfig{
			DefaultSymbols: []string{"SCHD", "SPYD", "JEPQ", "JEPI", "VYM"},
			DataFetch: DataFetchConfig{
				RetryTimes:     3,
				RetryDelay:     2 * time.Second,
				RequestTimeout: 30 * time.Second,
				RateLimitDelay: 1 * time.Second,
				MaxWorkers:     3,
			},
			Cache: CacheConfig{
				RealtimeTTL:   5 * time.Minute,
				HistoricalTTL: 24 * time.Hour,
				MetricsTTL:    1 * time.Hour,
				ComparisonTTL: 30 * time.Minute,
			},
		},
		Schedule: ScheduleConfig{
			DailyUpdateTime:   "09:30",
			MarketCloseUpdate: "16:30",
			Timezone:          "America/New_York",
			ExchangeRateTime:  "10:30",
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: "json",
		},
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Warnf("Config file not found: %s, using default config", path)
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// GetDSN 获取数据库DSN
func (c *DatabaseConfig) GetDSN() string {
	if c.DSN != "" {
		return c.DSN
	}

	if c.Driver == "mysql" {
		return c.Username + ":" + c.Password + "@tcp(" + c.Host + ":" + 
			strconv.Itoa(c.Port) + ")/" + c.Database + "?charset=" + c.Charset + 
			"&parseTime=True&loc=Local"
	}

	return c.Database + ".db"
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetRedisAddr() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

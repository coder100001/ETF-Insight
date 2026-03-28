package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// JSON 自定义JSON类型
type JSON map[string]interface{}

// Value 实现driver.Valuer接口
func (j JSON) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return errors.New("cannot scan type into JSON")
	}
}

// ==================== 工作流相关模型 ====================

// Workflow 工作流定义
type Workflow struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"size:100;not null" json:"name"`
	Description   string         `gorm:"type:text" json:"description"`
	Category      string         `gorm:"size:50" json:"category"`
	Status        int            `gorm:"default:1" json:"status"` // 0-禁用, 1-启用, 2-归档
	TriggerType   int            `json:"trigger_type"`            // 1-定时, 2-手动, 3-事件
	TriggerConfig datatypes.JSON `gorm:"type:json" json:"trigger_config"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	
	Steps     []WorkflowStep      `gorm:"foreignKey:WorkflowID" json:"steps,omitempty"`
	Instances []WorkflowInstance  `gorm:"foreignKey:WorkflowID" json:"instances,omitempty"`
}

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	WorkflowID    uint           `gorm:"not null;index" json:"workflow_id"`
	Name          string         `gorm:"size:100;not null" json:"name"`
	StepType      string         `gorm:"size:50" json:"step_type"`
	OrderIndex    int            `json:"order_index"`
	HandlerType   int            `json:"handler_type"` // 1-脚本, 2-函数, 3-API调用
	HandlerConfig datatypes.JSON `gorm:"type:json" json:"handler_config"`
	RetryTimes    int            `gorm:"default:3" json:"retry_times"`
	RetryInterval int            `gorm:"default:5" json:"retry_interval"` // 秒
	Timeout       int            `gorm:"default:300" json:"timeout"`        // 秒
	IsCritical    bool           `gorm:"default:false" json:"is_critical"`
	OnFailure     string         `gorm:"size:50" json:"on_failure"`
	DependsOn     datatypes.JSON `gorm:"type:json" json:"depends_on"`
	ExtraConfig   datatypes.JSON `gorm:"type:json" json:"extra_config"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// WorkflowInstance 工作流实例
type WorkflowInstance struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	WorkflowID    uint           `gorm:"not null;index" json:"workflow_id"`
	TriggerType   int            `json:"trigger_type"`
	TriggerBy     string         `gorm:"size:100" json:"trigger_by"`
	Status        int            `gorm:"default:0" json:"status"` // 0-等待, 1-运行, 2-成功, 3-失败, 4-取消
	StartTime     *time.Time     `json:"start_time"`
	EndTime       *time.Time     `json:"end_time"`
	Duration      int            `json:"duration"` // 秒
	ContextData   datatypes.JSON `gorm:"type:json" json:"context_data"`
	ErrorMessage  string         `gorm:"type:text" json:"error_message"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	
	Workflow     Workflow              `gorm:"foreignKey:WorkflowID" json:"workflow,omitempty"`
	StepInstances []WorkflowInstanceStep `gorm:"foreignKey:WorkflowInstanceID" json:"step_instances,omitempty"`
	Logs         []SystemLog           `gorm:"foreignKey:WorkflowInstanceID" json:"logs,omitempty"`
}

// WorkflowInstanceStep 工作流实例步骤
type WorkflowInstanceStep struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	WorkflowInstanceID uint          `gorm:"not null;index" json:"workflow_instance_id"`
	WorkflowStepID    uint           `json:"workflow_step_id"`
	StepName          string         `gorm:"size:100" json:"step_name"`
	Status            int            `gorm:"default:0" json:"status"` // 0-等待, 1-运行, 2-成功, 3-失败, 4-跳过
	RetryCount        int            `gorm:"default:0" json:"retry_count"`
	AssignedTo        *int64         `json:"assigned_to"`
	StartTime         *time.Time     `json:"start_time"`
	EndTime           *time.Time     `json:"end_time"`
	Duration          int            `json:"duration"`
	InputData         datatypes.JSON `gorm:"type:json" json:"input_data"`
	OutputData        datatypes.JSON `gorm:"type:json" json:"output_data"`
	ErrorMessage      string         `gorm:"type:text" json:"error_message"`
	Logs              string         `gorm:"type:text" json:"logs"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// ==================== ETF相关模型 ====================

// ETFConfig ETF配置
type ETFConfig struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	Symbol       string          `gorm:"size:20;uniqueIndex;not null" json:"symbol"`
	Name         string          `gorm:"size:200" json:"name"`
	Market       string          `gorm:"size:10;default:'US'" json:"market"` // US, CN, HK
	Strategy     string          `gorm:"size:100" json:"strategy"`
	Description  string          `gorm:"type:text" json:"description"`
	Focus        string          `gorm:"size:50" json:"focus"`
	ExpenseRatio decimal.Decimal `gorm:"type:decimal(5,4)" json:"expense_ratio"`
	Status       int             `gorm:"default:1" json:"status"` // 0-禁用, 1-启用
	SortOrder    int             `gorm:"default:0" json:"sort_order"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// ETFData ETF历史数据
type ETFData struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	Symbol       string          `gorm:"size:20;not null;index" json:"symbol"`
	Date         time.Time       `gorm:"type:date;not null;index" json:"date"`
	OpenPrice    decimal.Decimal `gorm:"type:decimal(10,4)" json:"open_price"`
	ClosePrice   decimal.Decimal `gorm:"type:decimal(10,4)" json:"close_price"`
	HighPrice    decimal.Decimal `gorm:"type:decimal(10,4)" json:"high_price"`
	LowPrice     decimal.Decimal `gorm:"type:decimal(10,4)" json:"low_price"`
	Volume       int64           `json:"volume"`
	Dividend     decimal.Decimal `gorm:"type:decimal(10,4)" json:"dividend"`
	DataSource   string          `gorm:"size:50" json:"data_source"`
	FetchInstanceID *uint        `json:"fetch_instance_id"`
	CreatedAt    time.Time       `json:"created_at"`
	
	FetchInstance *WorkflowInstance `gorm:"foreignKey:FetchInstanceID" json:"fetch_instance,omitempty"`
}

// ETFBaseInfo ETF基础信息
type ETFBaseInfo struct {
	ID                 uint            `gorm:"primaryKey" json:"id"`
	Symbol             string          `gorm:"size:20;uniqueIndex;not null" json:"symbol"`
	Name               string          `gorm:"size:200" json:"name"`
	NameEn             string          `gorm:"size:200" json:"name_en"`
	Market             string          `gorm:"size:10" json:"market"`
	AssetClass         string          `gorm:"size:20;default:'EQUITY'" json:"asset_class"`
	Category           string          `gorm:"size:50" json:"category"`
	Issuer             string          `gorm:"size:100" json:"issuer"`
	IssuerWebsite      string          `gorm:"size:500" json:"issuer_website"`
	TrackingIndex      string          `gorm:"size:200" json:"tracking_index"`
	TrackingIndexSymbol string         `gorm:"size:50" json:"tracking_index_symbol"`
	TrackingMethod     string          `gorm:"size:50" json:"tracking_method"`
	InceptionDate      *time.Time      `json:"inception_date"`
	AUM                decimal.Decimal `gorm:"type:decimal(20,2)" json:"aum"`
	AUMCurrency        string          `gorm:"size:10;default:'USD'" json:"aum_currency"`
	SharesOutstanding  int64           `json:"shares_outstanding"`
	ExpenseRatio       decimal.Decimal `gorm:"type:decimal(6,4)" json:"expense_ratio"`
	ManagementFee      decimal.Decimal `gorm:"type:decimal(6,4)" json:"management_fee"`
	OtherExpenses      decimal.Decimal `gorm:"type:decimal(6,4)" json:"other_expenses"`
	ListingExchange    string          `gorm:"size:50" json:"listing_exchange"`
	TradingCurrency    string          `gorm:"size:10;default:'USD'" json:"trading_currency"`
	LotSize            int             `gorm:"default:1" json:"lot_size"`
	IsLeveraged        bool            `gorm:"default:false" json:"is_leveraged"`
	LeverageRatio      decimal.Decimal `gorm:"type:decimal(4,2)" json:"leverage_ratio"`
	IsInverse          bool            `gorm:"default:false" json:"is_inverse"`
	InvestmentStrategy string          `gorm:"type:text" json:"investment_strategy"`
	InvestmentObjective string         `gorm:"type:text" json:"investment_objective"`
	Benchmark          string          `gorm:"size:200" json:"benchmark"`
	Status             int             `gorm:"default:1" json:"status"`
	SortOrder          int             `gorm:"default:0" json:"sort_order"`
	DataSource         string          `gorm:"size:50" json:"data_source"`
	LastUpdated        time.Time       `json:"last_updated"`
	CreatedAt          time.Time       `json:"created_at"`
}

// ETFPrice ETF价格数据
type ETFPrice struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	ETFID        uint            `gorm:"not null;index" json:"etf_id"`
	Symbol       string          `gorm:"size:20;not null" json:"symbol"`
	TradeDate    time.Time       `gorm:"type:date;not null" json:"trade_date"`
	TradeTime    *time.Time      `json:"trade_time"`
	Interval     string          `gorm:"size:10;default:'1d'" json:"interval"`
	OpenPrice    decimal.Decimal `gorm:"type:decimal(12,4);not null" json:"open_price"`
	HighPrice    decimal.Decimal `gorm:"type:decimal(12,4);not null" json:"high_price"`
	LowPrice     decimal.Decimal `gorm:"type:decimal(12,4);not null" json:"low_price"`
	ClosePrice   decimal.Decimal `gorm:"type:decimal(12,4);not null" json:"close_price"`
	PreClose     decimal.Decimal `gorm:"type:decimal(12,4)" json:"pre_close"`
	Volume       int64           `gorm:"not null" json:"volume"`
	Turnover     decimal.Decimal `gorm:"type:decimal(20,4)" json:"turnover"`
	ChangeAmount decimal.Decimal `gorm:"type:decimal(12,4)" json:"change_amount"`
	ChangePercent decimal.Decimal `gorm:"type:decimal(8,4)" json:"change_percent"`
	TurnoverRate decimal.Decimal `gorm:"type:decimal(8,4)" json:"turnover_rate"`
	BidPrice     decimal.Decimal `gorm:"type:decimal(12,4)" json:"bid_price"`
	AskPrice     decimal.Decimal `gorm:"type:decimal(12,4)" json:"ask_price"`
	BidVolume    int64           `json:"bid_volume"`
	AskVolume    int64           `json:"ask_volume"`
	DataSource   string          `gorm:"size:50" json:"data_source"`
	IsAdjusted   bool            `gorm:"default:false" json:"is_adjusted"`
	CreatedAt    time.Time       `json:"created_at"`
	
	ETF ETFBaseInfo `gorm:"foreignKey:ETFID" json:"etf,omitempty"`
}

// ETFDividend ETF分红数据
type ETFDividend struct {
	ID               uint            `gorm:"primaryKey" json:"id"`
	ETFID            uint            `gorm:"not null;index" json:"etf_id"`
	Symbol           string          `gorm:"size:20;not null" json:"symbol"`
	DividendType     string          `gorm:"size:20;default:'CASH'" json:"dividend_type"`
	DividendAmount   decimal.Decimal `gorm:"type:decimal(12,6);not null" json:"dividend_amount"`
	DividendCurrency string          `gorm:"size:10;default:'USD'" json:"dividend_currency"`
	ExDividendDate   time.Time       `gorm:"type:date;not null" json:"ex_dividend_date"`
	RecordDate       *time.Time      `json:"record_date"`
	PaymentDate      *time.Time      `json:"payment_date"`
	Frequency        string          `gorm:"size:20" json:"frequency"`
	DividendYield    decimal.Decimal `gorm:"type:decimal(8,4)" json:"dividend_yield"`
	AnnualizedYield  decimal.Decimal `gorm:"type:decimal(8,4)" json:"annualized_yield"`
	DataSource       string          `gorm:"size:50" json:"data_source"`
	CreatedAt        time.Time       `json:"created_at"`
}

// ==================== 投资组合相关模型 ====================

// PortfolioConfig 投资组合配置
type PortfolioConfig struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	Name           string          `gorm:"size:100;not null" json:"name"`
	Description    string          `gorm:"type:text" json:"description"`
	Allocation     datatypes.JSON  `gorm:"type:json;not null" json:"allocation"`
	TotalInvestment decimal.Decimal `gorm:"type:decimal(15,2)" json:"total_investment"`
	Status         int             `gorm:"default:1" json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// ==================== 汇率相关模型 ====================

// ExchangeRate 汇率表
type ExchangeRate struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	FromCurrency string          `gorm:"size:10;not null" json:"from_currency"`
	ToCurrency   string          `gorm:"size:10;not null" json:"to_currency"`
	Rate         decimal.Decimal `gorm:"type:decimal(15,6);not null" json:"rate"`
	RateDate     time.Time       `gorm:"type:date;not null" json:"rate_date"`
	DataSource   string          `gorm:"size:50" json:"data_source"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// ==================== 日志相关模型 ====================

// OperationLog 操作记录
type OperationLog struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	WorkflowInstanceID *uint       `json:"workflow_instance_id"`
	OperationType   string         `gorm:"size:50;not null" json:"operation_type"`
	OperationName   string         `gorm:"size:200;not null" json:"operation_name"`
	Operator        string         `gorm:"size:100" json:"operator"`
	Status          int            `gorm:"default:0" json:"status"` // 0-进行中, 1-成功, 2-失败
	StartTime       time.Time      `json:"start_time"`
	EndTime         *time.Time     `json:"end_time"`
	DurationMs      int            `json:"duration_ms"`
	InputParams     datatypes.JSON `gorm:"type:json" json:"input_params"`
	OutputResult    datatypes.JSON `gorm:"type:json" json:"output_result"`
	ErrorMessage    string         `gorm:"type:text" json:"error_message"`
	IPAddress       string         `gorm:"size:50" json:"ip_address"`
	UserAgent       string         `gorm:"size:500" json:"user_agent"`
	ExtraData       datatypes.JSON `gorm:"type:json" json:"extra_data"`
}

// SystemLog 系统日志
type SystemLog struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	WorkflowInstanceID *uint         `gorm:"index" json:"workflow_instance_id"`
	LogLevel          string         `gorm:"size:20" json:"log_level"`
	Module            string         `gorm:"size:100" json:"module"`
	Message           string         `gorm:"type:text" json:"message"`
	StackTrace        string         `gorm:"type:text" json:"stack_trace"`
	ExtraData         datatypes.JSON `gorm:"type:json" json:"extra_data"`
	CreatedAt         time.Time      `json:"created_at"`
}

// Notification 通知记录
type Notification struct {
	ID                     uint           `gorm:"primaryKey" json:"id"`
	WorkflowInstanceStepID *uint          `json:"workflow_instance_step_id"`
	WorkflowInstanceID     *uint          `json:"workflow_instance_id"`
	NotificationType       int            `json:"notification_type"` // 1-邮件, 2-短信, 3-APP推送, 4-Webhook
	Recipient              string         `gorm:"size:200" json:"recipient"`
	Title                  string         `gorm:"size:200" json:"title"`
	Content                string         `gorm:"type:text" json:"content"`
	Status                 int            `gorm:"default:0" json:"status"` // 0-待发送, 1-已发送, 2-发送失败
	ServerID               string         `gorm:"size:100" json:"server_id"`
	SendAt                 *time.Time     `json:"send_at"`
	RetryCount             int            `gorm:"default:0" json:"retry_count"`
	ErrorMessage           string         `gorm:"type:text" json:"error_message"`
	CreatedAt              time.Time      `json:"created_at"`
}

// AnalysisReport 分析报告
type AnalysisReport struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	WorkflowInstanceID *uint        `json:"workflow_instance_id"`
	PortfolioConfigID *uint         `json:"portfolio_config_id"`
	ReportType       string         `gorm:"size:50" json:"report_type"`
	ReportDate       time.Time      `gorm:"type:date" json:"report_date"`
	FilePath         string         `gorm:"size:500" json:"file_path"`
	Metrics          datatypes.JSON `gorm:"type:json" json:"metrics"`
	Status           int            `gorm:"default:1" json:"status"`
	CreatedAt        time.Time      `json:"created_at"`
}

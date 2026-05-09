package models

import (
	"time"
)

// ReportCategory 报告分类
type ReportCategory string

const (
	ReportCategoryPortfolio     ReportCategory = "portfolio"      // 投资组合分析报告
	ReportCategoryRisk          ReportCategory = "risk"           // 风险分析报告
	ReportCategoryETFComparison ReportCategory = "etf_comparison" // ETF对比报告
	ReportCategoryMarket        ReportCategory = "market"         // 市场分析报告
	ReportCategoryCustom        ReportCategory = "custom"         // 自定义报告
)

// ReportStatus 报告生成状态
type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "pending"    // 待生成
	ReportStatusGenerating ReportStatus = "generating" // 生成中
	ReportStatusCompleted  ReportStatus = "completed"  // 已完成
	ReportStatusFailed     ReportStatus = "failed"     // 失败
)

// ReportFormat 报告导出格式
type ReportFormat string

const (
	ReportFormatHTML  ReportFormat = "html"  // HTML格式
	ReportFormatPDF   ReportFormat = "pdf"   // PDF格式
	ReportFormatExcel ReportFormat = "excel" // Excel格式
)

// ReportTemplate 报告模板
type ReportTemplate struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:255;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Category    ReportCategory `json:"category" gorm:"size:50;index"`
	IsDefault   bool           `json:"is_default" gorm:"default:false"`

	// 模板配置
	Config JSONMap `json:"config" gorm:"type:json"` // 模板配置JSON
	// 章节定义
	Sections []ReportSection `json:"sections" gorm:"foreignKey:TemplateID"`
	// 参数定义
	Parameters []ReportParameter `json:"parameters" gorm:"foreignKey:TemplateID"`

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ReportTemplate) TableName() string {
	return "report_templates"
}

// ReportSection 报告章节
type ReportSection struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	TemplateID uint   `json:"template_id" gorm:"index;not null"`
	Title      string `json:"title" gorm:"size:255;not null"`
	Type       string `json:"type" gorm:"size:50"` // text/chart/table/metric/executive_summary

	// 章节配置
	Content  JSONMap `json:"content" gorm:"type:json"` // 章节内容配置
	Order    int     `json:"order" gorm:"default:0"`   // 排序
	Required bool    `json:"required" gorm:"default:true"`

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (ReportSection) TableName() string {
	return "report_sections"
}

// GeneratedReport 生成的报告
type GeneratedReport struct {
	ID         uint         `json:"id" gorm:"primaryKey"`
	TemplateID uint         `json:"template_id" gorm:"index"`
	Title      string       `json:"title" gorm:"size:255"`
	Format     ReportFormat `json:"format" gorm:"size:20"`

	// 文件信息
	FilePath string `json:"file_path" gorm:"size:500"`
	FileSize int64  `json:"file_size"`

	// 状态信息
	Status       ReportStatus `json:"status" gorm:"size:20;default:'pending'"`
	ErrorMessage string       `json:"error_message" gorm:"type:text"`

	// 报告数据
	Data JSONMap `json:"data" gorm:"type:json"` // 生成报告使用的数据

	// 时间戳
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`

	// 关联
	Template ReportTemplate `json:"template" gorm:"foreignKey:TemplateID"`
}

// TableName 指定表名
func (GeneratedReport) TableName() string {
	return "generated_reports"
}

// ReportParameter 报告参数定义
type ReportParameter struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	TemplateID  uint    `json:"template_id" gorm:"index;not null"`
	Name        string  `json:"name" gorm:"size:100"`
	Label       string  `json:"label" gorm:"size:255"`
	Type        string  `json:"type" gorm:"size:50"` // string/number/date/select/multi_select
	Required    bool    `json:"required" gorm:"default:true"`
	Default     string  `json:"default" gorm:"type:text"`
	Options     JSONMap `json:"options" gorm:"type:json"` // 选项JSON
	Description string  `json:"description" gorm:"type:text"`
	Order       int     `json:"order" gorm:"default:0"`

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ReportParameter) TableName() string {
	return "report_parameters"
}

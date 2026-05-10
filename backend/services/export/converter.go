package export

import "time"

// ExportFormat 导出格式
type ExportFormat string

const (
	FormatHTML     ExportFormat = "html"
	FormatPDF      ExportFormat = "pdf"
	FormatExcel    ExportFormat = "excel"
	FormatMarkdown ExportFormat = "markdown"
)

// ExportRequest 导出请求
type ExportRequest struct {
	Format  ExportFormat    `json:"format"`
	Title   string          `json:"title"`
	Data    map[string]any  `json:"data"`
}

// ExportResponse 导出响应
type ExportResponse struct {
	Success  bool   `json:"success"`
	Data     *ExportData `json:"data,omitempty"`
	Error    string `json:"error,omitempty"`
}

// ExportData 导出数据
type ExportData struct {
	Content  string `json:"content"`  // base64编码的内容
	Filename string `json:"filename"` // 文件名
	MimeType string `json:"mime_type"` // MIME类型
	FileSize int64  `json:"file_size"` // 文件大小
}

// ExportMetadata 导出元数据
type ExportMetadata struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	PageType  string    `json:"page_type"`
	Format    ExportFormat `json:"format"`
	Title     string    `json:"title"`
	DataSize  int       `json:"data_size"`
	Timestamp time.Time `json:"timestamp"`
}

// DataConverter 数据转换器接口
type DataConverter interface {
	// Convert 将原始数据转换为可导出的格式
	Convert(data map[string]any) (*ExportableData, error)

	// Validate 验证数据是否有效
	Validate(data map[string]any) error

	// GetSupportedTypes 获取支持的数据类型
	GetSupportedTypes() []string
}

// ExportableData 可导出的数据
type ExportableData struct {
	Title    string         `json:"title"`
	Sections []DataSection  `json:"sections"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// DataSection 数据段
type DataSection struct {
	Title   string         `json:"title"`
	Type    string         `json:"type"` // table, chart, text, list
	Content any            `json:"content"`
	Options map[string]any `json:"options,omitempty"`
}

// Generator 格式生成器接口
type Generator interface {
	// Generate 生成导出内容
	Generate(data *ExportableData) ([]byte, error)

	// GetFormat 获取生成的格式
	GetFormat() ExportFormat

	// GetMimeType 获取MIME类型
	GetMimeType() string

	// GetFileExtension 获取文件扩展名
	GetFileExtension() string
}

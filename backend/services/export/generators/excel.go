package generators

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"etf-insight/services/export"
	"fmt"
	"strings"
)

// ExcelGenerator Excel格式生成器（CSV格式）
type ExcelGenerator struct{}

// NewExcelGenerator 创建Excel生成器
func NewExcelGenerator() *ExcelGenerator {
	return &ExcelGenerator{}
}

// Generate 生成Excel内容（CSV格式）
func (g *ExcelGenerator) Generate(data *export.ExportableData) ([]byte, error) {
	if data == nil {
		return nil, fmt.Errorf("数据不能为空")
	}

	var buf bytes.Buffer

	// 添加BOM以支持中文
	buf.WriteString("\xEF\xBB\xBF")

	writer := csv.NewWriter(&buf)

	// 写入标题
	writer.Write([]string{data.Title})
	writer.Write([]string{}) // 空行

	// 写入各个部分
	for _, section := range data.Sections {
		// 写入部分标题
		writer.Write([]string{section.Title})

		if section.Type == "table" {
			// 写入表格数据
			if tableData, ok := section.Content.(map[string]any); ok {
				// 写入表头
				if headers, ok := tableData["headers"].([]string); ok {
					writer.Write(headers)
				}

				// 写入数据行
				if rows, ok := tableData["rows"].([][]any); ok {
					for _, row := range rows {
						strRow := make([]string, len(row))
						for i, cell := range row {
							strRow[i] = fmt.Sprintf("%v", cell)
						}
						writer.Write(strRow)
					}
				}
			}
		} else if section.Type == "text" {
			// 写入文本内容
			writer.Write([]string{fmt.Sprintf("%v", section.Content)})
		} else if section.Type == "list" {
			// 写入列表内容
			if list, ok := section.Content.([]any); ok {
				for _, item := range list {
					writer.Write([]string{fmt.Sprintf("%v", item)})
				}
			}
		}

		writer.Write([]string{}) // 空行分隔
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("写入CSV失败: %w", err)
	}

	return buf.Bytes(), nil
}

// GetFormat 获取格式
func (g *ExcelGenerator) GetFormat() export.ExportFormat {
	return export.FormatExcel
}

// GetMimeType 获取MIME类型
func (g *ExcelGenerator) GetMimeType() string {
	return "text/csv"
}

// GetFileExtension 获取文件扩展名
func (g *ExcelGenerator) GetFileExtension() string {
	return ".csv"
}

// GenerateExcelWithBase64 生成带Base64编码的Excel
func GenerateExcelWithBase64(data *export.ExportableData) (*export.ExportData, error) {
	generator := NewExcelGenerator()
	content, err := generator.Generate(data)
	if err != nil {
		return nil, err
	}

	encoded := base64.StdEncoding.EncodeToString(content)
	filename := fmt.Sprintf("%s.csv", strings.ReplaceAll(data.Title, " ", "_"))

	return &export.ExportData{
		Content:  encoded,
		Filename: filename,
		MimeType: generator.GetMimeType(),
		FileSize: int64(len(content)),
	}, nil
}

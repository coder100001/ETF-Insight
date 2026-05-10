package generators

import (
	"bytes"
	"encoding/base64"
	"etf-insight/services/export"
	"fmt"
	"strings"
)

// MarkdownGenerator Markdown格式生成器
type MarkdownGenerator struct{}

// NewMarkdownGenerator 创建Markdown生成器
func NewMarkdownGenerator() *MarkdownGenerator {
	return &MarkdownGenerator{}
}

// Generate 生成Markdown内容
func (g *MarkdownGenerator) Generate(data *export.ExportableData) ([]byte, error) {
	if data == nil {
		return nil, fmt.Errorf("数据不能为空")
	}

	var buf bytes.Buffer

	// 写入标题
	buf.WriteString(fmt.Sprintf("# %s\n\n", data.Title))

	// 写入各个部分
	for _, section := range data.Sections {
		buf.WriteString(fmt.Sprintf("## %s\n\n", section.Title))

		if section.Type == "table" {
			// 写入表格
			if tableData, ok := section.Content.(map[string]any); ok {
				// 写入表头
				if headers, ok := tableData["headers"].([]string); ok {
					buf.WriteString("| ")
					buf.WriteString(strings.Join(headers, " | "))
					buf.WriteString(" |\n")

					// 写入分隔线
					buf.WriteString("| ")
					for i := 0; i < len(headers); i++ {
						buf.WriteString("---")
						if i < len(headers)-1 {
							buf.WriteString(" | ")
						}
					}
					buf.WriteString(" |\n")
				}

				// 写入数据行
				if rows, ok := tableData["rows"].([][]any); ok {
					for _, row := range rows {
						buf.WriteString("| ")
						for i, cell := range row {
							buf.WriteString(fmt.Sprintf("%v", cell))
							if i < len(row)-1 {
								buf.WriteString(" | ")
							}
						}
						buf.WriteString(" |\n")
					}
				}
			}
		} else if section.Type == "text" {
			// 写入文本内容
			buf.WriteString(fmt.Sprintf("%v\n", section.Content))
		} else if section.Type == "list" {
			// 写入列表内容
			if list, ok := section.Content.([]any); ok {
				for _, item := range list {
					buf.WriteString(fmt.Sprintf("- %v\n", item))
				}
			}
		}

		buf.WriteString("\n")
	}

	// 写入页脚
	buf.WriteString("---\n\n")
	buf.WriteString("*报告生成时间: ETF-Insight 投资分析平台*\n")

	return buf.Bytes(), nil
}

// GetFormat 获取格式
func (g *MarkdownGenerator) GetFormat() export.ExportFormat {
	return export.FormatMarkdown
}

// GetMimeType 获取MIME类型
func (g *MarkdownGenerator) GetMimeType() string {
	return "text/markdown"
}

// GetFileExtension 获取文件扩展名
func (g *MarkdownGenerator) GetFileExtension() string {
	return ".md"
}

// GenerateMarkdownWithBase64 生成带Base64编码的Markdown
func GenerateMarkdownWithBase64(data *export.ExportableData) (*export.ExportData, error) {
	generator := NewMarkdownGenerator()
	content, err := generator.Generate(data)
	if err != nil {
		return nil, err
	}

	encoded := base64.StdEncoding.EncodeToString(content)
	filename := fmt.Sprintf("%s.md", strings.ReplaceAll(data.Title, " ", "_"))

	return &export.ExportData{
		Content:  encoded,
		Filename: filename,
		MimeType: generator.GetMimeType(),
		FileSize: int64(len(content)),
	}, nil
}

package generators

import (
	"bytes"
	"encoding/base64"
	"etf-insight/services/export"
	"fmt"
	"strings"
)

// PDFGenerator PDF格式生成器
type PDFGenerator struct{}

// NewPDFGenerator 创建PDF生成器
func NewPDFGenerator() *PDFGenerator {
	return &PDFGenerator{}
}

// Generate 生成PDF内容
func (g *PDFGenerator) Generate(data *export.ExportableData) ([]byte, error) {
	if data == nil {
		return nil, fmt.Errorf("数据不能为空")
	}

	// 注意：这里需要实现真正的PDF生成逻辑
	// 由于依赖问题，暂时返回一个简单的文本PDF
	// 实际项目中应该使用gofpdf或其他PDF库

	var buf bytes.Buffer

	// 写入PDF头部
	buf.WriteString("%PDF-1.4\n")
	buf.WriteString("1 0 obj\n")
	buf.WriteString("<< /Type /Catalog /Pages 2 0 R >>\n")
	buf.WriteString("endobj\n")

	// 写入页面
	buf.WriteString("2 0 obj\n")
	buf.WriteString("<< /Type /Pages /Kids [3 0 R] /Count 1 >>\n")
	buf.WriteString("endobj\n")

	// 写入页面内容
	buf.WriteString("3 0 obj\n")
	buf.WriteString("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\n")
	buf.WriteString("endobj\n")

	// 写入内容流
	content := fmt.Sprintf("BT\n/F1 24 Tf\n100 700 Td\n(%s) Tj\nET\n", data.Title)
	buf.WriteString("4 0 obj\n")
	buf.WriteString(fmt.Sprintf("<< /Length %d >>\n", len(content)))
	buf.WriteString("stream\n")
	buf.WriteString(content)
	buf.WriteString("endstream\n")
	buf.WriteString("endobj\n")

	// 写入字体
	buf.WriteString("5 0 obj\n")
	buf.WriteString("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\n")
	buf.WriteString("endobj\n")

	// 写入交叉引用表
	buf.WriteString("xref\n")
	buf.WriteString("0 6\n")
	buf.WriteString("0000000000 65535 f \n")
	buf.WriteString("0000000009 00000 n \n")
	buf.WriteString("0000000058 00000 n \n")
	buf.WriteString("0000000115 00000 n \n")
	buf.WriteString("0000000266 00000 n \n")
	buf.WriteString("0000000366 00000 n \n")

	// 写入尾部
	buf.WriteString("trailer\n")
	buf.WriteString("<< /Size 6 /Root 1 0 R >>\n")
	buf.WriteString("startxref\n")
	buf.WriteString("446\n")
	buf.WriteString("%%EOF\n")

	return buf.Bytes(), nil
}

// GetFormat 获取格式
func (g *PDFGenerator) GetFormat() export.ExportFormat {
	return export.FormatPDF
}

// GetMimeType 获取MIME类型
func (g *PDFGenerator) GetMimeType() string {
	return "application/pdf"
}

// GetFileExtension 获取文件扩展名
func (g *PDFGenerator) GetFileExtension() string {
	return ".pdf"
}

// GeneratePDFWithBase64 生成带Base64编码的PDF
func GeneratePDFWithBase64(data *export.ExportableData) (*export.ExportData, error) {
	generator := NewPDFGenerator()
	content, err := generator.Generate(data)
	if err != nil {
		return nil, err
	}

	encoded := base64.StdEncoding.EncodeToString(content)
	filename := fmt.Sprintf("%s.pdf", strings.ReplaceAll(data.Title, " ", "_"))

	return &export.ExportData{
		Content:  encoded,
		Filename: filename,
		MimeType: generator.GetMimeType(),
		FileSize: int64(len(content)),
	}, nil
}

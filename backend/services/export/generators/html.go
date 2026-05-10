package generators

import (
	"bytes"
	"encoding/base64"
	"etf-insight/services/export"
	"fmt"
	"html/template"
)

// HTMLGenerator HTML格式生成器
type HTMLGenerator struct{}

// NewHTMLGenerator 创建HTML生成器
func NewHTMLGenerator() *HTMLGenerator {
	return &HTMLGenerator{}
}

// Generate 生成HTML内容
func (g *HTMLGenerator) Generate(data *export.ExportableData) ([]byte, error) {
	if data == nil {
		return nil, fmt.Errorf("数据不能为空")
	}

	// 准备模板数据
	templateData := struct {
		Title    string
		Sections []export.DataSection
	}{
		Title:    data.Title,
		Sections: data.Sections,
	}

	// 解析模板
	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return nil, fmt.Errorf("解析HTML模板失败: %w", err)
	}

	// 执行模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return nil, fmt.Errorf("生成HTML失败: %w", err)
	}

	return buf.Bytes(), nil
}

// GetFormat 获取格式
func (g *HTMLGenerator) GetFormat() export.ExportFormat {
	return export.FormatHTML
}

// GetMimeType 获取MIME类型
func (g *HTMLGenerator) GetMimeType() string {
	return "text/html"
}

// GetFileExtension 获取文件扩展名
func (g *HTMLGenerator) GetFileExtension() string {
	return ".html"
}

// htmlTemplate HTML模板
const htmlTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        h1 {
            color: #2c3e50;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }
        h2 {
            color: #34495e;
            margin-top: 30px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        th, td {
            padding: 12px 15px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #3498db;
            color: white;
            font-weight: 600;
        }
        tr:nth-child(even) {
            background-color: #f8f9fa;
        }
        tr:hover {
            background-color: #e9ecef;
        }
        .text-content {
            background-color: #f8f9fa;
            padding: 15px;
            border-radius: 5px;
            border-left: 4px solid #3498db;
            margin: 15px 0;
        }
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #eee;
            color: #7f8c8d;
            font-size: 14px;
            text-align: center;
        }
        @media print {
            body {
                padding: 0;
            }
            table {
                box-shadow: none;
            }
        }
    </style>
</head>
<body>
    <h1>{{.Title}}</h1>

    {{range .Sections}}
    <div class="section">
        <h2>{{.Title}}</h2>

        {{if eq .Type "table"}}
        <table>
            <thead>
                <tr>
                    {{range .Content.headers}}
                    <th>{{.}}</th>
                    {{end}}
                </tr>
            </thead>
            <tbody>
                {{range .Content.rows}}
                <tr>
                    {{range .}}
                    <td>{{.}}</td>
                    {{end}}
                </tr>
                {{end}}
            </tbody>
        </table>
        {{else if eq .Type "text"}}
        <div class="text-content">
            {{.Content}}
        </div>
        {{else if eq .Type "list"}}
        <ul>
            {{range .Content}}
            <li>{{.}}</li>
            {{end}}
        </ul>
        {{end}}
    </div>
    {{end}}

    <div class="footer">
        <p>报告生成时间: {{.Title}}</p>
        <p>ETF-Insight 投资分析平台</p>
    </div>
</body>
</html>`

// GenerateHTMLWithBase64 生成带Base64编码的HTML
func GenerateHTMLWithBase64(data *export.ExportableData) (*export.ExportData, error) {
	generator := NewHTMLGenerator()
	content, err := generator.Generate(data)
	if err != nil {
		return nil, err
	}

	encoded := base64.StdEncoding.EncodeToString(content)
	filename := fmt.Sprintf("%s.html", data.Title)

	return &export.ExportData{
		Content:  encoded,
		Filename: filename,
		MimeType: generator.GetMimeType(),
		FileSize: int64(len(content)),
	}, nil
}

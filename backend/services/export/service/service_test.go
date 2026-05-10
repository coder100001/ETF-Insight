package service

import (
	"testing"

	"etf-insight/services/export"
)

func TestExportService_Export(t *testing.T) {
	service := NewExportService()

	// 测试空请求
	t.Run("空请求", func(t *testing.T) {
		request := &export.ExportRequest{
			Format: export.FormatHTML,
			Data:   map[string]any{},
		}
		response, err := service.Export("user1", "testuser", "portfolio", request)
		if err != nil {
			t.Errorf("期望无错误，得到: %v", err)
		}
		if !response.Success {
			t.Errorf("期望成功响应，得到失败: %s", response.Error)
		}
	})

	// 测试无效格式
	t.Run("无效格式", func(t *testing.T) {
		request := &export.ExportRequest{
			Format: "invalid",
			Data:   map[string]any{"test": "data"},
		}
		response, err := service.Export("user1", "testuser", "portfolio", request)
		if err != nil {
			t.Errorf("期望无错误，得到: %v", err)
		}
		if response.Success {
			t.Error("期望失败响应，得到成功")
		}
	})

	// 测试空数据
	t.Run("空数据", func(t *testing.T) {
		request := &export.ExportRequest{
			Format: export.FormatHTML,
			Data:   nil,
		}
		response, err := service.Export("user1", "testuser", "portfolio", request)
		if err != nil {
			t.Errorf("期望无错误，得到: %v", err)
		}
		if response.Success {
			t.Error("期望失败响应，得到成功")
		}
	})

	// 测试有效导出
	t.Run("有效导出", func(t *testing.T) {
		request := &export.ExportRequest{
			Format: export.FormatHTML,
			Title:  "测试报告",
			Data: map[string]any{
				"overview": "这是一个测试投资组合",
				"allocations": []any{
					map[string]any{
						"name":   "ETF A",
						"symbol": "ETF_A",
						"weight": 0.6,
						"amount": 10000,
						"return": 0.15,
					},
				},
			},
		}
		response, err := service.Export("user1", "testuser", "portfolio", request)
		if err != nil {
			t.Errorf("期望无错误，得到: %v", err)
		}
		if !response.Success {
			t.Errorf("期望成功响应，得到失败: %s", response.Error)
		}
		if response.Data == nil {
			t.Error("期望有导出数据")
		}
		if response.Data.Content == "" {
			t.Error("期望有内容")
		}
		if response.Data.Filename == "" {
			t.Error("期望有文件名")
		}
	})
}

func TestExportService_GetSupportedFormats(t *testing.T) {
	service := NewExportService()
	formats := service.GetSupportedFormats()

	if len(formats) == 0 {
		t.Error("期望有支持的格式")
	}

	expectedFormats := map[export.ExportFormat]bool{
		export.FormatHTML:     true,
		export.FormatPDF:      true,
		export.FormatExcel:    true,
		export.FormatMarkdown: true,
	}

	for _, format := range formats {
		if !expectedFormats[format] {
			t.Errorf("不期望的格式: %s", format)
		}
	}
}

func TestExportService_GetSupportedTypes(t *testing.T) {
	service := NewExportService()
	types := service.GetSupportedTypes()

	if len(types) == 0 {
		t.Error("期望有支持的类型")
	}

	expectedTypes := map[string]bool{
		"portfolio": true,
		"risk":      true,
		"default":   true,
	}

	for _, pageType := range types {
		if !expectedTypes[pageType] {
			t.Errorf("不期望的类型: %s", pageType)
		}
	}
}

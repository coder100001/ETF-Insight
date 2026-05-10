package converters

import (
	"etf-insight/services/export"
	"fmt"
)

// ConverterRegistry 转换器注册表
type ConverterRegistry struct {
	converters map[string]export.DataConverter
}

// NewConverterRegistry 创建转换器注册表
func NewConverterRegistry() *ConverterRegistry {
	registry := &ConverterRegistry{
		converters: make(map[string]export.DataConverter),
	}

	// 注册默认转换器
	registry.Register("portfolio", NewPortfolioConverter())
	registry.Register("risk", NewRiskConverter())
	registry.Register("default", NewDefaultConverter())

	return registry
}

// Register 注册转换器
func (r *ConverterRegistry) Register(pageType string, converter export.DataConverter) {
	r.converters[pageType] = converter
}

// GetConverter 获取转换器
func (r *ConverterRegistry) GetConverter(pageType string) (export.DataConverter, error) {
	converter, exists := r.converters[pageType]
	if !exists {
		// 返回默认转换器
		return r.converters["default"], nil
	}
	return converter, nil
}

// GetSupportedTypes 获取所有支持的类型
func (r *ConverterRegistry) GetSupportedTypes() []string {
	types := make([]string, 0, len(r.converters))
	for pageType := range r.converters {
		types = append(types, pageType)
	}
	return types
}

// DefaultConverter 默认转换器
type DefaultConverter struct{}

// NewDefaultConverter 创建默认转换器
func NewDefaultConverter() *DefaultConverter {
	return &DefaultConverter{}
}

// Convert 转换数据
func (c *DefaultConverter) Convert(data map[string]any) (*export.ExportableData, error) {
	if err := c.Validate(data); err != nil {
		return nil, err
	}

	result := &export.ExportableData{
		Title:    "数据导出报告",
		Sections: make([]export.DataSection, 0),
	}

	// 将数据转换为表格格式
	if len(data) > 0 {
		tableData := c.convertMapToTable(data)
		result.Sections = append(result.Sections, export.DataSection{
			Title:   "数据详情",
			Type:    "table",
			Content: tableData,
		})
	}

	return result, nil
}

// Validate 验证数据
func (c *DefaultConverter) Validate(data map[string]any) error {
	if data == nil {
		return fmt.Errorf("数据不能为空")
	}

	// 检查数据大小
	dataSize := estimateDataSize(data)
	if dataSize > 1024*1024 { // 1MB限制
		return fmt.Errorf("数据过大，超过1MB限制")
	}

	return nil
}

// GetSupportedTypes 获取支持的数据类型
func (c *DefaultConverter) GetSupportedTypes() []string {
	return []string{"default"}
}

// convertMapToTable 将Map转换为表格格式
func (c *DefaultConverter) convertMapToTable(data map[string]any) map[string]any {
	headers := []string{"键", "值"}
	rows := make([][]any, 0)

	for key, value := range data {
		row := []any{key, value}
		rows = append(rows, row)
	}

	return map[string]any{
		"headers": headers,
		"rows":    rows,
	}
}

// estimateDataSize 估算数据大小
func estimateDataSize(data map[string]any) int {
	size := 0
	for _, v := range data {
		switch val := v.(type) {
		case string:
			size += len(val)
		case map[string]any:
			size += estimateDataSize(val)
		case []any:
			for _, item := range val {
				if itemMap, ok := item.(map[string]any); ok {
					size += estimateDataSize(itemMap)
				}
			}
		default:
			size += 8
		}
	}
	return size
}

package converters

import (
	"etf-insight/services/export"
	"fmt"
)

// PortfolioConverter 投资组合数据转换器
type PortfolioConverter struct{}

// NewPortfolioConverter 创建投资组合转换器
func NewPortfolioConverter() *PortfolioConverter {
	return &PortfolioConverter{}
}

// Convert 转换投资组合数据
func (c *PortfolioConverter) Convert(data map[string]any) (*export.ExportableData, error) {
	if err := c.Validate(data); err != nil {
		return nil, err
	}

	result := &export.ExportableData{
		Title:    "投资组合分析报告",
		Sections: make([]export.DataSection, 0),
	}

	// 添加概览部分
	if overview, ok := data["overview"]; ok {
		result.Sections = append(result.Sections, export.DataSection{
			Title:   "投资组合概览",
			Type:    "text",
			Content: overview,
		})
	}

	// 添加资产配置表格
	if allocations, ok := data["allocations"]; ok {
		if allocList, ok := allocations.([]any); ok {
			tableData := c.convertAllocationsToTable(allocList)
			result.Sections = append(result.Sections, export.DataSection{
				Title:   "资产配置",
				Type:    "table",
				Content: tableData,
			})
		}
	}

	// 添加收益统计
	if returns, ok := data["returns"]; ok {
		result.Sections = append(result.Sections, export.DataSection{
			Title:   "收益统计",
			Type:    "table",
			Content: returns,
		})
	}

	// 添加风险指标
	if risk, ok := data["risk"]; ok {
		result.Sections = append(result.Sections, export.DataSection{
			Title:   "风险指标",
			Type:    "table",
			Content: risk,
		})
	}

	return result, nil
}

// Validate 验证数据
func (c *PortfolioConverter) Validate(data map[string]any) error {
	if data == nil {
		return fmt.Errorf("数据不能为空")
	}

	// 检查数据大小（粗略估计）
	dataSize := estimateDataSize(data)
	if dataSize > 1024*1024 { // 1MB限制
		return fmt.Errorf("数据过大，超过1MB限制")
	}

	return nil
}

// GetSupportedTypes 获取支持的数据类型
func (c *PortfolioConverter) GetSupportedTypes() []string {
	return []string{"portfolio", "investment"}
}

// convertAllocationsToTable 将资产配置转换为表格格式
func (c *PortfolioConverter) convertAllocationsToTable(allocations []any) map[string]any {
	if len(allocations) == 0 {
		return nil
	}

	// 假设每个allocation是一个map
	headers := []string{"资产名称", "代码", "权重", "金额", "收益率"}
	rows := make([][]any, 0)

	for _, alloc := range allocations {
		if allocMap, ok := alloc.(map[string]any); ok {
			row := []any{
				allocMap["name"],
				allocMap["symbol"],
				allocMap["weight"],
				allocMap["amount"],
				allocMap["return"],
			}
			rows = append(rows, row)
		}
	}

	return map[string]any{
		"headers": headers,
		"rows":    rows,
	}
}

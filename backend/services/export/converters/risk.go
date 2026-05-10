package converters

import (
	"etf-insight/services/export"
	"fmt"
)

// RiskConverter 风险分析数据转换器
type RiskConverter struct{}

// NewRiskConverter 创建风险分析转换器
func NewRiskConverter() *RiskConverter {
	return &RiskConverter{}
}

// Convert 转换风险分析数据
func (c *RiskConverter) Convert(data map[string]any) (*export.ExportableData, error) {
	if err := c.Validate(data); err != nil {
		return nil, err
	}

	result := &export.ExportableData{
		Title:    "风险分析报告",
		Sections: make([]export.DataSection, 0),
	}

	// 添加VaR分析
	if varData, ok := data["var"]; ok {
		result.Sections = append(result.Sections, export.DataSection{
			Title:   "VaR分析",
			Type:    "table",
			Content: varData,
		})
	}

	// 添加回撤分析
	if drawdown, ok := data["drawdown"]; ok {
		result.Sections = append(result.Sections, export.DataSection{
			Title:   "回撤分析",
			Type:    "table",
			Content: drawdown,
		})
	}

	// 添加波动率分析
	if volatility, ok := data["volatility"]; ok {
		result.Sections = append(result.Sections, export.DataSection{
			Title:   "波动率分析",
			Type:    "table",
			Content: volatility,
		})
	}

	// 添加相关性分析
	if correlation, ok := data["correlation"]; ok {
		result.Sections = append(result.Sections, export.DataSection{
			Title:   "相关性分析",
			Type:    "table",
			Content: correlation,
		})
	}

	// 添加风险指标汇总
	if summary, ok := data["summary"]; ok {
		result.Sections = append(result.Sections, export.DataSection{
			Title:   "风险指标汇总",
			Type:    "text",
			Content: summary,
		})
	}

	return result, nil
}

// Validate 验证数据
func (c *RiskConverter) Validate(data map[string]any) error {
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
func (c *RiskConverter) GetSupportedTypes() []string {
	return []string{"risk", "risk_analysis"}
}

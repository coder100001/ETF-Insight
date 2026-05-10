package generators

import (
	"etf-insight/services/export"
	"fmt"
)

// GeneratorRegistry 生成器注册表
type GeneratorRegistry struct {
	generators map[export.ExportFormat]export.Generator
}

// NewGeneratorRegistry 创建生成器注册表
func NewGeneratorRegistry() *GeneratorRegistry {
	registry := &GeneratorRegistry{
		generators: make(map[export.ExportFormat]export.Generator),
	}

	// 注册默认生成器
	registry.Register(NewHTMLGenerator())
	registry.Register(NewPDFGenerator())
	registry.Register(NewExcelGenerator())
	registry.Register(NewMarkdownGenerator())

	return registry
}

// Register 注册生成器
func (r *GeneratorRegistry) Register(generator export.Generator) {
	r.generators[generator.GetFormat()] = generator
}

// GetGenerator 获取生成器
func (r *GeneratorRegistry) GetGenerator(format export.ExportFormat) (export.Generator, error) {
	generator, exists := r.generators[format]
	if !exists {
		return nil, fmt.Errorf("不支持的导出格式: %s", format)
	}
	return generator, nil
}

// GetSupportedFormats 获取所有支持的格式
func (r *GeneratorRegistry) GetSupportedFormats() []export.ExportFormat {
	formats := make([]export.ExportFormat, 0, len(r.generators))
	for format := range r.generators {
		formats = append(formats, format)
	}
	return formats
}

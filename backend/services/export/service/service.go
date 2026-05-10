package service

import (
	"encoding/base64"
	"fmt"
	"time"

	"etf-insight/services/export"
	"etf-insight/services/export/converters"
	"etf-insight/services/export/generators"
)

// ExportService 导出服务
type ExportService struct {
	converterRegistry *converters.ConverterRegistry
	generatorRegistry *generators.GeneratorRegistry
}

// NewExportService 创建导出服务
func NewExportService() *ExportService {
	return &ExportService{
		converterRegistry: converters.NewConverterRegistry(),
		generatorRegistry: generators.NewGeneratorRegistry(),
	}
}

// Export 执行导出
func (s *ExportService) Export(userID, username, pageType string, request *export.ExportRequest) (*export.ExportResponse, error) {
	startTime := time.Now()

	// 记录开始
	export.LogExportStart(userID, username, pageType, string(request.Format))

	// 验证请求
	if err := s.validateRequest(request); err != nil {
		export.LogExportValidation(userID, username, pageType, err)
		export.RecordExport(userID, username, pageType, string(request.Format), 400, err)
		return &export.ExportResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 获取转换器
	converter, err := s.converterRegistry.GetConverter(pageType)
	if err != nil {
		export.LogExport(userID, username, pageType, string(request.Format), 0, err, time.Since(startTime))
		export.RecordExport(userID, username, pageType, string(request.Format), 400, err)
		return &export.ExportResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 转换数据
	exportableData, err := converter.Convert(request.Data)
	if err != nil {
		export.LogExport(userID, username, pageType, string(request.Format), 0, err, time.Since(startTime))
		export.RecordExport(userID, username, pageType, string(request.Format), 500, err)
		return &export.ExportResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 设置标题
	if request.Title != "" {
		exportableData.Title = request.Title
	}

	// 获取生成器
	generator, err := s.generatorRegistry.GetGenerator(request.Format)
	if err != nil {
		export.LogExport(userID, username, pageType, string(request.Format), 0, err, time.Since(startTime))
		export.RecordExport(userID, username, pageType, string(request.Format), 400, err)
		return &export.ExportResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 生成内容
	content, err := generator.Generate(exportableData)
	if err != nil {
		export.LogExport(userID, username, pageType, string(request.Format), 0, err, time.Since(startTime))
		export.RecordExport(userID, username, pageType, string(request.Format), 500, err)
		return &export.ExportResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 编码为Base64
	encoded := base64.StdEncoding.EncodeToString(content)

	// 生成文件名
	filename := fmt.Sprintf("%s_%s%s",
		exportableData.Title,
		time.Now().Format("20060102_150405"),
		generator.GetFileExtension())

	// 计算数据大小
	dataSize := len(request.Data)

	// 记录成功
	export.LogExport(userID, username, pageType, string(request.Format), dataSize, nil, time.Since(startTime))
	export.RecordExport(userID, username, pageType, string(request.Format), 200, nil)

	return &export.ExportResponse{
		Success: true,
		Data: &export.ExportData{
			Content:  encoded,
			Filename: filename,
			MimeType: generator.GetMimeType(),
			FileSize: int64(len(content)),
		},
	}, nil
}

// validateRequest 验证请求
func (s *ExportService) validateRequest(request *export.ExportRequest) error {
	if request == nil {
		return export.ErrDataMissing
	}

	if request.Format == "" {
		return export.WrapError(export.ErrFormatNotSupport, "导出格式不能为空")
	}

	// 检查格式是否支持
	if _, err := s.generatorRegistry.GetGenerator(request.Format); err != nil {
		return export.WrapError(export.ErrFormatNotSupport, string(request.Format))
	}

	if request.Data == nil {
		return export.ErrDataMissing
	}

	// 检查数据大小
	dataSize := estimateDataSize(request.Data)
	if dataSize > 1024*1024 { // 1MB限制
		return export.WrapError(export.ErrDataTooLarge, fmt.Sprintf("数据大小 %d 字节，超过1MB限制", dataSize))
	}

	return nil
}

// GetSupportedFormats 获取支持的格式
func (s *ExportService) GetSupportedFormats() []export.ExportFormat {
	return s.generatorRegistry.GetSupportedFormats()
}

// GetSupportedTypes 获取支持的类型
func (s *ExportService) GetSupportedTypes() []string {
	return s.converterRegistry.GetSupportedTypes()
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

package export

import (
	"time"

	"etf-insight/utils"
)

// LogExport 记录导出操作日志
func LogExport(userID, username, pageType, format string, dataSize int, err error, duration time.Duration) {
	if err != nil {
		utils.Error("Export failed", err,
			"user_id", userID,
			"username", username,
			"page_type", pageType,
			"format", format,
			"data_size", dataSize,
			"duration_ms", duration.Milliseconds(),
		)
	} else {
		utils.Info("Export success",
			"user_id", userID,
			"username", username,
			"page_type", pageType,
			"format", format,
			"data_size", dataSize,
			"duration_ms", duration.Milliseconds(),
		)
	}
}

// LogExportStart 记录导出开始
func LogExportStart(userID, username, pageType, format string) {
	utils.Info("Export started",
		"user_id", userID,
		"username", username,
		"page_type", pageType,
		"format", format,
	)
}

// LogExportValidation 记录数据验证失败
func LogExportValidation(userID, username, pageType string, err error) {
	utils.Warn("Export validation failed",
		"user_id", userID,
		"username", username,
		"page_type", pageType,
		"error", err.Error(),
	)
}

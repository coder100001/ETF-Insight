# 报告导出功能设计文档

**创建日期**: 2026-05-10
**状态**: 已确认
**复杂度**: L2

---

## 1. 背景与目标

### 1.1 背景
ETF-Insight 项目需要完善报告导出功能，支持多种格式（HTML、PDF、Excel、Markdown），并集成错误处理、日志记录和操作记录。

### 1.2 目标
1. 实现多格式报告导出（HTML、PDF、Excel、Markdown）
2. 统一错误处理机制
3. 日志记录（复用现有 Logger）
4. 操作记录（复用现有 AuditLog）
5. 预留全链路追踪扩展点

### 1.3 非目标
- 不实现复杂的权限控制
- 不实现异步导出队列
- 不实现邮件发送功能

---

## 2. 架构设计

### 2.1 整体架构

```
前端页面
├── ExportButton 组件
└── 调用 /api/export/:type

后端服务
├── ExportHandler (API 入口)
├── ExportService (业务逻辑)
├── DataConverter (数据转换)
├── Generator (格式生成)
├── ErrorHandler (错误处理)
├── Logger (日志记录)
└── AuditRecorder (操作记录)
```

### 2.2 文件结构

```
backend/services/export/
├── errors.go          # 错误定义
├── logger.go          # 日志记录
├── audit.go           # 操作记录
├── converter.go       # 转换器接口
├── generators/
│   ├── html.go        # HTML 生成器
│   ├── pdf.go         # PDF 生成器
│   ├── excel.go       # Excel 生成器
│   └── markdown.go    # Markdown 生成器
└── converters/
    ├── portfolio.go   # 投资组合转换器
    ├── risk.go        # 风险分析转换器
    └── ...            # 其他转换器

frontend/src/
├── components/
│   └── ExportButton.tsx   # 导出按钮组件
└── services/
    └── exportApi.ts       # 导出 API 服务
```

---

## 3. 详细设计

### 3.1 错误处理

```go
// errors.go
type ExportError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

var (
    ErrDataMissing      = NewError("EXPORT_001", "数据缺失", "")
    ErrDataInvalid      = NewError("EXPORT_002", "数据无效", "")
    ErrDataTooLarge     = NewError("EXPORT_003", "数据过大", "")
    ErrFormatNotSupport = NewError("EXPORT_004", "格式不支持", "")
    ErrGenerateFailed   = NewError("EXPORT_005", "生成失败", "")
    ErrTimeout          = NewError("EXPORT_006", "操作超时", "")
)
```

### 3.2 日志记录

```go
// logger.go
func LogExport(userID, username, pageType, format string, dataSize int, err error, duration time.Duration) {
    if err != nil {
        utils.Error("Export failed", err, ...)
    } else {
        utils.Info("Export success", ...)
    }
}
```

### 3.3 操作记录

```go
// audit.go
func RecordExport(userID, username, pageType, format string, statusCode int, err error) {
    log := &models.AuditLog{...}
    go models.DB.Create(log)
}
```

### 3.4 API 设计

```
POST /api/export/:type
Request:
{
    "format": "pdf",
    "title": "报告标题",
    "data": {...}
}

Response:
{
    "success": true,
    "data": {
        "content": "base64...",
        "filename": "report.pdf"
    }
}
```

---

## 4. 边界条件

### 4.1 数据验证
- 数据不能为空
- 数据大小限制 1MB
- 历史数据最多 1000 条

### 4.2 错误处理
- 输入验证失败：返回 400
- 生成失败：返回 500
- 超时：返回 504

### 4.3 性能
- 导出超时：30 秒
- 内存限制：512MB

---

## 5. 扩展点

### 5.1 用户信息获取
```go
// TODO: 实现从 context 获取 userID 和 username
func GetUserID(ctx context.Context) string
func GetUsername(ctx context.Context) string
```

### 5.2 全链路追踪
```go
// TODO: 实现 request_id 传递
// 1. 在中间件中生成 request_id
// 2. 注入 context
// 3. 在日志中记录
```

---

## 6. 测试策略

### 6.1 单元测试
- 测试各格式生成器
- 测试数据转换器
- 测试错误处理

### 6.2 集成测试
- 测试完整导出流程
- 测试边界条件
- 测试错误场景

---

## 7. 风险评估

| 风险 | 等级 | 应对措施 |
|------|------|---------|
| 数据过大 | 中 | 限制数据大小，提示用户 |
| 生成超时 | 中 | 设置超时，异步处理 |
| 内存不足 | 低 | 限制数据量，优化生成 |

---

**文档版本**: 1.0
**最后更新**: 2026-05-10

# Design Doc 003: 项目向金融分析工具演进规划

## 元数据

- **编号**: 003
- **标题**: 项目向专业金融分析工具演进规划
- **状态**: draft
- **创建日期**: 2026-04-30
- **最后更新**: 2026-04-30
- **关联任务**: 项目战略演进分析
- **复杂度级别**: L3
- **涉及端**: 双端（后端 + 前端）
- **前置 Design Doc**: 001-code-quality-analysis-and-improvement.md

***

## 1. 背景与动机

### 为什么需要这个规划？

当前 ETF-Insight 已经是一个功能丰富的 ETF 量化分析平台，但要成为真正的**专业金融分析工具**，还需要在多个维度进行升级：

1. **用户体验** - 更专业的报告生成、自定义分析流程
2. **功能深度** - 更丰富的金融分析能力、策略回测完善
3. **生态建设** - 插件系统、社区策略分享、API 服务
4. **商业价值** - 为机构客户提供专业分析报告

### 当前系统的痛点

| 痛点      | 影响   | 当前状态        |
| ------- | ---- | ----------- |
| 报告生成能力弱 | 🟡 中 | 缺乏专业报告导出    |
| 自定义分析流程 | 🟡 中 | 分析步骤固定，不够灵活 |
| 策略分享机制  | 🟡 中 | 没有策略仓库      |
| 移动端支持   | 🟡 中 | 无移动端适配      |
| 实时推送通知  | 🔴 高 | 无推送功能       |

### 业务/技术驱动因素

- **市场趋势** - 个人投资者对专业分析工具需求增长
- **技术成熟** - AI/ML、实时推送等技术可以提升体验
- **社区生态** - 开源项目需要更丰富的生态系统

***

## 2. 调研与现状分析

### 2.1 现有实现

让我们分析项目当前的能力矩阵：

| 功能领域         | 当前能力                  | 成熟度   |
| ------------ | --------------------- | ----- |
| **ETF 数据管理** | ✅ 完整的 ETF CRUD、实时数据同步 | ⭐⭐⭐⭐⭐ |
| **投资组合分析**   | ✅ 情景分析、风险指标、优化        | ⭐⭐⭐⭐  |
| **量化回测**     | ✅ 事件驱动回测引擎            | ⭐⭐⭐   |
| **技术分析**     | ✅ RSI/MACD/布林带        | ⭐⭐⭐⭐  |
| **因子分析**     | ✅ Fama-French 三因子/五因子 | ⭐⭐⭐⭐  |
| **持仓穿透**     | ✅ 底层持仓分析、重叠度计算        | ⭐⭐⭐⭐⭐ |
| **报告导出**     | ❌ 无                   | ⭐     |
| **自定义流程**    | ❌ 分析步骤固定              | ⭐     |
| **策略分享**     | ❌ 无                   | ⭐     |
| **实时通知**     | ❌ 无                   | ⭐     |

### 2.2 业界实践

让我们分析成功的金融分析工具：

| 产品                     | 优势             | 可借鉴点      |
| ---------------------- | -------------- | --------- |
| **Bloomberg Terminal** | 专业数据、实时推送、深度分析 | 报告生成、专业指标 |
| **TradingView**        | 技术分析、社区策略、社交功能 | 图表交互、策略分享 |
| **QuantConnect**       | 策略回测、开源社区、云平台  | 策略仓库、云服务  |
| **Morningstar**        | 专业报告、评级系统、深度研究 | 报告模板、评级体系 |

### 2.3 技术约束

- 保持向后兼容
- 增量演进，避免大规模重构
- 保持开源特性
- 性能优先原则

***

## 3. 可选方案

### 方案 A: 渐进式演进（推荐）

**描述**: 在现有基础上逐步增强，分阶段实现目标

- **阶段 1**: 报告生成系统（3个月）
- **阶段 2**: 自定义分析流程（3个月）
- **阶段 3**: 插件系统与生态建设（6个月）
- **阶段 4**: AI 增强与智能分析（12个月）

**优点**:

- ✅ 风险可控，每阶段都有可用成果
- ✅ 可以持续收集用户反馈
- ✅ 避免大规模重构
- ✅ 与现有架构兼容

**缺点**:

- ❌ 演进周期较长
- ❌ 需要持续的开发投入

**工作量**: 🟡 中等（需要 12-18 个月）

***

### 方案 B: 大版本重构

**描述**: 重新设计架构，一次性实现所有目标

- 完全重写前端架构
- 重写后端服务设计
- 新的数据模型

**优点**:

- ✅ 架构更合理、更现代化
- ✅ 可以一次性解决所有问题

**缺点**:

- ❌ 风险极高，可能引入新问题
- ❌ 开发周期长，用户等待久
- ❌ 与现有功能兼容性问题
- ❌ 需要大量资源投入

**工作量**: 🔴 高（需要 24 个月以上）

***

### 方案 C: 专注核心，有限扩展

**描述**: 只聚焦在 ETF 分析领域，不扩展到全金融工具

- 保持当前功能范围
- 只做质量优化和 bug 修复
- 不添加新的大功能

**优点**:

- ✅ 开发成本低
- ✅ 风险最小
- ✅ 可以专注做好现有功能

**缺点**:

- ❌ 项目发展受限
- ❌ 无法满足用户更多需求
- ❌ 可能被竞品超越

**工作量**: 🟢 低（只需要维护）

***

## 4. 决策

### 选定方案

**方案 A: 渐进式演进**

### 决策理由

1. **风险可控** - 分阶段实施，每阶段都有可交付成果
2. **用户导向** - 可以根据用户反馈调整方向
3. **资源合理** - 适合开源项目的开发节奏
4. **兼容性好** - 保持与现有功能的兼容

### 权衡取舍

- **时间换质量** - 虽然周期长，但质量更有保障
- **功能换稳定** - 不追求一次性大而全，而是稳定可用
- **社区参与** - 鼓励社区贡献，共同建设生态

***

## 5. 后端设计

### 5.1 报告生成系统

```go
// ReportTemplate - 报告模板配置
type ReportTemplate struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Sections    []ReportSection        `json:"sections"`
    OutputFormats []string             `json:"output_formats"` // PDF/Excel/HTML
    CreatedAt   time.Time              `json:"created_at"`
}

// ReportSection - 报告章节
type ReportSection struct {
    ID          string                 `json:"id"`
    Title       string                 `json:"title"`
    Type        string                 `json:"type"` // text/chart/table/metric
    Content     map[string]interface{} `json:"content"`
    Order       int                    `json:"order"`
}

// ReportGenerator - 报告生成器接口
type ReportGenerator interface {
    Generate(templateID string, data interface{}) ([]byte, error)
    GetTemplates() ([]ReportTemplate, error)
    CreateTemplate(template ReportTemplate) error
}
```

### 5.2 分析工作流引擎

```go
// AnalysisWorkflow - 自定义分析流程
type AnalysisWorkflow struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Steps       []WorkflowStep         `json:"steps"`
    Inputs      map[string]interface{} `json:"inputs"`
    Outputs     map[string]interface{} `json:"outputs"`
    CreatedAt   time.Time              `json:"created_at"`
}

// WorkflowStep - 工作流步骤
type WorkflowStep struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Type        string                 `json:"type"`
    Config      map[string]interface{} `json:"config"`
    DependsOn   []string               `json:"depends_on"`
    Order       int                    `json:"order"`
}

// WorkflowEngine - 工作流执行引擎
type WorkflowEngine interface {
    Execute(workflowID string, inputs map[string]interface{}) (map[string]interface{}, error)
    GetStatus(executionID string) (WorkflowExecutionStatus, error)
    ListExecutions(workflowID string) ([]WorkflowExecution, error)
}
```

### 5.3 策略分享系统

```go
// Strategy - 分享策略
type Strategy struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Description     string                 `json:"description"`
    Type            string                 `json:"type"` // trading/portfolio/risk
    Code            string                 `json:"code"`
    Config          map[string]interface{} `json:"config"`
    BacktestResults *BacktestResult        `json:"backtest_results"`
    AuthorID        string                 `json:"author_id"`
    Rating          float64                `json:"rating"`
    Downloads       int                    `json:"downloads"`
    CreatedAt       time.Time              `json:"created_at"`
    UpdatedAt       time.Time              `json:"updated_at"`
}

// StrategyRepository - 策略仓库
type StrategyRepository interface {
    List(page, pageSize int) ([]Strategy, error)
    Get(id string) (*Strategy, error)
    Create(strategy Strategy) error
    Update(strategy Strategy) error
    Rate(id string, rating float64) error
    Search(query string) ([]Strategy, error)
}
```

### 5.4 实时通知系统

```go
// Notification - 通知消息
type Notification struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    Type      string    `json:"type"` // alert/report/update
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    Data      map[string]interface{} `json:"data"`
    Read      bool      `json:"read"`
    CreatedAt time.Time `json:"created_at"`
}

// NotificationService - 通知服务
type NotificationService interface {
    Send(userID string, notification Notification) error
    List(userID string, page, pageSize int) ([]Notification, error)
    MarkAsRead(id string) error
    GetUnreadCount(userID string) (int, error)
}
```

***

## 6. 前端设计

### 6.1 报告生成器 UI

```
Report Builder/
├── TemplateSelector/     // 报告模板选择
├── PreviewPane/          // 实时预览
├── ExportOptions/        // 导出格式选项
└── History/              // 历史报告
```

### 6.2 工作流编辑器

```typescript
// WorkflowEditor - 可视化工作流编辑
interface WorkflowEditorProps {
    workflow: AnalysisWorkflow;
    onSave: (workflow: AnalysisWorkflow) => void;
    onExecute: (workflow: AnalysisWorkflow) => void;
}

// StepNode - 工作流步骤节点
interface StepNode {
    id: string;
    type: string;
    config: any;
    position: { x: number; y: number };
}
```

### 6.3 策略市场 UI

```
StrategyMarket/
├── SearchBar/            // 搜索和筛选
├── StrategyCard/         // 策略卡片
├── DetailView/           // 策略详情
└── UploadModal/          // 策略上传
```

### 6.4 通知中心

```typescript
// NotificationCenter - 通知中心组件
interface NotificationCenterProps {
    userId: string;
}
```

***

## 7. 前后端交互

### 新增 API 端点

```
POST   /api/reports/templates          // 创建报告模板
GET    /api/reports/templates          // 获取模板列表
POST   /api/reports/generate           // 生成报告
GET    /api/reports/:id                // 获取报告

POST   /api/workflows                  // 创建工作流
GET    /api/workflows                  // 获取工作流列表
POST   /api/workflows/:id/execute      // 执行工作流
GET    /api/workflows/executions/:id   // 获取执行状态

POST   /api/strategies                 // 上传策略
GET    /api/strategies                 // 获取策略列表
GET    /api/strategies/:id             // 获取策略详情
POST   /api/strategies/:id/rate        // 评分策略

GET    /api/notifications              // 获取通知列表
POST   /api/notifications/:id/read     // 标记已读
```

### 请求/响应格式

```json
{
  "success": true,
  "data": {
    "report_id": "xxx",
    "download_url": "xxx"
  },
  "message": "报告生成成功"
}
```

***

## 8. 影响范围

### 影响的模块/包

- **后端新增**: `services/report/`, `services/workflow/`, `services/strategy/`, `services/notification/`
- **后端修改**: `models/` 新增数据模型，`handlers/` 新增接口，`router/` 新增路由
- **前端新增**: `pages/ReportBuilder/`, `pages/WorkflowEditor/`, `pages/StrategyMarket/`, `components/NotificationCenter/`
- **前端修改**: `App.tsx` 新增路由，`services/api.ts` 新增 API 封装

### 影响的接口

- 新增约 15 个 API 端点
- 无现有接口修改

### 影响的配置

- 新增通知服务配置
- 新增报告导出配置
- 新增策略上传限制配置

### 影响的部署

- 需要通知推送服务（WebSocket/SSE）
- 需要文件存储服务（报告文件）
- 需要更大的数据库空间

***

## 9. 实施路线图

### Phase 1: 报告生成系统（0-3个月）

- [ ] 报告模板系统设计
- [ ] PDF/Excel/HTML 导出实现
- [ ] 前端报告构建器 UI
- [ ] 预设报告模板（组合分析、风险报告、ETF对比）

### Phase 2: 自定义分析流程（3-6个月）

- [ ] 工作流引擎实现
- [ ] 可视化工作流编辑器
- [ ] 预设分析流程模板
- [ ] 工作流分享功能

### Phase 3: 插件系统与生态（6-12个月）

- [ ] 策略仓库实现
- [ ] 策略上传/下载/评分
- [ ] 插件架构基础
- [ ] 社区建设

### Phase 4: AI 增强与智能分析（12-18个月）

- [ ] AI 报告生成
- [ ] 智能投顾建议
- [ ] 异常检测与预警
- [ ] 自然语言查询

***

## 10. 风险评估

| 风险类型   | 概率   | 影响   | 缓解措施          |
| ------ | ---- | ---- | ------------- |
| 技术复杂度高 | 🟡 中 | 🔴 高 | 分阶段实施，每阶段验证   |
| 性能问题   | 🟡 中 | 🟡 中 | 性能基准测试，优化关键路径 |
| 用户接受度  | 🟡 中 | 🟡 中 | 持续收集用户反馈，迭代优化 |
| 维护成本上升 | 🟡 中 | 🟡 中 | 良好的架构设计，文档完善  |
| 安全风险   | 🟢 低 | 🔴 高 | 严格的权限控制，输入验证  |

***

## 11. 开放问题

- [ ] 是否需要支持自定义指标？如何设计？
- [ ] 策略代码的安全性如何保证？（沙箱运行？）
- [ ] AI 增强部分使用什么技术栈？（OpenAI API？自建模型？）
- [ ] 通知推送选择什么技术？（WebSocket？SSE？Push？）
- [ ] 报告文件存储方案？（本地？云存储？）

***

**文档状态**: draft
**下一步**: 创建详细的 PLAN.md，开始 Phase 1 实施

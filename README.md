# ETF-Insight (v2.5.0) 🚀

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.21-blue)](https://golang.org/)
[![React Version](https://img.shields.io/badge/React-19.2.4-61DAFB)](https://reactjs.org/)
[![Test Coverage](https://img.shields.io/badge/coverage-55%25-yellowgreen)](https://github.com/coder100001/ETF-Insight)

**开源专业的 ETF 量化分析平台**

ETF-Insight 是一个面向专业投资者、量化研究员和金融机构的开源 ETF 分析平台。基于 Go + React 技术栈，提供机构级的 ETF 数据洞察、多维度量化分析、投资组合优化等专业功能。

> 🎯 **愿景**: 成为开源社区最专业的 ETF 量化分析工具，为投资者提供透明、可验证的分析能力

## 📢 最新动态

**v2.5 量化分析与测试覆盖升级**:
- ✅ **技术指标库**: RSI、MACD、布林带、移动平均线
- ✅ **风险模型**: VaR/CVaR (历史法/参数法)、组合风险分析
- ✅ **风险指标**: 夏普比率、索提诺比率、最大回撤、Beta/Alpha
- ✅ **前端分析页面**: 技术分析(雷达图)、风险分析(VaR可视化)
- ✅ **测试覆盖率**: middleware 68.8%, utils 81.2%
- ✅ **CI/CD**: 覆盖率检测、Codecov集成

**v2.4 安全与API文档升级**:
- ✅ **JWT身份认证**: 完整的认证中间件，支持Token生成/验证/角色控制
- ✅ **审计日志**: 异步写入，敏感信息自动脱敏，Request ID追踪
- ✅ **数据验证**: 通用验证中间件，支持string/number/email多种类型
- ✅ **API分页**: 通用分页响应结构，支持page/pageSize参数
- ✅ **速率限制**: IP级别的请求频率限制
- ✅ **股票代码验证**: 防止非法字符注入
- ✅ **Swagger API文档**: OpenAPI 3.0规范，交互式API测试

> **📚 开发者必读**: [agents.md](./agents.md) - 架构设计、数据模型、编码规则等核心文档
> **📖 API文档**: http://localhost:8080/swagger - 交互式API文档
> **📊 实现文档**: [docs/development/v2.5_phase1_implementation.md](./docs/development/v2.5_phase1_implementation.md) - v2.5详细实现文档

## 🎯 开源定位

ETF-Insight 坚持**开源、专业、透明**的理念：

- 🔓 **完全开源**: MIT 协议，代码透明可审计
- 📊 **专业分析**: 机构级量化指标，支持学术研究
- 🔧 **可扩展**: 插件化架构，支持自定义数据源和算法
- 🏛️ **社区驱动**: 欢迎贡献代码、策略和数据源

### 适用场景

| 用户类型 | 应用场景 |
|---------|---------|
| **量化研究员** | 策略回测、因子分析、学术验证 |
| **专业投资者** | 组合优化、风险管理、资产配置 |
| **金融机构** | 内部研究平台、客户报告生成 |
| **开发者** | 学习量化金融、构建自定义分析工具 |
| **数据科学家** | ETF 数据分析、机器学习特征工程 |

## ✨ 核心特性

### 📈 量化技术分析 (v2.5 新增)
- **技术指标库** - RSI、MACD、布林带、移动平均线
- **多因子雷达图** - 多维度技术指标可视化对比
- **趋势分析** - MACD趋势图、布林带位置分析
- **超买超卖提示** - RSI阈值预警

### 🛡️ 风险分析 (v2.5 新增)
- **VaR/CVaR计算** - 历史模拟法和参数法风险价值
- **组合风险分解** - 成分VaR、边际VaR分析
- **风险调整收益** - 夏普比率、索提诺比率、卡尔玛比率
- **市场风险指标** - Beta、Alpha、最大回撤
- **风险等级评估** - 保守/平衡/激进组合风险评级

### 📊 ETF 对比分析（ETF Comparison）
- **并排对比** - 最多支持 5 只 ETF 同时对比
- **多维度指标** - 费率、AUM、股息率、业绩表现、风险指标
- **智能股息率** - 根据 ETF 类型自动设置合理股息率（高股息 3.5%、覆盖收益型 7%、债券 4%）
- **持仓重叠分析** - 识别 ETF 间的持仓重合度，避免过度集中
- **业绩回测对比** - 不同时间周期的收益表现对比

### 🔍 持仓深度解构（Holdings Analysis）
- **前十大持仓** - 穿透底层资产，了解核心持仓
- **行业分布** - sector 权重分布及变化趋势
- **地区分布** - 国家/地区配置比例
- **市值分布** - 大/中/小盘股配置比例
- **风格分析** - 价值/成长风格暴露度

### 💼 A股红利ETF投资组合
- **A股ETF管理** - 支持中证红利、红利低波等主流红利ETF
- **投资占比分布** - 饼状图可视化展示投资组合配置
- **分红数据追踪** - 股息率、分红频率等关键指标

### 💱 汇率数据管理
- **实时汇率** - USD/CNY、USD/HKD 等主要货币对
- **多数据源支持** - Open Exchange Rates、CurrencyAPI、Frankfurter 三数据源
- **自动故障转移** - 主数据源不可用时自动切换到备用数据源
- **健康监控** - 实时监控数据源可用性
- **自动同步** - 定时任务自动更新汇率数据（每5分钟）
- **货币转换** - 支持多种货币间的换算功能
- **同步日志** - 完整的汇率同步批次记录与明细追踪

### ⚙️ ETF 配置管理
- **CRUD 操作** - 增删改查 ETF 配置信息
- **状态管理** - 启用/禁用 ETF 数据自动更新
- **数据源配置** - **Finage 唯一真实数据源** (v2.0 架构)

### 📈 投资组合配置
- **组合构建** - 自定义投资组合及权重分配
- **收益分析** - 基于历史数据的组合收益模拟
- **资本利得计算** - 基于真实历史数据计算资本利得和收益率
- **预设组合** - 内置多种投资策略组合模板

## 🛠️ 技术栈

### 后端 (Go)
| 技术 | 版本 | 用途 |
|------|------|------|
| Go | >= 1.21 | 核心语言 |
| Gin | v1.12.0 | Web 框架 |
| GORM | v1.30.0 | ORM 框架 (SQLite/PostgreSQL) |
| go-cache | v2.1.0 | 内存缓存 |
| cron/v3 | v3.0.1 | 定时任务调度 |
| JWT/v5 | v5.3.1 | 身份认证 |
| uuid | v1.6.0 | 唯一标识生成 |

### 前端 (React)
| 技术 | 版本 | 用途 |
|------|------|------|
| React | ^19.2.4 | UI 框架 |
| TypeScript | ^5.x | 类型安全 |
| Vite | latest | 构建工具 |
| Ant Design | ^6.3.4 | UI 组件库 |
| ECharts | ^6.0.0 | 数据可视化 |
| Recharts | ^3.8.1 | 图表组件 |
| React Router | ^7.13.2 | 路由管理 |

### 数据存储
- **SQLite** - 默认本地数据库（开发环境）
- **PostgreSQL** - 生产数据库支持

## 🚀 快速开始

### 方式一：一键启动（推荐）

```bash
# 克隆项目
git clone <repository-url>
cd py_project

# 一键启动（macOS/Linux）
./start.sh

# 一键启动（Windows）
start.bat
```

### 方式二：手动启动

#### 1. 后端服务启动
```bash
cd backend

# 安装依赖
go mod tidy

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件，配置 Finage API Key

# 启动后端服务
go run main.go
```

#### 2. 前端服务启动
```bash
cd frontend

# 安装依赖
npm install

# 启动前端服务
npm run dev
```

### 环境要求
- **Go**: >= 1.21
- **Node.js**: >= 18.0.0
- **npm**: >= 9.0.0
- **SQLite**: 3.35.0+

## 📊 数据源配置

### ETF 数据源
- **主数据源**: Finage API (必须配置)
- **环境变量**: `FINAGE_API_KEY=your_api_key_here`

### 汇率数据源
- **主数据源**: Open Exchange Rates
- **备用数据源**: CurrencyAPI、Frankfurter
- **故障转移**: 自动切换，无需手动配置

## 🔧 开发指南

### 代码规范
- **Go**: 遵循官方代码规范，使用 `gofmt` 格式化
- **TypeScript**: 严格类型检查，禁用 `any` 类型
- **React**: 函数式组件，Hooks 规范使用
- **Pre-commit钩子**: 提交前自动执行TypeScript和ESLint检查

### 项目结构
```
ETF-Insight/
├── backend/          # Go 后端服务
├── frontend/         # React 前端应用
├── agents.md         # 项目核心上下文文档
├── README.md         # 中文文档
└── README_EN.md      # 英文文档
```

### 核心文档
- **[agents.md](./agents.md)** - 项目架构、数据模型、开发规范
- **[docs/security/](./docs/security/)** - 安全文档和改进指南
- **[docs/reviews/](./docs/reviews/)** - 代码审查报告
- **[docs/roadmap/](./docs/roadmap/)** - 演进路线图
- **[docs/guides/](./docs/guides/)** - 使用指南
- **API 文档**: http://localhost:8080/swagger - OpenAPI 3.0 交互式文档

## 🎯 演进路线图

### v2.4 (当前版本) ✅
- ✅ 安全功能全面升级（JWT、审计、限流）
- ✅ Swagger/OpenAPI 3.0 API文档
- ✅ 代码质量优化

### v2.5 (进行中) 🔄
- 🔄 测试覆盖率提升至80%
- 🔄 技术指标库（RSI/MACD/布林带）
- 🔄 风险模型（VaR/CVaR）

### v2.6-2.8 (3-6个月) 📋
- 📋 回测引擎开发
- 📋 策略框架实现
- 📋 性能监控集成

### v3.0 (6-12个月) 🚀
- 🚀 插件系统架构
- 🚀 开源生态建设
- 🚀 学术合作支持

> 详细路线图: [docs/roadmap/PROFESSIONAL_ENHANCEMENT.md](./docs/roadmap/PROFESSIONAL_ENHANCEMENT.md)

## 📞 技术支持

### 常见问题
1. **数据源连接失败**: 检查网络连接和 API Key 配置
2. **汇率数据不一致**: 系统会自动故障转移，检查日志确认当前数据源
3. **性能问题**: 启用 Redis 缓存提升性能

### 日志查看
```bash
# 查看后端日志
tail -f backend/logs/app.log

# 查看汇率同步日志
tail -f backend/logs/exchange_rate.log
```

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！请确保：
1. 遵循项目代码规范
2. 新功能包含单元测试
3. 更新相关文档
4. 通过代码审查

## 📄 许可证

本项目采用 MIT 许可证。

---

**立即体验**: [http://localhost:8080](http://localhost:8080)
**API 文档**: [http://localhost:8080/swagger](http://localhost:8080/swagger)

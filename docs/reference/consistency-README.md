# 文档-代码一致性自动化机制

## 概述

本机制确保项目文档（AGENTS.md、README.md等）与代码实现保持强一致性，通过Go编写的自动化工具实现文档的实时同步和验证。

## 核心组件

### 1. 一致性检查工具 (`tools/doccheck/`)
Go编写的原生检查工具，与项目技术栈一致：
- 自动扫描Go后端和TypeScript前端代码
- 解析Markdown文档结构和内容
- 智能匹配代码元素与文档描述（支持驼峰拆分关键词匹配）
- 检测版本号一致性
- 生成差异报告

### 2. 映射规则 (`mapping_rules.json`)
文档与代码的关联映射定义，包括：
- 后端模块（handlers/services/models/middleware）→ 文档章节
- 前端模块（components/pages/services/types）→ 文档章节
- 功能特性 ↔ 代码指示器映射

### 3. 评审机制 (`review_process.md`)
- 多角色评审流程
- 技术准确性和内容质量检查标准
- 评审结果处理和跟踪

### 4. 自动化集成
- Git预提交钩子（`scripts/pre_commit_hook.sh`）
- GitHub Actions CI/CD（`.github/workflows/docs-consistency.yml`）
- Makefile命令（`make doccheck`）

## 快速开始

```bash
# 方式1: 使用Makefile（推荐）
make doccheck

# 方式2: 直接运行
./tools/doccheck/doccheck --project-root .

# 生成报告文件
./tools/doccheck/doccheck --project-root . --output report.md

# JSON格式输出
./tools/doccheck/doccheck --project-root . --format json

# 严格模式（高严重性问题返回非零退出码）
./tools/doccheck/doccheck --project-root . --strict

# 安装Git预提交钩子
make doccheck-hook
```

## 检查项目

| 检查项 | 说明 | 严重程度 |
|--------|------|----------|
| **未文档化的Handler/Service** | 核心处理器和服务未在文档中提及 | 中 |
| **版本号不一致** | README版本与package.json版本不匹配 | 高 |
| **未实现的功能** | 文档中描述的功能代码中不存在 | 高 |
| **必需文档缺失** | 必需的文档文件不存在 | 高 |

## 智能匹配策略

工具采用关键词匹配而非精确类名匹配：
- `ETFHandler` → 检查文档中是否包含 "ETF" 或 "ETFHandler"
- `PortfolioOptimizerService` → 检查 "PortfolioOptimizer"、"Portfolio"、"Optimizer"
- `AShareDataHandler` → 检查 "AShareData"、"AShare"、"Data"

## 一致性评分

评分基于文档覆盖率计算：
- 覆盖率 = 已文档化的核心元素数 / 核心元素总数
- 核心元素 = Handler结构体 + Service结构体 + 前端组件
- 评分 = (覆盖率 - 问题惩罚) × 100

## 支持的文档类型

- `AGENTS.md` / `agents.md` - 核心架构文档
- `README.md` - 项目总览文档
- `README_EN.md` - 英文项目文档
- `backend/README.md` - 后端文档
- `frontend/README.md` - 前端文档
- 所有以`readme`开头的文件

## 项目结构

```
tools/doccheck/
├── main.go              # 入口，命令行参数解析
├── models/models.go     # 数据模型定义
├── scanner/scanner.go   # Go/TS代码扫描器
├── parser/parser.go     # Markdown文档解析器
├── checker/checker.go   # 一致性检查引擎
└── reporter/reporter.go # 报告生成器
```

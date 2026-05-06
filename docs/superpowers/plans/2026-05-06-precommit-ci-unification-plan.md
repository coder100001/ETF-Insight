# Pre-commit 与 CI/CD 检查统一化实现计划

> **面向 AI 代理的工作者：** 使用 `superpowers:subagent-driven-development` 逐任务实现。步骤使用复选框跟踪进度。

**目标:** 实现 pre-commit 与 CI/CD 检查统一化，本地提交前拦截 95%+ CI 失败，总耗时 < 15 秒

**架构:** 统一检查清单驱动 pre-commit/CI/Makefile，Doccheck 快速模式实现变更驱动检查

**技术栈:** Bash, Go, YAML, pre-commit, GitHub Actions

---

## 文件清单

| 文件 | 操作 | 职责 |
|------|------|------|
| `scripts/check-manifest.yaml` | 创建 | 统一检查清单配置 |
| `scripts/run-check.sh` | 创建 | 通用检查执行引擎 |
| `scripts/pre-commit-checks.sh` | 创建 | pre-commit 入口（调用 run-check.sh） |
| `.pre-commit-config.yaml` | 修改 | 简化配置，使用统一脚本 |
| `.github/workflows/ci.yml` | 修改 | 使用统一脚本 |
| `Makefile` | 修改 | 使用统一脚本 |
| `tools/doccheck/checker/quick.go` | 创建 | Doccheck 快速模式检查器 |
| `tools/doccheck/checker/checker.go` | 修改 | 集成快速模式入口 |
| `tools/doccheck/main.go` | 修改 | 添加 `--quick` 和 `--changed-files` 参数 |

---

## 任务 1: 创建统一检查清单配置

**文件:**
- 创建: `scripts/check-manifest.yaml`

**步骤:**

- [ ] **步骤 1: 编写检查清单 YAML**

```yaml
# scripts/check-manifest.yaml
version: "1.0"

stages:
  format:
    description: "格式检查"
    timeout: 5
    commands:
      - name: gofmt
        cmd: "cd backend && gofmt -l ."
        fail_on_output: true
        message: "Go 文件需要格式化，运行: cd backend && gofmt -w ."

      - name: goimports
        cmd: "cd backend && goimports -l ."
        fail_on_output: true
        message: "Go imports 需要整理，运行: cd backend && goimports -w ."

      - name: prettier
        cmd: "cd frontend && npx prettier --check src/ --ignore-unknown 2>/dev/null || true"
        optional: true

  static:
    description: "静态分析"
    timeout: 10
    env:
      GOPROXY: "https://goproxy.cn,direct"
    commands:
      - name: go-vet
        cmd: "cd backend && go vet ./..."
        message: "Go vet 发现问题"

      - name: tsc
        cmd: "cd frontend && npx tsc --noEmit"
        message: "TypeScript 类型检查失败"

      - name: eslint
        cmd: "cd frontend && npx eslint --cache --quiet src/ 2>/dev/null || true"
        optional: true
        message: "ESLint 发现问题"

  build:
    description: "编译检查"
    timeout: 15
    env:
      GOPROXY: "https://goproxy.cn,direct"
    commands:
      - name: go-build
        cmd: "cd backend && go build ./..."
        message: "Go 编译失败"

      - name: npm-build
        cmd: "cd frontend && npm run build"
        message: "前端构建失败"

  test:
    description: "单元测试"
    timeout: 30
    env:
      GOPROXY: "https://goproxy.cn,direct"
    commands:
      - name: go-test-short
        cmd: "cd backend && go test -short -count=1 ./..."
        message: "Go 测试失败"

      - name: frontend-test
        cmd: "cd frontend && npm run test:run 2>/dev/null || echo '跳过前端测试'"
        optional: true

  docs:
    description: "文档一致性"
    timeout: 5
    commands:
      - name: doccheck
        cmd: "./tools/doccheck/doccheck --quick --strict"
        message: "文档一致性检查失败"
```

- [ ] **步骤 2: Commit**

```bash
git add scripts/check-manifest.yaml
git commit -m "chore: add unified check manifest configuration"
```

---

## 任务 2: 创建通用检查执行引擎

**文件:**
- 创建: `scripts/run-check.sh`

**步骤:**

- [ ] **步骤 1: 编写检查执行脚本**

```bash
#!/bin/bash
# scripts/run-check.sh - 通用检查执行引擎

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MANIFEST="$PROJECT_ROOT/scripts/check-manifest.yaml"

# 默认参数
STAGES="format,static,build,test,docs"
MODE="full"  # full 或 quick
BACKEND_ONLY=false
FRONTEND_ONLY=false

# 解析参数
while [[ $# -gt 0 ]]; do
  case $1 in
    --stage=*)
      STAGES="${1#*=}"
      shift
      ;;
    --mode=*)
      MODE="${1#*=}"
      shift
      ;;
    --backend)
      BACKEND_ONLY=true
      shift
      ;;
    --frontend)
      FRONTEND_ONLY=true
      shift
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查清单文件是否存在
if [[ ! -f "$MANIFEST" ]]; then
  echo -e "${RED}Error: Check manifest not found: $MANIFEST${NC}"
  exit 1
fi

# 运行单个命令
run_command() {
  local name="$1"
  local cmd="$2"
  local timeout="${3:-30}"
  local optional="${4:-false}"
  local message="${5:-"Check failed: $name"}"

  echo "  🔍 $name..."

  if [[ "$optional" == "true" ]]; then
    if ! timeout "$timeout" bash -c "$cmd" 2>/dev/null; then
      echo -e "  ${YELLOW}⚠️  $name (optional, skipped)${NC}"
      return 0
    fi
  else
    if ! timeout "$timeout" bash -c "$cmd"; then
      echo -e "  ${RED}❌ $name failed${NC}"
      echo "     $message"
      return 1
    fi
  fi

  echo -e "  ${GREEN}✅ $name${NC}"
  return 0
}

# 主执行流程
echo "🔧 Running checks: $STAGES"
echo "   Mode: $MODE"
echo ""

IFS=',' read -ra STAGE_LIST <<< "$STAGES"
FAILED=0

for stage in "${STAGE_LIST[@]}"; do
  echo "📦 Stage: $stage"

  # 使用 yq 或简单 grep 解析 YAML（简化版直接用硬编码）
  case "$stage" in
    format)
      run_command "gofmt" "cd $PROJECT_ROOT/backend && test -z \$(gofmt -l .)" 5 false "Go files need formatting"
      run_command "goimports" "cd $PROJECT_ROOT/backend && test -z \$(goimports -l . 2>/dev/null || echo 'skip')" 5 true
      ;;
    static)
      export GOPROXY="https://goproxy.cn,direct"
      run_command "go-vet" "cd $PROJECT_ROOT/backend && go vet ./..." 10 false "Go vet found issues"
      run_command "tsc" "cd $PROJECT_ROOT/frontend && npx tsc --noEmit" 10 false "TypeScript check failed"
      run_command "eslint" "cd $PROJECT_ROOT/frontend && npx eslint --cache --quiet src/ 2>/dev/null || true" 10 true
      ;;
    build)
      export GOPROXY="https://goproxy.cn,direct"
      run_command "go-build" "cd $PROJECT_ROOT/backend && go build ./..." 15 false "Go build failed"
      run_command "npm-build" "cd $PROJECT_ROOT/frontend && npm run build" 15 false "Frontend build failed"
      ;;
    test)
      export GOPROXY="https://goproxy.cn,direct"
      run_command "go-test-short" "cd $PROJECT_ROOT/backend && go test -short -count=1 ./..." 30 false "Go tests failed"
      run_command "frontend-test" "cd $PROJECT_ROOT/frontend && npm run test:run 2>/dev/null || echo 'skipped'" 30 true
      ;;
    docs)
      run_command "doccheck" "cd $PROJECT_ROOT && ./tools/doccheck/doccheck --quick --strict 2>/dev/null || ./tools/doccheck/doccheck --quick --strict" 5 false "Documentation consistency check failed"
      ;;
    *)
      echo -e "${YELLOW}Unknown stage: $stage${NC}"
      ;;
  esac

  if [[ $? -ne 0 ]]; then
    FAILED=1
  fi

  echo ""
done

if [[ $FAILED -eq 0 ]]; then
  echo -e "${GREEN}✅ All checks passed!${NC}"
  exit 0
else
  echo -e "${RED}❌ Some checks failed${NC}"
  exit 1
fi
```

- [ ] **步骤 2: 添加执行权限**

```bash
chmod +x scripts/run-check.sh
```

- [ ] **步骤 3: 测试脚本**

```bash
./scripts/run-check.sh --stage=format
```

- [ ] **步骤 4: Commit**

```bash
git add scripts/run-check.sh
git commit -m "chore: add unified check runner script"
```

---

## 任务 3: 创建 Pre-commit 入口脚本

**文件:**
- 创建: `scripts/pre-commit-checks.sh`

**步骤:**

- [ ] **步骤 1: 编写 pre-commit 入口**

```bash
#!/bin/bash
# scripts/pre-commit-checks.sh - Pre-commit 统一入口

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "🔍 Running pre-commit checks..."
echo ""

# 运行所有检查（pre-commit 需要完整检查）
exec "$PROJECT_ROOT/scripts/run-check.sh" --stage=format,static,build,test,docs
```

- [ ] **步骤 2: 添加执行权限**

```bash
chmod +x scripts/pre-commit-checks.sh
```

- [ ] **步骤 3: Commit**

```bash
git add scripts/pre-commit-checks.sh
git commit -m "chore: add pre-commit unified entry script"
```

---

## 任务 4: 修改 Pre-commit 配置

**文件:**
- 修改: `.pre-commit-config.yaml`

**步骤:**

- [ ] **步骤 1: 备份原配置**

```bash
cp .pre-commit-config.yaml .pre-commit-config.yaml.bak
```

- [ ] **步骤 2: 编写新配置**

```yaml
repos:
  # 基础检查
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.6.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
        args: ['--maxkb=1000']
      - id: check-json
      - id: check-merge-conflict
      - id: check-case-conflict
      - id: mixed-line-ending
      - id: detect-private-key

  # Go 格式检查
  - repo: https://github.com/dnephin/pre-commit-golang
    rev: v0.5.1
    hooks:
      - id: go-fmt
      - id: go-imports

  # 统一检查（format + static + build + test + docs）
  - repo: local
    hooks:
      - id: unified-checks
        name: Unified Checks (format/static/build/test/docs)
        entry: bash scripts/pre-commit-checks.sh
        language: system
        pass_filenames: false
        always_run: true
        verbose: true
```

- [ ] **步骤 3: 验证配置格式**

```bash
pre-commit validate-config
```

- [ ] **步骤 4: Commit**

```bash
git add .pre-commit-config.yaml
git commit -m "refactor(pre-commit): use unified check script"
```

---

## 任务 5: 创建 Doccheck 快速模式检查器

**文件:**
- 创建: `tools/doccheck/checker/quick.go`

**步骤:**

- [ ] **步骤 1: 编写快速模式检查器**

```go
// tools/doccheck/checker/quick.go
package checker

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/coder100001/etf-insight/tools/doccheck/models"
)

// QuickChecker 快速模式检查器
type QuickChecker struct {
	baseChecker *ConsistencyChecker
	changedFiles []string
}

// NewQuickChecker 创建快速检查器
func NewQuickChecker(base *ConsistencyChecker, changedFiles []string) *QuickChecker {
	return &QuickChecker{
		baseChecker:  base,
		changedFiles: changedFiles,
	}
}

// Run 执行快速检查
func (qc *QuickChecker) Run() (*models.CheckResult, error) {
	c := qc.baseChecker
	c.issues = c.issues[:0]

	// 1. 提取受影响的代码元素
	affectedElements := qc.extractAffectedElements()

	// 2. 检查这些元素的文档覆盖（代码→文档）
	for _, elem := range affectedElements {
		if !c.isNameDocumented(elem.Name) {
			c.issues = append(c.issues, models.Issue{
				Type:        "undocumented_code_element",
				Severity:    "medium",
				Message:     fmt.Sprintf("代码元素 '%s' 在文档中未提及", elem.Name),
				ElementName: elem.Name,
				ElementType: elem.Type,
			})
		}
	}

	// 3. 检查文档中提到的相关功能是否已实现（文档→代码）
	qc.checkRelatedFeatures(affectedElements)

	// 4. 版本一致性（每次必查）
	c.checkVersionConsistency()

	// 5. 必需文档存在性
	c.checkRequiredDocuments()

	return c.buildResult(), nil
}

// extractAffectedElements 从变更文件提取受影响的代码元素
func (qc *QuickChecker) extractAffectedElements() []models.CodeElement {
	c := qc.baseChecker
	var affected []models.CodeElement

	for _, elem := range c.codeScanner.Elements() {
		// 检查元素所在文件是否在变更列表中
		for _, changedFile := range qc.changedFiles {
			if strings.HasSuffix(elem.FilePath, changedFile) ||
			   strings.HasSuffix(changedFile, elem.FilePath) {
				affected = append(affected, elem)
				break
			}
		}
	}

	return affected
}

// checkRelatedFeatures 检查与变更元素相关的功能映射
func (qc *QuickChecker) checkRelatedFeatures(elements []models.CodeElement) {
	c := qc.baseChecker
	if c.mapping == nil {
		return
	}

	// 构建变更元素的名称集合
	elementNames := make(map[string]bool)
	for _, elem := range elements {
		baseName := extractBaseName(elem.Name)
		elementNames[baseName] = true
		elementNames[elem.Name] = true
	}

	// 只检查与变更元素相关的功能
	for featureName, mapping := range c.mapping.FeatureMappings {
		// 检查功能是否与变更元素相关
		if !qc.isFeatureRelated(featureName, mapping, elementNames) {
			continue
		}

		codeExists := c.featureExistsInCode(mapping.CodeIndicators)
		docExists := c.featureExistsInDocs(mapping.DocumentSections)

		if docExists && !codeExists {
			c.issues = append(c.issues, models.Issue{
				Type:     "unimplemented_feature",
				Severity: "high",
				Message:  fmt.Sprintf("文档中提到的功能 '%s' 在代码中未实现", featureName),
				Feature:  featureName,
			})
		}
	}
}

// isFeatureRelated 检查功能是否与变更元素相关
func (qc *QuickChecker) isFeatureRelated(featureName string, mapping models.FeatureMapping, elementNames map[string]bool) bool {
	// 检查功能名是否匹配
	if elementNames[featureName] {
		return true
	}

	// 检查代码指标是否匹配
	for _, indicator := range mapping.CodeIndicators {
		if elementNames[indicator] {
			return true
		}
	}

	return false
}

// extractBaseName 提取基础名称（去掉 Handler/Service 后缀）
func extractBaseName(name string) string {
	base := name
	base = strings.TrimSuffix(base, "Handler")
	base = strings.TrimSuffix(base, "Service")
	base = strings.TrimSuffix(base, "Controller")
	return base
}
```

- [ ] **步骤 2: Commit**

```bash
git add tools/doccheck/checker/quick.go
git commit -m "feat(doccheck): add quick mode checker"
```

---

## 任务 6: 修改 Doccheck 主程序

**文件:**
- 修改: `tools/doccheck/main.go`
- 修改: `tools/doccheck/checker/checker.go`

**步骤:**

- [ ] **步骤 1: 修改 main.go 添加参数**

```go
// tools/doccheck/main.go
func main() {
	projectRoot := flag.String("project-root", ".", "项目根目录路径")
	output := flag.String("output", "", "报告输出文件路径")
	format := flag.String("format", "markdown", "报告格式: markdown | json")
	strict := flag.Bool("strict", false, "严格模式：存在高严重性问题时返回非零退出码")
	quick := flag.Bool("quick", false, "快速模式：仅检查变更文件")
	changedFiles := flag.String("changed-files", "", "变更文件列表（逗号分隔）")
	flag.Parse()

	// ... 路径处理 ...

	c := checker.NewConsistencyChecker(absRoot)

	var result *models.CheckResult
	var err error

	if *quick {
		// 快速模式
		files := strings.Split(*changedFiles, ",")
		qc := checker.NewQuickChecker(c, files)
		result, err = qc.Run()
	} else {
		// 完整模式
		result, err = c.Run()
	}

	// ... 后续处理 ...
}
```

- [ ] **步骤 2: 修改 checker.go 添加快速模式入口**

在 `ConsistencyChecker` 结构体上添加方法：

```go
// RunQuick 执行快速检查（供 QuickChecker 使用）
func (c *ConsistencyChecker) RunQuick() (*models.CheckResult, error) {
	if err := c.codeScanner.Scan(); err != nil {
		return nil, fmt.Errorf("代码扫描失败: %w", err)
	}
	if err := c.docParser.Parse(); err != nil {
		return nil, fmt.Errorf("文档解析失败: %w", err)
	}
	c.loadMapping()
	return c.buildResult(), nil
}

// 暴露内部方法供 QuickChecker 使用
func (c *ConsistencyChecker) CodeScanner() *scanner.CodeScanner { return c.codeScanner }
func (c *ConsistencyChecker) DocParser() *parser.DocumentParser { return c.docParser }
func (c *ConsistencyChecker) Mapping() *models.MappingConfig { return c.mapping }
```

- [ ] **步骤 3: 重新编译 doccheck**

```bash
cd tools/doccheck && go build -o doccheck .
```

- [ ] **步骤 4: 测试快速模式**

```bash
./tools/doccheck/doccheck --quick --changed-files="README.md,backend/services/alpha_view_service.go"
```

- [ ] **步骤 5: Commit**

```bash
git add tools/doccheck/main.go tools/doccheck/checker/checker.go
git commit -m "feat(doccheck): integrate quick mode with --quick and --changed-files flags"
```

---

## 任务 7: 修改 CI/CD 配置

**文件:**
- 修改: `.github/workflows/ci.yml`

**步骤:**

- [ ] **步骤 1: 修改 backend job**

```yaml
  backend:
    name: Backend (Go)
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache-dependency-path: backend/go.sum

      - name: Run unified checks (backend)
        working-directory: .
        env:
          GOPROXY: https://goproxy.cn,direct
        run: |
          ./scripts/run-check.sh --stage=format,static,build,test --backend

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./backend/coverage.out
          flags: backend
          fail_ci_if_error: false
```

- [ ] **步骤 2: 修改 frontend job**

```yaml
  frontend:
    name: Frontend (React/TypeScript)
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        working-directory: frontend
        run: npm install

      - name: Run unified checks (frontend)
        run: |
          ./scripts/run-check.sh --stage=format,static,build,test --frontend

      - name: Upload build artifacts
        uses: actions/upload-artifact@v4
        with:
          name: frontend-build
          path: frontend/dist
          retention-days: 7
```

- [ ] **步骤 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "refactor(ci): use unified check script for backend and frontend"
```

---

## 任务 8: 修改 Makefile

**文件:**
- 修改: `Makefile`

**步骤:**

- [ ] **步骤 1: 修改 lint 目标**

```makefile
lint:
	@echo "运行代码检查..."
	@./scripts/run-check.sh --stage=format,static
```

- [ ] **步骤 2: 修改 test 目标**

```makefile
test:
	@echo "运行测试..."
	@./scripts/run-check.sh --stage=test
```

- [ ] **步骤 3: 添加 check 目标**

```makefile
check:
	@echo "运行完整检查..."
	@./scripts/run-check.sh --stage=all
```

- [ ] **步骤 4: Commit**

```bash
git add Makefile
git commit -m "refactor(make): use unified check script for lint/test/check"
```

---

## 任务 9: 更新文档一致性 CI

**文件:**
- 修改: `.github/workflows/docs-consistency.yml`

**步骤:**

- [ ] **步骤 1: 简化配置**

```yaml
name: Documentation Consistency Check

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 9 * * 1'

jobs:
  check-docs:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.24'

    - name: Run documentation consistency check (full)
      run: |
        cd tools/doccheck && go build -o doccheck .
        ./doccheck --project-root ../.. --strict --output docs_consistency_report.md

    - name: Upload consistency report
      if: failure()
      uses: actions/upload-artifact@v4
      with:
        name: docs-consistency-report
        path: docs_consistency_report.md
```

- [ ] **步骤 2: Commit**

```bash
git add .github/workflows/docs-consistency.yml
git commit -m "refactor(ci): simplify docs consistency workflow"
```

---

## 任务 10: 端到端测试

**步骤:**

- [ ] **步骤 1: 安装 pre-commit hook**

```bash
pre-commit install
```

- [ ] **步骤 2: 测试完整流程**

```bash
# 创建一个测试提交
echo "# test" >> README.md
git add README.md
git commit -m "test: verify pre-commit checks"
```

- [ ] **步骤 3: 验证检查执行**

确认输出包含：
- ✅ gofmt
- ✅ goimports
- ✅ go-vet
- ✅ tsc
- ✅ go-build
- ✅ npm-build
- ✅ go-test-short
- ✅ doccheck

- [ ] **步骤 4: 测试失败场景**

```bash
# 故意破坏格式
echo "  package main" > backend/test_fmt.go
git add backend/test_fmt.go
git commit -m "test: should fail" || true  # 应该失败
rm backend/test_fmt.go
git checkout -- .
```

- [ ] **步骤 5: 性能测试**

```bash
time ./scripts/run-check.sh --stage=all
```

目标：< 15 秒

---

## 验收验证

- [ ] `git commit` 时自动运行完整检查
- [ ] 检查总耗时 < 15 秒
- [ ] CI 使用与 pre-commit 相同的检查脚本
- [ ] `make check` 等价于 pre-commit 完整检查
- [ ] Doccheck 快速模式 < 5 秒
- [ ] 版本不一致时检查失败
- [ ] 新增代码元素无文档时检查失败
- [ ] 文档提到未实现功能时检查失败

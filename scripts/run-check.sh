#!/bin/bash
# ETF-Insight 统一检查执行引擎
# 驱动 pre-commit、CI/CD 和 Makefile
# v2.0: 按需加载，前后端解耦

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# 补全 PATH: pre-commit 环境可能缺少开发工具路径
NVM_NODE_BIN="$HOME/.nvm/versions/node/v24.11.0/bin"
export PATH="$HOME/go1.26/go/bin:$NVM_NODE_BIN:/usr/local/go/bin:/usr/local/bin:/opt/homebrew/bin:$HOME/go/bin:$PATH"

# 默认参数
STAGES="format,static,build,test,docs"
MODE="full"

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
BLUE='\033[0;34m'
NC='\033[0m'

# 检测变更文件类型（用于按需加载）
detect_changed_files() {
  local changed_files="$1"

  HAS_BACKEND=false
  HAS_FRONTEND=false
  HAS_DOCS=false

  if [[ -z "$changed_files" ]]; then
    HAS_BACKEND=true
    HAS_FRONTEND=true
    HAS_DOCS=true
    return
  fi

  IFS=',' read -ra FILES <<< "$changed_files"
  for file in "${FILES[@]}"; do
    file=$(echo "$file" | xargs)
    [[ -z "$file" ]] && continue

    if [[ "$file" == backend/* ]]; then
      HAS_BACKEND=true
    elif [[ "$file" == frontend/* ]]; then
      HAS_FRONTEND=true
    elif [[ "$file" == *.md ]] || [[ "$file" == docs/* ]]; then
      HAS_DOCS=true
    fi
  done

  if [[ "$HAS_BACKEND" == "false" ]] && [[ "$HAS_FRONTEND" == "false" ]] && [[ "$HAS_DOCS" == "false" ]]; then
    HAS_BACKEND=true
    HAS_FRONTEND=true
    HAS_DOCS=true
  fi
}

# 运行单个命令
run_command() {
  local name="$1"
  local cmd="$2"
  local optional="${3:-false}"
  local message="${4:-Check failed: $name}"

  echo "  [CHECK] $name..."

  if [[ "$optional" == "true" ]]; then
    if ! bash -c "$cmd" 2>/dev/null; then
      echo -e "  ${YELLOW}[WARN] $name - optional, skipped${NC}"
      return 0
    fi
  else
    if ! bash -c "$cmd"; then
      echo -e "  ${RED}[FAIL] $name failed${NC}"
      echo "     $message"
      return 1
    fi
  fi

  echo -e "  ${GREEN}[PASS] $name${NC}"
  return 0
}

# 检测变更文件类型（从环境变量获取）
detect_changed_files "${DOCCHECK_CHANGED_FILES:-}"

# 主执行流程
echo ""
echo -e "${BLUE}=== ETF-Insight Unified Checks ===${NC}"
echo -e "${BLUE}   Stages: $STAGES | Mode: $MODE${NC}"
echo -e "${BLUE}   Backend: $HAS_BACKEND | Frontend: $HAS_FRONTEND | Docs: $HAS_DOCS${NC}"
echo ""

IFS=',' read -ra STAGE_LIST <<< "$STAGES"
FAILED=0
TOTAL_START=$(date +%s)

for stage in "${STAGE_LIST[@]}"; do
  stage=$(echo "$stage" | xargs)
  echo -e "${BLUE}>>> Stage: $stage${NC}"
  STAGE_FAILED=0

  case "$stage" in
    format)
      if [[ "$HAS_BACKEND" == "true" ]]; then
        run_command "gofmt" "cd $PROJECT_ROOT/backend && test -z \$(gofmt -l .)" false "Go files need formatting: cd backend && gofmt -w ." || STAGE_FAILED=1
      else
        echo -e "  ${YELLOW}[SKIP] gofmt - no backend changes${NC}"
      fi
      ;;
    static)
      if [[ "$HAS_BACKEND" == "true" ]]; then
        export GOPROXY="https://goproxy.cn,direct"
        run_command "go-vet" "cd $PROJECT_ROOT/backend && go vet ./..." false "Go vet found issues" || STAGE_FAILED=1
        # Wire DI 生成代码一致性检查（wire_gen.go 必须与 wire.go/providers.go 同步）
        # 使用 @v0.7.0 固定版本，避免依赖项目 go.sum（wire 为 build-tag 工具依赖，tidy 不为其写入 go.sum）
        run_command "wire-regen" "cd $PROJECT_ROOT/backend && go run github.com/google/wire/cmd/wire@v0.7.0 ./di/ && git diff --exit-code -- di/" false "Wire DI regeneration mismatch: cd backend && go run github.com/google/wire/cmd/wire@v0.7.0 ./di/" || STAGE_FAILED=1
      else
        echo -e "  ${YELLOW}[SKIP] go-vet - no backend changes${NC}"
      fi

      if [[ "$HAS_FRONTEND" == "true" ]]; then
        run_command "tsc" "cd $PROJECT_ROOT/frontend && npx tsc --noEmit" false "TypeScript check failed" || STAGE_FAILED=1
        run_command "eslint" "cd $PROJECT_ROOT/frontend && npx eslint --cache --quiet src/" false "ESLint check failed" || STAGE_FAILED=1
        # 图表库单一性检查（recharts / @ant-design/charts 必须已移除）
        # 匹配 import 语句（含 from 前缀），避免误伤注释中的库名
        run_command "chart-lib-consistency" "cd $PROJECT_ROOT/frontend && test -z \"\$(grep -rn -e \"from 'recharts'\" -e \"from '@ant-design/charts'\" src/)\"" false "Chart library inconsistency: recharts/@ant-design/charts must be removed" || STAGE_FAILED=1
      else
        echo -e "  ${YELLOW}[SKIP] tsc/eslint - no frontend changes${NC}"
      fi
      ;;
    build)
      if [[ "$HAS_BACKEND" == "true" ]]; then
        export GOPROXY="https://goproxy.cn,direct"
        run_command "go-build" "cd $PROJECT_ROOT/backend && go build ./..." false "Go build failed" || STAGE_FAILED=1
      else
        echo -e "  ${YELLOW}[SKIP] go-build - no backend changes${NC}"
      fi

      if [[ "$HAS_FRONTEND" == "true" ]]; then
        run_command "npm-build" "cd $PROJECT_ROOT/frontend && npm run build" false "Frontend build failed" || STAGE_FAILED=1
      else
        echo -e "  ${YELLOW}[SKIP] npm-build - no frontend changes${NC}"
      fi
      ;;
    test)
      if [[ "$HAS_BACKEND" == "true" ]]; then
        export GOPROXY="https://goproxy.cn,direct"
        run_command "go-test-short" "cd $PROJECT_ROOT/backend && go test -short -count=1 ./..." false "Go tests failed" || STAGE_FAILED=1
      else
        echo -e "  ${YELLOW}[SKIP] go-test - no backend changes${NC}"
      fi

      if [[ "$HAS_FRONTEND" == "true" ]]; then
        run_command "frontend-test" "cd $PROJECT_ROOT/frontend && npx vitest run" false "Frontend tests failed" || STAGE_FAILED=1
      else
        echo -e "  ${YELLOW}[SKIP] frontend-test - no frontend changes${NC}"
      fi
      ;;
    docs)
      if [[ "$HAS_DOCS" == "true" ]] || [[ "$HAS_BACKEND" == "true" ]] || [[ "$HAS_FRONTEND" == "true" ]]; then
        run_command "doccheck" "cd $PROJECT_ROOT/tools/doccheck && go run . --quick --strict" false "Documentation consistency check failed" || STAGE_FAILED=1
      else
        echo -e "  ${YELLOW}[SKIP] doccheck - no doc changes${NC}"
      fi
      ;;
    *)
      echo -e "  ${YELLOW}[WARN] Unknown stage: $stage${NC}"
      ;;
  esac

  if [[ $STAGE_FAILED -ne 0 ]]; then
    FAILED=1
  fi

  echo ""
done

TOTAL_END=$(date +%s)
TOTAL_TIME=$((TOTAL_END - TOTAL_START))

echo -e "${BLUE}>>> Total time: ${TOTAL_TIME}s${NC}"
echo ""

if [[ $FAILED -eq 0 ]]; then
  echo -e "${GREEN}[SUCCESS] All checks passed!${NC}"
  exit 0
else
  echo -e "${RED}[FAILED] Some checks failed${NC}"
  echo ""
  echo -e "${YELLOW}[TIP] Fix the issues above and try again.${NC}"
  echo -e "${YELLOW}     To bypass - not recommended: git commit --no-verify${NC}"
  exit 1
fi

#!/bin/bash
# ETF-Insight 统一检查执行引擎
# 驱动 pre-commit、CI/CD 和 Makefile

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

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

# 运行单个命令
run_command() {
  local name="$1"
  local cmd="$2"
  local timeout_sec="${3:-30}"
  local optional="${4:-false}"
  local message="${5:-"Check failed: $name"}"

  echo "  🔍 $name..."

  if [[ "$optional" == "true" ]]; then
    if ! timeout "$timeout_sec" bash -c "$cmd" 2>/dev/null; then
      echo -e "  ${YELLOW}⚠️  $name (optional, skipped)${NC}"
      return 0
    fi
  else
    if ! timeout "$timeout_sec" bash -c "$cmd"; then
      echo -e "  ${RED}❌ $name failed${NC}"
      echo "     $message"
      return 1
    fi
  fi

  echo -e "  ${GREEN}✅ $name${NC}"
  return 0
}

# 主执行流程
echo ""
echo -e "${BLUE}🔧 ETF-Insight Unified Checks${NC}"
echo -e "${BLUE}   Stages: $STAGES | Mode: $MODE${NC}"
echo ""

IFS=',' read -ra STAGE_LIST <<< "$STAGES"
FAILED=0
TOTAL_START=$(date +%s)

for stage in "${STAGE_LIST[@]}"; do
  stage=$(echo "$stage" | xargs)  # trim whitespace
  echo -e "${BLUE}📦 Stage: $stage${NC}"
  STAGE_FAILED=0

  case "$stage" in
    format)
      run_command "gofmt" "cd $PROJECT_ROOT/backend && test -z \$(gofmt -l .)" 5 false "Go files need formatting: cd backend && gofmt -w ." || STAGE_FAILED=1
      run_command "goimports" "cd $PROJECT_ROOT/backend && test -z \$(goimports -l . 2>/dev/null || echo 'skip')" 5 true || true
      ;;
    static)
      export GOPROXY="https://goproxy.cn,direct"
      run_command "go-vet" "cd $PROJECT_ROOT/backend && go vet ./..." 10 false "Go vet found issues" || STAGE_FAILED=1
      run_command "tsc" "cd $PROJECT_ROOT/frontend && npx tsc --noEmit" 10 false "TypeScript check failed" || STAGE_FAILED=1
      run_command "eslint" "cd $PROJECT_ROOT/frontend && npx eslint --cache --quiet src/" 10 false "ESLint check failed" || STAGE_FAILED=1
      ;;
    build)
      export GOPROXY="https://goproxy.cn,direct"
      run_command "go-build" "cd $PROJECT_ROOT/backend && go build ./..." 15 false "Go build failed" || STAGE_FAILED=1
      run_command "npm-build" "cd $PROJECT_ROOT/frontend && npm run build" 15 false "Frontend build failed" || STAGE_FAILED=1
      ;;
    test)
      export GOPROXY="https://goproxy.cn,direct"
      run_command "go-test-short" "cd $PROJECT_ROOT/backend && go test -short -count=1 ./..." 30 false "Go tests failed" || STAGE_FAILED=1
      run_command "frontend-test" "cd $PROJECT_ROOT/frontend && npx vitest run" 30 false "Frontend tests failed" || STAGE_FAILED=1
      ;;
    docs)
      run_command "doccheck" "cd $PROJECT_ROOT/tools/doccheck && go run . --quick --strict" 30 false "Documentation consistency check failed" || STAGE_FAILED=1
      ;;
    *)
      echo -e "  ${YELLOW}⚠️  Unknown stage: $stage${NC}"
      ;;
  esac

  if [[ $STAGE_FAILED -ne 0 ]]; then
    FAILED=1
  fi

  echo ""
done

TOTAL_END=$(date +%s)
TOTAL_TIME=$((TOTAL_END - TOTAL_START))

echo -e "${BLUE}⏱️  Total time: ${TOTAL_TIME}s${NC}"
echo ""

if [[ $FAILED -eq 0 ]]; then
  echo -e "${GREEN}✅ All checks passed!${NC}"
  exit 0
else
  echo -e "${RED}❌ Some checks failed${NC}"
  echo ""
  echo -e "${YELLOW}💡 Tip: Fix the issues above and try again.${NC}"
  echo -e "${YELLOW}   To bypass (not recommended): git commit --no-verify${NC}"
  exit 1
fi

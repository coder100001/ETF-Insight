#!/bin/bash
# ETF-Insight Pre-commit 统一入口
# 调用 run-check.sh 执行完整检查

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "🔍 Running pre-commit checks..."
echo ""

# 获取变更文件列表（供 doccheck 快速模式使用）
CHANGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | tr '\n' ',')
export DOCCHECK_CHANGED_FILES="$CHANGED_FILES"

# 运行所有检查
exec "$PROJECT_ROOT/scripts/run-check.sh" --stage=format,static,build,test,docs

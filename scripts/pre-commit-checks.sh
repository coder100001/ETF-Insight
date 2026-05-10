#!/bin/bash
# ETF-Insight Pre-commit 统一入口
# v2.1: 快速模式 - 只运行 format + static 检查
# build 和 test 交给 CI/CD

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "[PRE-COMMIT] Running quick checks..."
echo ""

# 获取变更文件列表（供按需加载使用）
CHANGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | tr '\n' ',')
export DOCCHECK_CHANGED_FILES="$CHANGED_FILES"

# 只运行 format + static（快速检查，约 7 秒）
# build 和 test 在 CI/CD 中运行
exec "$PROJECT_ROOT/scripts/run-check.sh" --stage=format,static

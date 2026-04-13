#!/bin/bash

set -e

echo "============================================"
echo "   ETF-Insight Git Hooks 安装程序"
echo "============================================"

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

install_pre_commit() {
    echo ""
    echo "[1/3] 检查 pre-commit..."
    if command -v pre-commit &> /dev/null; then
        echo "✅ pre-commit 已安装: $(pre-commit --version)"
    else
        echo "📦 正在安装 pre-commit..."
        if command -v pip3 &> /dev/null; then
            pip3 install pre-commit
        elif command -v pip &> /dev/null; then
            pip install pre-commit
        elif command -v brew &> /dev/null; then
            brew install pre-commit
        else
            echo "❌ 无法安装 pre-commit，请手动安装:"
            echo "   pip install pre-commit"
            echo "   或: brew install pre-commit"
            return 1
        fi
    fi
}

install_frontend_deps() {
    echo ""
    echo "[2/3] 检查前端依赖..."
    if [ -d "frontend/node_modules" ]; then
        echo "✅ 前端依赖已安装"
    else
        echo "📦 正在安装前端依赖..."
        cd frontend && npm install && cd ..
    fi
}

install_git_hooks() {
    echo ""
    echo "[3/3] 安装 Git hooks..."
    pre-commit install
    pre-commit install --hook-type pre-push
    echo "✅ Git hooks 安装成功"
}

show_instructions() {
    echo ""
    echo "============================================"
    echo "   安装完成！"
    echo "============================================"
    echo ""
    echo "已配置以下钩子:"
    echo "  ✓ pre-commit: TypeScript类型检查 + ESLint + Go格式化"
    echo "  ✓ pre-push: Go代码检查"
    echo ""
    echo "每次提交将自动运行:"
    echo "  1. TypeScript类型检查 (tsc --noEmit)"
    echo "  2. ESLint代码检查"
    echo "  3. Go代码格式化检查"
    echo "  4. Go代码lint检查"
    echo ""
    echo "如果检查失败，提交将被阻止。"
    echo "修复错误后重新提交即可。"
}

main() {
    install_pre_commit
    install_frontend_deps
    install_git_hooks
    show_instructions
}

main "$@"

#!/bin/bash
# Git预提交钩子 - 文档一致性检查

PROJECT_ROOT="$(git rev-parse --show-toplevel)"
DOCCHECK="$PROJECT_ROOT/tools/doccheck/doccheck"

echo "🔍 运行文档一致性检查..."

if [ ! -f "$DOCCHECK" ]; then
    echo "⚙️  编译文档检查工具..."
    (cd "$PROJECT_ROOT/tools/doccheck" && go build -o doccheck .)
    if [ $? -ne 0 ]; then
        echo "⚠️  编译失败，跳过文档检查"
        exit 0
    fi
fi

$DOCCHECK --project-root "$PROJECT_ROOT" --strict

if [ $? -ne 0 ]; then
    echo ""
    echo "❌ 文档一致性检查失败！"
    echo ""
    echo "请先更新文档以确保与代码保持一致："
    echo "1. 查看详细报告: $DOCCHECK --project-root $PROJECT_ROOT"
    echo "2. 导出报告: $DOCCHECK --project-root $PROJECT_ROOT --output report.md"
    echo "3. 根据报告更新相关文档"
    echo "4. 重新运行检查确认问题已解决"
    echo ""
    echo "如需跳过检查（不推荐），使用: git commit --no-verify"
    exit 1
fi

echo "✅ 文档一致性检查通过"
exit 0

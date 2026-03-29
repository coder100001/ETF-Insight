#!/bin/bash

# ETF-Insight 启动脚本
# 用途：解决依赖安装问题，等待网络恢复后启动服务

set -e

echo "=========================================="
echo "ETF-Insight 启动脚本"
echo "=========================================="

cd "$(dirname "$0")/backend"

# 清除可能的代理设置
unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY

# 配置 Go 代理
export GOPROXY=https://mirrors.aliyun.com/goproxy,direct
export GOSUMDB=sum.golang.google.cn

echo "1. 检查 Gin 依赖..."
if ! go list github.com/gin-gonic/gin > /dev/null 2>&1; then
    echo "   正在下载 Gin 依赖..."
    go mod download github.com/gin-gonic/gin || {
        echo "   警告: Gin 依赖下载失败，尝试使用现有代码..."
    }
fi

echo "2. 运行 go mod tidy..."
go mod tidy || {
    echo "   警告: go mod tidy 失败，继续尝试运行..."
}

echo "3. 检查代码编译..."
go build -o /tmp/etf-insight-check . || {
    echo "   错误: 代码编译失败，请检查错误信息"
    exit 1
}

echo "4. 启动服务..."
echo "   访问地址: http://localhost:8080"
echo "   健康检查: http://localhost:8080/health"
echo "=========================================="

# 启动服务
go run main.go

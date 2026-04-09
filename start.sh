#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$SCRIPT_DIR"
BACKEND_DIR="$PROJECT_DIR/backend"
FRONTEND_DIR="$PROJECT_DIR/frontend"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

BACKEND_PORT=8080
FRONTEND_PORT=5173

error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
skip() { echo -e "${YELLOW}[SKIP]${NC} $1"; }
info() { echo -e "${BLUE}[INFO]${NC} $1"; }

check_port() {
    if lsof -i :$1 > /dev/null 2>&1; then
        return 0
    fi
    return 1
}

kill_port_process() {
    local port=$1
    if lsof -i :$port > /dev/null 2>&1; then
        local pids=$(lsof -t -i :$port 2>/dev/null)
        if [ -n "$pids" ]; then
            info "释放端口 $port..."
            kill -9 $pids 2>/dev/null || true
            sleep 1
        fi
    fi
}

cleanup() {
    info "停止服务..."
    pkill -f "etf-insight" 2>/dev/null || true
    pkill -f "vite" 2>/dev/null || true
}

trap cleanup EXIT INT TERM

echo ""
echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}   ETF-Insight 快速启动${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""

# 检查端口
if check_port $BACKEND_PORT; then
    info "后端端口 $BACKEND_PORT 已被占用，尝试释放..."
    kill_port_process $BACKEND_PORT
fi

if check_port $FRONTEND_PORT; then
    info "前端端口 $FRONTEND_PORT 已被占用，尝试释放..."
    kill_port_process $FRONTEND_PORT
fi

# ===== 后端 =====
cd "$BACKEND_DIR"

# 设置 Go 代理
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=sum.golang.google.cn

# 检查是否需要编译
NEED_BUILD=true
if [ -f "etf-insight" ]; then
    # 检查源码是否比二进制文件新
    newest_src=$(find . -name "*.go" -newer etf-insight 2>/dev/null | head -1)
    if [ -z "$newest_src" ]; then
        skip "后端二进制已存在，跳过编译"
        NEED_BUILD=false
    else
        info "检测到源码更新，重新编译..."
    fi
fi

if [ "$NEED_BUILD" = true ]; then
    go build -o etf-insight . || error "后端编译失败"
    success "后端编译完成"
fi

# 启动后端
./etf-insight &
BACKEND_PID=$!
sleep 2

if kill -0 $BACKEND_PID 2>/dev/null; then
    success "后端已启动 (http://localhost:$BACKEND_PORT)"
else
    error "后端启动失败"
fi

# ===== 前端 =====
cd "$FRONTEND_DIR"

# 启动前端
npm run dev &
FRONTEND_PID=$!
sleep 3

if kill -0 $FRONTEND_PID 2>/dev/null; then
    success "前端已启动 (http://localhost:$FRONTEND_PORT)"
else
    error "前端启动失败"
fi

echo ""
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}   ETF-Insight 启动成功！${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""
echo -e "  前端:  http://localhost:$FRONTEND_PORT"
echo -e "  后端:  http://localhost:$BACKEND_PORT"
echo -e "  ${YELLOW}按 Ctrl+C 停止${NC}"
echo ""

wait

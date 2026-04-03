#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
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
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
info() { echo -e "${BLUE}[INFO]${NC} $1"; }

check_port() {
    if lsof -i :$1 > /dev/null 2>&1 || netstat -tuln 2>/dev/null | grep ":$1 " > /dev/null; then
        return 0
    fi
    return 1
}

cleanup() {
    info "正在清理进程..."
    pkill -f "go run main.go" 2>/dev/null || true
    pkill -f "vite" 2>/dev/null || true
}

trap cleanup EXIT INT TERM

echo ""
echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}   ETF-Insight 一键启动脚本${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""

info "项目目录: $PROJECT_DIR"

info "检查环境..."

if ! command -v go &> /dev/null; then
    error "Go 未安装，请先安装 Go (>= 1.21)"
fi

GO_VERSION=$(go version | grep -oP '\d+\.\d+' | head -1)
success "Go 版本: $GO_VERSION"

if ! command -v node &> /dev/null; then
    error "Node.js 未安装，请先安装 Node.js (>= 18)"
fi

NODE_VERSION=$(node --version)
success "Node.js 版本: $NODE_VERSION"

if ! command -v npm &> /dev/null; then
    error "npm 未安装"
fi
success "npm 已安装"

if check_port $BACKEND_PORT; then
    warn "端口 $BACKEND_PORT 已被占用，请先释放端口或停止现有服务"
fi

if check_port $FRONTEND_PORT; then
    warn "端口 $FRONTEND_PORT 已被占用，请先释放端口或停止现有服务"
fi

echo ""
info "=========================================="
info "步骤 1/5: 安装后端依赖"
info "=========================================="
cd "$BACKEND_DIR"

unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=sum.golang.google.cn

if [ -f "go.sum" ]; then
    info "检测到 go.sum，验证依赖完整性..."
    go mod verify > /dev/null 2>&1 || {
        warn "依赖校验失败，重新下载..."
        rm -f go.sum
    }
fi

info "执行 go mod download..."
go mod download || {
    error "后端依赖下载失败，请检查网络连接"
}
success "后端依赖安装完成"

echo ""
info "=========================================="
info "步骤 2/5: 编译后端项目"
info "=========================================="

info "执行代码格式化..."
gofmt -w . > /dev/null 2>&1 || true

info "编译后端..."
go build -o etf-insight . || {
    error "后端编译失败，请检查代码错误"
}
success "后端编译成功"

echo ""
info "=========================================="
info "步骤 3/5: 安装前端依赖"
info "=========================================="
cd "$FRONTEND_DIR"

if [ ! -d "node_modules" ]; then
    info "首次安装，正在下载 npm 依赖..."
    npm install || {
        error "前端依赖安装失败，请检查网络连接"
    }
else
    info "node_modules 已存在，检查更新..."
    npm install --prefer-offline || {
        warn "前端依赖更新失败，尝试使用缓存..."
    }
fi
success "前端依赖安装完成"

echo ""
info "=========================================="
info "步骤 4/5: 启动后端服务"
info "=========================================="
cd "$BACKEND_DIR"

if [ -f "etf_insight.db" ]; then
    info "数据库文件已存在"
else
    info "初始化数据库..."
fi

./etf-insight &
BACKEND_PID=$!
sleep 3

if kill -0 $BACKEND_PID 2>/dev/null; then
    success "后端服务已启动 (PID: $BACKEND_PID)"
    info "后端地址: http://localhost:$BACKEND_PORT"
    info "健康检查: http://localhost:$BACKEND_PORT/health"
else
    error "后端服务启动失败，请查看日志"
fi

echo ""
info "=========================================="
info "步骤 5/5: 启动前端开发服务器"
info "=========================================="
cd "$FRONTEND_DIR"

npx vite --host &
FRONTEND_PID=$!
sleep 3

if kill -0 $FRONTEND_PID 2>/dev/null; then
    success "前端服务已启动 (PID: $FRONTEND_PID)"
    info "前端地址: http://localhost:$FRONTEND_PORT"
else
    error "前端服务启动失败，请查看日志"
fi

echo ""
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}   🎉 ETF-Insight 启动成功！${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""
echo -e "  ${BLUE}前端地址:${NC}  http://localhost:$FRONTEND_PORT"
echo -e "  ${BLUE}后端地址:${NC}  http://localhost:$BACKEND_PORT"
echo -e "  ${BLUE}健康检查:${NC}  http://localhost:$BACKEND_PORT/health"
echo -e "  ${BLUE}API 文档:${NC}  http://localhost:$BACKEND_PORT/api/exchange-rates"
echo ""
echo -e "  ${YELLOW}按 Ctrl+C 停止所有服务${NC}"
echo ""

wait
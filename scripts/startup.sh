#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BACKEND_DIR="$PROJECT_DIR/backend"
FRONTEND_DIR="$PROJECT_DIR/frontend"
LOG_DIR="$PROJECT_DIR/logs"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

BACKEND_PORT=8080
FRONTEND_PORT=5173
BACKEND_TIMEOUT=60
FRONTEND_TIMEOUT=30
HEALTH_CHECK_INTERVAL=2

BACKEND_PID=""
FRONTEND_PID=""

usage() {
    cat << EOF
ETF-Insight 信号量启动脚本

用法: $0 [选项]

选项:
    -h, --help              显示帮助信息
    -b, --backend-only      仅启动后端服务
    -f, --frontend-only     仅启动前端服务（需要后端已运行）
    -t, --timeout SECONDS   设置启动超时时间（默认：后端60秒，前端30秒）
    -l, --log-dir DIR       设置日志目录（默认：./logs）
    -v, --verbose           显示详细日志
    --skip-health-check     跳过健康检查
    --dry-run               仅检查环境，不启动服务

信号量机制:
    1. 后端服务启动后，通过 /ready 接口报告就绪状态
    2. 前端服务等待后端就绪信号后才启动
    3. 超时机制确保启动过程不会无限等待
    4. 详细日志记录启动过程中的关键节点

示例:
    $0                      # 完整启动（推荐）
    $0 -b                   # 仅启动后端
    $0 -f                   # 仅启动前端（后端必须已运行）
    $0 -t 120               # 设置超时为120秒
    $0 -v                   # 显示详细日志

EOF
    exit 0
}

log() {
    local level=$1
    local message=$2
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case $level in
        INFO)
            echo -e "${BLUE}[$timestamp INFO]${NC} $message"
            ;;
        SUCCESS)
            echo -e "${GREEN}[$timestamp SUCCESS]${NC} $message"
            ;;
        WARN)
            echo -e "${YELLOW}[$timestamp WARN]${NC} $message"
            ;;
        ERROR)
            echo -e "${RED}[$timestamp ERROR]${NC} $message"
            ;;
        SEMAPHORE)
            echo -e "${CYAN}[$timestamp SEMAPHORE]${NC} $message"
            ;;
    esac
    
    # 写入日志文件
    if [ -d "$LOG_DIR" ]; then
        echo "[$timestamp $level] $message" >> "$LOG_DIR/startup.log"
    fi
}

# 解析命令行参数
BACKEND_ONLY=false
FRONTEND_ONLY=false
VERBOSE=false
SKIP_HEALTH_CHECK=false
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            ;;
        -b|--backend-only)
            BACKEND_ONLY=true
            shift
            ;;
        -f|--frontend-only)
            FRONTEND_ONLY=true
            shift
            ;;
        -t|--timeout)
            BACKEND_TIMEOUT="$2"
            FRONTEND_TIMEOUT="$2"
            shift 2
            ;;
        -l|--log-dir)
            LOG_DIR="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --skip-health-check)
            SKIP_HEALTH_CHECK=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            log ERROR "未知选项: $1"
            usage
            ;;
    esac
done

# 验证参数
if [ "$BACKEND_ONLY" = true ] && [ "$FRONTEND_ONLY" = true ]; then
    log ERROR "不能同时指定 --backend-only 和 --frontend-only"
    exit 1
fi

# 创建日志目录
mkdir -p "$LOG_DIR"

log INFO "=========================================="
log INFO "ETF-Insight 信号量启动脚本"
log INFO "=========================================="
log INFO "项目目录: $PROJECT_DIR"
log INFO "日志目录: $LOG_DIR"
log INFO "后端超时: ${BACKEND_TIMEOUT}秒"
log INFO "前端超时: ${FRONTEND_TIMEOUT}秒"

# 详细模式
if [ "$VERBOSE" = true ]; then
    log INFO "详细模式: 启用"
fi

# 环境检查
check_environment() {
    log INFO "=========================================="
    log INFO "阶段 1/3: 环境检查"
    log INFO "=========================================="
    
    local errors=0
    
    # 检查 Go
    if ! command -v go &> /dev/null; then
        log ERROR "Go 未安装，请先安装 Go (>= 1.21)"
        ((errors++))
    else
        local go_version=$(go version | awk '{print $3}')
        log SUCCESS "Go 已安装: $go_version"
    fi
    
    # 检查 Node.js
    if ! command -v node &> /dev/null; then
        log ERROR "Node.js 未安装，请先安装 Node.js (>= 18)"
        ((errors++))
    else
        local node_version=$(node --version)
        log SUCCESS "Node.js 已安装: $node_version"
    fi
    
    # 检查 npm
    if ! command -v npm &> /dev/null; then
        log ERROR "npm 未安装"
        ((errors++))
    else
        local npm_version=$(npm --version)
        log SUCCESS "npm 已安装: $npm_version"
    fi
    
    # 检查 curl（用于健康检查）
    if ! command -v curl &> /dev/null; then
        log WARN "curl 未安装，将使用 nc 进行健康检查"
    else
        log SUCCESS "curl 已安装"
    fi
    
    # 检查目录
    if [ ! -d "$BACKEND_DIR" ]; then
        log ERROR "后端目录不存在: $BACKEND_DIR"
        ((errors++))
    else
        log SUCCESS "后端目录: $BACKEND_DIR"
    fi
    
    if [ ! -d "$FRONTEND_DIR" ]; then
        log ERROR "前端目录不存在: $FRONTEND_DIR"
        ((errors++))
    else
        log SUCCESS "前端目录: $FRONTEND_DIR"
    fi
    
    if [ $errors -gt 0 ]; then
        log ERROR "环境检查失败，发现 $errors 个错误"
        exit 1
    fi
    
    log SUCCESS "环境检查通过"
}

# 释放端口
release_port() {
    local port=$1
    if lsof -i :$port > /dev/null 2>&1; then
        local pids=$(lsof -t -i :$port 2>/dev/null)
        if [ -n "$pids" ]; then
            log WARN "端口 $port 被占用，正在释放..."
            kill -9 $pids 2>/dev/null || true
            sleep 1
            log SUCCESS "端口 $port 已释放"
        fi
    fi
}

# 等待后端就绪（信号量机制）
wait_for_backend_ready() {
    log SEMAPHORE "=========================================="
    log SEMAPHORE "阶段 2/3: 后端就绪信号量检测"
    log SEMAPHORE "=========================================="
    log SEMAPHORE "等待后端服务就绪（超时: ${BACKEND_TIMEOUT}秒）..."
    
    local elapsed=0
    local check_count=0
    
    while [ $elapsed -lt $BACKEND_TIMEOUT ]; do
        ((check_count++))
        
        if [ "$VERBOSE" = true ]; then
            log SEMAPHORE "检查 #$check_count: 等待后端就绪信号..."
        fi
        
        # 使用 curl 或 nc 检查健康接口
        local ready=false
        if command -v curl &> /dev/null; then
            if curl -s http://localhost:$BACKEND_PORT/ready > /dev/null 2>&1; then
                ready=true
            fi
        else
            if nc -z localhost $BACKEND_PORT 2>/dev/null; then
                ready=true
            fi
        fi
        
        if [ "$ready" = true ]; then
            log SEMAPHORE "✓ 收到后端就绪信号！"
            log SEMAPHORE "  检查次数: $check_count"
            log SEMAPHORE "  等待时间: ${elapsed}秒"
            
            # 获取详细状态
            if command -v curl &> /dev/null; then
                local status=$(curl -s http://localhost:$BACKEND_PORT/ready 2>/dev/null | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
                if [ -n "$status" ]; then
                    log SEMAPHORE "  服务状态: $status"
                fi
            fi
            
            return 0
        fi
        
        sleep $HEALTH_CHECK_INTERVAL
        ((elapsed+=HEALTH_CHECK_INTERVAL))
        
        if [ "$VERBOSE" = true ] && [ $((elapsed % 10)) -eq 0 ]; then
            log SEMAPHORE "已等待 ${elapsed}秒..."
        fi
    done
    
    log ERROR "=========================================="
    log ERROR "后端服务启动超时！"
    log ERROR "=========================================="
    log ERROR "等待时间: ${BACKEND_TIMEOUT}秒"
    log ERROR "检查次数: $check_count"
    log ERROR ""
    log ERROR "故障排查指引:"
    log ERROR "1. 检查后端日志: tail -f $LOG_DIR/backend.log"
    log ERROR "2. 手动检查健康接口: curl http://localhost:$BACKEND_PORT/ready"
    log ERROR "3. 检查端口占用: lsof -i :$BACKEND_PORT"
    log ERROR "4. 检查数据库连接是否正常"
    log ERROR "5. 尝试重启后端服务"
    
    return 1
}

# 启动后端服务
start_backend() {
    log INFO "=========================================="
    log INFO "启动后端服务"
    log INFO "=========================================="
    
    cd "$BACKEND_DIR"
    
    # 释放端口
    release_port $BACKEND_PORT
    
    # 设置环境变量
    unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY
    export GOPROXY=https://goproxy.cn,direct
    export GOSUMDB=sum.golang.google.cn
    
    # 安装依赖
    log INFO "安装后端依赖..."
    go mod download >> "$LOG_DIR/backend.log" 2>&1 || {
        log ERROR "后端依赖安装失败"
        return 1
    }
    log SUCCESS "后端依赖安装完成"
    
    # 编译
    log INFO "编译后端项目..."
    go build -o etf-insight . >> "$LOG_DIR/backend.log" 2>&1 || {
        log ERROR "后端编译失败"
        return 1
    }
    log SUCCESS "后端编译成功"
    
    # 启动服务
    log INFO "启动后端服务..."
    log SEMAPHORE "发送后端启动信号..."
    
    ./etf-insight >> "$LOG_DIR/backend.log" 2>&1 &
    BACKEND_PID=$!
    
    log SEMAPHORE "后端进程已启动 (PID: $BACKEND_PID)"
    log SEMAPHORE "等待后端就绪信号..."
    
    # 等待就绪
    if [ "$SKIP_HEALTH_CHECK" = false ]; then
        if ! wait_for_backend_ready; then
            return 1
        fi
    else
        log WARN "跳过健康检查，等待 5 秒..."
        sleep 5
    fi
    
    log SUCCESS "后端服务启动完成"
    log INFO "后端地址: http://localhost:$BACKEND_PORT"
    log INFO "健康检查: http://localhost:$BACKEND_PORT/health"
    log INFO "就绪检查: http://localhost:$BACKEND_PORT/ready"
    
    return 0
}

# 启动前端服务
start_frontend() {
    log SEMAPHORE "=========================================="
    log SEMAPHORE "阶段 3/3: 前端服务启动（收到后端就绪信号）"
    log SEMAPHORE "=========================================="
    
    cd "$FRONTEND_DIR"
    
    # 释放端口
    release_port $FRONTEND_PORT
    
    # 安装依赖
    if [ ! -d "node_modules" ]; then
        log INFO "安装前端依赖..."
        npm install >> "$LOG_DIR/frontend.log" 2>&1 || {
            log ERROR "前端依赖安装失败"
            return 1
        }
        log SUCCESS "前端依赖安装完成"
    else
        log INFO "前端依赖已存在"
    fi
    
    # 启动服务
    log INFO "启动前端开发服务器..."
    log SEMAPHORE "前端服务正在启动..."
    
    npx vite --host >> "$LOG_DIR/frontend.log" 2>&1 &
    FRONTEND_PID=$!
    
    log SEMAPHORE "前端进程已启动 (PID: $FRONTEND_PID)"
    
    # 等待前端就绪
    local elapsed=0
    while [ $elapsed -lt $FRONTEND_TIMEOUT ]; do
        if lsof -i :$FRONTEND_PORT > /dev/null 2>&1; then
            log SEMAPHORE "✓ 前端服务就绪！"
            log SEMAPHORE "  等待时间: ${elapsed}秒"
            break
        fi
        sleep 1
        ((elapsed++))
    done
    
    if [ $elapsed -ge $FRONTEND_TIMEOUT ]; then
        log ERROR "前端服务启动超时"
        return 1
    fi
    
    log SUCCESS "前端服务启动完成"
    log INFO "前端地址: http://localhost:$FRONTEND_PORT"
    
    return 0
}

# 清理函数
cleanup() {
    log INFO "=========================================="
    log INFO "清理进程..."
    log INFO "=========================================="
    
    if [ -n "$FRONTEND_PID" ]; then
        log INFO "停止前端服务 (PID: $FRONTEND_PID)..."
        kill $FRONTEND_PID 2>/dev/null || true
        wait $FRONTEND_PID 2>/dev/null || true
    fi
    
    if [ -n "$BACKEND_PID" ]; then
        log INFO "停止后端服务 (PID: $BACKEND_PID)..."
        kill $BACKEND_PID 2>/dev/null || true
        wait $BACKEND_PID 2>/dev/null || true
    fi
    
    log SUCCESS "清理完成"
}

trap cleanup EXIT INT TERM

# 主流程
main() {
    # 环境检查
    check_environment
    
    # 仅检查模式
    if [ "$DRY_RUN" = true ]; then
        log INFO "=========================================="
        log INFO "Dry-run 模式：仅检查环境，不启动服务"
        log INFO "=========================================="
        exit 0
    fi
    
    # 仅前端模式
    if [ "$FRONTEND_ONLY" = true ]; then
        log INFO "=========================================="
        log INFO "仅启动前端模式"
        log INFO "=========================================="
        
        # 检查后端是否已运行
        if ! curl -s http://localhost:$BACKEND_PORT/ready > /dev/null 2>&1; then
            log ERROR "后端服务未运行，无法启动前端"
            log ERROR "请先启动后端服务，或使用完整启动模式"
            exit 1
        fi
        
        log SUCCESS "检测到后端服务已就绪"
        start_frontend || exit 1
        
        log SUCCESS "=========================================="
        log SUCCESS "前端服务启动成功！"
        log SUCCESS "=========================================="
        log INFO "访问地址: http://localhost:$FRONTEND_PORT"
        
        wait
        exit 0
    fi
    
    # 启动后端
    if [ "$BACKEND_ONLY" = true ] || [ "$FRONTEND_ONLY" = false ]; then
        start_backend || exit 1
    fi
    
    # 启动前端（非仅后端模式）
    if [ "$BACKEND_ONLY" = false ]; then
        start_frontend || exit 1
    fi
    
    # 启动完成
    log SUCCESS "=========================================="
    log SUCCESS "🎉 ETF-Insight 启动成功！"
    log SUCCESS "=========================================="
    log INFO ""
    log INFO "服务状态:"
    log INFO "  后端: http://localhost:$BACKEND_PORT (PID: $BACKEND_PID)"
    log INFO "  前端: http://localhost:$FRONTEND_PORT (PID: $FRONTEND_PID)"
    log INFO ""
    log INFO "监控端点:"
    log INFO "  健康检查: http://localhost:$BACKEND_PORT/health"
    log INFO "  就绪检查: http://localhost:$BACKEND_PORT/ready"
    log INFO "  存活检查: http://localhost:$BACKEND_PORT/live"
    log INFO ""
    log INFO "日志文件:"
    log INFO "  启动日志: $LOG_DIR/startup.log"
    log INFO "  后端日志: $LOG_DIR/backend.log"
    log INFO "  前端日志: $LOG_DIR/frontend.log"
    log INFO ""
    log INFO "按 Ctrl+C 停止所有服务"
    
    wait
}

# 执行主流程
main
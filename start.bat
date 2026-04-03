@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion

set "SCRIPT_DIR=%~dp0"
set "PROJECT_DIR=%SCRIPT_DIR%.."
set "BACKEND_DIR=%PROJECT_DIR%\backend"
set "FRONTEND_DIR=%PROJECT_DIR%\frontend"

set BACKEND_PORT=8080
set FRONTEND_PORT=5173

echo.
echo ============================================
echo    ETF-Insight 一键启动脚本 (Windows)
echo ============================================
echo.

echo [INFO] 项目目录: %PROJECT_DIR%

echo [INFO] 检查环境...
where go >nul 2>nul
if errorlevel 1 (
    echo [ERROR] Go 未安装，请先安装 Go (>= 1.21)
    exit /b 1
)

for /f "tokens=3" %%v in ('go version') do set GO_VERSION=%%v
echo [OK] Go 版本: %GO_VERSION%

where node >nul 2>nul
if errorlevel 1 (
    echo [ERROR] Node.js 未安装，请先安装 Node.js (>= 18)
    exit /b 1
)

for /f "tokens=*" %%v in ('node --version') do set NODE_VERSION=%%v
echo [OK] Node.js 版本: %NODE_VERSION%

where npm >nul 2>nul
if errorlevel 1 (
    echo [ERROR] npm 未安装
    exit /b 1
)
echo [OK] npm 已安装

netstat -ano | findstr ":%BACKEND_PORT% " >nul 2>nul
if not errorlevel 1 (
    echo [WARN] 端口 %BACKEND_PORT% 已被占用
)

netstat -ano | findstr ":%FRONTEND_PORT% " >nul 2>nul
if not errorlevel 1 (
    echo [WARN] 端口 %FRONTEND_PORT% 已被占用
)

echo.
echo ============================================
echo 步骤 1/5: 安装后端依赖
echo ============================================
cd /d "%BACKEND_DIR%"

set GOPROXY=https://goproxy.cn,direct
set GOSUMDB=sum.golang.google.cn

if exist go.sum (
    echo [INFO] 检测到 go.sum，验证依赖完整性...
    go mod verify >nul 2>nul || (
        echo [WARN] 依赖校验失败，重新下载...
        del /q go.sum >nul 2>nul
    )
)

echo [INFO] 执行 go mod download...
go mod download
if errorlevel 1 (
    echo [ERROR] 后端依赖下载失败
    exit /b 1
)
echo [OK] 后端依赖安装完成

echo.
echo ============================================
echo 步骤 2/5: 编译后端项目
echo ============================================

echo [INFO] 编译后端...
go build -o etf-insight.exe .
if errorlevel 1 (
    echo [ERROR] 后端编译失败
    exit /b 1
)
echo [OK] 后端编译成功

echo.
echo ============================================
echo 步骤 3/5: 安装前端依赖
echo ============================================
cd /d "%FRONTEND_DIR%"

if not exist node_modules (
    echo [INFO] 首次安装，正在下载 npm 依赖...
    call npm install
    if errorlevel 1 (
        echo [ERROR] 前端依赖安装失败
        exit /b 1
    )
) else (
    echo [INFO] node_modules 已存在，检查更新...
    call npm install --prefer-offline >nul 2>nul || echo [WARN] 使用缓存依赖
)
echo [OK] 前端依赖安装完成

echo.
echo ============================================
echo 步骤 4/5: 启动后端服务
echo ============================================
cd /d "%BACKEND_DIR%"

start "ETF-Insight Backend" /B etf-insight.exe
timeout /t 3 /nobreak >nul

echo [OK] 后端服务已启动
echo [INFO] 后端地址: http://localhost:%BACKEND_PORT%
echo [INFO] 健康检查: http://localhost:%BACKEND_PORT%/health

echo.
echo ============================================
echo 步骤 5/5: 启动前端开发服务器
echo ============================================
cd /d "%FRONTEND_DIR%"

start "ETF-Insight Frontend" /B npx vite --host
timeout /t 3 /nobreak >nul

echo [OK] 前端服务已启动
echo [INFO] 前端地址: http://localhost:%FRONTEND_PORT%

echo.
echo ============================================
echo    ETF-Insight 启动成功！
echo ============================================
echo.
echo   前端地址:  http://localhost:%FRONTEND_PORT%
echo   后端地址:  http://localhost:%BACKEND_PORT%
echo   健康检查:  http://localhost:%BACKEND_PORT%/health
echo.
echo   按 Ctrl+C 停止所有服务
echo.

pause
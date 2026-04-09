# ETF-Insight 系统部署指南

## 系统架构概述

ETF-Insight 是一个完整的金融数据分析和可视化系统，包含以下组件：

1. **后端服务 (Go)**: 数据同步、API接口、定时任务
2. **前端界面 (React + TypeScript)**: 数据可视化、用户界面
3. **数据库 (SQLite)**: ETF配置和交易数据存储
4. **数据同步服务**: 定时同步ETF市场数据

## 系统启动步骤

### 1. 环境准备

确保已安装以下依赖：
- Go 1.25+ (当前: 1.26.1)
- Node.js 20+ (当前: v24.11.0)
- npm 9+ (当前: 11.6.1)

### 2. 启动后端服务

#### 选项A: 启动完整后端服务（API + 定时任务）
```bash
cd /Users/liunian/Desktop/dnmp/py_project/backend
go run main.go
```

**服务信息:**
- API服务器: http://localhost:8080
- 健康检查: http://localhost:8080/health
- 就绪检查: http://localhost:8080/ready

#### 选项B: 启动数据同步服务（仅数据同步）
```bash
cd /Users/liunian/Desktop/dnmp/py_project/backend
go run cmd/syncetf/main.go
```

**功能:** 同步15只ETF的实时数据和配置信息到数据库

### 3. 启动前端服务

```bash
cd /Users/liunian/Desktop/dnmp/py_project/frontend
npm run dev
```

**服务信息:**
- 前端地址: http://localhost:5173
- API代理: 自动代理到 http://localhost:8080/api

### 4. 同时启动所有服务（推荐）

```bash
# 终端1: 启动后端API服务
cd /Users/liunian/Desktop/dnmp/py_project/backend && go run main.go

# 终端2: 启动前端开发服务器
cd /Users/liunian/Desktop/dnmp/py_project/frontend && npm run dev

# 终端3: 启动数据同步服务（可选）
cd /Users/liunian/Desktop/dnmp/py_project/backend && go run cmd/syncetf/main.go
```

## 已验证功能

### 后端API接口
- ✓ `/health` - 健康检查
- ✓ `/ready` - 就绪检查
- ✓ `/api/etf/list` - ETF列表
- ✓ `/api/etf-configs/` - ETF配置管理
- ✓ `/api/portfolio-configs/` - 投资组合配置
- ✓ `/api/a-share/etfs` - A股ETF数据
- ✓ `/api/exchange-rates` - 汇率数据

### 前端功能
- ✓ 前端页面加载: http://localhost:5173
- ✓ API代理配置: Vite配置正确转发到后端
- ✓ React应用: 使用Ant Design组件库

### 数据同步
- ✓ 数据同步服务: 成功同步15只ETF数据
- ✓ 数据库: SQLite文件已创建并包含数据
- ✓ 定时任务: 每小时自动更新数据

## 配置说明

### 后端配置 (`backend/config.yaml`)
```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  dsn: "etf_insight.db"  # SQLite数据库文件

log:
  level: "info"  # 日志级别: debug, info, warn, error
```

### 前端代理配置 (`frontend/vite.config.ts`)
```typescript
server: {
  port: 5173,
  proxy: {
    '/api': {
      target: 'http://127.0.0.1:8080',
      changeOrigin: true,
    },
  },
}
```

## 数据库结构

系统使用以下主要表：
1. `etf_configs` - ETF配置信息
2. `etf_data` - ETF历史交易数据
3. `portfolio_configs` - 投资组合配置
4. `operation_logs` - 操作日志
5. `exchange_rates` - 汇率数据
6. `currency_pairs` - 货币对配置

## 故障排除

### 常见问题

1. **端口冲突**
   - 8080端口: 后端API服务
   - 5173端口: 前端开发服务器
   - 如端口被占用，修改 `config.yaml` 或 `vite.config.ts` 中的端口配置

2. **Go依赖问题**
   ```bash
   cd backend && go mod tidy
   cd backend && go build
   ```

3. **Node依赖问题**
   ```bash
   cd frontend && npm install
   cd frontend && npm run build
   ```

4. **API连接问题**
   - 检查后端服务是否运行: `curl http://localhost:8080/health`
   - 检查前端代理配置是否正确
   - 确认后端监听地址为 `0.0.0.0` 而不是 `127.0.0.1`

### 日志查看

1. **后端日志**: 服务启动时输出到控制台
2. **前端日志**: 浏览器开发者工具控制台
3. **数据库日志**: SQLite文件位于 `backend/etf_insight.db`

## 监控和健康检查

### 健康检查端点
- `GET /health` - 服务状态
- `GET /ready` - 服务就绪状态
- `GET /live` - 服务存活状态

### 系统状态
```bash
# 检查服务状态
curl http://localhost:8080/health
curl http://localhost:8080/ready

# 检查ETF数据
curl http://localhost:8080/api/etf/list

# 检查数据库连接
# 使用SQLite命令行工具查看数据库
sqlite3 backend/etf_insight.db ".tables"
```

## 扩展功能

### 数据源扩展
1. 添加新的ETF配置到 `etf_configs` 表
2. 修改数据同步逻辑以支持新的数据源
3. 更新前端以显示新的ETF类型

### 功能扩展
1. 添加新的分析指标
2. 增加投资组合优化算法
3. 集成实时市场数据源
4. 添加用户认证和权限管理

## 部署建议

### 生产环境部署
1. 使用Nginx反向代理前端和后端
2. 配置HTTPS证书
3. 使用PostgreSQL或MySQL替代SQLite
4. 设置进程守护（如systemd或supervisor）
5. 配置监控和告警

### 性能优化
1. 数据库索引优化
2. 查询缓存
3. 前端资源压缩
4. 静态资源CDN

---

## 当前状态总结

✅ **已完成的部署:**
- 后端API服务: 运行正常
- 前端开发服务器: 运行正常
- 数据同步服务: 运行正常
- API代理: 配置正确
- 数据库: 包含15只ETF数据

🚀 **系统已就绪:**
- 访问 http://localhost:5173 查看前端界面
- API服务运行在 http://localhost:8080
- 所有组件已集成并协同工作
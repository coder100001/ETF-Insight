# ETF-Insight

一个专业的 ETF 分析与对比平台，对标 Trackinsight、ETF Insider 等国际知名 ETF 分析工具。基于 Go + React 技术栈，提供深度的 ETF 数据洞察、多维度对比分析、持仓解构、风险评估和投资组合优化等一站式解决方案。

## 🎯 产品定位

ETF-Insight 致力于成为专业投资者和机构用户的 ETF 分析利器：

- **ETF 对比分析** - 多维度并排对比，发现最优投资标的
- **持仓深度解构** - 穿透底层资产，了解真实风险敞口
- **风险指标评估** - 波动率、夏普比率、最大回撤、Beta 等专业指标
- **投资组合优化** - 基于现代投资组合理论，构建最优资产配置

## ✨ 核心特性

### 📊 ETF 对比分析（ETF Comparison）
- **并排对比** - 最多支持 5 只 ETF 同时对比
- **多维度指标** - 费率、AUM、股息率、业绩表现、风险指标
- **持仓重叠分析** - 识别 ETF 间的持仓重合度，避免过度集中
- **业绩回测对比** - 不同时间周期的收益表现对比

### 🔍 持仓深度解构（Holdings Analysis）
- **前十大持仓** - 穿透底层资产，了解核心持仓
- **行业分布** - sector 权重分布及变化趋势
- **地区分布** - 国家/地区配置比例
- **市值分布** - 大/中/小盘股配置比例
- **风格分析** - 价值/成长风格暴露度

### 💼 A股红利ETF投资组合
- **A股ETF管理** - 支持中证红利、红利低波等主流红利ETF
- **投资占比分布** - 饼状图可视化展示投资组合配置
- **分红数据追踪** - 股息率、分红频率等关键指标

### 💱 汇率数据管理
- **实时汇率** - USD/CNY、HKD/CNY等主要货币对
- **自动同步** - 定时任务自动更新汇率数据
- **货币转换** - 支持多种货币间的换算功能

## 🛠️ 技术栈

### 后端 (Go)
| 技术 | 版本 | 用途 |
|------|------|------|
| Go | >= 1.21 | 核心语言 |
| Gin | v1.12.0 | Web 框架 |
| GORM | v1.30.0 | ORM 框架 (SQLite/PostgreSQL) |
| go-cache | v2.1.0 | 内存缓存 |
| cron/v3 | v3.0.1 | 定时任务调度 |

### 前端 (React)
| 技术 | 版本 | 用途 |
|------|------|------|
| React | ^19.2.4 | UI 框架 |
| TypeScript | ^5.x | 类型安全 |
| Vite | latest | 构建工具 |
| Ant Design | ^6.3.4 | UI 组件库 |
| ECharts | ^6.0.0 | 数据可视化 |
| Recharts | ^3.8.1 | 图表组件 |
| React Router | ^7.13.2 | 路由管理 |

### 数据存储
- **SQLite** - 默认本地数据库（开发环境）
- **PostgreSQL** - 生产数据库支持

## 🚀 快速开始

### 方式一：一键启动（推荐）

```bash
# 克隆项目
git clone https://github.com/coder100001/ETF-Insight.git
cd ETF-Insight

# macOS / Linux
chmod +x start.sh
./start.sh

# Windows
start.bat
```

启动脚本会自动完成以下操作：
1. ✅ 检查运行环境（Go、Node.js）
2. ✅ 安装后端依赖（go mod download）
3. ✅ 编译后端项目
4. ✅ 安装前端依赖（npm install）
5. ✅ 启动后端服务（端口 8080）
6. ✅ 启动前端服务（端口 5173）

### 方式二：Docker 部署

```bash
git clone https://github.com/coder100001/ETF-Insight.git
cd ETF-Insight
docker-compose up -d
```

访问 http://localhost:8080

### 方式三：手动启动

```bash
# 后端
cd backend
go mod download
go build -o etf-insight .
./etf-insight

# 新终端 - 前端
cd frontend
npm install
npm run dev
```

## 📋 环境要求

| 工具 | 最低版本 | 推荐版本 |
|------|----------|----------|
| Go | 1.21+ | 1.25+ |
| Node.js | 18+ | 20+ |
| npm | 9+ | 10+ |

## 🔧 配置说明

后端配置文件位于 `backend/config.yaml`：

```yaml
server:
  port: 8080        # 服务端口
  mode: release     # 运行模式: debug/release

database:
  dsn: "etf_insight.db"  # SQLite 数据库路径

schedule:
  exchange_rate_sync: "0 */5 * * * *"    # 汇率同步频率
  etf_update: "0 30 10 * * *"           # ETF 数据更新时间
```

## 📁 项目结构

```
ETF-Insight/
├── start.sh              # 一键启动脚本 (macOS/Linux)
├── start.bat             # 一键启动脚本 (Windows)
├── backend/
│   ├── main.go           # 入口文件
│   ├── config/           # 配置管理
│   ├── handlers/         # API 处理器
│   │   ├── exchange_rate.go      # 汇率接口
│   │   ├── a_share_portfolio_handler.go  # A股组合接口
│   │   └── etf_handler.go       # ETF 接口
│   ├── models/           # 数据模型
│   ├── services/         # 业务逻辑
│   │   └── exchange_rate.go      # 汇率服务
│   ├── tasks/            # 定时任务
│   ├── cmd/              # 命令行工具
│   └── config.yaml       # 配置文件
├── frontend/
│   ├── src/
│   │   ├── pages/        # 页面组件
│   │   │   ├── ExchangeRate.tsx      # 汇率页面
│   │   │   ├── ASharePortfolio.tsx  # A股组合页面
│   │   │   └── Dashboard.tsx        # 仪表盘
│   │   ├── components/  # 公共组件
│   │   └── services/     # API 服务
│   └── package.json
├── docker-compose.yml    # Docker 编排
└── Dockerfile            # Docker 构建
```

## 🌐 API 接口

### 汇率相关
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/exchange-rates` | 获取汇率列表 |
| GET | `/api/exchange-rates/:from/:to` | 获取指定汇率 |
| POST | `/api/exchange-rates/sync` | 触发汇率同步 |
| POST | `/api/exchange-rates/convert` | 货币转换 |

### A股组合相关
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/a-share/portfolio` | 获取投资组合 |
| GET | `/api/a-share/holdings` | 获取持仓明细 |

### ETF 相关
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/etfs` | 获取 ETF 列表 |
| GET | `/api/etfs/:symbol` | 获取 ETF 详情 |

## 📖 使用指南

### 启动项目

```bash
# 一键启动（推荐）
./start.sh

# 或手动启动
cd backend && ./etf-insight &
cd frontend && npm run dev
```

### 访问地址

- **前端界面**: http://localhost:5173
- **后端API**: http://localhost:8080
- **健康检查**: http://localhost:8080/health

### 常见问题

**Q: 端口被占用怎么办？**

修改 `backend/config.yaml` 中的端口配置，或停止占用端口的进程：

```bash
# macOS/Linux
lsof -i :8080 | grep LISTEN | awk '{print $2}' | xargs kill

# Windows
netstat -ano | findstr :8080
taskkill /PID <进程ID> /F
```

**Q: 依赖安装失败？**

国内用户可设置代理：
```bash
export GOPROXY=https://goproxy.cn,direct
npm config set registry https://registry.npmmirror.com
```

## 🗺️ 开发路线图

### Phase 1: 基础功能 ✅
- [x] ETF 基础信息管理
- [x] 实时行情数据展示
- [x] 汇率数据管理
- [x] A股红利ETF组合
- [x] 投资占比饼状图

### Phase 2: 深度分析 🚧
- [ ] 持仓重叠分析
- [ ] 行业/地区分布可视化
- [ ] 风险指标计算
- [ ] 相关性矩阵

### Phase 3: 组合优化 📋
- [ ] 投资组合构建器
- [ ] 有效前沿分析
- [ ] 再平衡策略建议

### Phase 4: 高级功能 📋
- [ ] 智能推荐系统
- [ ] 历史回测功能
- [ ] 投资报告导出
- [ ] 移动端适配

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

## 📄 License

MIT License
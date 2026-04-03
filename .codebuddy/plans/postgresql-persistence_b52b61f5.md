---
name: postgresql-persistence
overview: 将 ETF-Insight 后端从 MockDB 内存存储迁移到 PostgreSQL 持久化存储，使用 GORM 作为 ORM 框架，保持现有 API 接口和业务逻辑不变。
todos:
  - id: add-gorm-deps
    content: 在 go.mod 中添加 gorm.io/gorm 和 gorm.io/driver/postgres 依赖并执行 go mod tidy
    status: completed
  - id: refactor-db-layer
    content: 重写 models/db.go：将 MockDB 替换为 gorm.DB，实现 InitDB(cfg)、AutoMigrate、InitDefaultData
    status: completed
    dependencies:
      - add-gorm-deps
  - id: update-config
    content: 扩展 config.DatabaseConfig 增加 Host/Port/User/Password/DBName 字段，支持 PG DSN 构建
    status: completed
  - id: update-init-main
    content: 修改 main.go 中 InitDB 调用，传入数据库 DSN 配置
    status: completed
    dependencies:
      - refactor-db-layer
      - update-config
  - id: update-services
    content: 修改 scheduler.go 和 exchange_rate.go 中的数据写入为 upsert 避免重复
    status: completed
    dependencies:
      - refactor-db-layer
  - id: update-docker
    content: 更新 docker-compose.yml 添加 PostgreSQL 16 服务并配置 backend 连接
    status: completed
    dependencies:
      - update-config
---

## 用户需求

将 ETF-Insight 后端的数据存储层从内存 MockDB 替换为 PostgreSQL 持久化数据库，保持现有功能和 API 接口不变。

## 产品概述

当前项目所有数据存储在内存中（MockDB），应用重启后数据全部丢失。需要引入 PostgreSQL 数据库，使 ETF 配置、历史价格、汇率、操作日志、A 股组合等数据能够持久化。

## 核心功能

- 引入 GORM ORM 和 PostgreSQL 驱动，替换内存 MockDB
- 复用现有模型已有的 gorm struct 标签，自动建表迁移
- 保持 `models.DB` 全局变量接口不变，所有 handler/service 链式调用无需修改
- 更新配置结构支持 PostgreSQL 连接参数（Host/Port/User/Password/DBName）
- 更新 docker-compose.yml 添加 PostgreSQL 服务
- 应用启动时初始化默认种子数据

## 技术栈

- **ORM**: GORM v2 (`gorm.io/gorm`)
- **PostgreSQL 驱动**: `gorm.io/driver/postgres`（底层使用 `lib/pq`）
- **数据库**: PostgreSQL 16
- **配置**: 复用现有 `config.DatabaseConfig`，扩展连接字段

## 实现方案

### 核心策略

将 `models.DB` 的类型从 `*MockDB` 替换为 `*gorm.DB`。由于所有调用方使用的是 GORM 风格的链式 API（`Where().First().Find().Updates().Save().Delete()`），而 MockDB 当初就是模拟 GORM 接口设计的，所以只需将底层实现换成真正的 GORM 实例即可。

### 关键技术决策

1. **使用 GORM 而非原生 database/sql**: 模型已有完整 gorm 标签，GORM 可直接自动迁移建表，最大化代码复用
2. **类型替换为 `*gorm.DB`**: 保持 `models.DB` 变量名不变，调用方代码零修改
3. **`InitDB` 接收配置参数**: 从 `main.go` 传入数据库 DSN，支持环境变量覆盖
4. **Upsert 种子数据**: `InitDefaultData` 使用 GORM 的 `FirstOrCreate` 避免重复插入
5. **Dockerfile 保持 CGO_ENABLED=1**: `lib/pq` 需要 CGO，当前已配置

### 数据流

```
应用启动 → InitDB(dsn) → gorm.Open(postgres.Open(dsn)) → AutoMigrate(所有模型) → InitDefaultData
                                                     ↓
Handler/Service → models.DB.Where(...).First(...) → GORM → PostgreSQL
```

## 实现细节

- **数据库连接配置**: 新增 `Host`, `Port`, `User`, `Password`, `DBName` 字段，DSN 格式为 `host=xxx port=xxx user=xxx password=xxx dbname=xxx sslmode=disable`
- **AutoMigrate 模型列表**: ETFConfig, ETFData, OperationLog, ExchangeRate, AShareDividendETF, AShareETFPortfolio, ASharePortfolioHolding
- **历史数据去重**: scheduler 中批量写入 ETFData 时，使用 `symbol + date` 联合唯一索引 + `clause.OnConflict` 做 upsert
- **汇率数据去重**: exchange_rate.go 中 `from_currency + to_currency + rate_date` 加联合唯一索引

## 架构设计

不需要新的架构模式，仅替换数据层实现。修改集中在 `models/db.go` 和 `config/config.go`。

## 目录结构

```
backend/
├── config/
│   └── config.go          # [MODIFY] 扩展 DatabaseConfig，增加 PostgreSQL 连接字段
├── models/
│   ├── db.go              # [MODIFY] MockDB → gorm.DB，实现真正的 InitDB/AutoMigrate/InitDefaultData
│   ├── models.go          # [MODIFY] 删除 MockDB 相关方法（已移到 db.go 重写）
│   └── a_share_dividend_etf.go  # [MODIFY] 无需改动，gorm 标签已完整
├── go.mod                 # [MODIFY] 添加 gorm 和 postgres 驱动依赖
├── go.sum                 # [MODIFY] 自动更新
├── main.go                # [MODIFY] InitDB 传入数据库 DSN
├── services/
│   └── exchange_rate.go   # [MODIFY] 历史数据 upsert 逻辑
├── tasks/
│   └── scheduler.go       # [MODIFY] 批量写入使用 upsert
docker-compose.yml         # [MODIFY] 添加 PostgreSQL 服务，更新 backend 环境变量
Dockerfile                 # [MODIFY] 构建阶段安装 CGO 依赖（已有）
```
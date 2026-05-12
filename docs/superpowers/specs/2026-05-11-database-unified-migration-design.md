# Design Doc 006: 数据库模型统一与数据迁移方案

## 元数据

- **编号**: 006
- **标题**: 数据库模型统一与数据迁移方案
- **状态**: draft
- **创建日期**: 2026-05-11
- **最后更新**: 2026-05-11
- **关联任务**: ETF-Insight 数据库设计优化
- **复杂度级别**: L3
- **涉及端**: 后端
- **前置 Design Doc**: 005 (export-report-design)

---

## 1. 背景与动机

### 1.1 为什么需要这个改动

经过全维度数据流审计，发现数据库模型层存在以下核心问题：

- **核心概念重复定义**: ETF 概念 4 张表、价格 4 张表、持仓 5 张表、组合 3 张表，对应同一业务含义的数据分散在多张表中，互不知晓
- **新旧模型脱节**: 新统一模型表 (`assets` / `prices` / `holdings` / `portfolios`) 已通过 GORM AutoMigrate 创建了空表结构，但从未写入数据。所有真实数据仍然在旧表中
- **字段精度不统一**: 同类型字段在不同表中使用不同精度 (decimal(10,3) vs decimal(20,8))
- **类型错误**: `ETFConfig.Inception` 使用 `string` 而非 `time.Time`
- **查询路径低效**: 美股 ETF 数据在 `etf_configs` + `etf_data`，A 股 ETF 数据在 `a_share_dividend_etfs`，前端需要两套查询逻辑

### 1.2 业务驱动因素

- 后续功能（Alpha+BL 闭环、跨资产组合分析）需要统一的 ETF 数据模型
- 新增数据源时需要一套统一的 Provider 接口，依赖统一的数据库模型
- 当前 59 张表中仅 15 张有实际数据，大量空表增加维护负担

---

## 2. 调研与现状分析

### 2.1 现有实现

**数据库类型**: SQLite（默认），也可切换到 PostgreSQL

**59 张表分为三类**:

| 类别 | 表数 | 说明 |
|------|------|------|
| 旧模型表（有数据） | 10 张 | `etf_configs`, `a_share_dividend_etfs`, `etf_data`, `portfolio_configs`, `a_share_etf_portfolios`, `a_share_portfolio_holdings`, `exchange_rates`, `exchange_rate_sync_details`, `exchange_rate_sync_logs`, `currency_pairs` |
| 新统一模型表（空） | 5 张 | `assets`, `prices`, `holdings`, `portfolios`, `portfolio_positions` |
| 其他零数据空表 | 44 张 | 因子、Alpha 观点、风险预算、插件、报告等模块的表 |

**实际数据量统计**:

| 表名 | 行数 | 是否存在风险 |
|------|------|-------------|
| `etf_configs` | 2 | 低（仅 QQQ/SCHD） |
| `a_share_dividend_etfs` | 8 | 低 |
| `etf_data` | 4 | 低（Mock 数据） |
| `portfolio_configs` | 3 | 低 |
| `a_share_etf_portfolios` | 1 | 低 |
| `a_share_portfolio_holdings` | 8 | 低 |
| `exchange_rate_sync_details` | 1400 | 低（保留不迁移） |
| `exchange_rate_sync_logs` | 200 | 低（保留不迁移） |

核心业务数据总量不到 30 行，迁移风险可控。

**迁移机制现状**:

- 使用 GORM `AutoMigrate()` 自动创建表结构（[models/db.go:L58-L124](file:///Users/liunian/Desktop/dnmp/py_project/backend/models/db.go#L58-L124)）
- 另有一套手动 SQL 迁移脚本（`backend/migrations/001-005*.sql`），但仅覆盖因子/Alpha/风险预算/插件/报告 5 个模块，与核心业务表无关
- **双重迁移机制并存**：GORM AutoMigrate 创建全部表，SQL 脚本覆盖部分表，两者一致性难以保证

### 2.2 业界实践

- **GORM AutoMigrate + 版本化迁移**: 业界推荐使用 `golang-migrate` 或 `pressly/goose` 做版本化迁移，配合 GORM 的 AutoMigrate 做开发期快速迭代
- **单表继承 (Single Table Inheritance)**: 使用 `type` 字段区分不同类型的资产，避免每类资产一张表
- **渐进式迁移**: 旧表只读不写，新表双写，观察期后废弃旧表

### 2.3 技术约束

- 当前使用 SQLite，迁移脚本需兼容 SQLite 和 PostgreSQL
- GORM AutoMigrate 不支持 SQLite 的 `ALTER COLUMN`，字段类型变更需要新建表 + 数据迁移
- 项目中大量 handler 直接使用 `models.DB` 操作旧表，迁移后需同步更新代码

---

## 3. 可选方案

### 方案 A: 一次性迁移（推荐）

**描述**: 编写一个迁移命令，在一个事务中将所有旧表数据映射到新表，删除旧表，更新代码中的数据库引用。

**优点**:
- 干脆利落，迁移后系统只维护一套模型
- 新表字段更丰富（如 `prices` 含 MA/调整后价格/质量评分）

**缺点**:
- 迁移期间需要停机或只读
- 代码修改面较大（handler 中 19 处 `models.DB` 旧表引用需同步修改）

**工作量**: 中

### 方案 B: 渐进式迁移 + 视图兼容

**描述**: 创建数据库视图 (View) 使旧表结构兼容新表，新旧代码同时运行，逐步迁移查询路径。

**优点**:
- 零停机风险
- 可逐步验证

**缺点**:
- SQLite 视图功能有限，复杂字段映射困难
- 维护期较长（新旧两套并存）

**工作量**: 大

### 方案 C: 仅废弃不迁移

**描述**: 旧表仅标记为废弃，新表从零开始写入，数据通过新代码逻辑逐步补齐。

**优点**:
- 零迁移风险
- 不需要回滚方案

**缺点**:
- 旧数据永久丢失（当前 30 行数据可接受）
- 需要重新录入 ETF 信息

**工作量**: 小

---

## 4. 决策

- **选定方案**: 方案 A（一次性迁移）
- **决策理由**:
  1. 核心业务数据仅 30 行，迁移风险极低
  2. 一次性完成比长期维护两套模型更经济
  3. 旧数据多为 Mock/默认数据，丢失可重新生成
  4. 方案 B 在 SQLite 上实现困难
- **权衡取舍**: 需要停机短时间迁移，但当前无生产用户，停机不是问题

---

## 5. 后端设计

### 5.1 新增模型

在现有 `assets` 表基础上，新增 `etf_profiles` 表存储 ETF 扩展属性：

```go
// models/etf_profile.go
type ETFProfile struct {
    ID               uint            `json:"id" gorm:"primaryKey"`
    AssetID          uint            `json:"asset_id" gorm:"uniqueIndex;not null"`
    ExpenseRatio     decimal.Decimal `json:"expense_ratio" gorm:"type:decimal(5,4)"`
    ManagementFee    decimal.Decimal `json:"management_fee" gorm:"type:decimal(5,4)"`
    AUM              decimal.Decimal `json:"aum" gorm:"type:decimal(20,2)"`
    Strategy         string          `json:"strategy" gorm:"size:100"`
    Focus            string          `json:"focus" gorm:"size:100"`
    Category         string          `json:"category" gorm:"size:50"`
    Provider         string          `json:"provider" gorm:"size:100"`
    Benchmark        string          `json:"benchmark" gorm:"size:100"`
    AssetClass       string          `json:"asset_class" gorm:"size:50"`
    Region           string          `json:"region" gorm:"size:50"`
    ETFType          string          `json:"etf_type" gorm:"size:50"`
    DividendYieldMin decimal.Decimal `json:"dividend_yield_min" gorm:"type:decimal(5,2)"`
    DividendYieldMax decimal.Decimal `json:"dividend_yield_max" gorm:"type:decimal(5,2)"`
    DividendFrequency string         `json:"dividend_frequency" gorm:"size:20"`
    InceptionDate    *time.Time      `json:"inception_date"`
    Description      string          `json:"description" gorm:"size:500"`
    CreatedAt        time.Time       `json:"created_at"`
    UpdatedAt        time.Time       `json:"updated_at"`
}
```

### 5.2 迁移脚本设计

**迁移命令**: `go run cmd/migrate_unified/main.go`

**执行顺序 DAG**:

```
Step 1: 迁移组合 (portfolio_configs + a_share_etf_portfolios → portfolios)
            ↓
Step 2: 迁移 ETF (etf_configs + a_share_dividend_etfs → assets + etf_profiles)
            ↓
Step 3: 迁移价格 (etf_data → prices)          (可并行)
Step 4: 迁移持仓 (a_share_portfolio_holdings → portfolio_positions)  (可并行)
            ↓
Step 5: 更新 AutoMigrate 注册表, 移除旧模型
```

#### Step 1: 组合迁移

```go
func migratePortfolios(tx *gorm.DB) error {
    // 1. portfolio_configs → portfolios
    var configs []PortfolioConfig
    tx.Find(&configs)
    for _, c := range configs {
        p := Portfolio{
            Name:            c.Name,
            Description:     c.Description,
            PortfolioType:   "config",
            Status:          "active",
            IsDefault:       c.IsDefault,
            InitialCapital:  c.TotalInvestment,
            BaseCurrency:    "USD",
            TargetAllocation: json.RawMessage(c.Allocation),
        }
        if err := tx.Create(&p).Error; err != nil {
            return fmt.Errorf("迁移组合 %s 失败: %w", c.Name, err)
        }
    }

    // 2. a_share_etf_portfolios → portfolios
    var aPortfolios []AShareETFPortfolio
    tx.Find(&aPortfolios)
    for _, ap := range aPortfolios {
        p := Portfolio{
            Name:           ap.Name,
            Description:    ap.Description,
            PortfolioType:  "a_share_dividend",
            Status:         "active",
            IsDefault:      ap.IsDefault,
            InitialCapital: ap.TotalInvestment,
            BaseCurrency:   "CNY",
        }
        if err := tx.Create(&p).Error; err != nil {
            return fmt.Errorf("迁移A股组合 %s 失败: %w", ap.Name, err)
        }
    }
    return nil
}
```

#### Step 2: ETF 迁移

```go
func migrateETFs(tx *gorm.DB) error {
    oldIDMap := make(map[uint]uint) // a_share_dividend_etfs.id → assets.id

    // 2a. a_share_dividend_etfs → assets + etf_profiles
    var aShareETFs []AShareDividendETF
    tx.Find(&aShareETFs)
    for _, e := range aShareETFs {
        asset := Asset{
            Symbol:   e.Symbol,
            Name:     e.Name,
            Type:     "etf",
            Currency: "CNY",
            Exchange: e.Exchange,
            Status:   1,
        }
        if err := tx.Create(&asset).Error; err != nil {
            return fmt.Errorf("创建资产 %s 失败: %w", e.Symbol, err)
        }
        oldIDMap[e.ID] = asset.ID

        profile := ETFProfile{
            AssetID:          asset.ID,
            ManagementFee:    e.ManagementFee,
            DividendYieldMin: e.DividendYieldMin,
            DividendYieldMax: e.DividendYieldMax,
            DividendFrequency: e.DividendFrequency,
            Benchmark:        e.Benchmark,
            Description:      e.Description,
        }
        if err := tx.Create(&profile).Error; err != nil {
            return fmt.Errorf("创建ETF配置 %s 失败: %w", e.Symbol, err)
        }
    }

    // 2b. etf_configs → assets + etf_profiles
    var configs []ETFConfig
    tx.Find(&configs)
    for _, c := range configs {
        asset := Asset{
            Symbol:   c.Symbol,
            Name:     c.Name,
            Type:     "etf",
            Currency: c.Currency,
            Exchange: c.Exchange,
            Status:   1,
        }
        if err := tx.Create(&asset).Error; err != nil {
            return fmt.Errorf("创建资产 %s 失败: %w", c.Symbol, err)
        }

        profile := ETFProfile{
            AssetID:       asset.ID,
            ExpenseRatio:  c.ExpenseRatio,
            AUM:           c.AUM,
            Strategy:      c.Strategy,
            Focus:         c.Focus,
            Category:      c.Category,
            Provider:      c.Provider,
            Description:   c.Description,
        }
        if err := tx.Create(&profile).Error; err != nil {
            return fmt.Errorf("创建ETF配置 %s 失败: %w", c.Symbol, err)
        }
    }

    return nil
}
```

#### Step 3: 价格迁移

```go
func migratePrices(tx *gorm.DB) error {
    var dataList []ETFData
    tx.Find(&dataList)
    for _, d := range dataList {
        var asset Asset
        if err := tx.Where("symbol = ?", d.Symbol).First(&asset).Error; err != nil {
            continue // ETF 未迁移，跳过
        }
        price := Price{
            AssetID:    asset.ID,
            Symbol:     d.Symbol,
            Date:       d.Date,
            PriceType:  "daily",
            Open:       d.OpenPrice,
            High:       d.HighPrice,
            Low:        d.LowPrice,
            Close:      d.ClosePrice,
            Volume:     int64(d.Volume),
            DataSource: d.DataSource,
            IsValid:    true,
            IsImputed:  false,
        }
        if err := tx.Create(&price).Error; err != nil {
            return fmt.Errorf("迁移价格数据 %s/%s 失败: %w", d.Symbol, d.Date, err)
        }
    }
    return nil
}
```

#### Step 4: 持仓迁移

```go
func migrateHoldings(tx *gorm.DB, oldETFIDMap map[uint]uint, oldPortfolioIDMap map[uint]uint) error {
    var holdings []ASharePortfolioHolding
    tx.Find(&holdings)
    for _, h := range holdings {
        newPortfolioID, ok := oldPortfolioIDMap[h.PortfolioID]
        if !ok {
            continue
        }
        newAssetID, ok := oldETFIDMap[h.EtfID]
        if !ok {
            continue
        }
        pos := PortfolioPosition{
            PortfolioID:  newPortfolioID,
            AssetID:      newAssetID,
            MarketValue:  h.Investment,
            Weight:       h.Weight,
            IsActive:     true,
        }
        if err := tx.Create(&pos).Error; err != nil {
            return fmt.Errorf("迁移持仓失败: %w", err)
        }
    }
    return nil
}
```

### 5.3 旧表废弃策略

迁移完成后，执行：

```go
// 从 AutoMigrate 注册表中移除旧模型
// 但暂时不 DROP TABLE，保留以作回滚
// 标记为 readonly 模式
```

**废弃计划**:

| 旧表 | 操作 | 时间点 |
|------|------|--------|
| `etf_configs` | 标记废弃，保留 30 天 | 迁移完成后 |
| `etf_data` | 标记废弃，保留 30 天 | 迁移完成后 |
| `a_share_dividend_etfs` | 标记废弃，保留 30 天 | 迁移完成后 |
| `a_share_etf_portfolios` | 标记废弃，保留 30 天 | 迁移完成后 |
| `a_share_portfolio_holdings` | 标记废弃，保留 30 天 | 迁移完成后 |
| `portfolio_configs` | 标记废弃，保留 30 天 | 迁移完成后 |

### 5.4 代码同步

迁移完成后，同步修改以下代码路径：

| 文件 | 修改内容 | 风险 |
|------|---------|------|
| `handlers/etf_handler.go` | 19 处 `models.DB` 旧表引用改为新表 | 高 |
| `handlers/a_share_data_handler.go` | 旧表引用改为 `assets` + `etf_profiles` | 中 |
| `handlers/a_share_portfolio_handler.go` | 旧表引用改为 `portfolios` + `portfolio_positions` | 中 |
| `services/ashare/etf_data_service.go` | 写入目标从旧表改为新表 | 中 |

### 5.5 索引设计

新表已有索引之外，建议补充：

```sql
-- ETF 查询高频
CREATE INDEX idx_etf_profiles_asset_id ON etf_profiles(asset_id);
CREATE INDEX idx_etf_profiles_asset_class ON etf_profiles(asset_class);
-- 价格查询
CREATE INDEX idx_prices_price_type ON prices(price_type);
-- 持仓穿透
CREATE INDEX idx_portfolio_positions_active ON portfolio_positions(portfolio_id, is_active);
```

---

## 6. 影响范围

| 模块 | 影响 | 变更类型 |
|------|------|---------|
| `models/db.go` | 从 AutoMigrate 移除旧模型，注册 `ETFProfile` | 修改 |
| `models/etf_profile.go` | 新增模型文件 | 新增 |
| `handlers/etf_handler.go` | 旧表→新表引用替换（19 处） | 修改 |
| `handlers/a_share_data_handler.go` | 旧表→新表引用替换 | 修改 |
| `handlers/a_share_portfolio_handler.go` | 旧表→新表引用替换 | 修改 |
| `services/ashare/etf_data_service.go` | 写入目标改为新表 | 修改 |
| `router/router.go` | 新增 `/api/migrate` 端点 | 新增 |
| `docs/superpowers/specs/README.md` | 添加 006 编号 | 修改 |

---

## 7. 迁移验证

### 7.1 数据一致性校验

迁移后运行以下校验：

| 校验项 | SQL 语句 | 预期结果 |
|--------|---------|---------|
| 资产总数 | `SELECT COUNT(*) FROM assets;` | = `etf_configs` + `a_share_dividend_etfs` 行数之和 |
| ETF 配置数 | `SELECT COUNT(*) FROM etf_profiles;` | 等于 `assets WHERE type='etf'` 行数 |
| 价格迁移完整 | `SELECT COUNT(*) FROM prices;` | 等于 `etf_data` 行数 |
| 持仓关联完整 | `SELECT COUNT(*) FROM portfolio_positions;` | 等于 `a_share_portfolio_holdings` 行数 |
| 组合迁移完整 | `SELECT COUNT(*) FROM portfolios;` | 等于 `portfolio_configs + a_share_etf_portfolios` 行数 |

### 7.2 功能回归测试

- 美股 ETF 列表 API (`GET /api/etf`) 返回新表数据
- A 股 ETF 列表 API (`GET /api/a-share/search`) 返回新表数据
- 价格查询 API 返回新表数据
- 组合查询 API 返回新表数据

---

## 8. 回滚策略

### 8.1 数据库回滚

```bash
# 回滚命令
cd backend && go run cmd/rollback_unified/main.go
```

回滚脚本逻辑：

```go
func rollback() error {
    // 1. 验证旧表是否还存在 (未被 DROP)
    // 2. DELETE 新表中的数据
    // 3. 从 AutoMigrate 恢复旧模型注册
    // 4. 回滚 handler 代码到迁移前版本
    return nil
}
```

### 8.2 代码回滚

```bash
git revert HEAD
```

### 8.3 保留期

| 资源 | 保留期 | 到期操作 |
|------|--------|---------|
| 旧表 | 30 天 | 确认无误后 DROP |
| 迁移脚本 | 永久保留 | 归档到 `backend/migrations/` |
| 旧代码版本 | 永久保留 | Git 历史 |

---

## 9. 开放问题

- [ ] `portfolio_configs.allocation` (JSON string) 到 `portfolios.target_allocation` (JSON type) 的 GORM 兼容性需验证
- [ ] handler 中的 `models.DB` 旧表引用替换后，需确认所有查询路径正确
- [ ] 迁移脚本是否需要 `--dry-run` 模式？（建议有）
- [ ] `etf_profiles` 是否还需要存储 `dividend_yield_min/max` 的历史变更记录？

---

**文档状态**: draft
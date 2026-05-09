# 数据库迁移指南

## 概述

本目录包含ETF-Insight数据层改造的数据库迁移脚本，用于支持v2.7到v2.8版本的演进。

## 迁移阶段

### 阶段一：因子数据层（P0）
- **文件**: `001_add_factor_tables.sql`
- **执行时间**: 2026-Q4 第1-2周
- **内容**:
  - 因子数据表（factor_data）
  - 因子择时信号表（factor_timing_signals）

### 阶段二：Alpha观点层（P0）
- **文件**: `002_add_alpha_view_tables.sql`
- **执行时间**: 2026-Q4 第3-4周
- **内容**:
  - Alpha观点表（alpha_views）
  - Alpha观点表现表（alpha_view_performances）
  - Black-Litterman配置表（black_litterman_configs）
  - BL后验收益表（bl_posterior_returns）

### 阶段三：风险预算层（P0）
- **文件**: `003_add_risk_budget_tables.sql`
- **执行时间**: 2026-Q4 第5-6周
- **内容**:
  - 风险预算配置表（risk_budget_configs）
  - 蒙特卡洛模拟表（monte_carlo_simulations）
  - 风险贡献表（risk_contributions）
  - 风险预算执行记录表（risk_budget_executions）

### 阶段四：插件架构层（P1）
- **文件**: `004_add_plugin_tables.sql`
- **执行时间**: 2027-Q1 第1-2周
- **内容**:
  - 插件注册表（plugin_registries）
  - 插件配置表（plugin_configurations）
  - 插件执行日志表（plugin_execution_logs）
  - 模型基准对比矩阵表（model_benchmark_matrices）
  - 策略实验表（strategy_experiments）

## 执行方法

### SQLite数据库

```bash
# 进入backend目录
cd backend

# 执行单个迁移脚本
sqlite3 etf_insight.db < migrations/001_add_factor_tables.sql

# 或批量执行所有迁移
for file in migrations/*.sql; do
    echo "Executing $file..."
    sqlite3 etf_insight.db < "$file"
done
```

### PostgreSQL数据库

```bash
# 设置数据库连接
export DB_DSN="host=localhost user=postgres password=yourpassword dbname=etf_insight sslmode=disable"

# 执行单个迁移脚本
psql $DB_DSN -f migrations/001_add_factor_tables.sql

# 或批量执行所有迁移
for file in migrations/*.sql; do
    echo "Executing $file..."
    psql $DB_DSN -f "$file"
done
```

## 使用GORM AutoMigrate

迁移脚本已集成到GORM的AutoMigrate中，启动应用时会自动执行：

```bash
cd backend
go run main.go
```

## 验证迁移

### 检查表是否创建成功

```sql
-- SQLite
SELECT name FROM sqlite_master WHERE type='table' AND name LIKE '%factor%' OR name LIKE '%alpha%' OR name LIKE '%risk%' OR name LIKE '%plugin%';

-- PostgreSQL
SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND (tablename LIKE '%factor%' OR tablename LIKE '%alpha%' OR tablename LIKE '%risk%' OR tablename LIKE '%plugin%');
```

### 检查索引

```sql
-- SQLite
SELECT name, tbl_name FROM sqlite_master WHERE type='index' AND (tbl_name LIKE '%factor%' OR tbl_name LIKE '%alpha%' OR tbl_name LIKE '%risk%' OR tbl_name LIKE '%plugin%');

-- PostgreSQL
SELECT indexname, tablename FROM pg_indexes WHERE schemaname = 'public' AND (tablename LIKE '%factor%' OR tablename LIKE '%alpha%' OR tablename LIKE '%risk%' OR tablename LIKE '%plugin%');
```

## 回滚策略

如果迁移失败，可以手动删除相关表：

```sql
-- 删除因子数据层表
DROP TABLE IF EXISTS factor_timing_signals;
DROP TABLE IF EXISTS factor_data;

-- 删除Alpha观点层表
DROP TABLE IF EXISTS bl_posterior_returns;
DROP TABLE IF EXISTS black_litterman_configs;
DROP TABLE IF EXISTS alpha_view_performances;
DROP TABLE IF EXISTS alpha_views;

-- 删除风险预算层表
DROP TABLE IF EXISTS risk_budget_executions;
DROP TABLE IF EXISTS risk_contributions;
DROP TABLE IF EXISTS monte_carlo_simulations;
DROP TABLE IF EXISTS risk_budget_configs;

-- 删除插件架构层表
DROP TABLE IF EXISTS strategy_experiments;
DROP TABLE IF EXISTS model_benchmark_matrices;
DROP TABLE IF EXISTS plugin_execution_logs;
DROP TABLE IF EXISTS plugin_configurations;
DROP TABLE IF EXISTS plugin_registries;
```

## 注意事项

1. **数据安全**: 迁移脚本仅创建新表，不会删除或修改现有数据
2. **执行顺序**: 必须按照001、002、003、004的顺序执行
3. **备份**: 执行迁移前建议备份数据库
4. **验证**: 每个阶段执行后都要验证数据完整性

## 相关文档

- [数据层演进改造方案](../docs/reference/DATA_LAYER_EVOLUTION_PLAN.md)
- [数据层实施指南](../docs/reference/DATA_LAYER_IMPLEMENTATION_GUIDE.md)
- [演进路线图](../docs/reference/EVOLUTION_ROADMAP_2026.md)

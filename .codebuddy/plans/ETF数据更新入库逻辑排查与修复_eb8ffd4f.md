---
name: ETF数据更新入库逻辑排查与修复
overview: 排查并修复ETF数据更新入库混乱、价格错误的问题，核心修复5个关键Bug：Finage数据转换丢失OHLCV、Fallback模拟数据不准确、change计算逻辑错误、update_etf_prices基于价格范围推断symbol导致错乱、多处硬编码模拟数据未清理
todos:
  - id: fix-finage-provider
    content: 重写 FinageProvider 使用聚合API获取完整OHLCV数据
    status: completed
  - id: fix-scheduler
    content: 修复 Scheduler.updateETFData 删除硬编码ETF列表
    status: completed
  - id: fix-change-calc
    content: 修复 GetETFList/GetETFRealtime 的 change 和 previous_close 计算逻辑
    status: completed
  - id: fix-update-prices-cli
    content: 修复 update_etf_prices 工具，改为逐个symbol请求聚合API
    status: completed
  - id: fix-fallback-mock
    content: 更新 Fallback Provider 基准价格并清理所有硬编码mock数据
    status: completed
  - id: verify-rebuild
    content: 编译验证并测试API数据正确性
    status: completed
    dependencies:
      - fix-finage-provider
      - fix-scheduler
      - fix-change-calc
      - fix-update-prices-cli
      - fix-fallback-mock
---

## 用户需求

排查并修复 ETF 数据更新入库逻辑混乱问题，导致价格数据错误。

## 问题概述

项目有多条数据入库路径，每条路径都存在不同的 Bug，导致 `etf_data` 表和 API 返回的价格数据混乱。

## 核心问题

1. **Finage Provider 数据转换严重错误** - 只从 ask/bid 计算 midPrice，OHLCV 全部伪造，open=high=low=close=midPrice
2. **Fallback Provider 基准价格过时** - 多个 ETF 基准价格与实际市场价严重不符（如 SCHD 26.85 vs 实际 85+）
3. **GetETFList change/previous_close 计算错误** - 用 ClosePrice-OpenPrice 替代 ClosePrice-PreviousClose
4. **Scheduler updateETFData 硬编码只更新2只ETF** - 只写死了 QQQ 和 SCHD
5. **update_etf_prices 工具 inferSymbolFromData 极不可靠** - 用价格范围推断 symbol，VEA/VWO 重叠
6. **多处硬编码 mock 数据** - cache.go MockRealtimeData、UpdateRealtimeData realistic_mock_data 均为过时固定值

## 技术栈

- 后端: Go 1.24 + Gin + GORM + SQLite
- 前端: React + Vite + Ant Design
- 数据源: Finage API (主要) / Finnhub (备用) / Yahoo Finance (定时任务)

## 实现方案

### 核心策略

统一数据入库链路，修复每条路径的数据质量问题，消除所有硬编码模拟数据。

### 修复优先级和方案

#### 1. 重写 FinageProvider.convertToQuoteData (最严重)

- **问题**: Finage `/last/stock/` API 只返回 ask/bid/timestamp，convertToQuoteData 将 midPrice 填充所有 OHLCV 字段
- **方案**: 改用 Finage 聚合 API `/agg/stock/{symbol}/1/day/{from}/{to}` 获取完整 OHLCV 数据
- 新增 `FinageAggregateResponse` 结构体解析聚合数据
- `GetQuote` 逐个 symbol 调用聚合 API 获取最近1天数据
- `GetQuotes` 保持批量逻辑不变，内部调用新的 GetQuote

#### 2. 修复 Scheduler.updateETFData 硬编码

- **问题**: 第136行硬编码只更新 QQQ 和 SCHD，忽略数据库查询结果
- **方案**: 删除硬编码，使用从数据库读取的完整 etfConfigs 列表

#### 3. 修复 GetETFList/GetETFRealtime 的 change 计算

- **问题**: `change = ClosePrice - OpenPrice`，`previous_close = OpenPrice`，这是错误的
- **方案**: 从 etf_data 表查询前一日记录获取 previous_close；若无前日数据，标记 previous_close 为 0 并计算 change
- 新增辅助方法 `getPreviousClose(symbol, currentDate)`

#### 4. 修复 update_etf_prices 工具

- **问题**: `inferSymbolFromData` 按价格范围推断 symbol 不可靠
- **方案**: 改为逐个 symbol 调用聚合 API（`/agg/stock/{symbol}/1/day/...`），每个请求返回明确的 symbol 标识，无需推断

#### 5. 更新 Fallback Provider 基准价格

- **问题**: 多个基准价格严重过时
- **方案**: 更新为 2026年4月合理参考价格，并添加注释说明需要定期更新

#### 6. 清理硬编码 mock 数据

- **问题**: cache.go MockRealtimeData、etf_handler.go realistic_mock_data 均为固定假数据
- **方案**: 删除 MockRealtimeData 和 realistic_mock_data；当 Yahoo/Finage 不可用时，统一使用 Fallback Provider（其基准价格已更新）；CacheService.GetRealtimeData 在 Yahoo 失败时不再回退到硬编码 mock

## 目录结构

```
backend/
├── services/datasource/
│   ├── finage_provider.go        # [MODIFY] 重写 GetQuote 使用聚合API获取完整OHLCV
│   └── fallback_provider.go      # [MODIFY] 更新 defaultBasePrices 基准价格
├── handlers/
│   └── etf_handler.go            # [MODIFY] 修复change计算、删除realistic_mock_data
├── services/
│   └── cache.go                  # [MODIFY] 删除 MockRealtimeData，修改 GetRealtimeData 回退逻辑
├── tasks/
│   └── scheduler.go              # [MODIFY] 删除硬编码ETF列表，使用完整数据库查询
└── cmd/update_etf_prices/
    └── main.go                   # [MODIFY] 改为逐个symbol请求，删除inferSymbolFromData
```

## 实施注意事项

- Finage 聚合 API 日期格式需使用 `YYYY-MM-DD`，from 设为 T-7，to 设为 T-1 以确保获取最近交易日数据
- change/previous_close 修复需要一次额外 DB 查询获取前日收盘价，可考虑批量查询优化
- 删除 mock 数据后需确保 Fallback Provider 始终可用作为最终兜底
- Scheduler 中历史数据更新使用 goroutine 但无 WaitGroup，需修复并发控制
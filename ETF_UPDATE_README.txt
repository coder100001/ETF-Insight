# ETF数据更新脚本使用说明

## 功能特性

1. **先清除缓存，再更新数据** - 确保数据新鲜度
2. **双层缓存架构** - Redis持久化 + 内存缓存备用
3. **自动保存** - 同时更新MySQL和Redis

## Redis缓存架构

### 缓存策略
- **实时数据** (`etf:realtime:*`): 1小时过期
- **历史数据** (`etf:historical:*`): 持久化，不过期
- **指标数据** (`etf:metrics:*`): 持久化，不过期
- **对比数据** (`etf:comparison:*`): 1小时过期

### 数据库配置
- **Redis DB1**: 通用缓存（Session等）
- **Redis DB2**: ETF数据专用（持久化）

## 使用方法

### 1. 查看缓存统计
```bash
python update_etf_data.py --stats
```

### 2. 更新单个ETF
```bash
# 先清除SPYD缓存，再更新MySQL和Redis
python update_etf_data.py --symbol SPYD
```

### 3. 更新所有ETF
```bash
python update_etf_data.py --all
```

### 4. 仅清除缓存
```bash
# 清除指定ETF
python update_etf_data.py --symbol SCHD --clear-only

# 清除所有ETF
python update_etf_data.py --clear-only
```

## 缓存管理器API

### Python代码中使用

```python
from workflow.cache_manager import etf_cache

# 获取缓存
realtime = etf_cache.get_realtime('SPYD')
historical = etf_cache.get_historical('SPYD', '1y')
metrics = etf_cache.get_metrics('SPYD', '1y')

# 设置缓存
etf_cache.set_realtime('SPYD', data)
etf_cache.set_historical('SPYD', '1y', data)

# 清除缓存
etf_cache.clear_symbol('SPYD')  # 清除单个ETF
etf_cache.clear_all()           # 清除所有
etf_cache.clear_comparison()    # 清除对比数据

# 查看统计
stats = etf_cache.get_cache_stats()
```

## 数据流程

```
更新数据流程:
1. 清除Redis缓存 (先删除旧数据)
2. 从Yahoo Finance拉取最新数据
3. 保存到MySQL数据库
4. 写入Redis持久化缓存
5. 清除对比数据缓存（触发重新计算）
```

## 性能优化

1. **批量下载**: 使用 `yf.download()` 一次获取多个ETF
2. **并行请求**: ThreadPoolExecutor并发获取实时数据
3. **Redis持久化**: 历史数据永久缓存，减少API调用
4. **双层缓存**: Redis故障时自动降级到内存缓存

## 注意事项

1. **Redis必须启动**: 确保Redis服务运行在 `127.0.0.1:6379`
2. **API限流**: Yahoo Finance有频率限制，建议合理控制更新频率
3. **缓存更新**: 修改ETF标的后记得清除对应缓存
4. **持久化配置**: Redis建议配置RDB/AOF持久化

## 查看Redis数据

```bash
# 连接Redis
redis-cli -n 2

# 查看所有ETF键
KEYS etf:*

# 查看某个ETF的实时数据
GET etf:realtime:SPYD

# 查看某个ETF的历史数据
GET "etf:historical:SPYD_1y"

# 清空数据库
FLUSHDB
```

## 日志输出

脚本执行时会显示详细日志：
- 缓存清除情况
- MySQL保存记录数
- Redis写入状态
- 总耗时统计
- 成功/失败汇总

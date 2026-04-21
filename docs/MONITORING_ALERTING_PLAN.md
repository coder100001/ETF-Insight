# ETF-Insight 监控告警方案

**版本**: v1.0
**制定日期**: 2026-04-21
**生效日期**: 2026-05-01
**维护团队**: DevOps Team

---

## 📋 执行摘要

### 监控目标

构建全方位、多层次的监控告警体系，确保 ETF-Insight 平台的稳定性、性能和安全性。

### 监控范围

| 维度 | 覆盖范围 | 状态 |
|------|----------|------|
| 基础设施 | 服务器、网络、存储 | 🔄 待实施 |
| 应用性能 | API响应、错误率、吞吐量 | 🔄 待实施 |
| 业务指标 | 用户活跃、功能使用、数据质量 | 🔄 待实施 |
| 安全监控 | 访问日志、异常行为、漏洞扫描 | 🔄 待实施 |

---

## 🏗️ 监控架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      数据采集层                              │
├─────────────┬─────────────┬─────────────┬─────────────────┤
│  Metrics    │    Logs     │   Traces    │   Health Check  │
│  (指标)     │   (日志)    │   (链路)    │    (健康检查)    │
└──────┬──────┴──────┬──────┴──────┬──────┴────────┬────────┘
       │             │             │               │
       └─────────────┴──────┬──────┴───────────────┘
                            │
                    ┌───────▼────────┐
                    │  数据存储层     │
                    │  Prometheus    │
                    │  Loki          │
                    │  Tempo/Jaeger  │
                    └───────┬────────┘
                            │
                    ┌───────▼────────┐
                    │  分析与告警层   │
                    │  Grafana       │
                    │  AlertManager  │
                    └───────┬────────┘
                            │
       ┌────────────────────┼────────────────────┐
       │                    │                    │
┌──────▼──────┐    ┌───────▼────────┐   ┌──────▼──────┐
│  可视化仪表板 │    │   告警通知渠道   │   │  自动化响应  │
│  Dashboard  │    │  Email/Slack   │   │  Auto-heal  │
│             │    │  SMS/PagerDuty │   │  Auto-scale │
└─────────────┘    └────────────────┘   └─────────────┘
```

### 技术栈选型

| 组件 | 工具 | 用途 | 部署方式 |
|------|------|------|----------|
| 指标采集 | Prometheus | 时序数据存储和查询 | Docker |
| 日志采集 | Loki + Promtail | 日志聚合和查询 | Docker |
| 链路追踪 | Jaeger | 分布式链路追踪 | Docker |
| 可视化 | Grafana | 仪表板和告警 | Docker |
| 告警管理 | AlertManager | 告警路由和通知 | Docker |
| APM | OpenTelemetry | 应用性能监控 | 代码集成 |

---

## 📊 监控指标体系

### 1. 基础设施监控 (Infrastructure)

#### 1.1 服务器资源

| 指标 | 采集频率 | 告警阈值 | 告警级别 |
|------|----------|----------|----------|
| **CPU 使用率** | 15s | > 80% 持续5分钟 | Warning |
| | | > 90% 持续2分钟 | Critical |
| **内存使用率** | 15s | > 80% 持续5分钟 | Warning |
| | | > 90% 持续2分钟 | Critical |
| **磁盘使用率** | 60s | > 80% | Warning |
| | | > 90% | Critical |
| **磁盘 I/O** | 15s | > 100MB/s 持续5分钟 | Warning |
| **网络流量** | 15s | > 1Gbps 持续5分钟 | Warning |
| **网络延迟** | 15s | > 100ms | Warning |
| **连接数** | 15s | > 10000 | Warning |

#### 1.2 数据库监控

| 指标 | 采集频率 | 告警阈值 | 告警级别 |
|------|----------|----------|----------|
| **连接池使用率** | 30s | > 80% | Warning |
| **慢查询数量** | 60s | > 10/分钟 | Warning |
| **查询响应时间** | 30s | P95 > 100ms | Warning |
| **事务死锁** | 60s | > 0 | Critical |
| **复制延迟** | 30s | > 1s | Warning |
| **存储空间** | 300s | > 80% | Warning |

#### 1.3 容器/K8s 监控

| 指标 | 采集频率 | 告警阈值 | 告警级别 |
|------|----------|----------|----------|
| **Pod 重启次数** | 60s | > 3/小时 | Warning |
| **Pod 状态异常** | 30s | Not Ready > 2分钟 | Critical |
| **资源限制触发** | 60s | OOMKilled | Critical |
| **HPA 扩缩容** | 60s | 频繁扩缩容(>5次/小时) | Warning |

### 2. 应用性能监控 (APM)

#### 2.1 API 性能指标

| 指标 | 采集频率 | 告警阈值 | 告警级别 |
|------|----------|----------|----------|
| **响应时间 (P50)** | 实时 | > 50ms | Info |
| **响应时间 (P95)** | 实时 | > 200ms | Warning |
| **响应时间 (P99)** | 实时 | > 500ms | Critical |
| **错误率** | 实时 | > 1% | Warning |
| | | > 5% | Critical |
| **吞吐量 (RPS)** | 实时 | 异常下降 > 50% | Warning |
| **并发连接数** | 实时 | > 1000 | Warning |

#### 2.2 关键 API 端点监控

```yaml
关键端点:
  - path: /api/etf/list
    sla: 50ms
    importance: high

  - path: /api/portfolio/optimize
    sla: 500ms
    importance: high

  - path: /api/factor/fama-french
    sla: 200ms
    importance: high

  - path: /api/backtest/run
    sla: 2000ms
    importance: medium

  - path: /health
    sla: 10ms
    importance: critical
```

#### 2.3 Go 运行时监控

| 指标 | 采集频率 | 告警阈值 | 告警级别 |
|------|----------|----------|----------|
| **Goroutine 数量** | 15s | > 10000 | Warning |
| **GC 暂停时间** | 15s | > 10ms | Warning |
| **GC 频率** | 60s | > 10/分钟 | Warning |
| **堆内存使用** | 15s | > 1GB | Warning |
| **线程数** | 15s | > 1000 | Warning |

#### 2.4 前端性能监控

| 指标 | 采集频率 | 告警阈值 | 告警级别 |
|------|----------|----------|----------|
| **首屏加载时间 (FCP)** | 实时 | > 2s | Warning |
| **可交互时间 (TTI)** | 实时 | > 3s | Warning |
| **LCP (最大内容绘制)** | 实时 | > 2.5s | Warning |
| **CLS (累积布局偏移)** | 实时 | > 0.1 | Warning |
| **JS 错误率** | 实时 | > 1% | Warning |
| **API 请求失败率** | 实时 | > 5% | Critical |

### 3. 业务指标监控 (Business)

#### 3.1 用户行为

| 指标 | 采集频率 | 告警阈值 | 告警级别 |
|------|----------|----------|----------|
| **日活跃用户 (DAU)** | 1h | 环比下降 > 30% | Warning |
| **活跃会话数** | 实时 | 异常波动 | Info |
| **用户留存率** | 1d | 下降 > 10% | Warning |
| **功能使用频率** | 1h | 核心功能使用下降 | Warning |

#### 3.2 数据质量

| 指标 | 采集频率 | 告警阈值 | 告警级别 |
|------|----------|----------|----------|
| **数据同步延迟** | 5m | > 30分钟 | Warning |
| | | > 2小时 | Critical |
| **数据完整性** | 1h | < 99% | Warning |
| **ETF 价格更新** | 5m | 超过1小时未更新 | Critical |
| **汇率数据新鲜度** | 5m | 超过30分钟未更新 | Warning |

#### 3.3 业务异常

| 指标 | 采集频率 | 告警阈值 | 告警级别 |
|------|----------|----------|----------|
| **计算失败率** | 实时 | > 5% | Warning |
| **无效请求数** | 实时 | > 100/分钟 | Warning |
| **限流触发次数** | 实时 | > 1000/小时 | Warning |

### 4. 安全监控 (Security)

#### 4.1 访问监控

| 指标 | 采集频率 | 告警阈值 | 告警级别 |
|------|----------|----------|----------|
| **失败登录尝试** | 实时 | > 5/分钟/用户 | Warning |
| | | > 20/分钟/IP | Critical |
| **异常访问模式** | 实时 | 检测到爬虫行为 | Warning |
| **敏感接口访问** | 实时 | 非授权访问 | Critical |
| **JWT 异常** | 实时 | > 10/分钟 | Warning |

#### 4.2 攻击检测

| 指标 | 采集频率 | 告警阈值 | 告警级别 |
|------|----------|----------|----------|
| **SQL 注入尝试** | 实时 | > 0 | Critical |
| **XSS 攻击尝试** | 实时 | > 0 | Critical |
| **暴力破解** | 实时 | > 10/分钟 | Critical |
| **DDoS 检测** | 实时 | 流量异常激增 | Critical |

---

## 🚨 告警策略

### 告警级别定义

| 级别 | 名称 | 响应时间 | 通知方式 | 升级策略 |
|------|------|----------|----------|----------|
| **P0** | Critical | 15分钟内 | 电话 + 短信 + 邮件 + Slack | 5分钟无响应升级 |
| **P1** | High | 1小时内 | 短信 + 邮件 + Slack | 30分钟无响应升级 |
| **P2** | Medium | 4小时内 | 邮件 + Slack | 2小时无响应升级 |
| **P3** | Low | 24小时内 | 邮件 | 无需升级 |
| **P4** | Info | 无需响应 | Slack | 不升级 |

### 告警规则配置

```yaml
# alertmanager.yml
groups:
  - name: infrastructure
    rules:
      - alert: HighCPUUsage
        expr: cpu_usage_percent > 80
        for: 5m
        labels:
          severity: warning
          team: devops
        annotations:
          summary: "High CPU usage detected"
          description: "CPU usage is above 80% for more than 5 minutes"

      - alert: CriticalCPUUsage
        expr: cpu_usage_percent > 90
        for: 2m
        labels:
          severity: critical
          team: devops
        annotations:
          summary: "Critical CPU usage detected"
          description: "CPU usage is above 90% for more than 2 minutes"
          runbook_url: "https://wiki/runbooks/high-cpu"

  - name: application
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.05
        for: 2m
        labels:
          severity: critical
          team: backend
        annotations:
          summary: "High error rate detected"
          description: "Error rate is above 5% for more than 2 minutes"

      - alert: SlowAPIResponse
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5
        for: 5m
        labels:
          severity: warning
          team: backend
        annotations:
          summary: "Slow API response time"
          description: "P95 response time is above 500ms"

  - name: business
    rules:
      - alert: DataSyncDelayed
        expr: time() - last_data_sync_timestamp > 1800
        for: 5m
        labels:
          severity: warning
          team: data
        annotations:
          summary: "Data sync is delayed"
          description: "Data has not been synced for more than 30 minutes"

      - alert: DataSyncFailed
        expr: data_sync_status == 0
        for: 1m
        labels:
          severity: critical
          team: data
        annotations:
          summary: "Data sync failed"
          description: "Data synchronization has failed"
```

### 告警通知渠道

#### 渠道配置

```yaml
# 通知渠道
notification_channels:
  email:
    enabled: true
    smtp_server: smtp.example.com
    from: alerts@etf-insight.com
    to:
      - oncall@etf-insight.com
      - team@etf-insight.com

  slack:
    enabled: true
    webhook_url: ${SLACK_WEBHOOK_URL}
    channel: "#alerts"
    mention_on_critical: "@channel"

  sms:
    enabled: true
    provider: twilio
    to:
      - "+86138xxxx1234"  # On-call engineer

  pagerduty:
    enabled: true
    service_key: ${PAGERDUTY_KEY}
    severity_mapping:
      critical: critical
      warning: warning
```

#### 路由规则

```yaml
# 告警路由
route:
  group_by: ['alertname', 'severity', 'team']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  receiver: default
  routes:
    - match:
        severity: critical
      receiver: critical_alerts
      group_wait: 10s
      repeat_interval: 30m

    - match:
        team: backend
      receiver: backend_team

    - match:
        team: frontend
      receiver: frontend_team

    - match:
        team: devops
      receiver: devops_team

    - match:
        team: security
      receiver: security_team

receivers:
  - name: default
    email_configs:
      - to: 'team@etf-insight.com'

  - name: critical_alerts
    email_configs:
      - to: 'oncall@etf-insight.com'
    slack_configs:
      - channel: '#critical-alerts'
        send_resolved: true
    pagerduty_configs:
      - service_key: '${PAGERDUTY_KEY}'
        severity: critical
```

### 告警抑制与静默

```yaml
# 抑制规则
inhibit_rules:
  # 高优先级告警抑制低优先级
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'instance']

  # 同类型告警抑制
  - source_match:
      alertname: 'InstanceDown'
    target_match_re:
      alertname: '.*High.*'
    equal: ['instance']

# 静默规则（维护窗口）
silences:
  - id: maintenance-window
    matchers:
      - name: severity
        value: warning
    startsAt: "2026-05-01T02:00:00Z"
    endsAt: "2026-05-01T04:00:00Z"
    createdBy: "devops"
    comment: "Scheduled maintenance"
```

---

## 📈 可视化仪表板

### 1. 系统概览仪表板

**URL**: `https://grafana.etf-insight.com/d/system-overview`

**面板内容**:
- 系统健康状态 (红/黄/绿)
- 关键指标概览 (QPS、错误率、延迟)
- 活跃告警列表
- 最近事件时间线

### 2. 应用性能仪表板

**URL**: `https://grafana.etf-insight.com/d/app-performance`

**面板内容**:
- API 响应时间分布 (P50/P95/P99)
- 吞吐量趋势
- 错误率趋势
- 各端点性能排行
- 依赖服务性能

### 3. 基础设施仪表板

**URL**: `https://grafana.etf-insight.com/d/infrastructure`

**面板内容**:
- CPU/内存/磁盘使用率
- 网络流量
- 数据库性能
- 容器资源使用

### 4. 业务指标仪表板

**URL**: `https://grafana.etf-insight.com/d/business-metrics`

**面板内容**:
- 用户活跃度
- 功能使用统计
- 数据质量指标
- 计算任务成功率

### 5. 安全监控仪表板

**URL**: `https://grafana.etf-insight.com/d/security`

**面板内容**:
- 登录失败统计
- 异常访问模式
- 攻击尝试检测
- 安全事件时间线

---

## 🔄 故障响应流程

### 响应流程图

```
┌─────────────┐
│  告警触发   │
└──────┬──────┘
       │
       ▼
┌─────────────┐     ┌─────────────┐
│  告警分级   │────▶│  P0 Critical │
└──────┬──────┘     └──────┬──────┘
       │                   │
       │                   ▼
       │          ┌─────────────────┐
       │          │ 立即响应(15min) │
       │          │ 电话+短信+Slack │
       │          └────────┬────────┘
       │                   │
       │                   ▼
       │          ┌─────────────────┐
       │          │  启动应急预案   │
       │          │  升级至管理层   │
       │          └────────┬────────┘
       │                   │
       │                   ▼
       │          ┌─────────────────┐
       │          │   故障复盘      │
       │          │   改进措施      │
       │          └─────────────────┘
       │
       ▼
┌─────────────┐
│ P1/P2/P3    │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│  标准响应流程   │
│  工单跟踪      │
│  定时更新      │
└─────────────────┘
```

### 响应时间 SLA

| 级别 | 响应时间 | 解决时间目标 | 升级时间 |
|------|----------|--------------|----------|
| P0 | 15分钟 | 2小时 | 5分钟 |
| P1 | 1小时 | 4小时 | 30分钟 |
| P2 | 4小时 | 24小时 | 2小时 |
| P3 | 24小时 | 72小时 | 无需升级 |

### 故障处理 checklist

#### P0 故障处理

- [ ] 收到告警通知
- [ ] 确认故障影响范围
- [ ] 启动应急响应小组
- [ ] 通知相关干系人
- [ ] 尝试快速恢复
- [ ] 如无法恢复，启动灾备方案
- [ ] 持续更新状态
- [ ] 故障恢复后验证
- [ ] 启动故障复盘

#### 故障复盘模板

```markdown
## 故障复盘报告

### 基本信息
- 故障时间: YYYY-MM-DD HH:MM
- 恢复时间: YYYY-MM-DD HH:MM
- 故障时长: XX 分钟
- 故障级别: P0/P1/P2
- 影响范围:

### 故障现象
描述故障的具体表现

### 根因分析
5 Whys 分析

### 处理过程
时间线记录

### 改进措施
- [ ] 短期措施
- [ ] 长期措施
- [ ] 负责人
- [ ] 完成时间

### 经验教训
总结和分享
```

---

## 🛠️ 实施计划

### Phase 1: 基础监控 (Week 1-2)

#### Week 1: 部署监控基础设施

| 任务 | 负责人 | 交付物 |
|------|--------|--------|
| 部署 Prometheus | DevOps | 指标采集系统 |
| 部署 Grafana | DevOps | 可视化平台 |
| 部署 Loki | DevOps | 日志系统 |
| 配置基础采集 | DevOps | 服务器指标 |

#### Week 2: 应用指标接入

| 任务 | 负责人 | 交付物 |
|------|--------|--------|
| 集成 Prometheus Client | Backend | 应用指标暴露 |
| 配置业务指标 | Backend | 自定义指标 |
| 创建基础仪表板 | DevOps | 系统概览面板 |

### Phase 2: 告警配置 (Week 3-4)

#### Week 3: 告警规则配置

| 任务 | 负责人 | 交付物 |
|------|--------|--------|
| 部署 AlertManager | DevOps | 告警管理 |
| 配置告警规则 | DevOps | 规则文件 |
| 配置通知渠道 | DevOps | 渠道集成 |

#### Week 4: 告警测试与优化

| 任务 | 负责人 | 交付物 |
|------|--------|--------|
| 告警测试 | DevOps | 测试报告 |
| 优化阈值 | Team | 调优后的规则 |
| 文档更新 | DevOps | 运维手册 |

### Phase 3: 高级监控 (Week 5-6)

#### Week 5: 链路追踪

| 任务 | 负责人 | 交付物 |
|------|--------|--------|
| 部署 Jaeger | DevOps | 链路追踪系统 |
| 集成 OpenTelemetry | Backend | 分布式追踪 |
| 配置追踪采样 | Backend | 采样策略 |

#### Week 6: 前端监控

| 任务 | 负责人 | 交付物 |
|------|--------|--------|
| 集成 RUM | Frontend | 真实用户监控 |
| 配置性能指标 | Frontend | Web Vitals |
| 错误追踪 | Frontend | 前端错误监控 |

### Phase 4: 完善与文档 (Week 7-8)

#### Week 7: 安全监控

| 任务 | 负责人 | 交付物 |
|------|--------|--------|
| 配置安全告警 | Security | 安全规则 |
| 集成 WAF 日志 | Security | 攻击检测 |
| 异常行为检测 | Security | 行为分析 |

#### Week 8: 文档与培训

| 任务 | 负责人 | 交付物 |
|------|--------|--------|
| 运维手册 | DevOps | 操作文档 |
| 告警响应手册 | Team | 响应流程 |
| 团队培训 | DevOps | 培训材料 |

---

## 📋 运维手册

### 日常巡检

#### 每日检查

```bash
#!/bin/bash
# daily-check.sh

echo "=== ETF-Insight Daily Check ==="
echo "Date: $(date)"

# 检查系统健康
echo "1. Checking system health..."
curl -s http://localhost:8080/health | jq .

# 检查关键指标
echo "2. Checking key metrics..."
# CPU
echo "CPU Usage: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}')"

# 内存
echo "Memory Usage: $(free | grep Mem | awk '{printf "%.2f%%", $3/$2 * 100.0}')"

# 磁盘
echo "Disk Usage: $(df -h / | tail -1 | awk '{print $5}')"

# 检查告警
echo "3. Checking active alerts..."
curl -s http://alertmanager:9093/api/v1/alerts | jq '.data | length'

echo "=== Check Complete ==="
```

#### 每周检查

- [ ] 审查告警历史
- [ ] 检查仪表板性能
- [ ] 验证备份完整性
- [ ] 更新监控阈值（如需）
- [ ] 审查日志轮转

### 常见问题处理

#### 1. 磁盘空间不足

```bash
# 检查磁盘使用
df -h

# 清理日志
find /var/log/etf-insight -name "*.log" -mtime +7 -delete

# 清理临时文件
rm -rf /tmp/etf-insight/*
```

#### 2. 内存不足

```bash
# 查看内存使用
free -h

# 查看内存占用最高的进程
ps aux --sort=-%mem | head -10

# 重启服务（如果必要）
systemctl restart etf-insight
```

#### 3. 数据库连接池耗尽

```bash
# 查看当前连接数
psql -c "SELECT count(*) FROM pg_stat_activity;"

# 查看活跃连接
psql -c "SELECT * FROM pg_stat_activity WHERE state = 'active';"

# 重启应用释放连接
systemctl restart etf-insight
```

---

## 📚 参考资源

### 官方文档

- [Prometheus Docs](https://prometheus.io/docs/)
- [Grafana Docs](https://grafana.com/docs/)
- [AlertManager Docs](https://prometheus.io/docs/alerting/latest/alertmanager/)
- [OpenTelemetry Docs](https://opentelemetry.io/docs/)

### 最佳实践

- [Google SRE Book](https://sre.google/sre-book/table-of-contents/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Grafana Dashboard Best Practices](https://grafana.com/docs/grafana/latest/best-practices/)

---

**制定**: DevOps Team
**审批**: Tech Lead
**生效日期**: 2026-05-01

**最后更新**: 2026-04-21
**下次评审**: 2026-05-21

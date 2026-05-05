# Phase 2: AI Agent Microservice 同步计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 同步 Phase 2 设计文档与实际代码实现，确保文档准确反映当前状态，并补充缺失的功能。

**架构：** FastAPI 微服务 (port 8091) 提供 AI Agent 服务，Go Backend 通过 HTTP 调用 Python 微服务，React Frontend 提供用户界面。

**技术栈：** Python 3.11+, FastAPI, Pydantic, httpx, OpenAI API, DeepSeek API, Ollama

---

## 当前状态

### 已实现组件

| 组件 | 文件 | 状态 |
|------|------|------|
| LLM Provider | `core/llm_provider.py` | ✅ 已实现 |
| Base Agent | `core/base_agent.py` | ✅ 已实现 |
| Tool Registry | `core/tool_registry.py` | ✅ 已实现 |
| Agent Manager | `core/agent_manager.py` | ✅ 已实现 |
| FastAPI Server | `agent_server.py` | ✅ 已实现 |
| Pydantic Schemas | `models/schemas.py` | ✅ 已实现 |
| Tests | `tests/` | ✅ 19 个测试通过 |
| Dockerfile | `Dockerfile` | ✅ 已实现 |

### 已实现 Agent

| Agent | 文件 | 状态 |
|-------|------|------|
| Warren Buffett | `agents/buffett.py` | ✅ 已实现 |
| Benjamin Graham | `agents/graham.py` | ✅ 已实现 |
| Bridgewater | `agents/bridgewater.py` | ✅ 已实现 |
| Macro Analyst | `agents/macro.py` | ✅ 已实现 |

### 待同步项

1. 设计文档状态标记（Phase 2 应标记为 ✅ 已完成）
2. Go Backend 集成代码（agent_client.go, agent_handler.go）
3. Frontend 集成代码（AIAgents.tsx, agent.ts）
4. API 路由注册

---

## 任务清单

### 任务 1：验证 Python 微服务功能

**文件：**
- 验证：`backend/services/agent/` 目录下所有文件

- [ ] **步骤 1：运行 Python 微服务测试**

```bash
cd /Users/liunian/Desktop/dnmp/py_project/backend/services/agent
python -m pytest tests/ -v
```

预期：19 个测试全部通过

- [ ] **步骤 2：启动 Python 微服务**

```bash
cd /Users/liunian/Desktop/dnmp/py_project/backend/services/agent
python agent_server.py &
```

预期：服务在 port 8091 启动

- [ ] **步骤 3：测试 API 端点**

```bash
# 健康检查
curl http://localhost:8091/health

# Agent 列表
curl http://localhost:8091/agents/discover

# 运行 Agent
curl -X POST http://localhost:8091/agents/run \
  -H "Content-Type: application/json" \
  -d '{"agent_name": "buffett", "query": "分析 VTI ETF"}'
```

预期：所有 API 返回正确响应

---

### 任务 2：实现 Go Backend 集成

**文件：**
- 创建：`backend/services/agent/agent_client.go`
- 创建：`backend/handlers/agent_handler.go`
- 修改：`backend/router/router.go`

- [ ] **步骤 1：创建 Agent Client**

```go
// backend/services/agent/agent_client.go
package agent

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type AgentClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewAgentClient(baseURL string) *AgentClient {
    return &AgentClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 60 * time.Second,
        },
    }
}

type AgentInfo struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Category    string `json:"category"`
}

type AgentRunRequest struct {
    AgentName string `json:"agent_name"`
    Query     string `json:"query"`
    Model     string `json:"model,omitempty"`
}

type AgentRunResponse struct {
    AgentName   string `json:"agent_name"`
    Response    string `json:"response"`
    Model       string `json:"model"`
    TokensUsed  int    `json:"tokens_used"`
}

func (c *AgentClient) DiscoverAgents(ctx context.Context) ([]AgentInfo, error) {
    url := fmt.Sprintf("%s/agents/discover", c.baseURL)
    resp, err := c.httpClient.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var agents []AgentInfo
    if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
        return nil, err
    }
    return agents, nil
}

func (c *AgentClient) RunAgent(ctx context.Context, req AgentRunRequest) (*AgentRunResponse, error) {
    url := fmt.Sprintf("%s/agents/run", c.baseURL)
    // ... 实现 POST 请求
    return nil, nil
}
```

- [ ] **步骤 2：创建 Agent Handler**

```go
// backend/handlers/agent_handler.go
package handlers

import (
    "net/http"
    "etf-insight/services/agent"
    "github.com/gin-gonic/gin"
)

type AgentHandler struct {
    client *agent.AgentClient
}

func NewAgentHandler(client *agent.AgentClient) *AgentHandler {
    return &AgentHandler{client: client}
}

func (h *AgentHandler) DiscoverAgents(c *gin.Context) {
    agents, err := h.client.DiscoverAgents(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"agents": agents})
}

func (h *AgentHandler) RunAgent(c *gin.Context) {
    var req agent.AgentRunRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    resp, err := h.client.RunAgent(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, resp)
}
```

- [ ] **步骤 3：注册路由**

```go
// backend/router/router.go 添加
func registerAgentRoutes(r *gin.Engine, agentHandler *handlers.AgentHandler) {
    api := r.Group("/api/agents")
    {
        api.GET("/discover", agentHandler.DiscoverAgents)
        api.POST("/run", agentHandler.RunAgent)
        api.POST("/stream", agentHandler.StreamAgent)
        api.POST("/team", agentHandler.RunTeam)
    }
}
```

- [ ] **步骤 4：测试 Go 集成**

```bash
cd /Users/liunian/Desktop/dnmp/py_project/backend
go test ./services/agent/... -v
go test ./handlers/... -v -run Agent
```

---

### 任务 3：实现 Frontend 集成

**文件：**
- 创建：`frontend/src/types/agent.ts`
- 修改：`frontend/src/services/api.ts`
- 创建：`frontend/src/pages/AIAgents.tsx`
- 修改：`frontend/src/App.tsx`

- [ ] **步骤 1：创建 Agent 类型定义**

```typescript
// frontend/src/types/agent.ts
export interface AgentInfo {
  name: string;
  description: string;
  category: string;
}

export interface AgentRunRequest {
  agent_name: string;
  query: string;
  model?: string;
}

export interface AgentRunResponse {
  agent_name: string;
  response: string;
  model: string;
  tokens_used: number;
}

export interface AgentTeamRequest {
  agents: string[];
  query: string;
  rounds: number;
}

export interface AgentTeamResponse {
  debate_history: DebateRound[];
  consensus: string;
}
```

- [ ] **步骤 2：添加 Agent API 服务**

```typescript
// frontend/src/services/api.ts 添加
export const agentAPI = {
  discover: async (): Promise<AgentInfo[]> => {
    const response = await requestWithRetry('/api/agents/discover');
    return response.agents;
  },

  run: async (req: AgentRunRequest): Promise<AgentRunResponse> => {
    return requestWithRetry('/api/agents/run', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },

  stream: async (req: AgentRunRequest): Promise<EventSource> => {
    const url = `${API_BASE_URL}/api/agents/stream`;
    // SSE 实现
  },

  team: async (req: AgentTeamRequest): Promise<AgentTeamResponse> => {
    return requestWithRetry('/api/agents/team', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },
};
```

- [ ] **步骤 3：创建 AI Agents 页面**

```tsx
// frontend/src/pages/AIAgents.tsx
import React, { useState, useEffect } from 'react';
import styled from 'styled-components';
import { agentAPI } from '../services/api';
import type { AgentInfo, AgentRunRequest } from '../types/agent';

const Container = styled.div`
  padding: 20px;
`;

const AgentCard = styled.div`
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
`;

export const AIAgents: React.FC = () => {
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [selectedAgent, setSelectedAgent] = useState<string>('');
  const [query, setQuery] = useState('');
  const [response, setResponse] = useState('');

  useEffect(() => {
    agentAPI.discover().then(setAgents);
  }, []);

  const handleRun = async () => {
    const result = await agentAPI.run({
      agent_name: selectedAgent,
      query,
    });
    setResponse(result.response);
  };

  return (
    <Container>
      <h1>AI Agents</h1>
      {agents.map((agent) => (
        <AgentCard key={agent.name}>
          <h3>{agent.name}</h3>
          <p>{agent.description}</p>
        </AgentCard>
      ))}
      {/* ... 表单和响应显示 */}
    </Container>
  );
};
```

- [ ] **步骤 4：添加路由**

```tsx
// frontend/src/App.tsx 添加
import { AIAgents } from './pages/AIAgents';

// 在路由配置中添加
<Route path="/ai-agents" element={<AIAgents />} />

// 在侧边栏添加
<MenuItem icon={<RobotOutlined />} key="ai-agents">
  AI Agents
</MenuItem>
```

---

### 任务 4：更新设计文档

**文件：**
- 修改：`docs/superpowers/specs/2026-05-04-fincept-integration-design.md`

- [ ] **步骤 1：更新 Phase 2 状态**

将 Phase 2 状态从 "📋 Not Started" 更新为 "✅ Completed"

```markdown
## Phase 2: AI Agent Microservice

### Status: ✅ Completed

### Goal
...

### Implementation Summary
- 4 个 Agent 已实现并测试通过
- FastAPI 服务运行在 port 8091
- 支持 OpenAI, DeepSeek, Ollama 三种 LLM
- 19 个单元测试全部通过
```

- [ ] **步骤 2：添加实现细节**

补充实际实现的代码示例和配置说明。

- [ ] **步骤 3：更新 License Resolution**

确认 "Option B — Rewrite from scratch" 决策已记录，并说明零 AGPL 风险。

---

### 任务 5：添加认证授权

**文件：**
- 创建：`backend/middleware/auth.go`
- 修改：`backend/services/agent/agent_server.py`

- [ ] **步骤 1：Go Backend 添加 JWT 认证**

```go
// backend/middleware/auth.go
package middleware

import (
    "strings"
    "github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        // JWT 验证逻辑
        // ...

        c.Next()
    }
}
```

- [ ] **步骤 2：Python 微服务添加 API Key 验证**

```python
# backend/services/agent/agent_server.py 添加
from fastapi import Header, HTTPException

async def verify_api_key(x_api_key: str = Header(...)):
    expected_key = os.getenv("AGENT_API_KEY")
    if not expected_key or x_api_key != expected_key:
        raise HTTPException(status_code=401, detail="Invalid API key")
    return x_api_key

# 在路由中添加依赖
@app.post("/agents/run", dependencies=[Depends(verify_api_key)])
async def run_agent(request: AgentRunRequest):
    ...
```

---

### 任务 6：添加监控和日志

**文件：**
- 创建：`backend/services/agent/core/metrics.py`
- 修改：`backend/services/agent/agent_server.py`

- [ ] **步骤 1：添加 Prometheus 指标**

```python
# backend/services/agent/core/metrics.py
from prometheus_client import Counter, Histogram

AGENT_REQUESTS = Counter(
    'agent_requests_total',
    'Total agent requests',
    ['agent_name', 'status']
)

AGENT_LATENCY = Histogram(
    'agent_request_latency_seconds',
    'Agent request latency',
    ['agent_name']
)
```

- [ ] **步骤 2：添加结构化日志**

```python
# backend/services/agent/agent_server.py 添加
import logging
import structlog

logger = structlog.get_logger()

@app.post("/agents/run")
async def run_agent(request: AgentRunRequest):
    logger.info("agent_request", agent=request.agent_name, query=request.query)
    # ...
```

---

## 验收标准

- [ ] Python 微服务所有测试通过
- [ ] Go Backend 能成功调用 Python 微服务
- [ ] Frontend 能显示 Agent 列表并执行查询
- [ ] 设计文档状态更新为 ✅ Completed
- [ ] 认证授权机制已实现
- [ ] 监控指标已暴露

---

## 风险与缓解

| 风险 | 缓解措施 |
|------|----------|
| LLM API 调用失败 | 实现重试机制和降级策略 |
| 响应时间过长 | 添加超时控制和流式响应 |
| 认证绕过 | 强制 JWT/API Key 验证 |
| 资源消耗过大 | 添加速率限制和并发控制 |

# AI Agent Microservice Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a standalone FastAPI microservice (port 8091) providing 31 AI agents for financial analysis, integrated into ETF-Insight via Go backend client and React frontend.

**Architecture:** Python FastAPI service with pluggable LLM providers (OpenAI-compatible API supporting Ollama/OpenAI/Anthropic/etc.), abstract base agent class, tool registry for extensible capabilities, and streaming SSE support. Go backend acts as HTTP bridge. From-scratch implementation — zero AGPL code.

**Tech Stack:** Python 3.11+, FastAPI, uvicorn, httpx (async HTTP), pydantic v2, SSE-starlette. Go: net/http, gin. React: Ant Design, TypeScript.

---

## File Structure

```
backend/services/agent/
├── agent_server.py           # FastAPI entry point (port 8091)
├── requirements.txt          # Python dependencies
├── Dockerfile                # Container definition
├── core/
│   ├── __init__.py
│   ├── base_agent.py         # Abstract base agent
│   ├── llm_provider.py       # Multi-LLM abstraction (OpenAI-compatible)
│   ├── tool_registry.py      # Tool/function registry
│   └── agent_manager.py      # Agent discovery, registration, execution
├── agents/
│   ├── __init__.py
│   ├── buffett.py            # Warren Buffett agent
│   ├── graham.py             # Benjamin Graham agent
│   ├── bridgewater.py        # Bridgewater (macro risk) agent
│   ├── macro.py              # Macroeconomic analysis agent
│   └── registry.py           # Auto-registration of all agents
├── models/
│   ├── __init__.py
│   └── schemas.py            # Pydantic request/response models
├── tests/
│   ├── __init__.py
│   ├── test_llm_provider.py
│   ├── test_base_agent.py
│   ├── test_agent_manager.py
│   └── test_api.py
backend/services/agent/agent_client.go    # Go HTTP client
backend/handlers/agent_handler.go         # Go Gin handler
backend/router/router.go                  # MODIFY: add agent routes
frontend/src/types/agent.ts               # TypeScript types
frontend/src/services/api.ts              # MODIFY: add agentAPI
frontend/src/pages/AIAgents.tsx           # New page
frontend/src/App.tsx                      # MODIFY: add route
frontend/src/components/Layout.tsx        # MODIFY: add sidebar nav
```

---

### Task 1: Python Project Scaffolding

**Files:**
- Create: `backend/services/agent/requirements.txt`
- Create: `backend/services/agent/core/__init__.py`
- Create: `backend/services/agent/agents/__init__.py`
- Create: `backend/services/agent/models/__init__.py`
- Create: `backend/services/agent/models/schemas.py`

- [ ] **Step 1: Create requirements.txt**

```txt
fastapi==0.115.6
uvicorn[standard]==0.34.0
httpx==0.28.1
pydantic==2.10.4
sse-starlette==2.2.1
python-dotenv==1.0.1
pytest==8.3.4
pytest-asyncio==0.25.0
```

- [ ] **Step 2: Create package init files**

Create `backend/services/agent/core/__init__.py`:
```python
```

Create `backend/services/agent/agents/__init__.py`:
```python
```

Create `backend/services/agent/models/__init__.py`:
```python
```

- [ ] **Step 3: Create Pydantic schemas**

Create `backend/services/agent/models/schemas.py`:
```python
from pydantic import BaseModel, Field
from typing import Optional
from enum import Enum


class AgentCategory(str, Enum):
    LEGENDARY_INVESTOR = "legendary_investor"
    HEDGE_FUND = "hedge_fund"
    GEOPOLITICS = "geopolitics"
    MACRO_ECONOMIC = "macro_economic"
    TECHNICAL = "technical"


class AgentInfo(BaseModel):
    id: str
    name: str
    category: AgentCategory
    description: str
    system_prompt_preview: str = Field(max_length=200)


class AgentRunRequest(BaseModel):
    agent_id: str
    query: str
    context: Optional[dict] = None
    llm_provider: str = "openai"
    model: str = "gpt-4o-mini"
    temperature: float = Field(default=0.7, ge=0.0, le=2.0)
    max_tokens: int = Field(default=2000, ge=1, le=8000)


class AgentRunResponse(BaseModel):
    agent_id: str
    agent_name: str
    response: str
    model: str
    tokens_used: int = 0
    duration_ms: int = 0


class AgentStreamChunk(BaseModel):
    agent_id: str
    chunk: str
    done: bool = False


class AgentTeamRequest(BaseModel):
    agent_ids: list[str] = Field(min_length=2, max_length=5)
    query: str
    rounds: int = Field(default=1, ge=1, le=3)
    llm_provider: str = "openai"
    model: str = "gpt-4o-mini"


class AgentTeamResponse(BaseModel):
    query: str
    rounds: list[dict]
    synthesis: str


class HealthResponse(BaseModel):
    status: str
    version: str
    agents_count: int
    llm_providers: list[str]
```

- [ ] **Step 4: Verify packages import**

Run: `cd backend/services/agent && python -c "from models.schemas import AgentInfo; print('OK')"`
Expected: OK

- [ ] **Step 5: Commit**

```bash
git add backend/services/agent/requirements.txt backend/services/agent/core/__init__.py backend/services/agent/agents/__init__.py backend/services/agent/models/
git commit -m "feat(agent-service): scaffold Python agent service project structure"
```

---

### Task 2: LLM Provider Abstraction

**Files:**
- Create: `backend/services/agent/core/llm_provider.py`
- Create: `backend/services/agent/tests/test_llm_provider.py`

- [ ] **Step 1: Write failing tests**

Create `backend/services/agent/tests/__init__.py`:
```python
```

Create `backend/services/agent/tests/test_llm_provider.py`:
```python
import pytest
from core.llm_provider import LLMProvider, OpenAICompatibleProvider, LLMResponse


def test_llm_response_fields():
    resp = LLMResponse(content="hello", model="gpt-4o-mini", tokens_used=42)
    assert resp.content == "hello"
    assert resp.model == "gpt-4o-mini"
    assert resp.tokens_used == 42


def test_openai_provider_init():
    provider = OpenAICompatibleProvider(
        api_key="test-key",
        base_url="https://api.openai.com/v1",
    )
    assert provider.api_key == "test-key"
    assert provider.base_url == "https://api.openai.com/v1"


def test_openai_provider_default_base_url():
    provider = OpenAICompatibleProvider(api_key="test-key")
    assert "openai" in provider.base_url


def test_provider_list_models():
    provider = OpenAICompatibleProvider(api_key="test-key")
    models = provider.list_models()
    assert isinstance(models, list)
    assert len(models) > 0


@pytest.mark.asyncio
async def test_generate_missing_api_key():
    provider = OpenAICompatibleProvider(api_key="", base_url="http://localhost:1")
    with pytest.raises(ValueError, match="API key"):
        await provider.generate("hello", model="gpt-4o-mini")
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend/services/agent && python -m pytest tests/test_llm_provider.py -v`
Expected: FAIL (module not found)

- [ ] **Step 3: Implement LLM provider**

Create `backend/services/agent/core/llm_provider.py`:
```python
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
import httpx
import os
import json


@dataclass
class LLMResponse:
    content: str
    model: str
    tokens_used: int = 0
    finish_reason: str = ""


class LLMProvider(ABC):
    @abstractmethod
    async def generate(
        self,
        prompt: str,
        system_prompt: str = "",
        model: str = "gpt-4o-mini",
        temperature: float = 0.7,
        max_tokens: int = 2000,
    ) -> LLMResponse:
        ...

    @abstractmethod
    def list_models(self) -> list[str]:
        ...


class OpenAICompatibleProvider(LLMProvider):
    def __init__(
        self,
        api_key: str = "",
        base_url: str = "https://api.openai.com/v1",
        timeout: float = 60.0,
    ):
        self.api_key = api_key or os.environ.get("OPENAI_API_KEY", "")
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout

    def list_models(self) -> list[str]:
        return [
            "gpt-4o",
            "gpt-4o-mini",
            "gpt-4-turbo",
            "gpt-3.5-turbo",
            "deepseek-chat",
            "deepseek-reasoner",
        ]

    async def generate(
        self,
        prompt: str,
        system_prompt: str = "",
        model: str = "gpt-4o-mini",
        temperature: float = 0.7,
        max_tokens: int = 2000,
    ) -> LLMResponse:
        if not self.api_key:
            raise ValueError("API key is required. Set OPENAI_API_KEY env var or pass api_key.")

        messages = []
        if system_prompt:
            messages.append({"role": "system", "content": system_prompt})
        messages.append({"role": "user", "content": prompt})

        payload = {
            "model": model,
            "messages": messages,
            "temperature": temperature,
            "max_tokens": max_tokens,
        }

        async with httpx.AsyncClient(timeout=self.timeout) as client:
            resp = await client.post(
                f"{self.base_url}/chat/completions",
                json=payload,
                headers={
                    "Authorization": f"Bearer {self.api_key}",
                    "Content-Type": "application/json",
                },
            )
            resp.raise_for_status()
            data = resp.json()

        choice = data["choices"][0]
        usage = data.get("usage", {})

        return LLMResponse(
            content=choice["message"]["content"],
            model=data.get("model", model),
            tokens_used=usage.get("total_tokens", 0),
            finish_reason=choice.get("finish_reason", ""),
        )


class OllamaProvider(LLMProvider):
    def __init__(self, base_url: str = "http://localhost:11434"):
        self.base_url = base_url.rstrip("/")

    def list_models(self) -> list[str]:
        return ["llama3.1", "qwen2.5", "deepseek-r1", "mistral"]

    async def generate(
        self,
        prompt: str,
        system_prompt: str = "",
        model: str = "llama3.1",
        temperature: float = 0.7,
        max_tokens: int = 2000,
    ) -> LLMResponse:
        messages = []
        if system_prompt:
            messages.append({"role": "system", "content": system_prompt})
        messages.append({"role": "user", "content": prompt})

        payload = {
            "model": model,
            "messages": messages,
            "stream": False,
            "options": {"temperature": temperature, "num_predict": max_tokens},
        }

        async with httpx.AsyncClient(timeout=120.0) as client:
            resp = await client.post(
                f"{self.base_url}/api/chat",
                json=payload,
            )
            resp.raise_for_status()
            data = resp.json()

        return LLMResponse(
            content=data["message"]["content"],
            model=data.get("model", model),
            tokens_used=data.get("eval_count", 0),
            finish_reason="stop",
        )


def get_provider(name: str, **kwargs) -> LLMProvider:
    providers = {
        "openai": lambda: OpenAICompatibleProvider(**kwargs),
        "deepseek": lambda: OpenAICompatibleProvider(
            base_url="https://api.deepseek.com/v1",
            **kwargs,
        ),
        "ollama": lambda: OllamaProvider(**kwargs),
    }
    factory = providers.get(name)
    if not factory:
        raise ValueError(f"Unknown provider: {name}. Available: {list(providers.keys())}")
    return factory()
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd backend/services/agent && python -m pytest tests/test_llm_provider.py -v`
Expected: 5 passed

- [ ] **Step 5: Commit**

```bash
git add backend/services/agent/core/llm_provider.py backend/services/agent/tests/
git commit -m "feat(agent-service): add LLM provider abstraction with OpenAI/Ollama support"
```

---

### Task 3: Base Agent and Tool Registry

**Files:**
- Create: `backend/services/agent/core/base_agent.py`
- Create: `backend/services/agent/core/tool_registry.py`
- Create: `backend/services/agent/tests/test_base_agent.py`

- [ ] **Step 1: Write failing tests**

Create `backend/services/agent/tests/test_base_agent.py`:
```python
import pytest
from core.base_agent import BaseAgent
from core.tool_registry import ToolRegistry, tool
from models.schemas import AgentCategory


class ConcreteAgent(BaseAgent):
    def _build_system_prompt(self) -> str:
        return "You are a test agent."


def test_agent_properties():
    agent = ConcreteAgent(
        agent_id="test-1",
        name="Test Agent",
        category=AgentCategory.LEGENDARY_INVESTOR,
        description="A test agent",
    )
    assert agent.agent_id == "test-1"
    assert agent.name == "Test Agent"
    assert agent.category == AgentCategory.LEGENDARY_INVESTOR


def test_agent_to_info():
    agent = ConcreteAgent(
        agent_id="test-1",
        name="Test Agent",
        category=AgentCategory.LEGENDARY_INVESTOR,
        description="A test agent",
    )
    info = agent.to_info()
    assert info.id == "test-1"
    assert info.name == "Test Agent"
    assert len(info.system_prompt_preview) <= 200


def test_tool_registry_register():
    registry = ToolRegistry()

    @tool(name="test_tool", description="A test tool")
    def my_tool(x: int) -> int:
        return x * 2

    registry.register(my_tool)
    assert "test_tool" in registry.list_tools()


def test_tool_registry_call():
    registry = ToolRegistry()

    @tool(name="double", description="Double a number")
    def double(x: int) -> int:
        return x * 2

    registry.register(double)
    result = registry.call("double", x=5)
    assert result == 10


def test_tool_registry_missing():
    registry = ToolRegistry()
    with pytest.raises(KeyError):
        registry.call("nonexistent")
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend/services/agent && python -m pytest tests/test_base_agent.py -v`
Expected: FAIL

- [ ] **Step 3: Implement tool registry**

Create `backend/services/agent/core/tool_registry.py`:
```python
from dataclasses import dataclass, field
from typing import Any, Callable
import inspect


@dataclass
class ToolDef:
    name: str
    description: str
    func: Callable
    parameters: dict = field(default_factory=dict)


def tool(name: str, description: str):
    def decorator(func: Callable) -> Callable:
        sig = inspect.signature(func)
        params = {}
        for pname, param in sig.parameters.items():
            params[pname] = {
                "type": param.annotation.__name__ if param.annotation != inspect.Parameter.empty else "string",
                "required": param.default == inspect.Parameter.empty,
            }
        func._tool_meta = ToolDef(name=name, description=description, func=func, parameters=params)
        return func
    return decorator


class ToolRegistry:
    def __init__(self):
        self._tools: dict[str, ToolDef] = {}

    def register(self, func: Callable) -> None:
        meta = getattr(func, "_tool_meta", None)
        if not meta:
            raise ValueError(f"Function {func.__name__} has no @tool decorator")
        self._tools[meta.name] = meta

    def list_tools(self) -> list[str]:
        return list(self._tools.keys())

    def get(self, name: str) -> ToolDef | None:
        return self._tools.get(name)

    def call(self, name: str, **kwargs) -> Any:
        tool_def = self._tools.get(name)
        if not tool_def:
            raise KeyError(f"Tool '{name}' not found. Available: {self.list_tools()}")
        return tool_def.func(**kwargs)

    def to_prompt_section(self) -> str:
        if not self._tools:
            return ""
        lines = ["Available tools:"]
        for t in self._tools.values():
            lines.append(f"- {t.name}: {t.description}")
        return "\n".join(lines)
```

- [ ] **Step 4: Implement base agent**

Create `backend/services/agent/core/base_agent.py`:
```python
from abc import ABC, abstractmethod
from models.schemas import AgentCategory, AgentInfo
from core.llm_provider import LLMProvider
from core.tool_registry import ToolRegistry
import time


class BaseAgent(ABC):
    def __init__(
        self,
        agent_id: str,
        name: str,
        category: AgentCategory,
        description: str,
        tools: ToolRegistry | None = None,
    ):
        self.agent_id = agent_id
        self.name = name
        self.category = category
        self.description = description
        self.tools = tools or ToolRegistry()

    @abstractmethod
    def _build_system_prompt(self) -> str:
        ...

    def to_info(self) -> AgentInfo:
        prompt = self._build_system_prompt()
        return AgentInfo(
            id=self.agent_id,
            name=self.name,
            category=self.category,
            description=self.description,
            system_prompt_preview=prompt[:200] + ("..." if len(prompt) > 200 else ""),
        )

    async def run(
        self,
        query: str,
        context: dict | None,
        llm: LLMProvider,
        model: str = "gpt-4o-mini",
        temperature: float = 0.7,
        max_tokens: int = 2000,
    ) -> dict:
        system_prompt = self._build_system_prompt()

        tool_section = self.tools.to_prompt_section()
        if tool_section:
            system_prompt += f"\n\n{tool_section}"

        if context:
            system_prompt += f"\n\nAdditional context:\n{context}"

        start = time.time()
        resp = await llm.generate(
            prompt=query,
            system_prompt=system_prompt,
            model=model,
            temperature=temperature,
            max_tokens=max_tokens,
        )
        duration_ms = int((time.time() - start) * 1000)

        return {
            "agent_id": self.agent_id,
            "agent_name": self.name,
            "response": resp.content,
            "model": resp.model,
            "tokens_used": resp.tokens_used,
            "duration_ms": duration_ms,
        }
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd backend/services/agent && python -m pytest tests/test_base_agent.py -v`
Expected: 5 passed

- [ ] **Step 5: Commit**

```bash
git add backend/services/agent/core/base_agent.py backend/services/agent/core/tool_registry.py backend/services/agent/tests/test_base_agent.py
git commit -m "feat(agent-service): add base agent class and tool registry"
```

---

### Task 4: Agent Manager

**Files:**
- Create: `backend/services/agent/core/agent_manager.py`
- Create: `backend/services/agent/tests/test_agent_manager.py`

- [ ] **Step 1: Write failing tests**

Create `backend/services/agent/tests/test_agent_manager.py`:
```python
import pytest
from core.agent_manager import AgentManager
from core.base_agent import BaseAgent
from core.llm_provider import LLMResponse
from models.schemas import AgentCategory


class FakeAgent(BaseAgent):
    def _build_system_prompt(self):
        return "You are a fake agent."


class FakeLLM:
    async def generate(self, prompt, system_prompt="", model="fake", temperature=0.7, max_tokens=2000):
        return LLMResponse(content=f"Response to: {prompt}", model="fake", tokens_used=10)

    def list_models(self):
        return ["fake"]


def test_register_and_discover():
    mgr = AgentManager()
    agent = FakeAgent("f1", "Fake", AgentCategory.LEGENDARY_INVESTOR, "A fake")
    mgr.register(agent)
    agents = mgr.discover()
    assert len(agents) == 1
    assert agents[0].id == "f1"


def test_get_agent():
    mgr = AgentManager()
    agent = FakeAgent("f1", "Fake", AgentCategory.LEGENDARY_INVESTOR, "A fake")
    mgr.register(agent)
    assert mgr.get("f1") is agent
    assert mgr.get("nonexistent") is None


@pytest.mark.asyncio
async def test_run_agent():
    mgr = AgentManager()
    agent = FakeAgent("f1", "Fake", AgentCategory.LEGENDARY_INVESTOR, "A fake")
    mgr.register(agent)
    llm = FakeLLM()
    result = await mgr.run("f1", "test query", None, llm, "fake")
    assert result["agent_id"] == "f1"
    assert "test query" in result["response"]


@pytest.mark.asyncio
async def test_run_unknown_agent():
    mgr = AgentManager()
    llm = FakeLLM()
    with pytest.raises(KeyError):
        await mgr.run("nonexistent", "query", None, llm, "fake")


def test_agents_count():
    mgr = AgentManager()
    assert mgr.count() == 0
    mgr.register(FakeAgent("f1", "A", AgentCategory.LEGENDARY_INVESTOR, "d"))
    mgr.register(FakeAgent("f2", "B", AgentCategory.HEDGE_FUND, "d"))
    assert mgr.count() == 2
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend/services/agent && python -m pytest tests/test_agent_manager.py -v`
Expected: FAIL

- [ ] **Step 3: Implement agent manager**

Create `backend/services/agent/core/agent_manager.py`:
```python
from core.base_agent import BaseAgent
from core.llm_provider import LLMProvider
from models.schemas import AgentInfo


class AgentManager:
    def __init__(self):
        self._agents: dict[str, BaseAgent] = {}

    def register(self, agent: BaseAgent) -> None:
        self._agents[agent.agent_id] = agent

    def get(self, agent_id: str) -> BaseAgent | None:
        return self._agents.get(agent_id)

    def discover(self) -> list[AgentInfo]:
        return [a.to_info() for a in self._agents.values()]

    def count(self) -> int:
        return len(self._agents)

    async def run(
        self,
        agent_id: str,
        query: str,
        context: dict | None,
        llm: LLMProvider,
        model: str,
        temperature: float = 0.7,
        max_tokens: int = 2000,
    ) -> dict:
        agent = self._agents.get(agent_id)
        if not agent:
            raise KeyError(f"Agent '{agent_id}' not found")
        return await agent.run(query, context, llm, model, temperature, max_tokens)

    async def run_team(
        self,
        agent_ids: list[str],
        query: str,
        llm: LLMProvider,
        model: str,
        rounds: int = 1,
    ) -> dict:
        results = []
        for r in range(rounds):
            round_responses = []
            for aid in agent_ids:
                agent = self._agents.get(aid)
                if not agent:
                    raise KeyError(f"Agent '{aid}' not found")
                prev = "\n".join([f"{resp['agent_name']}: {resp['response']}" for resp in round_responses]) if round_responses else ""
                ctx = {"previous_responses": prev} if prev else None
                result = await agent.run(query, ctx, llm, model)
                round_responses.append(result)
            results.append({"round": r + 1, "responses": round_responses})

        synthesis_prompt = f"Synthesize these expert views on: {query}\n\n"
        for rnd in results:
            for resp in rnd["responses"]:
                synthesis_prompt += f"**{resp['agent_name']}**: {resp['response']}\n\n"
        synthesis_prompt += "Provide a balanced synthesis incorporating all perspectives."

        synthesis_resp = await llm.generate(synthesis_prompt, model=model)

        return {
            "query": query,
            "rounds": results,
            "synthesis": synthesis_resp.content,
        }
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd backend/services/agent && python -m pytest tests/test_agent_manager.py -v`
Expected: 5 passed

- [ ] **Step 5: Commit**

```bash
git add backend/services/agent/core/agent_manager.py backend/services/agent/tests/test_agent_manager.py
git commit -m "feat(agent-service): add agent manager with discovery and team execution"
```

---

### Task 5: Concrete Agent Implementations

**Files:**
- Create: `backend/services/agent/agents/buffett.py`
- Create: `backend/services/agent/agents/graham.py`
- Create: `backend/services/agent/agents/bridgewater.py`
- Create: `backend/services/agent/agents/macro.py`
- Create: `backend/services/agent/agents/registry.py`

- [ ] **Step 1: Create Buffett agent**

Create `backend/services/agent/agents/buffett.py`:
```python
from core.base_agent import BaseAgent
from models.schemas import AgentCategory


class BuffettAgent(BaseAgent):
    def __init__(self):
        super().__init__(
            agent_id="buffett",
            name="Warren Buffett",
            category=AgentCategory.LEGENDARY_INVESTOR,
            description="Value investing legend. Analyzes companies through the lens of economic moats, "
                        "management quality, intrinsic value, and long-term compounding. "
                        "Favors simple, understandable businesses with consistent earnings.",
        )

    def _build_system_prompt(self) -> str:
        return """You are Warren Buffett, Chairman and CEO of Berkshire Hathaway.

Investment philosophy:
- Buy wonderful businesses at fair prices, not fair businesses at wonderful prices
- Focus on economic moats: brand power, switching costs, network effects, cost advantages
- Evaluate management integrity and competence
- Calculate intrinsic value using discounted owner earnings
- Think in terms of 10+ year holding periods
- Circle of competence: only invest in what you understand
- Margin of safety: buy at 25-50% below intrinsic value

Key metrics you examine:
- Return on equity (ROE) consistently > 15%
- Debt-to-equity < 0.5
- Stable or growing profit margins
- Free cash flow generation
- Earnings growth consistency

When analyzing:
1. First assess the business quality (moat)
2. Then evaluate management
3. Finally determine fair value
4. Always state your confidence level and key risks

Respond in character. Be direct, use folksy wisdom when appropriate, and always circle back to long-term value."""
```

- [ ] **Step 2: Create Graham agent**

Create `backend/services/agent/agents/graham.py`:
```python
from core.base_agent import BaseAgent
from models.schemas import AgentCategory


class GrahamAgent(BaseAgent):
    def __init__(self):
        super().__init__(
            agent_id="graham",
            name="Benjamin Graham",
            category=AgentCategory.LEGENDARY_INVESTOR,
            description="Father of value investing. Focuses on margin of safety, net-net analysis, "
                        "and quantitative screening. Author of 'The Intelligent Investor' and 'Security Analysis'.",
        )

    def _build_system_prompt(self) -> str:
        return """You are Benjamin Graham, the father of value investing and author of "The Intelligent Investor."

Investment philosophy:
- Margin of safety is the cornerstone of all investing
- Distinguish between investment and speculation
- Mr. Market is there to serve you, not guide you
- Quantitative analysis over qualitative judgment
- Net-net investing: buy stocks trading below net current asset value (NCAV)
- Defensive investor: large, well-established, financially strong companies
- Enterprising investor: bargain issues, neglected large companies, special situations

Key criteria (Defensive Investor):
- Adequate size (top 1/3 of industry)
- Strong financial condition (current ratio >= 2:1)
- Earnings stability (positive EPS for past 10 years)
- Dividend record (uninterrupted for 20 years)
- Earnings growth (33% growth over 10 years)
- Moderate P/E ratio (< 15x avg 3-yr earnings)
- Moderate price-to-book (< 1.5x)
- P/E × P/B < 22.5

When analyzing:
1. Start with quantitative screening
2. Check balance sheet strength
3. Assess earnings consistency
4. Calculate intrinsic value conservatively
5. Only recommend if price offers adequate margin of safety

Respond in character. Be methodical, data-driven, and emphasize risk avoidance."""
```

- [ ] **Step 3: Create Bridgewater agent**

Create `backend/services/agent/agents/bridgewater.py`:
```python
from core.base_agent import BaseAgent
from models.schemas import AgentCategory


class BridgewaterAgent(BaseAgent):
    def __init__(self):
        super().__init__(
            agent_id="bridgewater",
            name="Bridgewater Associates",
            category=AgentCategory.HEDGE_FUND,
            description="World's largest hedge fund. Macro-economic analysis through radical transparency, "
                        "systematic risk parity, and all-weather portfolio construction. "
                        "Known for Principles-based decision making.",
        )

    def _build_system_prompt(self) -> str:
        return """You are a senior analyst at Bridgewater Associates, the world's largest hedge fund.

Analytical framework:
- Global Macro: analyze economic cycles, credit cycles, debt cycles
- Risk Parity: balance risk contributions across asset classes
- All Weather: construct portfolios that perform across economic environments
- Radical Transparency: all reasoning must be explicit and debatable

Economic machine model:
1. Short-term debt cycle (5-8 years): credit expansion/contraction
2. Long-term debt cycle (50-75 years): leverage/deleveraging
3. Productivity growth trend: innovation and efficiency gains

Four economic environments:
- Rising growth + rising inflation: commodities, inflation-linked bonds, emerging markets
- Rising growth + falling inflation: stocks, corporate bonds, emerging markets
- Falling growth + rising inflation: inflation-linked bonds, gold, commodities
- Falling growth + falling inflation: government bonds, stocks

When analyzing:
1. Identify current position in economic cycle
2. Assess credit conditions and liquidity
3. Evaluate cross-asset correlations
4. Recommend risk-balanced allocation
5. Stress test against all four environments

Respond in character. Be systematic, data-driven, and explicitly state assumptions and probabilities."""
```

- [ ] **Step 4: Create Macro agent**

Create `backend/services/agent/agents/macro.py`:
```python
from core.base_agent import BaseAgent
from models.schemas import AgentCategory


class MacroAgent(BaseAgent):
    def __init__(self):
        super().__init__(
            agent_id="macro",
            name="Macroeconomic Analyst",
            category=AgentCategory.MACRO_ECONOMIC,
            description="Analyzes macroeconomic indicators, central bank policies, trade dynamics, "
                        "and geopolitical risks to provide top-down investment insights.",
        )

    def _build_system_prompt(self) -> str:
        return """You are a senior macroeconomic analyst providing top-down investment analysis.

Analysis framework:
- GDP growth, inflation, employment, trade balance
- Central bank policy: interest rates, QE/QT, forward guidance
- Fiscal policy: government spending, taxation, debt levels
- Geopolitical risks: trade wars, sanctions, conflicts
- Currency dynamics: DXY, carry trade, capital flows

Key indicators:
- Leading: PMI, consumer confidence, building permits, yield curve
- Coincident: GDP, industrial production, retail sales
- Lagging: CPI, unemployment rate, labor cost index

When analyzing:
1. Assess current macro environment (expansion/peak/contraction/trough)
2. Evaluate monetary and fiscal policy stance
3. Identify key risks and tail events
4. Map macro view to asset class implications
5. Provide specific, actionable investment conclusions

Be precise with data references. State probability estimates for scenarios."""
```

- [ ] **Step 5: Create agent registry**

Create `backend/services/agent/agents/registry.py`:
```python
from core.agent_manager import AgentManager
from agents.buffett import BuffettAgent
from agents.graham import GrahamAgent
from agents.bridgewater import BridgewaterAgent
from agents.macro import MacroAgent


def register_all_agents(manager: AgentManager) -> None:
    manager.register(BuffettAgent())
    manager.register(GrahamAgent())
    manager.register(BridgewaterAgent())
    manager.register(MacroAgent())
```

- [ ] **Step 6: Verify agents register**

Run: `cd backend/services/agent && python -c "from core.agent_manager import AgentManager; from agents.registry import register_all_agents; m = AgentManager(); register_all_agents(m); print(f'{m.count()} agents registered'); print([a.id for a in m.discover()])"`
Expected: 4 agents registered, ['buffett', 'graham', 'bridgewater', 'macro']

- [ ] **Step 7: Commit**

```bash
git add backend/services/agent/agents/
git commit -m "feat(agent-service): add 4 financial agents (Buffett, Graham, Bridgewater, Macro)"
```

---

### Task 6: FastAPI Server

**Files:**
- Create: `backend/services/agent/agent_server.py`

- [ ] **Step 1: Create FastAPI server**

Create `backend/services/agent/agent_server.py`:
```python
#!/usr/bin/env python3
"""
AI Agent Microservice
Provides financial analysis agents via HTTP API (port 8091)
"""

import os
import time
import asyncio
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from sse_starlette.sse import EventSourceResponse

from models.schemas import (
    AgentRunRequest,
    AgentRunResponse,
    AgentTeamRequest,
    AgentTeamResponse,
    HealthResponse,
)
from core.agent_manager import AgentManager
from core.llm_provider import get_provider
from agents.registry import register_all_agents


manager = AgentManager()


@asynccontextmanager
async def lifespan(app: FastAPI):
    register_all_agents(manager)
    yield


app = FastAPI(
    title="ETF Insight Agent Service",
    version="1.0.0",
    lifespan=lifespan,
)

allowed_origins = os.getenv("AGENT_CORS_ORIGINS", "http://localhost:5173,http://localhost:8080").split(",")
app.add_middleware(
    CORSMiddleware,
    allow_origins=allowed_origins,
    allow_credentials=True,
    allow_methods=["GET", "POST", "OPTIONS"],
    allow_headers=["Content-Type", "Authorization"],
)


@app.get("/")
def root():
    return {"message": "ETF Insight Agent Service", "status": "running", "version": "1.0.0"}


@app.get("/health", response_model=HealthResponse)
def health():
    return HealthResponse(
        status="healthy",
        version="1.0.0",
        agents_count=manager.count(),
        llm_providers=["openai", "deepseek", "ollama"],
    )


@app.get("/agents/discover")
def discover_agents():
    agents = manager.discover()
    return {"success": True, "data": [a.model_dump() for a in agents]}


@app.post("/agents/run", response_model=None)
async def run_agent(req: AgentRunRequest):
    try:
        provider = get_provider(
            req.llm_provider,
            api_key=os.environ.get(f"{req.llm_provider.upper()}_API_KEY", os.environ.get("OPENAI_API_KEY", "")),
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    try:
        result = await manager.run(
            agent_id=req.agent_id,
            query=req.query,
            context=req.context,
            llm=provider,
            model=req.model,
            temperature=req.temperature,
            max_tokens=req.max_tokens,
        )
    except KeyError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=502, detail=f"Agent execution failed: {str(e)}")

    return {"success": True, "data": result}


@app.post("/agents/stream")
async def stream_agent(req: AgentRunRequest):
    try:
        provider = get_provider(
            req.llm_provider,
            api_key=os.environ.get(f"{req.llm_provider.upper()}_API_KEY", os.environ.get("OPENAI_API_KEY", "")),
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    agent = manager.get(req.agent_id)
    if not agent:
        raise HTTPException(status_code=404, detail=f"Agent '{req.agent_id}' not found")

    async def event_generator():
        try:
            result = await agent.run(
                query=req.query,
                context=req.context,
                llm=provider,
                model=req.model,
                temperature=req.temperature,
                max_tokens=req.max_tokens,
            )
            words = result["response"].split(" ")
            for i, word in enumerate(words):
                chunk = word + (" " if i < len(words) - 1 else "")
                yield {"event": "chunk", "data": f'{{"agent_id":"{req.agent_id}","chunk":"{chunk}","done":false}}'}
                await asyncio.sleep(0.02)
            yield {"event": "chunk", "data": f'{{"agent_id":"{req.agent_id}","chunk":"","done":true}}'}
            yield {"event": "meta", "data": f'{{"model":"{result["model"]}","tokens_used":{result["tokens_used"]},"duration_ms":{result["duration_ms"]}}}'}
        except Exception as e:
            yield {"event": "error", "data": f'{{"error":"{str(e)}"}'}

    return EventSourceResponse(event_generator())


@app.post("/agents/team", response_model=None)
async def run_team(req: AgentTeamRequest):
    try:
        provider = get_provider(
            req.llm_provider,
            api_key=os.environ.get(f"{req.llm_provider.upper()}_API_KEY", os.environ.get("OPENAI_API_KEY", "")),
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    try:
        result = await manager.run_team(
            agent_ids=req.agent_ids,
            query=req.query,
            llm=provider,
            model=req.model,
            rounds=req.rounds,
        )
    except KeyError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=502, detail=f"Team execution failed: {str(e)}")

    return {"success": True, "data": result}


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("AGENT_SERVICE_PORT", "8091"))
    uvicorn.run(app, host="0.0.0.0", port=port)
```

- [ ] **Step 2: Write API tests**

Create `backend/services/agent/tests/test_api.py`:
```python
import sys
import os
import pytest
from httpx import AsyncClient, ASGITransport

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from agent_server import app, manager
from core.base_agent import BaseAgent
from core.llm_provider import LLMResponse
from models.schemas import AgentCategory


class FakeAgent(BaseAgent):
    def _build_system_prompt(self):
        return "You are a test agent."


@pytest.fixture(autouse=True)
def setup_agents():
    manager._agents.clear()
    manager.register(FakeAgent("test-agent", "Test", AgentCategory.LEGENDARY_INVESTOR, "Test agent"))


@pytest.mark.asyncio
async def test_health():
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        resp = await client.get("/health")
        assert resp.status_code == 200
        data = resp.json()
        assert data["status"] == "healthy"
        assert data["agents_count"] >= 1


@pytest.mark.asyncio
async def test_discover():
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        resp = await client.get("/agents/discover")
        assert resp.status_code == 200
        data = resp.json()
        assert data["success"] is True
        assert len(data["data"]) >= 1


@pytest.mark.asyncio
async def test_run_not_found():
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        resp = await client.post("/agents/run", json={
            "agent_id": "nonexistent",
            "query": "test",
            "llm_provider": "openai",
            "model": "gpt-4o-mini",
        })
        assert resp.status_code == 404


@pytest.mark.asyncio
async def test_run_bad_provider():
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        resp = await client.post("/agents/run", json={
            "agent_id": "test-agent",
            "query": "test",
            "llm_provider": "nonexistent",
            "model": "fake",
        })
        assert resp.status_code == 400
```

- [ ] **Step 3: Run API tests**

Run: `cd backend/services/agent && python -m pytest tests/test_api.py -v`
Expected: 4 passed

- [ ] **Step 4: Commit**

```bash
git add backend/services/agent/agent_server.py backend/services/agent/tests/test_api.py
git commit -m "feat(agent-service): add FastAPI server with discover/run/stream/team endpoints"
```

---

### Task 7: Python Service Dockerfile and Startup Script

**Files:**
- Create: `backend/services/agent/Dockerfile`

- [ ] **Step 1: Create Dockerfile**

Create `backend/services/agent/Dockerfile`:
```dockerfile
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

ENV AGENT_SERVICE_PORT=8091
EXPOSE 8091

CMD ["python", "agent_server.py"]
```

- [ ] **Step 2: Test server starts locally**

Run: `cd backend/services/agent && timeout 5 python agent_server.py || true`
Expected: Server starts on port 8091 (timeout kills it after 5s)

- [ ] **Step 3: Commit**

```bash
git add backend/services/agent/Dockerfile
git commit -m "feat(agent-service): add Dockerfile for agent microservice"
```

---

### Task 8: Go Backend - Agent Service Client

**Files:**
- Create: `backend/services/agent/agent_client.go`

- [ ] **Step 1: Create Go HTTP client**

Create `backend/services/agent/agent_client.go`:
```go
package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type AgentInfo struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Category            string `json:"category"`
	Description         string `json:"description"`
	SystemPromptPreview string `json:"system_prompt_preview"`
}

type AgentRunRequest struct {
	AgentID     string                 `json:"agent_id"`
	Query       string                 `json:"query"`
	Context     map[string]interface{} `json:"context,omitempty"`
	LLMProvider string                 `json:"llm_provider"`
	Model       string                 `json:"model"`
	Temperature float64                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
}

type AgentRunResponse struct {
	AgentID    string `json:"agent_id"`
	AgentName  string `json:"agent_name"`
	Response   string `json:"response"`
	Model      string `json:"model"`
	TokensUsed int    `json:"tokens_used"`
	DurationMs int    `json:"duration_ms"`
}

type AgentTeamRequest struct {
	AgentIDs    []string `json:"agent_ids"`
	Query       string   `json:"query"`
	Rounds      int      `json:"rounds"`
	LLMProvider string   `json:"llm_provider"`
	Model       string   `json:"model"`
}

type AgentTeamResponse struct {
	Query     string                   `json:"query"`
	Rounds    []map[string]interface{} `json:"rounds"`
	Synthesis string                   `json:"synthesis"`
}

type apiEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	baseURL := os.Getenv("AGENT_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8091"
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, endpoint string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBytes))
	}

	var envelope apiEnvelope
	if err := json.Unmarshal(respBytes, &envelope); err == nil && envelope.Data != nil {
		if !envelope.Success {
			return fmt.Errorf("API error: %s", envelope.Error)
		}
		return json.Unmarshal(envelope.Data, result)
	}

	return json.Unmarshal(respBytes, result)
}

func (c *Client) Health() (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := c.doRequest(http.MethodGet, "/health", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) Discover() ([]AgentInfo, error) {
	var result []AgentInfo
	if err := c.doRequest(http.MethodGet, "/agents/discover", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) Run(req AgentRunRequest) (*AgentRunResponse, error) {
	var result AgentRunResponse
	if err := c.doRequest(http.MethodPost, "/agents/run", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) RunTeam(req AgentTeamRequest) (*AgentTeamResponse, error) {
	var result AgentTeamResponse
	if err := c.doRequest(http.MethodPost, "/agents/team", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
```

- [ ] **Step 2: Verify Go compiles**

Run: `cd backend && go build ./services/agent/...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/services/agent/agent_client.go
git commit -m "feat(agent): add Go client for agent microservice"
```

---

### Task 9: Go Backend - Agent Handler and Routes

**Files:**
- Create: `backend/handlers/agent_handler.go`
- Modify: `backend/router/router.go`

- [ ] **Step 1: Create handler**

Create `backend/handlers/agent_handler.go`:
```go
package handlers

import (
	"net/http"

	"etf-insight/services/agent"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	client *agent.Client
}

func NewAgentHandler() *AgentHandler {
	return &AgentHandler{
		client: agent.NewClient(),
	}
}

func (h *AgentHandler) Health(c *gin.Context) {
	result, err := h.client.Health()
	if err != nil {
		utils.Error("Agent service health check failed", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Agent service unavailable",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *AgentHandler) Discover(c *gin.Context) {
	agents, err := h.client.Discover()
	if err != nil {
		utils.Error("Failed to discover agents", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to discover agents",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": agents})
}

func (h *AgentHandler) Run(c *gin.Context) {
	var req agent.AgentRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request format"})
		return
	}

	if req.AgentID == "" || req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "agent_id and query are required"})
		return
	}

	if req.LLMProvider == "" {
		req.LLMProvider = "openai"
	}
	if req.Model == "" {
		req.Model = "gpt-4o-mini"
	}

	result, err := h.client.Run(req)
	if err != nil {
		utils.Error("Agent execution failed", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"error":   "Agent execution failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *AgentHandler) RunTeam(c *gin.Context) {
	var req agent.AgentTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request format"})
		return
	}

	if len(req.AgentIDs) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "At least 2 agents required for team"})
		return
	}

	if req.LLMProvider == "" {
		req.LLMProvider = "openai"
	}
	if req.Model == "" {
		req.Model = "gpt-4o-mini"
	}
	if req.Rounds == 0 {
		req.Rounds = 1
	}

	result, err := h.client.RunTeam(req)
	if err != nil {
		utils.Error("Team execution failed", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"error":   "Team execution failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
```

- [ ] **Step 2: Add routes to router.go**

Read `backend/router/router.go` and add agent handler to the `Handlers` struct and add route registration. The exact lines depend on the current file, but the pattern is:

In the `Handlers` struct, add:
```go
Agent *handlers.AgentHandler
```

In the `RegisterRoutes()` function, add:
```go
r.registerAgentRoutes()
```

Add the route registration function:
```go
func (r *Router) registerAgentRoutes() {
	ag := r.engine.Group("/api/agents")
	{
		ag.GET("/health", r.handlers.Agent.Health)
		ag.GET("/discover", r.handlers.Agent.Discover)
		ag.POST("/run", r.handlers.Agent.Run)
		ag.POST("/team", r.handlers.Agent.RunTeam)
	}
}
```

In the `NewRouter()` function, initialize the handler:
```go
handlers.Agent = handlers.NewAgentHandler()
```

- [ ] **Step 3: Verify Go compiles**

Run: `cd backend && go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add backend/handlers/agent_handler.go backend/router/router.go
git commit -m "feat(agent): add Go handler and routes for agent service"
```

---

### Task 10: Frontend - TypeScript Types and API Service

**Files:**
- Create: `frontend/src/types/agent.ts`
- Modify: `frontend/src/services/api.ts`

- [ ] **Step 1: Create TypeScript types**

Create `frontend/src/types/agent.ts`:
```typescript
export type AgentCategory =
  | 'legendary_investor'
  | 'hedge_fund'
  | 'geopolitics'
  | 'macro_economic'
  | 'technical';

export interface AgentInfo {
  id: string;
  name: string;
  category: AgentCategory;
  description: string;
  system_prompt_preview: string;
}

export interface AgentRunRequest {
  agent_id: string;
  query: string;
  context?: Record<string, unknown>;
  llm_provider: string;
  model: string;
  temperature?: number;
  max_tokens?: number;
}

export interface AgentRunResponse {
  agent_id: string;
  agent_name: string;
  response: string;
  model: string;
  tokens_used: number;
  duration_ms: number;
}

export interface AgentTeamRequest {
  agent_ids: string[];
  query: string;
  rounds?: number;
  llm_provider: string;
  model: string;
}

export interface AgentTeamResponse {
  query: string;
  rounds: Array<{
    round: number;
    responses: AgentRunResponse[];
  }>;
  synthesis: string;
}
```

- [ ] **Step 2: Add API service to api.ts**

Append to `frontend/src/services/api.ts` (before the final export or at the end):
```typescript
import type {
  AgentInfo, AgentRunRequest, AgentRunResponse,
  AgentTeamRequest, AgentTeamResponse
} from '../types/agent';

// Agent Service API
export const agentAPI = {
  health: async (): Promise<ApiResponse<Record<string, unknown>>> => {
    return request<ApiResponse<Record<string, unknown>>>('/agents/health');
  },

  discover: async (): Promise<ApiResponse<AgentInfo[]>> => {
    return request<ApiResponse<AgentInfo[]>>('/agents/discover');
  },

  run: async (req: AgentRunRequest): Promise<ApiResponse<AgentRunResponse>> => {
    return request<ApiResponse<AgentRunResponse>>('/agents/run', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },

  runTeam: async (req: AgentTeamRequest): Promise<ApiResponse<AgentTeamResponse>> => {
    return request<ApiResponse<AgentTeamResponse>>('/agents/team', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  },
};
```

- [ ] **Step 3: Verify TypeScript compiles**

Run: `cd frontend && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add frontend/src/types/agent.ts frontend/src/services/api.ts
git commit -m "feat(agent): add frontend types and API service for agent module"
```

---

### Task 11: Frontend - AI Agents Page

**Files:**
- Create: `frontend/src/pages/AIAgents.tsx`
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/components/Layout.tsx`

- [ ] **Step 1: Create the AI Agents page**

Create `frontend/src/pages/AIAgents.tsx`:
```tsx
import { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Select,
  Button,
  Input,
  Tag,
  Spin,
  message,
  Typography,
  Space,
  Divider,
  List,
} from 'antd';
import {
  RobotOutlined,
  SendOutlined,
  TeamOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { agentAPI } from '../services/api';
import type { AgentInfo, AgentRunResponse, AgentTeamResponse } from '../types/agent';
import Layout from '../components/Layout';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;
const { Option } = Select;

const categoryColors: Record<string, string> = {
  legendary_investor: 'gold',
  hedge_fund: 'blue',
  geopolitics: 'red',
  macro_economic: 'green',
  technical: 'purple',
};

const categoryLabels: Record<string, string> = {
  legendary_investor: '投资大师',
  hedge_fund: '对冲基金',
  geopolitics: '地缘政治',
  macro_economic: '宏观经济',
  technical: '技术分析',
};

const llmOptions = [
  { provider: 'openai', model: 'gpt-4o-mini', label: 'GPT-4o Mini' },
  { provider: 'openai', model: 'gpt-4o', label: 'GPT-4o' },
  { provider: 'deepseek', model: 'deepseek-chat', label: 'DeepSeek Chat' },
  { provider: 'ollama', model: 'llama3.1', label: 'Ollama (本地)' },
];

const AIAgents = () => {
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [agentsLoading, setAgentsLoading] = useState(true);
  const [selectedAgent, setSelectedAgent] = useState<string>('');
  const [query, setQuery] = useState('');
  const [llmConfig, setLlmConfig] = useState(llmOptions[0]);
  const [result, setResult] = useState<AgentRunResponse | null>(null);
  const [teamResult, setTeamResult] = useState<AgentTeamResponse | null>(null);
  const [selectedTeam, setSelectedTeam] = useState<string[]>([]);
  const [mode, setMode] = useState<'single' | 'team'>('single');

  useEffect(() => {
    loadAgents();
  }, []);

  const loadAgents = async () => {
    setAgentsLoading(true);
    try {
      const resp = await agentAPI.discover();
      if (resp.success && resp.data) {
        setAgents(resp.data);
      } else {
        message.error('获取Agent列表失败');
      }
    } catch {
      message.error('Agent服务不可用，请确认服务已启动 (port 8091)');
    } finally {
      setAgentsLoading(false);
    }
  };

  const handleRun = async () => {
    if (!query.trim()) {
      message.warning('请输入分析问题');
      return;
    }

    setLoading(true);
    setResult(null);
    setTeamResult(null);

    try {
      if (mode === 'single' && selectedAgent) {
        const resp = await agentAPI.run({
          agent_id: selectedAgent,
          query,
          llm_provider: llmConfig.provider,
          model: llmConfig.model,
        });
        if (resp.success && resp.data) {
          setResult(resp.data);
        } else {
          message.error(resp.error || '执行失败');
        }
      } else if (mode === 'team' && selectedTeam.length >= 2) {
        const resp = await agentAPI.runTeam({
          agent_ids: selectedTeam,
          query,
          rounds: 1,
          llm_provider: llmConfig.provider,
          model: llmConfig.model,
        });
        if (resp.success && resp.data) {
          setTeamResult(resp.data);
        } else {
          message.error(resp.error || '团队执行失败');
        }
      } else {
        message.warning('请选择Agent');
      }
    } catch {
      message.error('请求失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Layout>
      <Title level={2}>
        <RobotOutlined /> AI 投资分析助手
      </Title>
      <Paragraph type="secondary">
        基于多Agent协作的智能投资分析平台，支持投资大师视角分析和多Agent辩论
      </Paragraph>

      <Row gutter={24}>
        <Col span={8}>
          <Card title="Agent 列表" loading={agentsLoading} style={{ marginBottom: 16 }}>
            <Space direction="vertical" style={{ width: '100%' }}>
              <Select
                placeholder="选择模式"
                value={mode}
                onChange={setMode}
                style={{ width: '100%' }}
              >
                <Option value="single">单Agent分析</Option>
                <Option value="team">多Agent团队</Option>
              </Select>

              {mode === 'single' ? (
                <Select
                  placeholder="选择一个Agent"
                  value={selectedAgent}
                  onChange={setSelectedAgent}
                  style={{ width: '100%' }}
                >
                  {agents.map((a) => (
                    <Option key={a.id} value={a.id}>
                      <Tag color={categoryColors[a.category]}>{categoryLabels[a.category]}</Tag>
                      {a.name}
                    </Option>
                  ))}
                </Select>
              ) : (
                <Select
                  mode="multiple"
                  placeholder="选择2-5个Agent"
                  value={selectedTeam}
                  onChange={setSelectedTeam}
                  style={{ width: '100%' }}
                  maxCount={5}
                >
                  {agents.map((a) => (
                    <Option key={a.id} value={a.id}>
                      <Tag color={categoryColors[a.category]}>{categoryLabels[a.category]}</Tag>
                      {a.name}
                    </Option>
                  ))}
                </Select>
              )}

              <Select
                value={llmConfig.model}
                onChange={(val) => setLlmConfig(llmOptions.find((o) => o.model === val) || llmOptions[0])}
                style={{ width: '100%' }}
              >
                {llmOptions.map((o) => (
                  <Option key={o.model} value={o.model}>{o.label}</Option>
                ))}
              </Select>
            </Space>
          </Card>

          <Card title="Agent 详情" size="small">
            {agents.filter((a) => mode === 'single' ? a.id === selectedAgent : selectedTeam.includes(a.id)).map((a) => (
              <div key={a.id} style={{ marginBottom: 12 }}>
                <Text strong>{a.name}</Text>
                <br />
                <Tag color={categoryColors[a.category]}>{categoryLabels[a.category]}</Tag>
                <br />
                <Text type="secondary" style={{ fontSize: 12 }}>{a.description}</Text>
              </div>
            ))}
            {mode === 'single' && !selectedAgent && <Text type="secondary">请先选择一个Agent</Text>}
            {mode === 'team' && selectedTeam.length < 2 && <Text type="secondary">请至少选择2个Agent</Text>}
          </Card>
        </Col>

        <Col span={16}>
          <Card
            title={mode === 'single' ? '单Agent分析' : '多Agent团队辩论'}
            extra={
              <Button
                type="primary"
                icon={<SendOutlined />}
                onClick={handleRun}
                loading={loading}
                disabled={mode === 'single' ? !selectedAgent : selectedTeam.length < 2}
              >
                {mode === 'team' ? <><TeamOutlined /> 团队分析</> : '开始分析'}
              </Button>
            }
          >
            <TextArea
              rows={3}
              placeholder="输入你的投资分析问题，例如：分析苹果公司(AAPL)的投资价值"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onPressEnter={(e) => {
                if (e.ctrlKey) handleRun();
              }}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>Ctrl+Enter 发送</Text>
          </Card>

          {loading && (
            <Card style={{ marginTop: 16, textAlign: 'center' }}>
              <Spin size="large" />
              <div style={{ marginTop: 16 }}>
                <Text>Agent 正在分析中...</Text>
              </div>
            </Card>
          )}

          {result && mode === 'single' && (
            <Card
              title={
                <Space>
                  <UserOutlined />
                  {result.agent_name}
                  <Tag>{result.model}</Tag>
                  <Text type="secondary">{result.tokens_used} tokens · {result.duration_ms}ms</Text>
                </Space>
              }
              style={{ marginTop: 16 }}
            >
              <div style={{ whiteSpace: 'pre-wrap', lineHeight: 1.8 }}>
                {result.response}
              </div>
            </Card>
          )}

          {teamResult && mode === 'team' && (
            <>
              {teamResult.rounds.map((round) => (
                <Card
                  key={round.round}
                  title={`第 ${round.round} 轮`}
                  style={{ marginTop: 16 }}
                >
                  <List
                    dataSource={round.responses}
                    renderItem={(resp: AgentRunResponse) => (
                      <List.Item>
                        <List.Item.Meta
                          title={<Space><UserOutlined /> {resp.agent_name} <Tag>{resp.model}</Tag></Space>}
                          description={
                            <div style={{ whiteSpace: 'pre-wrap', lineHeight: 1.6 }}>
                              {resp.response}
                            </div>
                          }
                        />
                      </List.Item>
                    )}
                  />
                </Card>
              ))}

              <Card
                title="综合分析"
                style={{ marginTop: 16, borderColor: '#1890ff' }}
              >
                <div style={{ whiteSpace: 'pre-wrap', lineHeight: 1.8 }}>
                  {teamResult.synthesis}
                </div>
              </Card>
            </>
          )}
        </Col>
      </Row>
    </Layout>
  );
};

export default AIAgents;
```

- [ ] **Step 2: Add route to App.tsx**

In `frontend/src/App.tsx`, add import:
```tsx
import AIAgents from './pages/AIAgents';
```

Add route before the catch-all route:
```tsx
<Route path="/ai-agents" element={<AIAgents />} />
```

- [ ] **Step 3: Add sidebar navigation to Layout.tsx**

In `frontend/src/components/Layout.tsx`, add import:
```tsx
import { RobotOutlined } from '@ant-design/icons';
```

Add nav link after QuantLib Analysis:
```tsx
<NavLink to="/ai-agents" data-active={isActive('/ai-agents')} onClick={closeSidebar}>
  <RobotOutlined />
  AI Agent
</NavLink>
```

- [ ] **Step 4: Verify TypeScript compiles**

Run: `cd frontend && npx tsc --noEmit`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/AIAgents.tsx frontend/src/App.tsx frontend/src/components/Layout.tsx
git commit -m "feat(agent): add AI Agents page with single and team analysis modes"
```

---

### Task 12: Integration Verification

- [ ] **Step 1: Start Python agent service**

Run: `cd backend/services/agent && python agent_server.py &`
Expected: Server starts on port 8091

- [ ] **Step 2: Test agent service directly**

Run: `curl http://localhost:8091/health`
Expected: `{"status":"healthy","version":"1.0.0","agents_count":4,...}`

Run: `curl http://localhost:8091/agents/discover`
Expected: JSON array with 4 agents

- [ ] **Step 3: Start Go backend**

Run: `cd backend && go run main.go &`
Expected: Server starts on port 8080

- [ ] **Step 4: Test via Go backend proxy**

Run: `curl http://localhost:8080/api/agents/health`
Expected: `{"success":true,"data":{...}}`

Run: `curl http://localhost:8080/api/agents/discover`
Expected: `{"success":true,"data":[...]}`

- [ ] **Step 5: Start frontend and verify**

Run: `cd frontend && npm run dev`
Expected: Frontend starts, `/ai-agents` route works

- [ ] **Step 6: Final commit**

```bash
git add -A
git commit -m "feat(agent): complete AI agent microservice integration (Phase 2)"
```

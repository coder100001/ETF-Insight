"""Pydantic v2 schemas for the AI Agent microservice.

Defines request/response models for single-agent runs, team discussions,
streaming responses, and health checks.
"""

from pydantic import BaseModel, Field
from typing import Any, Optional
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
    query: str = Field(max_length=10000)
    context: Optional[dict[str, Any]] = None
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
    query: str = Field(max_length=10000)
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

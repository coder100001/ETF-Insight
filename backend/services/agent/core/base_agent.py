from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Optional

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
        tools: Optional[ToolRegistry] = None,
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
        context: Optional[dict],
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

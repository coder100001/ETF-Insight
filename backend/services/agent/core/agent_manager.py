from __future__ import annotations

from typing import Optional

from core.base_agent import BaseAgent
from core.llm_provider import LLMProvider
from models.schemas import AgentInfo


class AgentManager:
    def __init__(self):
        self._agents: dict[str, BaseAgent] = {}

    def register(self, agent: BaseAgent) -> None:
        self._agents[agent.agent_id] = agent

    def get(self, agent_id: str) -> Optional[BaseAgent]:
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

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

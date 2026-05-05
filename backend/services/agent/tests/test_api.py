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

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

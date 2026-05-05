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

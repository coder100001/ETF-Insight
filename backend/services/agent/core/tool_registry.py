from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Callable, Optional
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

    def get(self, name: str) -> Optional[ToolDef]:
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

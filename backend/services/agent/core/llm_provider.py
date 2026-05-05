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

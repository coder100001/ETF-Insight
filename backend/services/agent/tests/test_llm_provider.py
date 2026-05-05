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

#!/usr/bin/env python3
"""
AI Agent Microservice
Provides financial analysis agents via HTTP API (port 8091)
"""

import os
import json
import asyncio
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from sse_starlette.sse import EventSourceResponse

from models.schemas import (
    AgentRunRequest,
    AgentRunResponse,
    AgentTeamRequest,
    AgentTeamResponse,
    HealthResponse,
)
from core.agent_manager import AgentManager
from core.llm_provider import get_provider
from agents.registry import register_all_agents


manager = AgentManager()


@asynccontextmanager
async def lifespan(app: FastAPI):
    register_all_agents(manager)
    yield


app = FastAPI(
    title="ETF Insight Agent Service",
    version="1.0.0",
    lifespan=lifespan,
)

allowed_origins = os.getenv("AGENT_CORS_ORIGINS", "http://localhost:5173,http://localhost:8080").split(",")
app.add_middleware(
    CORSMiddleware,
    allow_origins=allowed_origins,
    allow_credentials=True,
    allow_methods=["GET", "POST", "OPTIONS"],
    allow_headers=["Content-Type", "Authorization"],
)


@app.get("/")
def root():
    return {"message": "ETF Insight Agent Service", "status": "running", "version": "1.0.0"}


@app.get("/health", response_model=HealthResponse)
def health():
    return HealthResponse(
        status="healthy",
        version="1.0.0",
        agents_count=manager.count(),
        llm_providers=["openai", "deepseek", "ollama"],
    )


@app.get("/agents/discover")
def discover_agents():
    agents = manager.discover()
    return {"success": True, "data": [a.model_dump() for a in agents]}


@app.post("/agents/run", response_model=None)
async def run_agent(req: AgentRunRequest):
    try:
        provider = get_provider(
            req.llm_provider,
            api_key=os.environ.get(f"{req.llm_provider.upper()}_API_KEY", os.environ.get("OPENAI_API_KEY", "")),
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    try:
        result = await manager.run(
            agent_id=req.agent_id,
            query=req.query,
            context=req.context,
            llm=provider,
            model=req.model,
            temperature=req.temperature,
            max_tokens=req.max_tokens,
        )
    except KeyError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=502, detail=f"Agent execution failed: {str(e)}")

    return {"success": True, "data": result}


@app.post("/agents/stream")
async def stream_agent(req: AgentRunRequest):
    try:
        provider = get_provider(
            req.llm_provider,
            api_key=os.environ.get(f"{req.llm_provider.upper()}_API_KEY", os.environ.get("OPENAI_API_KEY", "")),
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    agent = manager.get(req.agent_id)
    if not agent:
        raise HTTPException(status_code=404, detail=f"Agent '{req.agent_id}' not found")

    async def event_generator():
        try:
            result = await agent.run(
                query=req.query,
                context=req.context,
                llm=provider,
                model=req.model,
                temperature=req.temperature,
                max_tokens=req.max_tokens,
            )
            words = result["response"].split(" ")
            for i, word in enumerate(words):
                chunk = word + (" " if i < len(words) - 1 else "")
                chunk_data = json.dumps({"agent_id": req.agent_id, "chunk": chunk, "done": False})
                yield {"event": "chunk", "data": chunk_data}
                await asyncio.sleep(0.02)
            done_data = json.dumps({"agent_id": req.agent_id, "chunk": "", "done": True})
            yield {"event": "chunk", "data": done_data}
            meta_data = json.dumps({"model": result["model"], "tokens_used": result["tokens_used"], "duration_ms": result["duration_ms"]})
            yield {"event": "meta", "data": meta_data}
        except Exception as e:
            error_data = json.dumps({"error": str(e)})
            yield {"event": "error", "data": error_data}

    return EventSourceResponse(event_generator())


@app.post("/agents/team", response_model=None)
async def run_team(req: AgentTeamRequest):
    try:
        provider = get_provider(
            req.llm_provider,
            api_key=os.environ.get(f"{req.llm_provider.upper()}_API_KEY", os.environ.get("OPENAI_API_KEY", "")),
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    try:
        result = await manager.run_team(
            agent_ids=req.agent_ids,
            query=req.query,
            llm=provider,
            model=req.model,
            rounds=req.rounds,
        )
    except KeyError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=502, detail=f"Team execution failed: {str(e)}")

    return {"success": True, "data": result}


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("AGENT_SERVICE_PORT", "8091"))
    uvicorn.run(app, host="0.0.0.0", port=port)

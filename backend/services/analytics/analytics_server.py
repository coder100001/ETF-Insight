#!/usr/bin/env python3
"""
Analytics Microservice - Unified analytics API (port 8093)
Directly wraps FinceptTerminal analytics modules.
"""

import os
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from routers import optimization, risk, analytics, planning


@asynccontextmanager
async def lifespan(app: FastAPI):
    yield


app = FastAPI(
    title="ETF Insight Analytics Service",
    version="1.0.0",
    lifespan=lifespan,
)

allowed_origins = os.getenv(
    "ANALYTICS_CORS_ORIGINS",
    "http://localhost:5173,http://localhost:8080"
).split(",")
app.add_middleware(
    CORSMiddleware,
    allow_origins=allowed_origins,
    allow_credentials=True,
    allow_methods=["GET", "POST"],
    allow_headers=["*"],
)

app.include_router(optimization.router, prefix="/api/optimization", tags=["Portfolio Optimization"])
app.include_router(risk.router, prefix="/api/risk", tags=["Risk Management"])
app.include_router(analytics.router, prefix="/api/analytics", tags=["Portfolio Analytics"])
app.include_router(planning.router, prefix="/api/planning", tags=["Portfolio Planning"])


@app.get("/health")
def health():
    return {
        "status": "healthy",
        "version": "1.0.0",
        "modules": [
            "portfolio_optimization",
            "risk_management",
            "portfolio_analytics",
            "portfolio_management",
            "portfolio_planning",
            "active_management",
            "economics_markets",
            "ffn_analysis",
            "math_engine",
            "behavioral_finance",
            "etf_analytics",
        ],
    }


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("ANALYTICS_SERVICE_PORT", "8093"))
    uvicorn.run(app, host="0.0.0.0", port=port)

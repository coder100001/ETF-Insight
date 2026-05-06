#!/usr/bin/env python3
"""
Data Source Microservice - Unified data API (port 8092)
Directly wraps FinceptTerminal data source scripts.
"""

import os
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from routers import fred, worldbank, imf, yfinance, akshare, coingecko


@asynccontextmanager
async def lifespan(app: FastAPI):
    yield


app = FastAPI(
    title="ETF Insight Data Service",
    version="1.0.0",
    lifespan=lifespan,
)

allowed_origins = os.getenv(
    "DATA_CORS_ORIGINS",
    "http://localhost:5173,http://localhost:8080"
).split(",")
app.add_middleware(
    CORSMiddleware,
    allow_origins=allowed_origins,
    allow_credentials=True,
    allow_methods=["GET", "POST"],
    allow_headers=["*"],
)

app.include_router(fred.router, prefix="/api/fred", tags=["FRED"])
app.include_router(worldbank.router, prefix="/api/worldbank", tags=["World Bank"])
app.include_router(imf.router, prefix="/api/imf", tags=["IMF"])
app.include_router(yfinance.router, prefix="/api/yfinance", tags=["Yahoo Finance"])
app.include_router(akshare.router, prefix="/api/akshare", tags=["AkShare"])
app.include_router(coingecko.router, prefix="/api/coingecko", tags=["CoinGecko"])


@app.get("/health")
def health():
    return {
        "status": "healthy",
        "version": "1.0.0",
        "sources": ["fred", "worldbank", "imf", "yfinance", "akshare", "coingecko"],
    }


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("DATA_SERVICE_PORT", "8092"))
    uvicorn.run(app, host="0.0.0.0", port=port)

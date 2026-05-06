from fastapi import APIRouter, Query
from sources.yfinance_data import (
    get_quote, get_historical, get_info,
    get_financials, get_batch_quotes, search_symbols,
)

router = APIRouter()


@router.get("/quote/{symbol}")
def yf_quote(symbol: str):
    return get_quote(symbol)


@router.get("/historical/{symbol}")
def yf_historical(
    symbol: str,
    start_date: str = Query(None),
    end_date: str = Query(None),
    interval: str = Query("1d"),
):
    return get_historical(symbol, start_date, end_date, interval)


@router.get("/info/{symbol}")
def yf_info(symbol: str):
    return get_info(symbol)


@router.get("/financials/{symbol}")
def yf_financials(symbol: str):
    return get_financials(symbol)


@router.post("/batch-quotes")
def yf_batch_quotes(symbols: list[str]):
    return get_batch_quotes(symbols)


@router.get("/search")
def yf_search(q: str, limit: int = Query(10)):
    return search_symbols(q, limit)

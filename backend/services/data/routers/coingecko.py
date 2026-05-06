from fastapi import APIRouter, Query
from sources.coingecko import (
    get_simple_price, get_coin_markets, get_coin_details,
    get_market_chart, get_trending_coins, get_global_data,
)

router = APIRouter()


@router.get("/price")
def cg_price(
    ids: str,
    vs_currencies: str = Query("usd"),
):
    return get_simple_price(ids, vs_currencies)


@router.get("/markets")
def cg_markets(
    vs_currency: str = Query("usd"),
    per_page: int = Query(20, ge=1, le=250),
):
    return get_coin_markets(vs_currency, per_page=per_page)


@router.get("/coin/{coin_id}")
def cg_coin(coin_id: str):
    return get_coin_details(coin_id)


@router.get("/chart/{coin_id}")
def cg_chart(
    coin_id: str,
    vs_currency: str = Query("usd"),
    days: str = Query("30"),
):
    return get_market_chart(coin_id, vs_currency, days)


@router.get("/trending")
def cg_trending():
    return get_trending_coins()


@router.get("/global")
def cg_global():
    return get_global_data()

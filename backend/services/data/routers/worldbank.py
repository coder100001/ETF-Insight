from fastapi import APIRouter, Query
from sources.worldbank_data import get_indicators, get_countries, get_economic_snapshot

router = APIRouter()


@router.get("/indicators/{country_code}/{indicator}")
def wb_indicators(
    country_code: str,
    indicator: str,
    date_range: str = Query(None),
):
    return get_indicators(country_code, indicator, date_range)


@router.get("/countries")
def wb_countries(
    region: str = Query(None),
    income_level: str = Query(None),
):
    return get_countries(region, income_level)


@router.get("/snapshot/{country_code}")
def wb_snapshot(country_code: str):
    return get_economic_snapshot(country_code)

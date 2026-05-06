from fastapi import APIRouter, Query
from sources.imf_data import IMFDataWrapper

router = APIRouter()
wrapper = IMFDataWrapper()


@router.get("/economic-indicators")
def imf_indicators(
    countries: str = Query(None),
    symbols: str = Query(None),
    frequency: str = Query("A"),
):
    return wrapper.get_economic_indicators(countries, symbols, frequency)


@router.get("/trade")
def imf_trade(
    countries: str = Query(None),
    counterparts: str = Query(None),
    direction: str = Query("E"),
):
    return wrapper.get_direction_of_trade(countries, counterparts, direction)

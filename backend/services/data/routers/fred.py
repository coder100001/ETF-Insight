from fastapi import APIRouter, Query
from sources.fred_data import get_series, search_series, get_categories

router = APIRouter()


@router.get("/series/{series_id}")
def fred_series(
    series_id: str,
    start_date: str = Query(None),
    end_date: str = Query(None),
    frequency: str = Query(None),
    transform: str = Query(None),
):
    return get_series(series_id, start_date, end_date, frequency, transform)


@router.get("/search")
def fred_search(q: str, limit: int = Query(10, ge=1, le=100)):
    return search_series(q, limit)


@router.get("/categories/{category_id}")
def fred_categories(category_id: str = "0"):
    return get_categories(category_id)

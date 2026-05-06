from fastapi import APIRouter, Query
from modules.portfolio_analytics import PortfolioAnalytics, CAPMAnalysis

router = APIRouter()
pa = PortfolioAnalytics()
capm = CAPMAnalysis()


@router.post("/capm")
def calculate_capm(
    risk_free_rate: float = Query(..., ge=0, le=1),
    market_return: float = Query(..., ge=0, le=1),
    beta: float = Query(..., ge=0, le=5),
):
    """Calculate expected return using CAPM."""
    try:
        expected_return = capm.calculate_expected_return(risk_free_rate, market_return, beta)
        return {
            "success": True,
            "data": {
                "expected_return": expected_return,
                "risk_free_rate": risk_free_rate,
                "market_return": market_return,
                "beta": beta,
            },
        }
    except Exception as e:
        return {"success": False, "error": str(e)}


@router.post("/metrics")
def calculate_metrics(
    symbols: list[str],
    period: str = Query("1y"),
):
    """Calculate portfolio metrics."""
    try:
        metrics = pa.calculate_portfolio_metrics(symbols, period)
        return {"success": True, "data": metrics}
    except Exception as e:
        return {"success": False, "error": str(e)}

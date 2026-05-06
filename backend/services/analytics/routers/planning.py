from fastapi import APIRouter, Query
from modules.portfolio_planning import PortfolioPlanning

router = APIRouter()
planning = PortfolioPlanning()


@router.post("/allocate")
def allocate_portfolio(
    symbols: list[str],
    risk_tolerance: str = Query("moderate", enum=["conservative", "moderate", "aggressive"]),
):
    """Allocate portfolio based on risk tolerance."""
    try:
        allocation = planning.allocate(symbols, risk_tolerance)
        return {"success": True, "data": allocation}
    except Exception as e:
        return {"success": False, "error": str(e)}

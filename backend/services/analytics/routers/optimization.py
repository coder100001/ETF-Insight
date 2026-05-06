from fastapi import APIRouter, Query
from modules.portfolio_optimization import optimize_portfolio, fetch_returns

router = APIRouter()


@router.post("/optimize")
def optimize(
    symbols: list[str],
    strategy: str = Query("max_sharpe", enum=["max_sharpe", "min_volatility", "equal_weight"]),
    period: str = Query("1y"),
):
    """Optimize portfolio with given strategy."""
    try:
        result = optimize_portfolio(symbols, strategy, {"period": period})
        return {"success": True, "data": result}
    except Exception as e:
        return {"success": False, "error": str(e)}


@router.post("/returns")
def get_returns(
    symbols: list[str],
    period: str = Query("1y"),
):
    """Fetch historical returns for given symbols."""
    try:
        returns = fetch_returns(symbols, period)
        return {"success": True, "data": returns.to_dict()}
    except Exception as e:
        return {"success": False, "error": str(e)}

import sys, os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from modules.portfolio_optimization import optimize_portfolio, fetch_returns


def test_fetch_returns():
    """Test fetching returns for given symbols."""
    symbols = ["AAPL", "MSFT", "GOOGL"]
    returns = fetch_returns(symbols, period="1mo")
    assert returns is not None
    assert len(returns) > 0
    assert all(s in returns.columns for s in symbols)


def test_optimize_portfolio_max_sharpe():
    """Test portfolio optimization with max Sharpe ratio strategy."""
    symbols = ["AAPL", "MSFT", "GOOGL"]
    result = optimize_portfolio(symbols, strategy="max_sharpe")
    assert result is not None
    assert "weights" in result
    assert "return" in result
    assert "volatility" in result
    assert "sharpe_ratio" in result
    # Weights should sum to approximately 1
    assert abs(sum(result["weights"]) - 1.0) < 0.01


def test_optimize_portfolio_min_volatility():
    """Test portfolio optimization with min volatility strategy."""
    symbols = ["AAPL", "MSFT", "GOOGL"]
    result = optimize_portfolio(symbols, strategy="min_volatility")
    assert result is not None
    assert "weights" in result
    assert "return" in result
    assert "volatility" in result
    # Min volatility should have lower volatility than max sharpe
    max_sharpe_result = optimize_portfolio(symbols, strategy="max_sharpe")
    assert result["volatility"] <= max_sharpe_result["volatility"]

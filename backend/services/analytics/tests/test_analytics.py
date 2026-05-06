import sys, os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest

# Skip if scipy is not installed
scipy = pytest.importorskip("scipy")

from modules.portfolio_analytics import PortfolioAnalytics, CAPMAnalysis


def test_portfolio_analytics_initialization():
    """Test PortfolioAnalytics class can be initialized."""
    pa = PortfolioAnalytics()
    assert pa is not None


def test_capm_analysis():
    """Test CAPM analysis."""
    capm = CAPMAnalysis()
    assert capm is not None

    # Test CAPM calculation
    risk_free_rate = 0.04
    market_return = 0.10
    beta = 1.2

    expected_return = capm.calculate_expected_return(risk_free_rate, market_return, beta)
    assert expected_return is not None
    # CAPM: E(R) = Rf + β(Rm - Rf)
    expected = risk_free_rate + beta * (market_return - risk_free_rate)
    assert abs(expected_return - expected) < 0.001

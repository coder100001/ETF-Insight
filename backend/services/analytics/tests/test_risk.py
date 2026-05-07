import sys, os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest

# Skip if scipy is not installed
scipy = pytest.importorskip("scipy")

from modules.risk_management import RiskManagement, VaRCalculations


def test_risk_management_initialization():
    """Test RiskManagement class can be initialized."""
    rm = RiskManagement()
    assert rm is not None


def test_var_calculations():
    """Test VaR calculations."""
    var_calc = VaRCalculations()
    assert var_calc is not None

    # Test with sample returns
    import numpy as np
    sample_returns = np.random.normal(0.001, 0.02, 100)

    # VaRCalculations.parametric_var is a static method
    result_95 = VaRCalculations.parametric_var(sample_returns, confidence_level=0.95)
    assert result_95 is not None
    assert "var" in result_95
    assert result_95["var"] > 0  # VaR is positive (loss amount)

    result_99 = VaRCalculations.parametric_var(sample_returns, confidence_level=0.99)
    assert result_99 is not None
    assert result_99["var"] > result_95["var"]  # 99% VaR should be larger than 95% VaR

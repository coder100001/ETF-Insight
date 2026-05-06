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

    var_95 = var_calc.calculate_var(sample_returns, confidence=0.95)
    assert var_95 is not None
    assert var_95 < 0  # VaR should be negative (loss)

    var_99 = var_calc.calculate_var(sample_returns, confidence=0.99)
    assert var_99 is not None
    assert var_99 < var_95  # 99% VaR should be more negative than 95% VaR

import sys, os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from sources.worldbank_data import get_countries, get_indicators


def test_get_countries():
    """Test getting list of countries from World Bank."""
    result = get_countries()
    assert result is not None
    # Should return a list or dict
    assert isinstance(result, (list, dict))


def test_get_indicators_china_gdp():
    """Test getting China GDP indicators."""
    result = get_indicators("CHN", "NY.GDP.MKTP.CD", "2020:2023")
    assert result is not None
    # Should contain data
    assert len(str(result)) > 10

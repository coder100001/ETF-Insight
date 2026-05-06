from fastapi import APIRouter, Query
from modules.risk_management import RiskManagement, VaRCalculations

router = APIRouter()
risk_mgmt = RiskManagement()
var_calc = VaRCalculations()


@router.post("/var")
def calculate_var(
    returns: list[float],
    confidence: float = Query(0.95, ge=0.9, le=0.99),
):
    """Calculate Value at Risk."""
    try:
        import numpy as np
        returns_array = np.array(returns)
        var = var_calc.calculate_var(returns_array, confidence)
        return {"success": True, "data": {"var": var, "confidence": confidence}}
    except Exception as e:
        return {"success": False, "error": str(e)}


@router.post("/cvar")
def calculate_cvar(
    returns: list[float],
    confidence: float = Query(0.95, ge=0.9, le=0.99),
):
    """Calculate Conditional Value at Risk."""
    try:
        import numpy as np
        returns_array = np.array(returns)
        cvar = var_calc.calculate_cvar(returns_array, confidence)
        return {"success": True, "data": {"cvar": cvar, "confidence": confidence}}
    except Exception as e:
        return {"success": False, "error": str(e)}

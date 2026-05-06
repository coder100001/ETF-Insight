from fastapi import APIRouter, Query
from sources.akshare_data import AKShareDataWrapper

router = APIRouter()
wrapper = AKShareDataWrapper()


# ---- Stock endpoints ----
@router.get("/stock/spot")
def ak_stock_spot():
    """A股实时行情"""
    return wrapper.get_stock_zh_a_spot()


@router.get("/stock/daily/{symbol}")
def ak_stock_daily(
    symbol: str,
    start_date: str = Query(None),
    end_date: str = Query(None),
):
    """A股日线数据"""
    return wrapper.get_stock_zh_a_daily(symbol, start_date, end_date)


@router.get("/stock/hk/spot")
def ak_stock_hk_spot():
    """港股实时行情"""
    from sources.akshare_stocks_realtime import get_stock_hk_spot
    return get_stock_hk_spot()


@router.get("/stock/us/spot")
def ak_stock_us_spot():
    """美股实时行情"""
    from sources.akshare_stocks_realtime import get_stock_us_spot
    return get_stock_us_spot()


# ---- Macro endpoints ----
@router.get("/macro/gdp")
def ak_macro_gdp():
    """中国GDP数据"""
    return wrapper.get_macro_china_gdp()


@router.get("/macro/{indicator}")
def ak_macro(indicator: str):
    """通用宏观经济指标"""
    func = getattr(wrapper, f"get_macro_china_{indicator}", None)
    if func:
        return func()
    return {"error": f"Unknown indicator: {indicator}"}


# ---- Bond endpoints ----
@router.get("/bonds/{bond_type}")
def ak_bonds(bond_type: str):
    """债券数据"""
    from sources.akshare_bonds import BondsWrapper
    bw = BondsWrapper()
    func = getattr(bw, f"get_{bond_type}", None)
    if func:
        return func()
    return {"error": f"Unknown bond type: {bond_type}"}


# ---- Crypto endpoints ----
@router.get("/crypto/spot")
def ak_crypto_spot():
    """加密货币现货"""
    from sources.akshare_crypto import get_crypto_js_spot
    return get_crypto_js_spot()

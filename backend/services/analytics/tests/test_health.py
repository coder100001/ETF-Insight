import sys, os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest

# Skip if scipy is not installed
scipy = pytest.importorskip("scipy")

from httpx import AsyncClient, ASGITransport
from analytics_server import app


@pytest.mark.asyncio
async def test_health():
    """Test health endpoint returns 200 and expected data."""
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as client:
        resp = await client.get("/health")
        assert resp.status_code == 200
        data = resp.json()
        assert data["status"] == "healthy"
        assert "modules" in data
        assert len(data["modules"]) > 0

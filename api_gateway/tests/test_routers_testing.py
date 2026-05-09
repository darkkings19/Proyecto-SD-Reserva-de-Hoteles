import pytest
from httpx import AsyncClient, ASGITransport
from main import app
from routers.testing import get_notification_client

class MockSuccessClient:
    async def send_confirmation(self, user_id, reservation_id, tipo):
        return True

class MockFailureClient:
    async def send_confirmation(self, user_id, reservation_id, tipo):
        return False

@pytest.mark.asyncio
async def test_notification_endpoint_success():
    app.dependency_overrides[get_notification_client] = lambda: MockSuccessClient()
    
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/test/notifications", json={
            "user_id": "1",
            "reservation_id": "100",
            "tipo": "CONFIRMACION"
        })
        
    assert response.status_code == 200
    assert response.json() == {"success": True}
    
    app.dependency_overrides.clear()

@pytest.mark.asyncio
async def test_notification_endpoint_failure():
    app.dependency_overrides[get_notification_client] = lambda: MockFailureClient()
    
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/api/v1/test/notifications", json={
            "user_id": "1",
            "reservation_id": "100",
            "tipo": "CONFIRMACION"
        })
        
    assert response.status_code == 503
    
    app.dependency_overrides.clear()
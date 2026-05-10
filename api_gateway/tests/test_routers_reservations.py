import pytest
from httpx import AsyncClient, ASGITransport
from main import app
from routers.reservations import get_reservations_client
from grpc_clients.reservations_client import CreateReservationResult, Reservation


class MockSuccessClient:
    async def create_reservation(self, user_id, hotel_id, room_type_id, fecha_inicio, fecha_fin):
        return CreateReservationResult(
            reservation_id="res-20260509120000",
            status="CONFIRMADA",
            monto_total=150.50,
        )

    async def list_reservations(self):
        return [
            Reservation(
                reservation_id="res-1",
                user_id="user-1",
                hotel_id="hotel-playa",
                room_type_id="suite",
                status="CONFIRMADA",
                monto_total=150.50,
            )
        ]


class MockEmptyListClient:
    async def create_reservation(self, user_id, hotel_id, room_type_id, fecha_inicio, fecha_fin):
        raise Exception("not used")

    async def list_reservations(self):
        return []


@pytest.mark.asyncio
async def test_create_reservation_endpoint_success():
    app.dependency_overrides[get_reservations_client] = lambda: MockSuccessClient()

    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.post("/reservations", json={
            "user_id": "user-1",
            "hotel_id": "hotel-playa",
            "room_type_id": "suite",
            "fecha_inicio": "2026-05-09",
            "fecha_fin": "2026-05-13",
        })

    assert response.status_code == 201
    data = response.json()
    assert data["reservation_id"] == "res-20260509120000"
    assert data["status"] == "CONFIRMADA"
    assert data["monto_total"] == 150.50

    app.dependency_overrides.clear()


@pytest.mark.asyncio
async def test_list_reservations_endpoint_success():
    app.dependency_overrides[get_reservations_client] = lambda: MockSuccessClient()

    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/reservations")

    assert response.status_code == 200
    data = response.json()
    assert len(data) == 1
    assert data[0]["reservation_id"] == "res-1"
    assert data[0]["hotel_id"] == "hotel-playa"

    app.dependency_overrides.clear()


@pytest.mark.asyncio
async def test_list_reservations_endpoint_empty():
    app.dependency_overrides[get_reservations_client] = lambda: MockEmptyListClient()

    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/reservations")

    assert response.status_code == 200
    data = response.json()
    assert data == []

    app.dependency_overrides.clear()


@pytest.mark.asyncio
async def test_health_endpoint():
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as ac:
        response = await ac.get("/health")

    assert response.status_code == 200
    assert response.json() == {"status": "ok"}

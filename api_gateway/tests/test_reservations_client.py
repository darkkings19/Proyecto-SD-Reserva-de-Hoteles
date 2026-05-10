import pytest
from unittest.mock import AsyncMock, patch, MagicMock
import grpc
from grpc_clients.reservations_client import ReservationsClient, CreateReservationResult


@pytest.mark.asyncio
async def test_create_reservation_success():
    with patch("grpc_clients.reservations_client.grpc.aio.insecure_channel"):
        mock_stub = AsyncMock()
        mock_response = AsyncMock()
        mock_response.reservation_id = "res-20260509120000"
        mock_response.status = "CONFIRMADA"
        mock_response.monto_total = 150.50
        mock_stub.CreateReservation.return_value = mock_response

        with patch(
            "grpc_clients.reservations_client.servicio_pb2_grpc.ReservationServiceStub",
            return_value=mock_stub,
        ):
            client = ReservationsClient("fake-host:50051")
            result = await client.create_reservation(
                user_id="user-1",
                hotel_id="hotel-playa",
                room_type_id="suite",
                fecha_inicio="2026-05-09",
                fecha_fin="2026-05-13",
            )

            assert isinstance(result, CreateReservationResult)
            assert result.reservation_id == "res-20260509120000"
            assert result.status == "CONFIRMADA"
            assert result.monto_total == 150.50
            mock_stub.CreateReservation.assert_called_once()


@pytest.mark.asyncio
async def test_create_reservation_grpc_error():
    with patch("grpc_clients.reservations_client.grpc.aio.insecure_channel"):
        mock_stub = AsyncMock()
        mock_stub.CreateReservation.side_effect = grpc.aio.AioRpcError(
            code=grpc.StatusCode.NOT_FOUND,
            initial_metadata=None,
            trailing_metadata=None,
            details="usuario no encontrado",
        )

        with patch(
            "grpc_clients.reservations_client.servicio_pb2_grpc.ReservationServiceStub",
            return_value=mock_stub,
        ):
            client = ReservationsClient("fake-host:50051")
            with pytest.raises(grpc.aio.AioRpcError):
                await client.create_reservation(
                    user_id="error_user",
                    hotel_id="hotel-playa",
                    room_type_id="suite",
                    fecha_inicio="2026-05-09",
                    fecha_fin="2026-05-13",
                )


@pytest.mark.asyncio
async def test_list_reservations_success():
    with patch("grpc_clients.reservations_client.grpc.aio.insecure_channel"):
        mock_stub = AsyncMock()
        mock_res1 = MagicMock()
        mock_res1.reservation_id = "res-1"
        mock_res1.user_id = "user-1"
        mock_res1.hotel_id = "hotel-playa"
        mock_res1.room_type_id = "suite"
        mock_res1.status = "CONFIRMADA"
        mock_res1.monto_total = 150.50

        mock_response = AsyncMock()
        mock_response.reservations = [mock_res1]
        mock_stub.ListReservations.return_value = mock_response

        with patch(
            "grpc_clients.reservations_client.servicio_pb2_grpc.ReservationServiceStub",
            return_value=mock_stub,
        ):
            client = ReservationsClient("fake-host:50051")
            reservations = await client.list_reservations()

            assert len(reservations) == 1
            assert reservations[0].reservation_id == "res-1"
            assert reservations[0].hotel_id == "hotel-playa"
            assert reservations[0].status == "CONFIRMADA"
            mock_stub.ListReservations.assert_called_once()


@pytest.mark.asyncio
async def test_list_reservations_empty():
    with patch("grpc_clients.reservations_client.grpc.aio.insecure_channel"):
        mock_stub = AsyncMock()
        mock_response = AsyncMock()
        mock_response.reservations = []
        mock_stub.ListReservations.return_value = mock_response

        with patch(
            "grpc_clients.reservations_client.servicio_pb2_grpc.ReservationServiceStub",
            return_value=mock_stub,
        ):
            client = ReservationsClient("fake-host:50051")
            reservations = await client.list_reservations()

            assert len(reservations) == 0


@pytest.mark.asyncio
async def test_list_reservations_grpc_error():
    with patch("grpc_clients.reservations_client.grpc.aio.insecure_channel"):
        mock_stub = AsyncMock()
        mock_stub.ListReservations.side_effect = grpc.aio.AioRpcError(
            code=grpc.StatusCode.UNAVAILABLE,
            initial_metadata=None,
            trailing_metadata=None,
            details="Service unavailable",
        )

        with patch(
            "grpc_clients.reservations_client.servicio_pb2_grpc.ReservationServiceStub",
            return_value=mock_stub,
        ):
            client = ReservationsClient("fake-host:50051")
            with pytest.raises(grpc.aio.AioRpcError):
                await client.list_reservations()

import grpc.aio
import logging
from typing import Optional
from proto import servicio_pb2
from proto import servicio_pb2_grpc


class Reservation:
    """Data class representing a reservation from the gRPC response."""
    def __init__(self, reservation_id: str, user_id: str, hotel_id: str,
                 room_type_id: str, status: str, monto_total: float):
        self.reservation_id = reservation_id
        self.user_id = user_id
        self.hotel_id = hotel_id
        self.room_type_id = room_type_id
        self.status = status
        self.monto_total = monto_total


class CreateReservationResult:
    """Result of creating a reservation."""
    def __init__(self, reservation_id: str, status: str, monto_total: float):
        self.reservation_id = reservation_id
        self.status = status
        self.monto_total = monto_total


class ReservationsClient:
    def __init__(self, host: str):
        self.host = host

    async def create_reservation(self, user_id: str, hotel_id: str,
                                  room_type_id: str, fecha_inicio: str,
                                  fecha_fin: str) -> Optional[CreateReservationResult]:
        try:
            async with grpc.aio.insecure_channel(self.host) as channel:
                stub = servicio_pb2_grpc.ReservationServiceStub(channel)
                request = servicio_pb2.CreateReservationRequest(
                    user_id=user_id,
                    hotel_id=hotel_id,
                    room_type_id=room_type_id,
                    fecha_inicio=fecha_inicio,
                    fecha_fin=fecha_fin,
                )
                response = await stub.CreateReservation(request, timeout=5.0)
                return CreateReservationResult(
                    reservation_id=response.reservation_id,
                    status=response.status,
                    monto_total=response.monto_total,
                )
        except grpc.aio.AioRpcError as e:
            logging.error(f"gRPC CreateReservation failed: {e.code()} - {e.details()}")
            raise
        except Exception as e:
            logging.error(f"Unexpected error calling Reservation Service: {e}")
            raise

    async def list_reservations(self) -> list[Reservation]:
        try:
            async with grpc.aio.insecure_channel(self.host) as channel:
                stub = servicio_pb2_grpc.ReservationServiceStub(channel)
                request = servicio_pb2.ListReservationsRequest()
                response = await stub.ListReservations(request, timeout=5.0)
                return [
                    Reservation(
                        reservation_id=r.reservation_id,
                        user_id=r.user_id,
                        hotel_id=r.hotel_id,
                        room_type_id=r.room_type_id,
                        status=r.status,
                        monto_total=r.monto_total,
                    )
                    for r in response.reservations
                ]
        except grpc.aio.AioRpcError as e:
            logging.error(f"gRPC ListReservations failed: {e.code()} - {e.details()}")
            raise
        except Exception as e:
            logging.error(f"Unexpected error calling Reservation Service: {e}")
            raise

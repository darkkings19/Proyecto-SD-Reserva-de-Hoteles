import grpc.aio
import os
import sys
from dataclasses import dataclass

sys.path.append(os.path.join(os.path.dirname(__file__), 'proto'))

try:
    import servicio_pb2, servicio_pb2_grpc
except ImportError:
    from proto import servicio_pb2, servicio_pb2_grpc

@dataclass
class CreateReservationResult:
    reservation_id: str
    status: str
    monto_total: float

class ReservationsClient:
    def __init__(self, host: str):
        self.host = host

    async def list_reservations(self):
        async with grpc.aio.insecure_channel(self.host) as channel:
            stub = servicio_pb2_grpc.ReservationServiceStub(channel)
            response = await stub.ListReservations(servicio_pb2.ListReservationsRequest())
            return response.reservations

    async def create_reservation(self, user_id, hotel_id, room_type_id, fecha_inicio, fecha_fin):
        async with grpc.aio.insecure_channel(self.host) as channel:
            stub = servicio_pb2_grpc.ReservationServiceStub(channel)
            request = servicio_pb2.CreateReservationRequest(
                user_id=user_id,
                hotel_id=hotel_id,
                room_type_id=room_type_id,
                fecha_inicio=fecha_inicio,
                fecha_fin=fecha_fin
            )
            response = await stub.CreateReservation(request)
            return CreateReservationResult(
                reservation_id=response.reservation_id,
                status=response.status,
                monto_total=float(response.monto_total)
            )

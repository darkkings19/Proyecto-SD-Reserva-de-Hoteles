from fastapi import APIRouter, HTTPException, Depends
from pydantic import BaseModel
from typing import Annotated
import grpc.aio
import os
import logging
from grpc_clients.reservations_client import ReservationsClient, CreateReservationResult

router = APIRouter()


class CreateReservationRequest(BaseModel):
    user_id: str
    hotel_id: str
    room_type_id: str
    fecha_inicio: str
    fecha_fin: str


def get_reservations_client() -> ReservationsClient:
    host = os.environ.get(
        "RESERVATION_SERVICE_HOST",
        os.environ.get("NOTIFICATION_SERVICE_HOST", "localhost:50051"),
    )
    return ReservationsClient(host)


@router.get("/reservations", status_code=200)
async def list_reservations(
    client: Annotated[ReservationsClient, Depends(get_reservations_client)],
):
    try:
        reservations = await client.list_reservations()
        return [
            {
                "reservation_id": r.reservation_id,
                "user_id": r.user_id,
                "hotel_id": r.hotel_id,
                "room_type_id": r.room_type_id,
                "status": r.status,
                "monto_total": r.monto_total,
            }
            for r in reservations
        ]
    except grpc.aio.AioRpcError as e:
        raise HTTPException(
            status_code=502,
            detail=f"Reservation Service error ({e.code()}): {e.details()}",
        )
    except Exception as e:
        logging.error(f"Unexpected error listing reservations: {e}")
        raise HTTPException(status_code=503, detail="Reservation Service unavailable")


@router.post("/reservations", status_code=201)
async def create_reservation(
    req: CreateReservationRequest,
    client: Annotated[ReservationsClient, Depends(get_reservations_client)],
):
    try:
        result = await client.create_reservation(
            user_id=req.user_id,
            hotel_id=req.hotel_id,
            room_type_id=req.room_type_id,
            fecha_inicio=req.fecha_inicio,
            fecha_fin=req.fecha_fin,
        )
        return {
            "reservation_id": result.reservation_id,
            "status": result.status,
            "monto_total": result.monto_total,
        }
    except grpc.aio.AioRpcError as e:
        if e.code() == grpc.StatusCode.NOT_FOUND:
            raise HTTPException(status_code=404, detail=e.details())
        if e.code() == grpc.StatusCode.RESOURCE_EXHAUSTED:
            raise HTTPException(status_code=409, detail=e.details())
        raise HTTPException(
            status_code=502,
            detail=f"Reservation Service error ({e.code()}): {e.details()}",
        )
    except Exception as e:
        logging.error(f"Unexpected error creating reservation: {e}")
        raise HTTPException(status_code=503, detail="Reservation Service unavailable")

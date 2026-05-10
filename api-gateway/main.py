from fastapi import FastAPI, HTTPException, Depends
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import Annotated
import os
import logging

# Clientes locales
from inventory_client import search_available_rooms
from reservations_client import ReservationsClient

app = FastAPI(title="Origen X - API Gateway Unificado (Slim)", version="1.1.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# --- Modelos de Datos ---
class SearchRoomsRequest(BaseModel):
    fecha_inicio: str
    fecha_fin: str
    ubicacion: str = ""
    precio_max: float = 0
    capacidad: int = 0

class CreateReservationRequest(BaseModel):
    user_id: str
    hotel_id: str
    room_type_id: str
    fecha_inicio: str
    fecha_fin: str

# --- Dependencias ---
def get_reservations_client() -> ReservationsClient:
    host = os.environ.get("RESERVATION_SERVICE_HOST", "localhost:50051")
    return ReservationsClient(host)

# --- Endpoints de Inventario ---
@app.post("/api/inventory/search")
async def search_rooms(req: SearchRoomsRequest):
    try:
        rooms = search_available_rooms(
            fecha_inicio=req.fecha_inicio,
            fecha_fin=req.fecha_fin,
            ubicacion=req.ubicacion,
            precio_max=req.precio_max,
            capacidad=req.capacidad
        )
        return {"rooms": rooms}
    except Exception as e:
        logging.error(f"Error en búsqueda: {e}")
        raise HTTPException(status_code=500, detail=str(e))

# --- Endpoints de Reservas ---
@app.get("/reservations")
async def list_reservations(client: Annotated[ReservationsClient, Depends(get_reservations_client)]):
    try:
        reservations = await client.list_reservations()
        return [
            {
                "reservation_id": r.reservation_id,
                "user_id": r.user_id,
                "hotel_id": r.hotel_id,
                "room_type_id": r.room_type_id,
                "status": r.status,
                "monto_total": float(r.monto_total),
            }
            for r in reservations
        ]
    except Exception as e:
        raise HTTPException(status_code=502, detail=str(e))

@app.post("/reservations", status_code=201)
async def create_reservation(req: CreateReservationRequest, client: Annotated[ReservationsClient, Depends(get_reservations_client)]):
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
    except Exception as e:
        raise HTTPException(status_code=409, detail=str(e))

@app.get("/health")
async def health_check():
    return {"status": "ok"}

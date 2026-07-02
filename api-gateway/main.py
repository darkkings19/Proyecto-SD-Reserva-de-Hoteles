from fastapi import FastAPI, HTTPException, Depends, Header
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import Annotated
import os
import logging

# Clientes locales
from inventory_client import search_available_rooms
from reservations_client import ReservationsClient
from users_client import UsersClient

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
    user_id: str = ""
    hotel_id: str
    room_type_id: str
    fecha_inicio: str
    fecha_fin: str

class UserRegistrationRequest(BaseModel):
    nombre: str
    email: str
    password: str
    telefono: str = ""

class UserUpdateRequest(BaseModel):
    nombre: str
    telefono: str = ""

class LoginRequest(BaseModel):
    email: str
    password: str

class TokenRequest(BaseModel):
    access_token: str

# --- Dependencias ---
def get_reservations_client() -> ReservationsClient:
    host = os.environ.get("RESERVATION_SERVICE_HOST", "localhost:50052")
    return ReservationsClient(host)

def get_users_client() -> UsersClient:
    host = os.environ.get("USER_SERVICE_HOST", "user-service:9090")
    return UsersClient(host)

def _extract_bearer_token(authorization: str | None) -> str:
    if not authorization:
        raise HTTPException(status_code=401, detail="Sesion requerida")

    scheme, _, token = authorization.partition(" ")
    if scheme.lower() != "bearer" or not token:
        raise HTTPException(status_code=401, detail="Token invalido")

    return token

def require_session_user(users_client: UsersClient, authorization: str | None):
    token = _extract_bearer_token(authorization)
    res = users_client.validate_token(access_token=token)
    if not res.success or not res.user.id:
        raise HTTPException(status_code=401, detail="Sesion invalida")

    return res.user

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
async def list_reservations(
    client: Annotated[ReservationsClient, Depends(get_reservations_client)],
    users_client: Annotated[UsersClient, Depends(get_users_client)],
    authorization: str | None = Header(default=None)
):
    try:
        session_user = require_session_user(users_client, authorization)
        reservations = await client.list_reservations(user_id=session_user.id)
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
        if isinstance(e, HTTPException):
            raise e
        raise HTTPException(status_code=502, detail=str(e))

@app.post("/reservations", status_code=201)
async def create_reservation(
    req: CreateReservationRequest,
    client: Annotated[ReservationsClient, Depends(get_reservations_client)],
    users_client: Annotated[UsersClient, Depends(get_users_client)],
    authorization: str | None = Header(default=None)
):
    try:
        session_user = require_session_user(users_client, authorization)
        result = await client.create_reservation(
            user_id=session_user.id,
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
        if isinstance(e, HTTPException):
            raise e
        raise HTTPException(status_code=409, detail=str(e))

@app.post("/users", status_code=201)
async def register_user(req: UserRegistrationRequest, client: Annotated[UsersClient, Depends(get_users_client)]):
    try:
        user = client.create_user(
            nombre=req.nombre,
            email=req.email,
            password=req.password,
            telefono=req.telefono
        )
        return {
            "id": user.id,
            "nombre": user.nombre,
            "email": user.email
        }
    except Exception as e:
        logging.error(f"Error registrando usuario: {e}")
        raise HTTPException(status_code=400, detail=str(e))

@app.get("/users/{user_id}")
async def get_user(
    user_id: str,
    client: Annotated[UsersClient, Depends(get_users_client)],
    authorization: str | None = Header(default=None)
):
    try:
        session_user = require_session_user(client, authorization)
        if session_user.id != user_id:
            raise HTTPException(status_code=403, detail="Solo puedes ver tu propio perfil")

        user = client.get_user(user_id)
        return {
            "id": user.id,
            "nombre": user.nombre,
            "email": user.email,
            "telefono": user.telefono,
            "created_at": user.created_at
        }
    except Exception as e:
        if isinstance(e, HTTPException):
            raise e
        logging.error(f"Error obteniendo usuario: {e}")
        raise HTTPException(status_code=404, detail=str(e))

@app.put("/users/{user_id}")
async def update_user(
    user_id: str,
    req: UserUpdateRequest,
    client: Annotated[UsersClient, Depends(get_users_client)],
    authorization: str | None = Header(default=None)
):
    try:
        session_user = require_session_user(client, authorization)
        if session_user.id != user_id:
            raise HTTPException(status_code=403, detail="Solo puedes editar tu propio perfil")

        user = client.update_user(
            user_id=user_id,
            nombre=req.nombre,
            telefono=req.telefono
        )
        return {
            "id": user.id,
            "nombre": user.nombre,
            "email": user.email,
            "telefono": user.telefono,
            "created_at": user.created_at
        }
    except Exception as e:
        if isinstance(e, HTTPException):
            raise e
        logging.error(f"Error actualizando usuario: {e}")
        raise HTTPException(status_code=400, detail=str(e))

@app.post("/login")
async def login(req: LoginRequest, client: Annotated[UsersClient, Depends(get_users_client)]):
    try:
        res = client.authenticate(email=req.email, password=req.password)
        if not res.success:
            raise HTTPException(status_code=401, detail="Credenciales inválidas")
        
        return {
            "id": res.user.id,
            "nombre": res.user.nombre,
            "email": res.user.email,
            "access_token": res.access_token,
            "expires_at": res.expires_at
        }
    except Exception as e:
        logging.error(f"Error en login: {e}")
        raise HTTPException(status_code=401, detail=str(e))

@app.post("/logout")
async def logout(req: TokenRequest, client: Annotated[UsersClient, Depends(get_users_client)]):
    try:
        res = client.logout(access_token=req.access_token)
        return {
            "success": res.success,
            "id": res.user.id,
            "email": res.user.email,
            "expires_at": res.expires_at
        }
    except Exception as e:
        logging.error(f"Error en logout: {e}")
        raise HTTPException(status_code=401, detail=str(e))

@app.post("/validate-token")
async def validate_token(req: TokenRequest, client: Annotated[UsersClient, Depends(get_users_client)]):
    try:
        res = client.validate_token(access_token=req.access_token)
        return {
            "success": res.success,
            "id": res.user.id,
            "nombre": res.user.nombre,
            "email": res.user.email,
            "expires_at": res.expires_at
        }
    except Exception as e:
        logging.error(f"Error validando token: {e}")
        raise HTTPException(status_code=401, detail=str(e))

@app.get("/health")
async def health_check():
    return {"status": "ok"}

from fastapi import APIRouter, HTTPException, Depends
from pydantic import BaseModel
from typing import Annotated
import os
from grpc_clients.notifications_client import NotificationClient

router = APIRouter()

class NotificationRequest(BaseModel):
    user_id: str
    reservation_id: str
    tipo: str

def get_notification_client():
    host = os.environ.get("NOTIFICATION_SERVICE_HOST", "localhost:50051")
    return NotificationClient(host)

# =====================================================================
# TEMPORAL - BORRAR POSTERIORMENTE EN EL MERGE
# Endpoint de prueba para validar que el API Gateway se comunica
# correctamente con el Servicio de Notificaciones via gRPC.
# =====================================================================
@router.post(
    "/api/v1/test/notifications", 
    status_code=200,
    responses={503: {"description": "Notification Service unavailable or failed"}}
)
async def test_notification(
    request: NotificationRequest,
    client: Annotated[NotificationClient, Depends(get_notification_client)]
):
    success = await client.send_confirmation(
        user_id=request.user_id,
        reservation_id=request.reservation_id,
        tipo=request.tipo
    )
    
    if not success:
        raise HTTPException(status_code=503, detail="Notification Service unavailable or failed")
        
    return {"success": True}
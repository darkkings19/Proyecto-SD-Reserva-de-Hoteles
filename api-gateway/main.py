from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import grpc
import servicio_pb2
import servicio_pb2_grpc
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI(title="API Gateway de Reservas")

# Habilitar CORS para que el frontend pueda comunicarse
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

class ReservationReq(BaseModel):
    user_id: str
    hotel_id: str
    room_type_id: str
    fecha_inicio: str
    fecha_fin: str

@app.post("/reservations", status_code=201)
def create_reservation(req: ReservationReq):
    try:
        # Conexión gRPC hacia el servicio de Go (puerto 50051)
        channel = grpc.insecure_channel('localhost:50051')
        stub = servicio_pb2_grpc.ReservationServiceStub(channel)
        
        # Llamar al método gRPC
        grpc_req = servicio_pb2.CreateReservationRequest(
            user_id=req.user_id,
            hotel_id=req.hotel_id,
            room_type_id=req.room_type_id,
            fecha_inicio=req.fecha_inicio,
            fecha_fin=req.fecha_fin
        )
        response = stub.CreateReservation(grpc_req)
        
        return {
            "reservation_id": response.reservation_id,
            "status": response.status,
            "monto_total": response.monto_total
        }
    except grpc.RpcError as e:
        raise HTTPException(status_code=400, detail=f"Error gRPC ({e.code()}): {e.details()}")

@app.get("/reservations")
def list_reservations():
    """Obtiene las reservas llamando directamente a Go por gRPC, delegando la base de datos a Go."""
    try:
        channel = grpc.insecure_channel('localhost:50051')
        stub = servicio_pb2_grpc.ReservationServiceStub(channel)
        
        # Llamar a Go para que él le pregunte a PostgreSQL
        response = stub.ListReservations(servicio_pb2.ListReservationsRequest())
        
        reservations = []
        for r in response.reservations:
            reservations.append({
                "reservation_id": r.reservation_id,
                "user_id": r.user_id,
                "hotel_id": r.hotel_id,
                "room_type_id": r.room_type_id,
                "status": r.status,
                "monto_total": r.monto_total
            })
        return reservations
    except grpc.RpcError as e:
        raise HTTPException(status_code=500, detail=f"Error de conexión con Go: {e.details()}")

if __name__ == "__main__":
    import uvicorn
    # Corre el servidor en localhost:8080
    uvicorn.run(app, host="127.0.0.1", port=8080)

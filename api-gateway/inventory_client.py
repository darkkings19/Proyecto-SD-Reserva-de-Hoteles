import grpc
import os
import sys

# Añadir la carpeta de protos al path para que los imports internos de gRPC funcionen
sys.path.append(os.path.join(os.path.dirname(__file__), 'proto'))

try:
    import inventory_pb2, inventory_pb2_grpc
except ImportError:
    from proto import inventory_pb2, inventory_pb2_grpc

def search_available_rooms(fecha_inicio, fecha_fin, ubicacion="", precio_max=0, capacidad=0):
    host = os.environ.get("INVENTORY_SERVICE_HOST", "localhost:50051")
    channel = grpc.insecure_channel(host)
    stub = inventory_pb2_grpc.InventoryServiceStub(channel)
    
    request = inventory_pb2.SearchRequest(
        ubicacion=ubicacion,
        precio_max=precio_max,
        capacidad=capacidad
    )
    
    response = stub.SearchAvailableRooms(request)
    
    results = []
    for r in response.rooms:
        results.append({
            "hotel_id": r.hotel_id,
            "nombre_hotel": r.nombre_hotel,
            "room_type_id": r.room_type_id,
            "precio_noche": float(r.precio_noche),
            "capacidad": int(r.capacidad),
            "stock_disponible": int(r.stock_disponible),
            "room_type_name": r.room_type_name
        })
    return results

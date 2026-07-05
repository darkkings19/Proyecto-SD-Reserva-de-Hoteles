import grpc
import os
import sys
from opentelemetry.propagate import inject

# Añadir la carpeta de protos al path
sys.path.append(os.path.join(os.path.dirname(__file__), 'proto'))

try:
    import servicio_pb2, servicio_pb2_grpc
except ImportError:
    from proto import servicio_pb2, servicio_pb2_grpc

def trace_metadata():
    carrier = {}
    inject(carrier)
    return tuple(carrier.items())

class UsersClient:
    def __init__(self, host: str):
        self.host = host

    def create_user(self, nombre, email, password, rol=1, telefono=""):
        channel = grpc.insecure_channel(self.host)
        stub = servicio_pb2_grpc.UserServiceStub(channel)
        
        request = servicio_pb2.CreateUserRequest(
            nombre=nombre,
            email=email,
            password=password,
            rol=rol,
            telefono=telefono
        )
        
        response = stub.CreateUser(request, metadata=trace_metadata())
        return response.user

    def authenticate(self, email, password):
        channel = grpc.insecure_channel(self.host)
        stub = servicio_pb2_grpc.UserServiceStub(channel)
        
        request = servicio_pb2.AuthenticateUserRequest(
            email=email,
            password=password
        )
        
        response = stub.AuthenticateUser(request, metadata=trace_metadata())
        return response

    def logout(self, access_token):
        channel = grpc.insecure_channel(self.host)
        stub = servicio_pb2_grpc.UserServiceStub(channel)

        request = servicio_pb2.LogoutRequest(access_token=access_token)
        return stub.LogoutUser(request, metadata=trace_metadata())

    def validate_token(self, access_token):
        channel = grpc.insecure_channel(self.host)
        stub = servicio_pb2_grpc.UserServiceStub(channel)

        request = servicio_pb2.ValidateTokenRequest(access_token=access_token)
        return stub.ValidateToken(request, metadata=trace_metadata())

FROM python:3.11-slim

WORKDIR /app

# Instalar dependencias del sistema para compilar protos
RUN apt-get update && apt-get install -y \
    protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PYTHONPATH=/app/src

# Instalar dependencias de Python
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copiar el resto del código y la carpeta proto
COPY . .

# Regenerar los archivos gRPC automáticamente
RUN python -m grpc_tools.protoc -I. --python_out=src --grpc_python_out=src proto/notifications.proto && \
    python -m grpc_tools.protoc -I. --python_out=src --grpc_python_out=src proto/user_service.proto

EXPOSE 50051

CMD ["python", "src/main.py"]

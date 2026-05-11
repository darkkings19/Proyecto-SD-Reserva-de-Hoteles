FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# Compilar protos en la carpeta proto
RUN python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. proto/inventory.proto
RUN python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. proto/servicio.proto

ENV PYTHONPATH=/app
ENV PORT=8080

EXPOSE 8080

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8080"]

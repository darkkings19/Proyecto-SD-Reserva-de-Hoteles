FROM golang:1.25-alpine AS builder

# Instalar protoc y plugins
RUN apk add --no-cache protobuf
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Limpiar archivos generados viejos para evitar conflictos de metadatos
RUN rm -rf proto/gen && mkdir -p proto/gen

# Generar archivos proto. 
# Usamos --go_out=. sin source_relative para que respete el path definido en 'option go_package'
# o simplemente forzamos la salida al directorio gen.
RUN protoc --go_out=. --go-grpc_out=. proto/inventory.proto

# Mover los archivos si protoc los creó en una ruta basada en el nombre del módulo
# (Depende de cómo esté configurado el entorno, pero esto asegura que terminen en proto/gen)
RUN if [ -d github.com ]; then \
        cp -r github.com/darkkings19/inventario-service/proto/gen/* proto/gen/ && \
        rm -rf github.com; \
    fi

RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /app/inventory-service .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/inventory-service .

EXPOSE 50053

CMD ["/app/inventory-service"]

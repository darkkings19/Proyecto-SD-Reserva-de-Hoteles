FROM golang:1.25-alpine AS builder

# Instalar protoc y plugins
RUN apk add --no-cache protobuf
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Generar archivos proto
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pb/servicio.proto

RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /app/server .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 50052

CMD ["./server"]

.PHONY: proto build run test clean

# Regenerar stubs de protobuf
proto:
	protoc --go_out=proto/gen --go_opt=paths=source_relative \
		--go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
		-I=proto proto/inventory.proto

# Compilar el binario
build:
	go build -o bin/server ./cmd/server

# Ejecutar en modo local (requiere .env y PostgreSQL corriendo)
run: build
	./bin/server

# Ejecutar tests
test:
	go test ./... -v -count=1

# Limpiar artefactos de compilación
clean:
	rm -rf bin/

FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy only necessary files to avoid scanning old folders
COPY go.mod go.sum main.go ./
COPY proto/ ./proto/

# Ensure dependencies are tidy and downloaded
RUN go mod tidy && go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/inventory-service .

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/inventory-service .

CMD ["/app/inventory-service"]

# Build stage
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy migrations from builder
COPY --from=builder /app/migrations ./migrations

# Copy built application from builder
COPY --from=builder /app/api .

# Expose port
EXPOSE 8080

# Run application
CMD ["./api"]

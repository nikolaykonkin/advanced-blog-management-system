# Build stage
FROM golang:1.26-alpine AS builder

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
FROM alpine:3.20

RUN adduser -D -u 10001 appuser

WORKDIR /app

COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/api .

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./api"]
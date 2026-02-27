# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o main cmd/main.go

# Run stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

# Expose port
EXPOSE 3002

# Run the application
CMD ["./main"]
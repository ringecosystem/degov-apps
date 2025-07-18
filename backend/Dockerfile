FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o bin/server .

# Final stage
FROM alpine:latest

# Install ca-certificates for TLS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/bin/server .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 1096

# Run the binary
CMD ["./server"]

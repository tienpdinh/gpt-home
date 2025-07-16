# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for go mod download
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gpt-home ./cmd/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/gpt-home .

# Copy web templates and static files if they exist
COPY --from=builder /app/web ./web

# Create directory for models and data
RUN mkdir -p /root/models /root/data

# Expose port
EXPOSE 8080

# Set environment variables
ENV SERVER_PORT=8080
ENV SERVER_HOST=0.0.0.0
ENV LOG_LEVEL=info

# Run the application
CMD ["./gpt-home"]
# Dockerfile
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /builder

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o push-service ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates
RUN apk update && apk add --no-cache ca-certificates tzdata && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy the built binary from builder stage
COPY --from=builder /builder/push-service /cmd/server

# EXPOSE port
EXPOSE 8081

# Run the application
ENTRYPOINT ["/cmd/server"]
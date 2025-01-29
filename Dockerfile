# Stage 1: Build the application
FROM golang:1.22-alpine AS builder

# Install required dependencies for SQLite and build
RUN apk add --no-cache gcc musl-dev git

WORKDIR /app

# Copy module files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy all source files
COPY . .

# Build the application with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o forum ./cmd/web/

# Stage 2: Create production image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache libc6-compat

WORKDIR /app

# Copy binaries and resources
COPY --from=builder /app/forum .
COPY --from=builder /app/ui ./ui
COPY --from=builder /app/tls ./tls
COPY --from=builder /app/data ./data

# Create necessary directories
RUN mkdir -p /data && \
    mkdir -p /app/ui/static/upload

# Expose port and set entrypoint
EXPOSE 4000
CMD ["./forum"]
# Multi-stage Dockerfile for Gorax
# Stage 1: Build frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web

# Copy package files
COPY web/package*.json ./

# Install dependencies
RUN npm ci

# Copy frontend source
COPY web/ ./

# Build frontend
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.23-alpine AS backend-builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /bin/gorax ./cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /bin/worker ./cmd/worker/main.go

# Stage 3: Final minimal image
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binaries from builder
COPY --from=backend-builder /bin/gorax /app/gorax
COPY --from=backend-builder /bin/worker /app/worker

# Copy frontend build
COPY --from=frontend-builder /app/web/dist /app/web/dist

# Copy migrations (if they exist)
COPY migrations/ /app/migrations/ 2>/dev/null || true

# Create non-root user
RUN addgroup -g 1000 gorax && \
    adduser -D -u 1000 -G gorax gorax && \
    chown -R gorax:gorax /app

USER gorax

# Expose ports
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command
CMD ["/app/gorax"]

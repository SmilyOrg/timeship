# Build stage for UI
FROM node:22-alpine AS ui-builder

WORKDIR /build/ui

# Copy UI package files
COPY ui/package*.json ./

# Install dependencies
RUN npm install

# Copy UI source
COPY ui/ ./

# Build UI
RUN npm run build

# Build stage for Go API
FROM golang:1.25-alpine AS api-builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY api/go.mod api/go.sum ./api/

# Download dependencies
WORKDIR /build/api
RUN go mod download

# Copy API source
WORKDIR /build
COPY api/ ./api/

# Copy built UI from previous stage
COPY --from=ui-builder /build/ui/dist ./api/ui/dist

# Build the API with embedded UI
WORKDIR /build/api
RUN CGO_ENABLED=0 GOOS=linux go build -tags embedui -ldflags="-w -s" -o timeship .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 timeship && \
    adduser -D -u 1000 -G timeship timeship

WORKDIR /app

# Copy binary from builder
COPY --from=api-builder /build/api/timeship .

# Use non-root user
USER timeship

# Expose port
EXPOSE 8080

# Run the application
CMD ["./timeship"]

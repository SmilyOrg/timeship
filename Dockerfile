# Build stage
FROM golang:1.25-alpine AS builder

ARG LDFLAGS=""

WORKDIR /build

# Go dependency caching
COPY api/go.mod api/go.sum ./
RUN go mod download

# Full API source & UI copy
COPY api/ ./
RUN \
  set -eou pipefail && \
  CGO_ENABLED=0 \
  go build \
    -ldflags "${LDFLAGS}" \
    -tags embedui \
    -o timeship .



# Runtime stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /build/timeship ./timeship

EXPOSE 8080
ENV TIMESHIP_ROOT=/mnt
ENTRYPOINT ["./timeship"]

# ==============================================================================
# STAGE 1: Build the optimized Go binary
# ==============================================================================
FROM golang:1.25-alpine AS builder

# Install build dependencies required for confluent-kafka-go (librdkafka requires CGO)
RUN apk add --no-cache \
    alpine-sdk \
    musl-dev \
    gcc \
    make \
    librdkafka-dev

WORKDIR /app

# Leverage Docker cache layers by copying dependency specifications first
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source tree
COPY . .

# Build the binary with CGO enabled (required for librdkafka), stripping debug assertions
RUN CGO_ENABLED=1 GOOS=linux go build \
    -tags musl \
    -ldflags="-w -s" \
    -o kafka-producer \
    ./cmd/producer/main.go

# ==============================================================================
# STAGE 2: Ultra-lightweight secure runtime environment
# ==============================================================================
FROM alpine:3.19

# Install runtime dependencies for shared libraries and SSL root certificates
RUN apk --no-cache add ca-certificates librdkafka

WORKDIR /root/

# Copy the pre-compiled binary from the builder stage
COPY --from=builder /app/kafka-producer .

# CRITICAL: Copy configuration and data directories for runtime
COPY --from=builder /app/profiles ./profiles
COPY --from=builder /app/data ./data

# Execute the binary
ENTRYPOINT ["./kafka-producer"]
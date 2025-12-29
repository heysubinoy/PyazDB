# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binaries
RUN go build -o /bin/kv-single ./cmd/kv-single
RUN go build -o /bin/kv-cli ./cmd/kv-cli
RUN go build -o /bin/mandi ./cmd/mandi
RUN go build -o /bin/sarpanch ./cmd/sarpanch

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binaries from builder
COPY --from=builder /bin/kv-single /usr/local/bin/
COPY --from=builder /bin/kv-cli /usr/local/bin/
COPY --from=builder /bin/mandi /usr/local/bin/
COPY --from=builder /bin/sarpanch /usr/local/bin/

# Create data directory
RUN mkdir -p /data

# Expose ports
# 7000: Mandi discovery service
# 8080-8082: HTTP API ports for nodes
# 9090-9092: gRPC ports for nodes
# 12000-12002: Raft communication ports
EXPOSE 7000 8080 8081 8082 9090 9091 9092 12000 12001 12002

# Default command runs kv-single
# Override with docker run command to run mandi or kv-cli
CMD ["kv-single"]

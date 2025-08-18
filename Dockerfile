# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    build-base

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download and verify dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build arguments for version information
ARG VERSION=dev
ARG BUILD_TIME
ARG COMMIT_HASH
ARG TARGETARCH

# Build the binary (no cross-compiler needed for native builds)
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -trimpath \
    -a \
    -ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.commitHash=${COMMIT_HASH} -s -w" \
    -o memory-calculator \
    ./cmd/memory-calculator

# Verify the binary is statically linked (important for Alpine)
RUN file memory-calculator

# Test the binary works in Alpine environment
RUN ./memory-calculator --version

# Alpine target for minimal runtime
FROM eclipse-temurin:21-jre-alpine AS alpine

# Bring build args to final stage
ARG VERSION=dev
ARG BUILD_TIME
ARG COMMIT_HASH

# Install minimal runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1001 memcalc && \
    adduser -D -u 1001 -G memcalc memcalc

# Set working directory
WORKDIR /home/memcalc

# Copy binary from builder stage
COPY --from=builder /app/memory-calculator /usr/local/bin/memory-calculator

# Ensure binary is executable
RUN chmod +x /usr/local/bin/memory-calculator

# Test the binary in the Alpine environment
RUN memory-calculator --version

# Switch to non-root user
USER memcalc

# Default command
ENTRYPOINT ["memory-calculator"]
CMD ["--help"]

# Metadata
LABEL maintainer="Patrick Baumgartner <contact@patbaumgartner.com>" \
      description="JVM Memory Calculator for Container Environments" \
      version="${VERSION}" \
      build.time="${BUILD_TIME}" \
      build.commit="${COMMIT_HASH}" \
      alpine.version="3.20" \
      go.version="1.24.5"

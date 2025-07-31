# Build stage
FROM golang:1.24.5-alpine AS builder

# Install git (needed for version info)
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG VERSION=dev
ARG BUILD_TIME
ARG COMMIT_HASH
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.commitHash=${COMMIT_HASH} -w -s" \
    -a -installsuffix cgo \
    -o memory-calculator ./cmd/memory-calculator

# Final stage
FROM alpine:latest

# Bring build args to final stage
ARG VERSION=dev
ARG BUILD_TIME
ARG COMMIT_HASH

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 memcalc && \
    adduser -D -u 1001 -G memcalc memcalc

# Set working directory
WORKDIR /home/memcalc

# Copy binary from builder stage
COPY --from=builder /app/memory-calculator /usr/local/bin/memory-calculator

# Make binary executable
RUN chmod +x /usr/local/bin/memory-calculator

# Switch to non-root user
USER memcalc

# Default command
ENTRYPOINT ["memory-calculator"]
CMD ["--help"]

# Metadata
LABEL maintainer="Patrick Baumgartner <contact@patbaumgartner.com>"
LABEL description="JVM Memory Calculator for Container Environments"
LABEL version="${VERSION}"
LABEL build.time="${BUILD_TIME}"
LABEL build.commit="${COMMIT_HASH}"

# examples/Dockerfile
# Production-ready Dockerfile using memory calculator

# Multi-stage build
FROM golang:1.21-alpine AS calculator-builder
WORKDIR /build
COPY . .
RUN apk add --no-cache make && make build-minimal

# Runtime image
FROM bellsoft/liberica-runtime-container:jdk-21-slim-musl
WORKDIR /app

# Install memory calculator
COPY --from=calculator-builder /build/memory-calculator /usr/local/bin/
RUN chmod +x /usr/local/bin/memory-calculator

# Copy entrypoint script
COPY examples/docker-entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Copy your application (replace with actual app)
# COPY myapp.jar /app/

# Add non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD /entrypoint.sh health

# Set entrypoint
ENTRYPOINT ["/entrypoint.sh"]

# Default command (replace with your app)
CMD ["java", "-jar", "myapp.jar"]

# Example build commands:
# docker build -t myapp:latest .
# docker run --memory=2g myapp:latest
# docker run myapp:latest java-opts  # Show calculated JVM options

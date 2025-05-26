# Custodian Killer - AWS Policy Management Tool
# Multi-stage build for optimal image size

# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (needed for go mod download with private repos)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o custodian-killer .

# Final stage
FROM alpine:3.18

# Install ca-certificates for SSL/TLS and other common tools
RUN apk --no-cache add ca-certificates tzdata curl

# Create a non-root user
RUN addgroup -g 1001 -S custodian && \
    adduser -u 1001 -S custodian -G custodian

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/custodian-killer .

# Create data directory for policies
RUN mkdir -p /home/custodian/.custodian-killer && \
    chown -R custodian:custodian /home/custodian

# Switch to non-root user
USER custodian

# Set home directory
ENV HOME=/home/custodian

# Expose port (if needed for future web interface)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./custodian-killer --help || exit 1

# Set entrypoint
ENTRYPOINT ["./custodian-killer"]

# Default command (interactive mode)
CMD ["interactive"]

# Labels for metadata
LABEL maintainer="Custodian Killer Team"
LABEL description="AWS Policy Management Tool - Making AWS compliance fun again!"
LABEL version="1.0.0"

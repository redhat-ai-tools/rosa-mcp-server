# Build stage using UBI9 Go toolset
FROM registry.access.redhat.com/ubi9/go-toolset:latest AS builder

# Switch to root to install dependencies
USER 0

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o rosa-mcp-server ./cmd/rosa-mcp-server

# Runtime stage using minimal UBI9
FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

# Install ca-certificates for HTTPS requests to OCM
RUN microdnf install -y ca-certificates && microdnf clean all

# Create non-root user
RUN groupadd -r rosa && useradd -r -g rosa rosa

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/rosa-mcp-server .

# Change ownership to non-root user
RUN chown rosa:rosa /app/rosa-mcp-server

# Switch to non-root user
USER rosa

# Expose port 8080 for SSE transport
EXPOSE 8080

# Default command runs with SSE transport on port 8080
CMD ["./rosa-mcp-server", "--transport=sse", "--port=8080"]
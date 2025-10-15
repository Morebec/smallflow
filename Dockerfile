# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -trimpath -o /bin/smallflow ./cmd/smallflow

# Final stage
FROM alpine:3.20

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 smallflow && \
    adduser -D -u 1000 -G smallflow smallflow

# Set working directory
WORKDIR /home/smallflow

# Copy binary from builder
COPY --from=builder /bin/smallflow /usr/local/bin/smallflow

# Change ownership
RUN chown -R smallflow:smallflow /home/smallflow

# Switch to non-root user
USER smallflow

# Run the application
ENTRYPOINT ["smallflow"]

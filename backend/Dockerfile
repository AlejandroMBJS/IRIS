# Use the official Golang image as a base image
FROM docker.io/library/golang:1.24-alpine AS builder

# Install required dependencies for CGO (sqlite) in separate steps to avoid package conflicts
RUN apk update && apk cache clean
RUN apk add --no-cache gcc musl-dev
RUN apk add --no-cache sqlite-dev
RUN apk add --no-cache make

# Set the working directory inside the container
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy ALL source files into the container
COPY . .

# Build the Go application with CGO enabled for sqlite
# Use tags to ensure proper sqlite build
RUN CGO_ENABLED=1 CGO_CFLAGS="-D_LARGEFILE64_SOURCE" go build -o /iris-payroll-backend cmd/api/main.go

# Start a new, smaller stage for the final image
FROM docker.io/library/alpine:latest

# Install runtime dependencies
RUN apk add --no-cache sqlite-libs wget ca-certificates tzdata

# Set the working directory
WORKDIR /app

# Create a non-root user and group
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Create directories for data storage with proper permissions
RUN mkdir -p /data/uploads /data/pdfs && \
    chown -R appuser:appgroup /data && \
    chmod -R 755 /data

# Copy the built binary from the builder stage
COPY --from=builder /iris-payroll-backend .

# Copy config files
COPY --from=builder /app/configs ./configs

# Change ownership of the app directory
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose the port the app runs on
EXPOSE 8080

# Set the command to run the application
CMD ["./iris-payroll-backend"]

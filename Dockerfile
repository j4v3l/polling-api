
# Stage 1: Build the Go application
FROM golang:1.20-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o polling-api ./cmd/server

# Stage 2: Run the Go application in a lightweight container
FROM scratch

# Set the working directory inside the container
WORKDIR /app

# Copy the compiled Go binary from the build stage
COPY --from=builder /app/polling-api /app/polling-api

# Copy the SQLite database file if needed (this assumes it's pre-seeded or will be created at runtime)
# COPY --from=builder /app/polls.db /app/polls.db

# Expose the application port
EXPOSE 8080

# Run the Go binary
ENTRYPOINT ["/app/polling-api"]

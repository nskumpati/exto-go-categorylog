# Start with the official Golang image to build the application
# Using a specific version for reproducibility
FROM golang:1.24.0-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies first
# This leverages Docker's layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application's source code
COPY . .

# Build the application. 
# -o /app/server will place the executable in the /app directory
# CGO_ENABLED=0 is used to build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server

# ---
# Start a new, smaller stage for the final image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/server .

# Copy the client_secret file needed for Google authentication
# Note: It's generally better to handle secrets securely, e.g., via environment variables or a secret management system.
# COPY client_secret_*.json ./

# Expose the port the app runs on.
# You might need to change this if your app runs on a different port.
EXPOSE 7070

# Command to run the executable
CMD ["./server"]

# Use the official Go image as the base image
FROM golang:1.19 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN go build -o tbd-bot ./cmd

# Use a minimal base image for the final container
FROM debian:bullseye-slim

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/tbd-bot .

# Expose the port your application listens on (if applicable)
EXPOSE 8080

# Command to run the application
CMD ["./tbd-bot"]
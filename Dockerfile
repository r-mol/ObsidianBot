# Stage 1: Build the Go application
FROM golang:1.22.1 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go source code into the container
COPY . .

# Download Go modules
RUN go mod download

# Build the Go application as a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app ./cmd/main.go

# Stage 2: Create a minimal image to run the Go application
FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the binary file from the previous stage
COPY --from=builder /app/app .

# Expose port 8081 to the outside world
EXPOSE 8081

# Command to run the executable
CMD ["./app", "start", "--config", "/app/config.yaml"]
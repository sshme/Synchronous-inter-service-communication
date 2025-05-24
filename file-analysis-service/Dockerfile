FROM golang:1.23-alpine AS builder

# Install git for dependency downloads
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Install and run Swagger
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/api/main.go

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api

# Create the final image with minimal size
FROM alpine:3.18

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/bin/api /app/api

# Copy swagger documentation
COPY --from=builder /app/docs /app/docs

# Copy .env file if it exists
COPY --from=builder /app/config/.env /app/config/.env

# Set working directory
WORKDIR /app

# Expose the application port
EXPOSE 8000

# Run the application
CMD ["./api"] 
# Build stage
FROM golang:1.24-alpine AS build

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Install Encore
RUN apk add --no-cache curl && \
    curl -L https://encore.dev/install.sh | sh

# Add Encore to PATH
ENV PATH="/root/.encore/bin:$PATH"

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# Encore build generates a standalone directory with the binary and assets
RUN encore build --log-level=debug --on-error=continue render .output

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the built application from the build stage
COPY --from=build /app/.output /app

# Expose the port the app runs on
EXPOSE 8081

# Command to run the application
CMD ["/app/encore-app"]

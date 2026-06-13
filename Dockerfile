# Build Stage
FROM golang:alpine AS builder

# Install build dependencies for go-sqlite3 (requires CGO)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -o who-owes-me main.go

# Final Stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/who-owes-me .
# Copy templates
COPY --from=builder /app/templates ./templates

# Expose port
EXPOSE 8080

# Set entrypoint
CMD ["./who-owes-me"]

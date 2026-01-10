FROM golang:1.24-alpine as builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git postgresql-client curl

# Download migrate tool
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.2/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/tracehub-server ./cmd/server

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates postgresql-client

# Copy binary and migrations
COPY --from=builder /app/tracehub-server .
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./tracehub-server"]

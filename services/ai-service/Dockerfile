# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ai-service ./cmd/main.go

# Runtime stage
FROM gcr.io/distroless/base-debian12:latest

WORKDIR /app

COPY --from=builder /app/ai-service /app/ai-service

EXPOSE 8084

CMD ["/app/ai-service"]

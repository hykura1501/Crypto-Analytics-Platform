FROM golang:1.24-alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o crawler-service ./cmd/main.go

FROM debian:12-slim

WORKDIR /app

# Install chromium and dependencies
RUN apt-get update && apt-get install -y \
    chromium \
    chromium-driver \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/crawler-service /app/crawler-service

COPY .env /app/.env

# Set Chrome path for chromedp
ENV CHROME_BIN=/usr/bin/chromium
ENV CHROME_PATH=/usr/bin/chromium

EXPOSE 8083

ENTRYPOINT ["/app/crawler-service"]


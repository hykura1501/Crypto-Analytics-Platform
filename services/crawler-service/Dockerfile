FROM golang:1.22 AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o crawler-service ./cmd/main.go

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/crawler-service /app/crawler-service

EXPOSE 8083

ENTRYPOINT ["/app/crawler-service"]


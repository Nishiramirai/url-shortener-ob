# Builder
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/url-shortener ./cmd/shortener/main.go

# Run
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/bin/url-shortener .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./url-shortener"]
# Dockerfile for AliceTraINT
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git build-base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o AliceTraINT ./cmd/AliceTraINT

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /app/AliceTraINT .

EXPOSE 8088

CMD ["./AliceTraINT"]


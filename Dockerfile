# Dockerfile for AliceTraINT
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git build-base curl make

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# install tailwind CLI
ENV TAILWIND_VERSION=v3.4.10
RUN curl -L "https://github.com/tailwindlabs/tailwindcss/releases/download/$TAILWIND_VERSION/tailwindcss-linux-x64" -o /usr/local/bin/tailwindcss
RUN chmod +x /usr/local/bin/tailwindcss

COPY . .

# generate css using tailwind CLI
RUN make css

RUN CGO_ENABLED=0 GOOS=linux go build -o AliceTraINT ./cmd/AliceTraINT

FROM python:3.12-alpine

RUN apk add --no-cache ca-certificates \
                                cmake  \
                                make   \
                                g++    \
                                zlib-dev \
                                libuuid  \
                                util-linux-dev \
                                openssl \
                                libcrypto3 \
                                openssl-dev \
                                openssl-libs-static \
                                musl-dev \
                                linux-headers \
                                git \
                                rsync
RUN pip install alienpy

WORKDIR /root/

COPY gridCertificate.p12 .
RUN mkdir ~/.globus
RUN openssl pkcs12 -clcerts -nokeys -in ./gridCertificate.p12 -out ~/.globus/usercert.pem -password pass:
RUN openssl pkcs12 -nocerts -nodes -in ./gridCertificate.p12 -out ~/.globus/userkey.pem -password pass:
RUN chmod 0400 ~/.globus/userkey.pem
ENV CCDB_SSL_CERT_PATH=/root/.globus/usercert.pem
ENV CCDB_SSL_KEY_PATH=/root/.globus/userkey.pem

WORKDIR /app
RUN alien.py getCAcerts
# for some reason first execution of ls is returning nil - from second and on it works as expected
RUN alien_ls /

COPY --from=builder /app/.env /app/AliceTraINT ./
COPY --from=builder /app/web ./web
COPY --from=builder /app/static/ ./static/

EXPOSE 8088

CMD ["./AliceTraINT"]


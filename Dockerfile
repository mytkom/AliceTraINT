FROM golang:1.22-alpine AS builder

# --- config that differs locally vs OpenShift ---
ARG CERT_PATH=./gridCertificate.p12   # default for local builds

# Dependencies for Go + Tailwind
RUN apk add --no-cache git build-base curl make openssl

USER root
WORKDIR /app

# Copy Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy only important files
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./.env* .
COPY ./Makefile ./*.p12 .
RUN touch .env

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o AliceTraINT ./cmd/AliceTraINT

# Generate certificates
RUN mkdir -p ~/.globus
RUN openssl pkcs12 -clcerts -nokeys -in "$CERT_PATH" -out ./usercert.pem -password pass:
RUN openssl pkcs12 -nocerts -nodes -in "$CERT_PATH" -out ./userkey.pem -password pass:
RUN chmod 0400 ./userkey.pem

RUN git clone --depth=1 --branch master https://github.com/alisw/alien-cas.git /app/alien-cas
RUN openssl rehash /app/alien-cas

# --- runtime stage (same for both) ---
FROM registry.access.redhat.com/ubi9 AS runtime

# Install Tailwind CLI
ENV TAILWIND_VERSION=v3.4.10
RUN curl -L "https://github.com/tailwindlabs/tailwindcss/releases/download/$TAILWIND_VERSION/tailwindcss-linux-x64" \
    -o /usr/local/bin/tailwindcss && \
    chmod +x /usr/local/bin/tailwindcss

USER 1001
WORKDIR /app

USER root

COPY --from=builder /app/alien-cas ./alien-cas
ENV JALIEN_CERT_CA_DIR=/app/alien-cas

# Copy Go binary and certificates
COPY --from=builder /app/AliceTraINT ./
COPY --from=builder /app/usercert.pem ./usercert.pem
COPY --from=builder /app/userkey.pem ./userkey.pem
COPY --from=builder /app/.env ./

RUN chmod 0400 ./userkey.pem
ENV GRID_CERT_PATH=./usercert.pem
ENV GRID_KEY_PATH=./userkey.pem

# Docs and web templates are loaded at the execution
# of Go binary so they can be copied here and css generated
COPY ./tailwind* .
COPY ./web ./web
COPY ./static ./static

# Generate CSS (it may be unnecessary, but it's better to have it here)
RUN tailwindcss -i static/css/input.css -o static/css/output.css --minify

COPY ./docs ./docs

RUN mkdir -p . && \
    chgrp -R 0 . && \
    chmod -R g=u .

USER 1001

EXPOSE 8088
CMD ["./AliceTraINT"]
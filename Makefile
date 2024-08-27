.PHONY: build run lint test docker

build:
	go build -o bin/AliceTraINT ./cmd/AliceTraINT

run:
	go run ./cmd/AliceTraINT

lint:
	golangci-lint run

test:
	go test ./...

docker:
	docker-compose up --build


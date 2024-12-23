.PHONY: build
build:
	go build -o bin/AliceTraINT ./cmd/AliceTraINT

.PHONY: run
run:
	go run ./cmd/AliceTraINT

.PHONY: run-test
run-test:
	go run ./cmd/AliceTraINT_test

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test ./...

.PHONY: docker
docker:
	docker-compose up --build

## css: build tailwindcss
.PHONY: css
css:
	tailwindcss -i static/css/input.css -o static/css/output.css --minify

## css-watch: watch build tailwindcss
.PHONY: css-watch
css-watch:
	tailwindcss -i static/css/input.css -o static/css/output.css --watch

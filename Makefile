.PHONY: build run test docker-up docker-down env_copy

build:
	go build -o server ./cmd/

run:
	go run ./cmd/

test:
	go test -race ./...

docker-up:
	docker compose up --build

docker-down:
	docker compose down

env_copy:
	cp .env.example .env

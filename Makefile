.PHONY: up down build restart logs ps test test-v cover unit-cover integration seed bench lint swagger

run:
	go run ./cmd/app/main.go

up:
	docker compose up -d --build

down:
	docker compose down

build:
	go build ./...

restart:
	docker compose down
	docker compose up -d --build

logs:
	docker compose logs -f

ps:
	docker compose ps

test:
	go test ./...

test-v:
	go test ./... -v

cover:
	go test ./... -coverpkg=./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

unit-cover:
	go test ./internal/usecase/... -coverpkg=./internal/usecase/... -coverprofile=unit.out
	go tool cover -func=unit.out

integration:
	go test ./internal/tests -v

seed:
	go run bench/seed.go

bench:
	bash bench/bench.sh

swagger:
	@echo "Open http://localhost:8080/swagger"

lint:
	golangci-lint run
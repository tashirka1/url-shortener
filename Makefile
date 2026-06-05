include .env

.PHONY: up
up:
	@docker compose -p url-shortener up -d --remove-orphans

.PHONY: down
down:
	@docker compose -p url-shortener down

.PHONY: build
build:
	@docker compose -p url-shortener build

.PHONY: build-bin
build-bin:
	@go mod download
	@go tool templ generate && go build -ldflags="-s -w" -o bin/http cmd/http/main.go

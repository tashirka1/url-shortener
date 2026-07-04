include .env

.PHONY: up
up:
	@docker compose -p url-shortener up -d --remove-orphans --force-recreate

.PHONY: down
down:
	@docker compose -p url-shortener down

.PHONY: build
build:
	@docker compose -p url-shortener build

.PHONY: build-bin
build-bin:
	@go tool templ generate && go build -tags fts5 -ldflags="-s -w" -o bin/url-shortener cmd/url-shortener/main.go

.PHONY: lint
lint:
	@golangci-lint run ./...

.PHONY: check
check:
	golangci-lint run ./... && go test -tags fts5 ./...

.PHONY: air
air:
	@go tool air

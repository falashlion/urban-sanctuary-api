.PHONY: dev migrate-up migrate-down sqlc test lint build docker-build seed copy-env help

# Load .env if present
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

GO := /usr/local/go/bin/go
DOCKER_COMPOSE := docker-compose

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

dev: ## Start local infra + air live-reload dev server
	$(DOCKER_COMPOSE) -f docker-compose.dev.yml up -d
	$(GO) run github.com/cosmtrek/air@latest

run: ## Run server without live reload
	$(GO) run ./cmd/api

migrate-up: ## Run pending database migrations
	migrate -path ./migrations -database "$(DB_DSN)" up

migrate-down: ## Roll back last database migration
	migrate -path ./migrations -database "$(DB_DSN)" down 1

migrate-create: ## Create a new migration file (usage: make migrate-create name=create_xyz)
	migrate create -ext sql -dir ./migrations -seq $(name)

sqlc: ## Regenerate sqlc type-safe query code
	sqlc generate

test: ## Run all tests with race detector
	$(GO) test -race -count=1 ./...

test-cover: ## Run tests with coverage report
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

lint: ## Run golangci-lint
	golangci-lint run ./...

vet: ## Run go vet
	$(GO) vet ./...

build: ## Compile production binary
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-w -s" -o ./server ./cmd/api

docker-build: ## Build Docker image locally
	docker build -f deployments/Dockerfile -t urban-sanctuary-api:latest .

docker-up: ## Start production Docker stack
	$(DOCKER_COMPOSE) -f deployments/docker-compose.yml up -d

docker-down: ## Stop production Docker stack
	$(DOCKER_COMPOSE) -f deployments/docker-compose.yml down

copy-env: ## Copy root .env to deployments folder
	cp .env deployments/.env

seed: ## Load development seed data
	$(GO) run ./cmd/seed

tidy: ## Tidy go modules
	$(GO) mod tidy

fmt: ## Format code
	$(GO) fmt ./...
	goimports -w .

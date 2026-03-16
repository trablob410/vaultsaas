.PHONY: help dev down build test lint migrate-up migrate-down seed clean

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Start dev environment
	docker compose up -d --build

down: ## Stop all services
	docker compose down

build: ## Build all services
	docker compose build

test: ## Run all tests
	cd server && go test ./... -v
	cd dashboard && npm test
	cd mcp-server && cargo test

lint: ## Run linters
	cd server && golangci-lint run
	cd dashboard && npm run lint
	cd mcp-server && cargo clippy -- -D warnings

migrate-up: ## Run database migrations
	docker compose exec server /usr/local/bin/valt-server migrate up

migrate-down: ## Rollback last migration
	docker compose exec server /usr/local/bin/valt-server migrate down

seed: ## Seed development data
	docker compose exec server /usr/local/bin/valt-server seed

logs: ## Tail all service logs
	docker compose logs -f

logs-server: ## Tail server logs
	docker compose logs -f server

ps: ## Show running services
	docker compose ps

clean: ## Remove all volumes and containers
	docker compose down -v --remove-orphans

restart: ## Restart all services
	docker compose restart

db-shell: ## Open psql shell
	docker compose exec postgres psql -U $${POSTGRES_USER:-valt} -d $${POSTGRES_DB:-valt}

minio-console: ## Open MinIO console URL
	@echo "MinIO Console: http://localhost:9001"

.PHONY: dev sqlc tidy dr docker-run dd docker-down help

APP_NAME   := Shoka
BACKEND    := ./backend
FRONTEND   := ./frontend

# ----- Development -----
dev: ## Start full stack in dev mode
	@echo "Starting FlareSolverr..."
	docker compose -f docker-compose.dev.yml up -d flaresolverr
	@echo "Starting backend with Air..."
	cd $(BACKEND) && air &
	@echo "Starting frontend..."
	cd $(FRONTEND) && npm run dev
 
dev/backend: ## Start only the backend with Air
	cd $(BACKEND) && air
 
dev/frontend: ## Start only the frontend dev server
	cd $(FRONTEND) && npm run dev
 
dev/services: ## Start only background services 
	docker compose -f docker-compose.dev.yml up

# ----- Production -----
run: ## Start full production stack
	docker compose up -d
 
stop: ## Stop all containers
	docker compose down
 
logs: ## Tail all container logs
	docker compose logs -f
 
logs/backend: ## Tail backend logs only
	docker compose logs -f backend


# ----- Database -----
migrate/up: ## Apply all pending migrations
	cd $(BACKEND) && goose sqlite3 ./data/shoka.db up
 
migrate/down: ## Roll back the last migration
	cd $(BACKEND) && goose sqlite3 ./data/shoka.db down

migrate/status: ## Show migration status
	cd $(BACKEND) && goose sqlite3 ./data/shoka.db status
 
migrate/create: ## Create a new migration — usage: make migrate/create NAME=add_collections
	@[ "${NAME}" ] || ( echo "Usage: make migrate/create NAME=your_migration_name"; exit 1 )
	cd $(BACKEND) && goose sqlite3 ./data/shoka.db -s -dir internal/database/migrations create $(NAME) sql

# ----- Code Generation -----
sqlc: ## Regenerate sqlc Go code from query files
	cd $(BACKEND) && sqlc generate

swag: ## Regenerate swaggo API docs
	cd $(BACKEND) && swag init --parseDependency --parseInternal

# ----- Code Quality -----
tidy: ## Tidy Go module dependencies
	cd $(BACKEND) && go mod tidy

# ----- Cache -----
cache/clear: ## Delete all cached thumbnails and pages
	rm -rf ./cache/*/thumbs ./cache/*/pages
	@echo "Cache cleared."
 
cache/clear/pages: ## Delete only full-resolution page cache (keeps thumbnails)
	rm -rf ./cache/*/pages
	@echo "Full-resolution page cache cleared."

# ----- Help -----
help: ## List available commands
	@grep -E '^[a-zA-Z_/]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'
 
.DEFAULT_GOAL := help

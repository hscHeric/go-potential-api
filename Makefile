.PHONY: help install build run clean test docker-up docker-down docker-reset docker-logs migrate-up migrate-down migrate-drop migrate-force migrate-version migrate-create swagger dev

# Load environment variables
include .env
export

# Variables
APP_NAME=potential-idiomas-api
BINARY_NAME=api
DOCKER_COMPOSE=docker-compose
MIGRATE=migrate
MIGRATIONS_PATH=./migrations
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Default target
help:
	@echo "Available commands:"
	@echo "  make install         Install Go dependencies"
	@echo "  make build          Build the application"
	@echo "  make run            Run the application"
	@echo "  make dev            Run with hot reload (requires air)"
	@echo "  make clean          Clean build artifacts"
	@echo "  make test           Run tests"
	@echo ""
	@echo "Docker commands:"
	@echo "  make docker-up      Start all Docker containers"
	@echo "  make docker-down    Stop all Docker containers"
	@echo "  make docker-reset   Stop, remove volumes and restart containers"
	@echo "  make docker-logs    Show Docker logs"
	@echo "  make docker-ps      Show running containers"
	@echo ""
	@echo "Database migration commands:"
	@echo "  make migrate-up     Apply all pending migrations"
	@echo "  make migrate-down   Revert last migration"
	@echo "  make migrate-drop   Drop all migrations (WARNING: deletes all data)"
	@echo "  make migrate-force  Force migration version (usage: make migrate-force VERSION=1)"
	@echo "  make migrate-version Show current migration version"
	@echo "  make migrate-create  Create new migration (usage: make migrate-create NAME=add_users)"
	@echo ""
	@echo "API documentation:"
	@echo "  make swagger        Generate Swagger documentation"

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed successfully"

# Build the application
build:
	@echo "Building application..."
	go build -o bin/$(BINARY_NAME) cmd/api/main.go
	@echo "Build complete: bin/$(BINARY_NAME)"

# Run the application
run: build
	@echo "Starting application..."
	./bin/$(BINARY_NAME)

# Run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	@echo "Starting development server with hot reload..."
	air

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf tmp/
	go clean
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	go test -v -cover ./...

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Docker commands
docker-up:
	@echo "Starting Docker containers..."
	$(DOCKER_COMPOSE) up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Docker containers are running"
	@$(MAKE) docker-ps

docker-down:
	@echo "Stopping Docker containers..."
	$(DOCKER_COMPOSE) down
	@echo "Docker containers stopped"

docker-reset:
	@echo "Resetting Docker environment..."
	$(DOCKER_COMPOSE) down -v
	@echo "Removing volumes and orphaned containers..."
	docker volume prune -f
	@echo "Starting fresh containers..."
	$(DOCKER_COMPOSE) up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Docker environment reset complete"
	@$(MAKE) migrate-up

docker-logs:
	@echo "Showing Docker logs (Ctrl+C to exit)..."
	$(DOCKER_COMPOSE) logs -f

docker-logs-api:
	@echo "Showing API logs (Ctrl+C to exit)..."
	$(DOCKER_COMPOSE) logs -f api

docker-logs-db:
	@echo "Showing database logs (Ctrl+C to exit)..."
	$(DOCKER_COMPOSE) logs -f postgres

docker-ps:
	@echo "Running containers:"
	@$(DOCKER_COMPOSE) ps

docker-restart:
	@echo "Restarting Docker containers..."
	$(DOCKER_COMPOSE) restart
	@echo "Containers restarted"

# Database migration commands
migrate-up:
	@echo "Applying migrations..."
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up
	@echo "Migrations applied successfully"

migrate-down:
	@echo "Reverting last migration..."
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1
	@echo "Migration reverted successfully"

migrate-drop:
	@echo "WARNING: This will drop all tables and data"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" drop -f; \
		echo "All migrations dropped"; \
	else \
		echo "Operation cancelled"; \
	fi

migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "ERROR: VERSION is required"; \
		echo "Usage: make migrate-force VERSION=1"; \
		exit 1; \
	fi
	@echo "Forcing migration version $(VERSION)..."
	$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $(VERSION)
	@echo "Migration version forced to $(VERSION)"

migrate-version:
	@echo "Current migration version:"
	@$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "ERROR: NAME is required"; \
		echo "Usage: make migrate-create NAME=add_users_table"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_PATH) -seq $(NAME)
	@echo "Migration files created"

# Database access
db-connect:
	@echo "Connecting to database..."
	docker exec -it potential_db psql -U $(DB_USER) -d $(DB_NAME)

db-dump:
	@echo "Creating database dump..."
	docker exec potential_db pg_dump -U $(DB_USER) $(DB_NAME) > backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "Dump created successfully"

db-restore:
	@if [ -z "$(FILE)" ]; then \
		echo "ERROR: FILE is required"; \
		echo "Usage: make db-restore FILE=backup.sql"; \
		exit 1; \
	fi
	@echo "Restoring database from $(FILE)..."
	docker exec -i potential_db psql -U $(DB_USER) -d $(DB_NAME) < $(FILE)
	@echo "Database restored successfully"

# Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/api/main.go -o docs
	@echo "Swagger documentation generated in docs/"
	@echo "Access docs at: http://localhost:8080/swagger/index.html"

# Development helpers
setup: install docker-up migrate-up
	@echo "Development environment setup complete"
	@echo "Database is ready with all migrations applied"

reset: clean docker-reset
	@echo "Environment reset complete"

# Linting and formatting
lint:
	@echo "Running linter..."
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development tools installed"

# Production build
build-prod:
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/$(BINARY_NAME) cmd/api/main.go
	@echo "Production build complete"

# Docker build for API
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .
	@echo "Docker image built successfully"

# Full reset and setup (dangerous - use with caution)
nuke: docker-down
	@echo "WARNING: This will remove all containers, volumes, and data"
	@read -p "Are you sure? Type 'yes' to confirm: " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		docker-compose down -v --remove-orphans; \
		docker system prune -f; \
		rm -rf bin/ tmp/; \
		echo "Environment nuked successfully"; \
		echo "Run 'make setup' to reinitialize"; \
	else \
		echo "Operation cancelled"; \
	fi

# SHBucket Makefile

# Variables
BINARY_NAME=shbucket
DOCKER_IMAGE=shbucket:latest
GO_VERSION=1.21
DATABASE_URL?=postgres://shbucket:shbucket_password@localhost:5432/shbucket?sslmode=disable

# Default target
.PHONY: help
help: ## Show this help message
	@echo 'Usage: make <target>'
	@echo ''
	@echo 'Targets:'
	@egrep '^(.+)\:\ ##\ (.+)' $(MAKEFILE_LIST) | column -t -c 2 -s ':#'

# Quick Start
.PHONY: master
master: ## Start SHBucket as master server with Docker Compose
	@echo "🚀 Starting SHBucket Master Server..."
	docker-compose up -d
	@echo ""
	@echo "✅ SHBucket Master Server is starting!"
	@echo "📊 Web Dashboard: http://localhost:3000"
	@echo "🌐 API Endpoint: http://localhost:8080/api/v1" 
	@echo "📚 API Docs: http://localhost:8080/swagger"
	@echo "🔑 Default Login: admin@shbucket.local / admin123"

.PHONY: master-web
master-web: ## Start master server with web interface
	@echo "🚀 Starting SHBucket Master with Web Interface..."
	docker-compose --profile web up -d
	@echo ""
	@echo "✅ SHBucket Master + Web is starting!"
	@echo "📊 Web Dashboard: http://localhost:3000"
	@echo "🌐 API Endpoint: http://localhost:8080/api/v1"

.PHONY: node
node: ## Start SHBucket as storage node
	@if [ -z "$(MASTER_URL)" ]; then \
		echo "❌ Error: MASTER_URL is required for node mode"; \
		echo "Usage: make node MASTER_URL=http://master-server:8080 NODE_PORT=8081"; \
		exit 1; \
	fi
	@echo "🗄️ Starting SHBucket Storage Node..."
	@echo "🔗 Connecting to master: $(MASTER_URL)"
	MASTER_URL=$(MASTER_URL) NODE_PORT=$(NODE_PORT) docker-compose -f docker-compose.node.yml up -d
	@echo ""
	@echo "✅ Storage Node is starting!"
	@echo "🔗 Master URL: $(MASTER_URL)"
	@echo "📡 Node Port: $(NODE_PORT)"

# Local Development
.PHONY: run-master
run-master: ## Run master server locally (requires PostgreSQL)
	@echo "🚀 Starting SHBucket Master Server locally..."
	@echo "📊 Ensure PostgreSQL is running on localhost:5432"
	DATABASE_URL="$(DATABASE_URL)" \
	STORAGE_PATH="./storage" \
	JWT_SECRET="development-secret-key" \
	SIGNATURE_SECRET="development-signature-secret" \
	ADMIN_PASSWORD="admin123" \
	PORT=8080 \
	go run cmd/server/main.go

.PHONY: run-node
run-node: ## Run storage node locally
	@if [ -z "$(MASTER_URL)" ]; then \
		echo "❌ Error: MASTER_URL is required"; \
		echo "Usage: make run-node MASTER_URL=http://localhost:8080 PORT=8081"; \
		exit 1; \
	fi
	@echo "🗄️ Starting Storage Node locally..."
	MASTER_URL=$(MASTER_URL) \
	PORT=$(PORT) \
	MODE=node \
	STORAGE_PATH="./node-storage" \
	go run cmd/server/main.go

# Build
.PHONY: build
build: ## Build the application binary
	@echo "🔨 Building SHBucket..."
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/$(BINARY_NAME) cmd/server/main.go
	@echo "✅ Binary built: bin/$(BINARY_NAME)"

.PHONY: build-web
build-web: ## Build the web UI
	cd web && npm ci && npm run build

.PHONY: clean
clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -rf web/build
	rm -rf web/node_modules

# Testing
.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: vet
vet: ## Run go vet
	go vet ./...

# Dependencies
.PHONY: deps
deps: ## Download dependencies
	go mod download
	go mod tidy

.PHONY: deps-web
deps-web: ## Install web dependencies
	cd web && npm ci

# Database
.PHONY: migrate
migrate: ## Run database migrations
	@echo "📊 Running database migrations..."
	DATABASE_URL="$(DATABASE_URL)" go run cmd/migrations/main.go migrations:up
	@echo "✅ Migrations completed"

.PHONY: migrate-status
migrate-status: ## Check migration status
	DATABASE_URL="$(DATABASE_URL)" go run cmd/migrations/main.go migrations:status

.PHONY: migrate-rollback
migrate-rollback: ## Rollback last migration
	DATABASE_URL="$(DATABASE_URL)" go run cmd/migrations/main.go migrations:rollback

.PHONY: migrate-create
migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Error: NAME is required"; \
		echo "Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	DATABASE_URL="$(DATABASE_URL)" go run cmd/migrations/main.go migration add $(NAME)

# Docker
.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-run
docker-run: ## Run Docker container
	docker run -p 8080:8080 $(DOCKER_IMAGE)

.PHONY: stop
stop: ## Stop all SHBucket services
	@echo "🛑 Stopping SHBucket services..."
	docker-compose down
	docker-compose -f docker-compose.node.yml down 2>/dev/null || true
	@echo "✅ All services stopped"

.PHONY: logs
logs: ## Show SHBucket logs
	docker-compose logs -f

.PHONY: logs-node
logs-node: ## Show storage node logs
	docker-compose -f docker-compose.node.yml logs -f

.PHONY: docker-clean
docker-clean: ## Clean Docker images and containers
	docker-compose down -v --remove-orphans
	docker system prune -f

# Production helpers
.PHONY: backup-db
backup-db: ## Backup database
	docker-compose exec -T db pg_dump -U shbucket shbucket > backup-$(shell date +%Y%m%d-%H%M%S).sql

.PHONY: restore-db
restore-db: ## Restore database (usage: make restore-db FILE=backup.sql)
	docker-compose exec -T db psql -U shbucket -d shbucket < $(FILE)

# Health checks
.PHONY: health
health: ## Check application health
	curl -f http://localhost:8080/health || exit 1

.PHONY: check-deps
check-deps: ## Check if required tools are installed
	@echo "Checking dependencies..."
	@command -v go >/dev/null 2>&1 || { echo "Go is not installed"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "Docker is not installed"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "Docker Compose is not installed"; exit 1; }
	@echo "All dependencies are installed!"

# Documentation
.PHONY: docs
docs: ## Generate API documentation
	@echo "API documentation available at:"
	@echo "- Swagger UI: http://localhost:8080/swagger/"
	@echo "- ReDoc: http://localhost:8080/docs/"

# Quick setup
.PHONY: setup
setup: ## Complete setup for development
	@echo "🔧 Setting up SHBucket development environment..."
	@echo "1. Installing dependencies..."
	make deps
	@echo "2. Building web interface..."
	make build-web  
	@echo "3. Building binary..."
	make build
	@echo ""
	@echo "✅ Setup complete!"
	@echo ""
	@echo "🚀 Quick start commands:"
	@echo "  make master     - Start master server with Docker"
	@echo "  make master-web - Start master + web interface" 
	@echo "  make node MASTER_URL=http://localhost:8080 - Start storage node"

.PHONY: quickstart
quickstart: check-deps master-web ## Complete quickstart with web interface
	@echo ""
	@echo "🎉 SHBucket is ready!"
	@echo ""
	@echo "📊 Web Dashboard: http://localhost:3000"
	@echo "🌐 API Endpoint: http://localhost:8080/api/v1"
	@echo "📚 API Docs: http://localhost:8080/swagger"
	@echo "🔑 Default Login: admin@shbucket.local / admin123"

# All-in-one targets
.PHONY: ci
ci: deps vet lint test ## Run CI pipeline

.PHONY: clean-all
clean-all: stop docker-clean clean ## Clean everything

# Version info
.PHONY: version
version: ## Show version information
	@echo "SHBucket v1.0.0"
	@echo "Go version: $(shell go version)"
	@echo "Git commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
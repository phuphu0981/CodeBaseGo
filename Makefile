.PHONY: run build setup wire swagger test clean install-tools migrate-up migrate-down migrate-status

# Migration configuration (Default: MySQL)
DB_DRIVER ?= mysql
DB_DSN ?= "root:password@tcp(127.0.0.1:3306)/codebasego?parseTime=true"

# Run the application (auto-validates module dependencies & syncs Wire DI)
run: setup
	go run ./cmd/api/

# Build binary (auto-validates module dependencies & syncs Wire DI)
build: setup
	go build -o bin/api ./cmd/api/

# Auto-detect modules, sync configs/modules.yaml, and generate Wire DI
setup:
	go run ./cmd/tools/modgen/

# Generate Wire dependency injection code (runs setup)
wire: setup

# Generate Swagger documentation
swagger:
	@which swag > /dev/null 2>&1 && swag init -g cmd/api/main.go -o docs/ || go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o docs/


# Run tests
test:
	go test ./... -v -cover

# Database migrations (using goose)
migrate-up:
	goose -dir migrations $(DB_DRIVER) $(DB_DSN) up

migrate-down:
	goose -dir migrations $(DB_DRIVER) $(DB_DSN) down

migrate-status:
	goose -dir migrations $(DB_DRIVER) $(DB_DSN) status

# Clean build artifacts
clean:
	rm -rf bin/ tmp/

# Install development tools
install-tools:
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run ./...

# Download dependencies
deps:
	go mod tidy
	go mod download

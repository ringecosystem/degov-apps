.PHONY: build run test clean docker-up docker-down generate help

help:
	@echo "Available commands:"
	@echo "  build       - Build the application"
	@echo "  serve       - Build and serve the application"
	@echo "  run         - Run the application"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean build artifacts"
	@echo "  docker-up   - Start Docker services"
	@echo "  docker-down - Stop Docker services"
	@echo "  generate    - Generate GraphQL code"

build:
	go build -o bin/degov-server cmd/main.go

run:
	go run cmd/main.go

serve: build
	./bin/degov-server

test:
	go test ./...

clean:
	rm -rf bin/
	rm -rf .data/

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

generate:
	go get github.com/99designs/gqlgen
	go generate ./...
	go mod tidy

deps:
	go mod tidy
	go mod download

# complete setup for development environment
setup: generate deps docker-up
	@echo "Setting up development environment..."
	@sleep 2
	@echo "Development environment ready!"
	@echo "Run 'make run' to start the application"

reset: docker-down clean setup serve
	@echo "Resetting development environment..."
	@sleep 2
	@echo "Development environment reset!"

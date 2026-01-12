.PHONY: all build clean test vet server client docker-build help

# Binary output directory
BIN_DIR := bin

# Default target
all: build

# Build both server and client
build: server client

# Build server
server:
	@echo "Building server..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/server ./cmd/server

# Build client
client:
	@echo "Building client..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/client ./cmd/client

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Build Docker images (requires Linux binaries)
docker-build:
	@echo "Building binaries for Docker..."
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o client ./cmd/client
	@echo "Building Docker images..."
	@docker build -t websocket-test-server:dev -f Dockerfile.server .
	@docker build -t websocket-test-client:dev -f Dockerfile.client .
	@echo "Cleaning up Linux binaries..."
	@rm -f server client

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@[ ! -f server ] || rm -f server
	@[ ! -f client ] || rm -f client
	@[ ! -d dist ] || rm -rf dist/
	@[ ! -f server/server ] || rm -f server/server
	@[ ! -f client/client ] || rm -f client/client

# Run server locally
run-server: server
	@$(BIN_DIR)/server

# Run client locally
run-client: client
	@$(BIN_DIR)/client

# Show help
help:
	@echo "Available targets:"
	@echo "  all          - Build both server and client (default)"
	@echo "  build        - Build both server and client"
	@echo "  server       - Build server only"
	@echo "  client       - Build client only"
	@echo "  test         - Run tests"
	@echo "  vet          - Run go vet"
	@echo "  docker-build - Build Docker images"
	@echo "  clean        - Remove build artifacts"
	@echo "  run-server   - Build and run server"
	@echo "  run-client   - Build and run client"
	@echo "  help         - Show this help message"


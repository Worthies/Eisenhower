.PHONY: build clean install test run help

# Binary name
BINARY_NAME=eisenhower
# Installation directory
INSTALL_DIR=$(HOME)/.local/bin
# Default database path
DB_PATH=$(HOME)/.config/eisenhower.db

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the MCP server
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME)
	@echo "✓ Build complete: ./$(BINARY_NAME)"

clean: ## Remove built binary
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@echo "✓ Clean complete"

install: build ## Build and install to ~/.local/bin
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BINARY_NAME) $(INSTALL_DIR)/
	@echo "✓ Installed to $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo ""
	@echo "Add this to your VS Code settings:"
	@echo '{'
	@echo '  "github.copilot.chat.mcp.servers": {'
	@echo '    "eisenhower": {'
	@echo '      "command": "$(INSTALL_DIR)/$(BINARY_NAME)",'
	@echo '      "args": ["-db", "$(DB_PATH)"]'
	@echo '    }'
	@echo '  }'
	@echo '}'

test: build ## Run basic server test
	@echo "Testing server..."
	@./test-server.sh

run: build ## Build and run the server (for testing)
	@echo "Starting server..."
	@echo "Press Ctrl+C to stop"
	@./$(BINARY_NAME)

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies ready"

fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Formatting complete"

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "✓ Vet complete"

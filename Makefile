# Custodian Killer - AWS Policy Management Tool
# Making AWS compliance fun again! 🔥

# Variables
BINARY_NAME=custodian-killer
VERSION?=1.0.0
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Directories
BUILD_DIR=build
DIST_DIR=dist

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
PURPLE=\033[0;35m
CYAN=\033[0;36m
NC=\033[0m # No Color

.PHONY: all build clean test deps fmt lint run help install uninstall docker

# Default target
all: clean deps fmt test build

# Help target
help: ## Show this help message
	@echo "$(CYAN)🔥 Custodian Killer - Making AWS compliance fun again!$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the binary
build: ## Build the application
	@echo "$(BLUE)🏗️  Building Custodian Killer...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)✅ Build complete! Binary: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Build for multiple platforms
build-all: ## Build for multiple platforms (Linux, macOS, Windows)
	@echo "$(BLUE)🌍 Building for multiple platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	
	# Linux AMD64
	@echo "$(PURPLE)  📦 Building for Linux AMD64...$(NC)"
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	
	# Linux ARM64
	@echo "$(PURPLE)  📦 Building for Linux ARM64...$(NC)"
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	# macOS AMD64
	@echo "$(PURPLE)  📦 Building for macOS AMD64...$(NC)"
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	
	# macOS ARM64 (Apple Silicon)
	@echo "$(PURPLE)  📦 Building for macOS ARM64...$(NC)"
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	
	# Windows AMD64
	@echo "$(PURPLE)  📦 Building for Windows AMD64...$(NC)"
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	
	@echo "$(GREEN)✅ Multi-platform build complete! Check $(DIST_DIR)/ directory$(NC)"

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "$(YELLOW)🧹 Cleaning build artifacts...$(NC)"
	$(GOCLEAN)
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	@echo "$(GREEN)✅ Clean complete!$(NC)"

# Download dependencies
deps: ## Download and verify dependencies
	@echo "$(BLUE)📦 Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) verify
	$(GOMOD) tidy
	@echo "$(GREEN)✅ Dependencies updated!$(NC)"

# Format code
fmt: ## Format Go code
	@echo "$(BLUE)🎨 Formatting code...$(NC)"
	$(GOFMT) ./...
	@echo "$(GREEN)✅ Code formatted!$(NC)"

# Run tests
test: ## Run all tests
	@echo "$(BLUE)🧪 Running tests...$(NC)"
	$(GOTEST) -v ./...
	@echo "$(GREEN)✅ Tests complete!$(NC)"

# Run tests with coverage
test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)🧪 Running tests with coverage...$(NC)"
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report generated: coverage.html$(NC)"

# Lint code (requires golangci-lint)
lint: ## Run linter (requires golangci-lint)
	@echo "$(BLUE)🔍 Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "$(GREEN)✅ Linting complete!$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint not found. Install it with:$(NC)"; \
		echo "$(YELLOW)   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.54.2$(NC)"; \
	fi

# Run the application
run: build ## Build and run the application
	@echo "$(CYAN)🚀 Starting Custodian Killer...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run in interactive mode
interactive: build ## Build and run in interactive mode
	@echo "$(CYAN)🎯 Starting Custodian Killer in interactive mode...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) interactive

# Install to system PATH
install: build ## Install binary to system PATH
	@echo "$(BLUE)📥 Installing Custodian Killer...$(NC)"
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)✅ Installed! You can now run 'custodian-killer' from anywhere$(NC)"

# Uninstall from system PATH
uninstall: ## Uninstall binary from system PATH
	@echo "$(YELLOW)🗑️  Uninstalling Custodian Killer...$(NC)"
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✅ Uninstalled!$(NC)"

# Development server (with file watching)
dev: ## Run in development mode with auto-restart (requires entr)
	@echo "$(CYAN)🔄 Starting development mode...$(NC)"
	@if command -v entr >/dev/null 2>&1; then \
		find . -name '*.go' | entr -r make run; \
	else \
		echo "$(YELLOW)⚠️  'entr' not found. Install it for auto-restart functionality$(NC)"; \
		echo "$(YELLOW)   On macOS: brew install entr$(NC)"; \
		echo "$(YELLOW)   On Ubuntu: apt-get install entr$(NC)"; \
		make run; \
	fi

# Docker targets
docker-build: ## Build Docker image
	@echo "$(BLUE)🐳 Building Docker image...$(NC)"
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest
	@echo "$(GREEN)✅ Docker image built: $(BINARY_NAME):$(VERSION)$(NC)"

docker-run: docker-build ## Build and run Docker container
	@echo "$(CYAN)🐳 Running Docker container...$(NC)"
	docker run -it --rm $(BINARY_NAME):latest

# Release targets
release: clean deps fmt test build-all ## Create a release (clean, test, build all platforms)
	@echo "$(PURPLE)🎉 Creating release $(VERSION)...$(NC)"
	@mkdir -p $(DIST_DIR)/release
	
	# Create archives for each platform
	@cd $(DIST_DIR) && tar -czf release/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	@cd $(DIST_DIR) && tar -czf release/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	@cd $(DIST_DIR) && tar -czf release/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	@cd $(DIST_DIR) && tar -czf release/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	@cd $(DIST_DIR) && zip -q release/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	
	@echo "$(GREEN)✅ Release $(VERSION) created in $(DIST_DIR)/release/$(NC)"

# Quick build and test
quick: ## Quick build and test (for development)
	@echo "$(CYAN)⚡ Quick build and test...$(NC)"
	$(GOTEST) ./... && $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)✅ Quick build complete!$(NC)"

# Benchmark tests
bench: ## Run benchmark tests
	@echo "$(BLUE)⏱️  Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

# Security scan (requires gosec)
security: ## Run security scan (requires gosec)
	@echo "$(BLUE)🔒 Running security scan...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
		echo "$(GREEN)✅ Security scan complete!$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  gosec not found. Install it with:$(NC)"; \
		echo "$(YELLOW)   go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest$(NC)"; \
	fi

# Generate documentation
docs: ## Generate documentation
	@echo "$(BLUE)📚 Generating documentation...$(NC)"
	$(GOCMD) doc ./...

# Show project statistics
stats: ## Show project statistics
	@echo "$(CYAN)📊 Project Statistics$(NC)"
	@echo "$(YELLOW)Lines of code:$(NC)"
	@find . -name '*.go' -not -path './vendor/*' | xargs wc -l | tail -1
	@echo "$(YELLOW)Go files:$(NC)"
	@find . -name '*.go' -not -path './vendor/*' | wc -l
	@echo "$(YELLOW)Packages:$(NC)"
	@find . -name '*.go' -not -path './vendor/*' -exec dirname {} \; | sort -u | wc -l

# Check Go version
check-go: ## Check Go version
	@echo "$(BLUE)🔍 Checking Go version...$(NC)"
	@$(GOCMD) version
	@echo "$(YELLOW)Required: Go 1.21 or later$(NC)"

# Initialize project (for new setups)
init: ## Initialize project dependencies and tools
	@echo "$(BLUE)🚀 Initializing Custodian Killer project...$(NC)"
	$(GOMOD) tidy
	@echo "$(GREEN)✅ Project initialized! Run 'make help' to see available commands$(NC)"

# Show build info
info: ## Show build information
	@echo "$(CYAN)ℹ️  Build Information$(NC)"
	@echo "$(YELLOW)Binary Name:$(NC) $(BINARY_NAME)"
	@echo "$(YELLOW)Version:$(NC) $(VERSION)"
	@echo "$(YELLOW)Build Time:$(NC) $(BUILD_TIME)"
	@echo "$(YELLOW)Git Commit:$(NC) $(GIT_COMMIT)"
	@echo "$(YELLOW)Go Version:$(NC) $$(go version)"

# Validate project structure
validate: ## Validate project structure and dependencies
	@echo "$(BLUE)✅ Validating project...$(NC)"
	@$(GOMOD) verify
	@$(GOCMD) vet ./...
	@echo "$(GREEN)✅ Project validation complete!$(NC)"

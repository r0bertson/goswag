.PHONY: help test build integration comprehensive integration-test install-deps clean test-coverage fmt lint docs

# Default target
help:
	@echo "Available targets:"
	@echo "  make integration       - Run the integration app (Gin framework, generates swagger and serves Swagger UI)"
	@echo "  make comprehensive     - Run the comprehensive example (HTTP framework, demonstrates all v2.0 features)"
	@echo "  make integration-test - Run integration tests"
	@echo "  make test            - Run all tests"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make build           - Build the project"
	@echo "  make install-deps    - Install required dependencies (swag)"
	@echo "  make clean           - Clean generated files"
	@echo "  make fmt             - Format code"
	@echo "  make lint            - Lint code (requires golangci-lint)"
	@echo "  make docs            - Generate swagger documentation"

# Run the integration app
integration:
	@echo "🚀 Starting integration app..."
	@cd examples && go run integration_app.go

# Run the comprehensive example (HTTP framework)
comprehensive:
	@echo "🚀 Starting comprehensive example..."
	@cd examples && go run comprehensive_example.go

# Run integration tests
integration-test:
	@echo "🧪 Running integration tests..."
	@go test ./examples -v -run TestIntegration

# Run all tests
test:
	@echo "🧪 Running all tests..."
	@go test ./... -v

# Build the project
build:
	@echo "🔨 Building project..."
	@go build ./...

# Install dependencies (swag)
install-deps:
	@echo "📦 Installing swag..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✓ swag installed successfully"

# Clean generated files
clean:
	@echo "🧹 Cleaning generated files..."
	@rm -f goswag.go
	@rm -rf docs
	@rm -rf examples/goswag
	@find . -type d -name "goswag-integration-*" -exec rm -rf {} + 2>/dev/null || true
	@echo "✓ Cleaned generated files"

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	@go test ./... -race -covermode atomic -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "📝 Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

# Lint code (requires golangci-lint)
lint:
	@echo "🔍 Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠ golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Generate swagger docs (for examples)
docs:
	@echo "📚 Generating swagger documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		cd examples && go run integration_app.go & \
		sleep 2 && \
		pkill -f integration_app || true; \
		echo "✓ Swagger docs generated"; \
	else \
		echo "⚠ swag not found. Install it with: make install-deps"; \
	fi

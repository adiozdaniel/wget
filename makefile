# Variables
APP_NAME = wget
MAIN_FILE = ./main.go
BUILD_OUTPUT = ./$(APP_NAME)
MODULE_NAME = wget  # Change this to your actual module name

# Default target: Reinitialize module and build the application
all:
	@echo "Cleaning up go.mod..."
	@rm -f go.mod go.sum
	@echo "Reinitializing Go module..."
	@go mod init $(MODULE_NAME)
	@echo "Downloading dependencies..."
	@go mod tidy
	@echo "Building the application..."
	go build -o $(BUILD_OUTPUT) .

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f $(BUILD_OUTPUT)

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@echo "  (default) - Reinitialize go.mod and build the project"
	@echo "  clean     - Remove build artifacts"
	@echo "  test      - Run tests"
	@echo "  fmt       - Format code"
	@echo "  help      - Show this help message"

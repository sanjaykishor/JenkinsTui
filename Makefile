.PHONY: build run clean test install

# Variables
BINARY_NAME=jenkinsTui
BUILD_DIR=./bin
VERSION=0.1.0
MAIN_PATH=./cmd/jenkinsTui
GOFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Default action: build
all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete!"

# Run the application
run:
	@go run $(GOFLAGS) $(MAIN_PATH)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Install the application
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(GOFLAGS) $(MAIN_PATH)
	@echo "Installation complete!"

# Generate a template config file
config:
	@echo "Generating template config file..."
	@mkdir -p $(HOME)/.jenkins-cli
	@[ -f $(HOME)/.jenkins-cli.yaml ] || \
		echo "current: default\njenkins_servers:\n  - name: default\n    url: https://jenkins.example.com\n    username: admin\n    token: your-api-token-here\n    proxy: \"\"\n    insecureSkipVerify: true" > $(HOME)/.jenkins-cli.yaml
	@echo "Generated config at $(HOME)/.jenkins-cli.yaml"

# Show help
help:
	@echo "Jenkins TUI - A terminal UI for Jenkins"
	@echo ""
	@echo "Usage:"
	@echo "  make build     - Build the application"
	@echo "  make run       - Run the application"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make test      - Run tests"
	@echo "  make install   - Install the application"
	@echo "  make config    - Generate a template config file"
	@echo "  make help      - Show this help"

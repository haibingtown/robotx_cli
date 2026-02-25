.PHONY: build install install-user install-binary install-go clean test deps tidy build-all help

# Build the CLI binary
build:
	@echo "Building RobotX CLI..."
	go build -o robotx ./cmd/robotx
	@echo "✅ Build complete: robotx"

# Install the CLI to /usr/local/bin
install: build
	@echo "Installing RobotX CLI..."
	sudo mv robotx /usr/local/bin/robotx
	@echo "✅ Installed to /usr/local/bin/robotx"

# Install for current user only (no sudo required)
install-user: build
	@echo "Installing RobotX CLI for current user..."
	mkdir -p ~/bin
	mv robotx ~/bin/robotx
	@echo "✅ Installed to ~/bin/robotx"
	@echo "⚠️  Make sure ~/bin is in your PATH"

# Install prebuilt binary from GitHub release (no Go toolchain required)
install-binary:
	@echo "Installing RobotX CLI from release binaries..."
	./scripts/install.sh

# Install with go install and auto PATH update
install-go:
	@echo "Installing RobotX CLI with go install..."
	./scripts/go-install.sh

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f robotx robotx-*
	@echo "✅ Clean complete"

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	@echo "✅ Dependencies downloaded"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	@echo "✅ Dependencies tidied"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=darwin GOARCH=amd64 go build -o robotx-darwin-amd64 ./cmd/robotx
	GOOS=darwin GOARCH=arm64 go build -o robotx-darwin-arm64 ./cmd/robotx
	GOOS=linux GOARCH=amd64 go build -o robotx-linux-amd64 ./cmd/robotx
	GOOS=linux GOARCH=arm64 go build -o robotx-linux-arm64 ./cmd/robotx
	GOOS=windows GOARCH=amd64 go build -o robotx-windows-amd64.exe ./cmd/robotx
	@echo "✅ Multi-platform build complete"


# Show help
help:
	@echo "RobotX CLI Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build        - Build the CLI binary"
	@echo "  install      - Install the CLI to /usr/local/bin (requires sudo)"
	@echo "  install-user - Install the CLI to ~/bin (no sudo required)"
	@echo "  install-binary - Install prebuilt binary from GitHub releases"
	@echo "  install-go   - Install with go install and auto PATH update"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  deps         - Download dependencies"
	@echo "  tidy         - Tidy dependencies"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  help         - Show this help message"

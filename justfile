# Finicky: macOS Browser Chooser

default:
    @just --list

# Install all dependencies (Node.js and Go)
install:
    ./scripts/install.sh

# Build the application and install to /Applications
build:
    ./scripts/build.sh

# Watch for changes and rebuild automatically (requires fd and entr)
watch:
    ./scripts/watch.sh

# Run the application
run:
    open /Applications/Finicky.app

# Build and run
build-run: build run

# Build config-api only
build-config:
    cd packages/config-api && npm run build && npm run generate-types

# Build finicky-ui only
build-ui:
    cd packages/finicky-ui && npm run build

# Build Go app only
build-go:
    cd apps/finicky/src && go build -o ../build/finicky

# Run tests for config-api
test:
    cd packages/config-api && npm test

# Run tests for Go app
test-go:
    cd apps/finicky/src && go test ./...

# Run tests with verbose output
test-go-v:
    cd apps/finicky/src && go test -v ./...

# Format Go code
fmt:
    cd apps/finicky/src && go fmt ./...
    shfmt -w -i 2 -ci ./scripts/*.sh 2>/dev/null || true

# Lint Go code
lint:
    cd apps/finicky/src && go vet ./...

# Clean build artifacts
clean:
    rm -rf apps/finicky/build
    rm -rf packages/config-api/dist
    rm -rf packages/finicky-ui/dist
    cd apps/finicky/src && go clean

# Update Go dependencies
deps:
    cd apps/finicky/src && go mod tidy

# Kill running Finicky instance
kill:
    killall Finicky || true

# Show logs (open Console.app filtered to Finicky)
logs:
    open /Applications/Finicky.app

# Run all checks (fmt, lint, test)
check: fmt lint test test-go

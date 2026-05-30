APP_NAME := gozone
BIN_DIR := ./bin

.PHONY: default build run test test-verbose clean fmt vet gosec deps update docker-build docker-up docker-down help

default: help

# build the binary
build:
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/gozone

# build and run locally
run: build
	$(BIN_DIR)/$(APP_NAME) -config config.yaml

# run tests
test:
	go test ./...

# run tests with verbose output
test-verbose:
	go test -v ./...

# remove build artifacts and database
clean:
	rm -rf $(BIN_DIR)/$(APP_NAME) ./data/gozone.db*

# format all source files
fmt:
	go fmt ./...

# run vet on all packages
vet:
	go vet ./...

# run gosec security analysis (optional tool)
gosec:
	@if command -v gosec > /dev/null 2>&1; then \
		gosec -exclude-dir='.cache|vendor|bin' -no-fail ./...; \
	else \
		echo "gosec not installed. Run: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

# download and tidy dependencies
deps:
	go mod download
	go mod tidy

# run update
update:
	go get -u ./...
	go mod tidy
	go mod vendor

# build Docker image
docker-build:
	docker build -t gozone .

# start services with docker-compose
docker-up:
	docker-compose up -d

# stop services
docker-down:
	docker-compose down

# show available commands
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build           Build the binary"
	@echo "  run             Build and run locally"
	@echo "  test            Run tests"
	@echo "  test-verbose    Run tests with verbose output"
	@echo "  clean           Remove build artifacts and database"
	@echo "  fmt             Format all source files"
	@echo "  vet             Run vet on all packages"
	@echo "  gosec           Run gosec security analysis"
	@echo "  deps            Download and tidy dependencies"
	@echo "  update          Update all dependencies"
	@echo "  docker-build    Build Docker image"
	@echo "  docker-up       Start services with docker-compose"
	@echo "  docker-down     Stop services"
	@echo "  help            Show this message"

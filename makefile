# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=myapp
COVERAGE_FILE=coverage.out
HTML_COVERAGE=coverage.html
GOLINT=golangci-lint
SRC_DIRS=./...
MAIN_PACKAGE := ./cmd

# Version handling that won't fail if git is not available
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags="-w -s -X main.version=$(VERSION)"

.DEFAULT_GOAL := help

# Fix the duplicate build-linux target and ensure proper flags
build-linux:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PACKAGE)
all: fmt vet lint test coverage build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PACKAGE)

test:
	$(GOTEST) $(SRC_DIRS)

coverage:
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) $(SRC_DIRS)
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(HTML_COVERAGE)

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(COVERAGE_FILE)
	rm -f $(HTML_COVERAGE)

run: build
	./$(BINARY_NAME)

deps:
	$(GOGET) $(SRC_DIRS)

# lint:
# 	$(GOLINT) run

fmt:
	$(GOCMD) fmt $(SRC_DIRS)

vet:
	$(GOCMD) vet $(SRC_DIRS)

benchmark:
	$(GOTEST) -bench=. -benchmem $(SRC_DIRS)

docker:
	docker build -t $(BINARY_NAME) .

generate:
	$(GOCMD) generate $(SRC_DIRS)


build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)_windows.exe -v

help:
	@echo "Available targets:"
	@echo "  build           : Build the binary"
	@echo "  test            : Run tests"
	@echo "  coverage        : Generate test coverage"
	@echo "  clean           : Clean up build artifacts"
	@echo "  run             : Build and run the binary"
	@echo "  deps            : Get dependencies"
	@echo "  lint            : Run linter"
	@echo "  fmt             : Format code"
	@echo "  vet             : Run go vet"
	@echo "  benchmark       : Run benchmarks"
	@echo "  docker          : Build Docker image"
	@echo "  generate        : Run go generate"
	@echo "  build-linux     : Cross-compile for Linux"
	@echo "  build-windows   : Cross-compile for Windows"
	@echo "  all             : Run fmt, vet, lint, test, coverage, and build"

.PHONY: all build test coverage clean run deps lint fmt vet benchmark docker generate build-linux build-windows help
PROTO_DIR=proto
PROTO_OUT=proto
PROTOC=protoc

# Build configuration
BINARY_NAME=podoai
BUILD_DIR=build
MAIN_PATH=./cmd/main.go

# List all proto files recursively
PROTO_FILES=$(shell find $(PROTO_DIR) -name '*.proto')

.PHONY: help gen clean build run test

# Default target
.DEFAULT_GOAL := help

help:
	@echo "Available targets:"
	@echo "  gen     - Generate Go code from proto files"
	@echo "  build   - Build the binary to build/$(BINARY_NAME)"
	@echo "  run     - Build and run the application"
	@echo "  test    - Run all tests"
	@echo "  clean   - Clean generated files and build directory"
	@echo "  help    - Show this help message"

gen:
	@echo "Generating Go code from proto files..."
	$(PROTOC) \
		--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		-I $(PROTO_DIR) \
		$(PROTO_FILES)

build: gen
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

test:
	@echo "Running tests..."
	go test ./...

clean:
	@echo "Cleaning generated files..."
	find $(PROTO_OUT) -name '*.pb.go' -delete
	@echo "Cleaning build directory..."
	rm -rf $(BUILD_DIR)

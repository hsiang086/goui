BINARY_NAME := goui
BUILD_DIR := ./bin
SRC_DIR := ./engine/game
MAIN_SRC := ./main.go

GO := go
GOFLAGS := -mod=readonly
LDFLAGS := 

.PHONY: all build clean run test

all: build

build:
	@echo "Building the project..."
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_SRC)
	@echo "Build completed."

run: build
	@echo "Running the project..."
	$(BUILD_DIR)/$(BINARY_NAME)

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)/*
	@echo "Cleanup completed."

test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) ./...
	@echo "Tests completed."

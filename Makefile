BINARY_NAME=mnemosyne
BUILD_DIR=bin
MAIN_PACKAGE=cmd/mnemosyne

.PHONY: build run run-dev clean test

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)/main.go

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

run-dev:
	go run $(MAIN_PACKAGE)

clean:
	rm -rf $(BUILD_DIR)

test:
	go test -v ./...

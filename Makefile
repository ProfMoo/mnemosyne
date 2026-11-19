BINARY_NAME=mnemosyne
BUILD_DIR=bin
MAIN_PACKAGE=cmd/mnemosyne

.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)/main.go

.PHONY: run
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: run-dev
run-dev:
	go run ./$(MAIN_PACKAGE)

.PHONY: gen
gen:
	go generate ./...

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY: test
test:
	go test -v ./...

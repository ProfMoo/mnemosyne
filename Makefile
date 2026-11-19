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

.PHONY: lint-api
lint-api:
	npx --yes @redocly/cli@2.80.0 lint
	npx --yes @stoplight/spectral-cli@6.17.1 lint api/openapi.yaml

.PHONY: gen
gen: lint-api
	rm -rf api/gen
	rm api/openapi.bundled.yaml
	go generate ./...

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY: test
test:
	go test -v ./...

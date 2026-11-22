BINARY_NAME=mnemosyne
BUILD_DIR=bin
MAIN_PACKAGE=cmd/mnemosyne

.PHONY: build
build: gen
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)/main.go

.PHONY: run
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: run-dev
run-dev: gen
	go run ./$(MAIN_PACKAGE)

.PHONY: lint-api
lint-api:
	REDOCLY_SUPPRESS_NOTIFICATIONS=true npx --yes @redocly/cli@2.80.0 lint
	npx --yes @stoplight/spectral-cli@6.17.1 lint api/openapi.yaml

.PHONY: lint
lint:
	go tool golangci-lint run ./...

.PHONY: fmt
fmt:
	go tool golangci-lint fmt ./...

.PHONY: fmt-check
fmt-check:
	go tool golangci-lint fmt --diff ./...

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

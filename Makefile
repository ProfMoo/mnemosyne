BINARY_NAME=mnemosyne
BUILD_DIR=bin
MAIN_PACKAGE=cmd/mnemosyne
GOLANGCI_LINT = $(shell go tool -n golangci-lint)
DOCKER_IMAGE ?= mnemosyne
DOCKER_CONTAINER ?= mnemosyne
DOCKER_PORT ?= 8080
DOCKER_VOLUME ?= mnemosyne-data

.PHONY: build
build: gen frontend-build
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)/main.go

.PHONY: run
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: run-dev
run-dev: gen frontend-build
	go run ./$(MAIN_PACKAGE)

.PHONY: lint-api
lint-api:
	REDOCLY_SUPPRESS_NOTIFICATIONS=true npx --yes @redocly/cli@2.80.0 lint
	npx --yes @stoplight/spectral-cli@6.17.1 lint api/openapi.yaml

.PHONY: lint
lint:
	go tool golangci-lint run ./...
	GOOS=js GOARCH=wasm "$(GOLANGCI_LINT)" run ./internal/ui/... ./cmd/frontend/...

.PHONY: fmt
fmt:
	go tool golangci-lint fmt ./...

.PHONY: fmt-check
fmt-check:
	go tool golangci-lint fmt --diff ./...

.PHONY: gen
gen: lint-api
	go generate ./...

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY: test
test:
	go test -v ./...

.PHONY: frontend-build
frontend-build:
	go run ./cmd/frontend-export

.PHONY: docker-start docker-stop
docker-start:
	docker build -t $(DOCKER_IMAGE) .
	docker run --detach --name $(DOCKER_CONTAINER) \
		--publish $(DOCKER_PORT):8080 \
		--mount type=volume,source=$(DOCKER_VOLUME),target=/data \
		$(DOCKER_IMAGE)

docker-stop:
	docker stop $(DOCKER_CONTAINER)
	docker rm $(DOCKER_CONTAINER)

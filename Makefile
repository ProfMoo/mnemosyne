BINARY_NAME=mnemosyne
BUILD_DIR=bin
MAIN_PACKAGE=cmd/mnemosyne
DOCKER_IMAGE ?= mnemosyne
DOCKER_CONTAINER ?= mnemosyne
DOCKER_PORT ?= 8080
DOCKER_VOLUME ?= mnemosyne-data

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

.PHONY: frontend-install frontend-dev frontend-gen frontend-check frontend-build
frontend-install:
	npm --prefix frontend ci

frontend-dev:
	npm --prefix frontend run dev

frontend-gen:
	npm --prefix frontend run generate

frontend-check:
	npm --prefix frontend run generate:check
	npm --prefix frontend run check

frontend-build:
	npm --prefix frontend run build

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

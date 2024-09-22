PROJECT_NAME=url-shortener
SRC_DIR=./cmd/$(PROJECT_NAME)
BINARY_NAME=$(PROJECT_NAME)
BUILD_DIR=./bin
MIGRATIONS_DIR=./migrations

CGO_ENABLED=0
GOARCH=amd64
GOOS=linux

.PHONY: all
all: clean tidy fmt lint test/unit build run

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: mock
mock:
	mockery

.PHONY: test/unit
test/unit:
	go test -cover -race ./internal/... ./pkg/...

.PHONY: test/integration
test/integration:
	go test -cover -race ./tests/integration/...

.PHONY: test/e2e
test/e2e:
	go test -cover -race ./tests/e2e/...

.PHONY: build
build:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)

.PHONY: run
run:
	@if [ ! -f $(BUILD_DIR)/$(BINARY_NAME) ]; then \
		echo "Error: binary '$(BUILD_DIR)/$(BINARY_NAME)' doesn't exist"; \
		exit 1; \
	fi
	$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: clean
clean:
	rm -fr $(BUILD_DIR)
	go clean -testcache

.PHONY: migrations/create
migrations/create:
	ifndef MIGRATION_NAME
		$(error MIGRATION_NAME is not defined)
	endif
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(MIGRATION_NAME)

.PHONY: migrations/up
migrations/up:
	@if [ -z "$(DATABASE_DSN)" ]; then \
		echo "Error: DATABASE_DSN is not defined"; \
		exit 1; \
	fi; \
	migrate -database $(DATABASE_DSN) -path $(MIGRATIONS_DIR) up

.PHONY: migrations/down
migrations/down:
	@if [ -z "$(DATABASE_DSN)" ]; then \
		echo "Error: DATABASE_DSN is not defined"; \
		exit 1; \
	fi; \
	migrate -database $(DATABASE_DSN) -path $(MIGRATIONS_DIR) down -all

.PHONY: ci
ci: tidy fmt build lint test/unit test/integration clean

.PHONY: git/push
git/push: ci
	@if [ -z "$(BRANCH)" ]; then \
		BRANCH=$(shell git rev-parse --abbrev-ref HEAD); \
	fi; \
	git push -u origin $(BRANCH)

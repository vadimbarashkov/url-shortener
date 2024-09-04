PROJECT_NAME=url-shortener
SRC_DIR=./cmd/$(PROJECT_NAME)
BINARY_NAME=$(PROJECT_NAME)
BUILD_DIR=./bin
MIGRATIONS_DIR=./migrations

CGO_ENABLED=0
GOARCH=amd64
GOOS=linux

.PHONY: all
all: clean tidy fmt lint test build

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test -cover ./...

.PHONY: test-race
test-race:
	go test -race -cover ./...

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
	ifndef DATABASE_DSN
		$(error DATABASE_DSN is not defined)
	endif
	migrate -database $(DATABASE_DSN) -path $(MIGRATIONS_DIR) up

.PHONY: migrations/down
migrations/down:
	ifndef DATABASE_DSN
		$(error DATABASE_DSN is not defined)
	endif
	migrate -database $(DATABASE_DSN) -path $(MIGRATIONS_DIR) down -all

.PHONY: ci
ci: tidy fmt build lint test-race clean

.PHONY: git/push
git/push: ci
	@if [ -z "$(BRANCH)" ]; then \
		BRANCH=$(shell git rev-parse --abbrev-ref HEAD); \
	fi; \
	git push -u origin $(BRANCH)

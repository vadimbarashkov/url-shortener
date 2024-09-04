PROJECT_NAME=url-shortener
SRC_DIR=./cmd/${PROJECT_NAME}
BINARY_NAME=${PROJECT_NAME}
BUILD_DIR=./bin
MIGRATIONS_DIR=./migrations

GGO_ENABLED=0
GOARCH=amd64
GOOS=linux

.PHONY: all
all: clean tidy fmt lint test build run

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golint ./...
	staticcheck ./...

.PHONY: test
test:
	go test -cover ./...

.PHONY: build
build:
	mkdir $(BUILD_DIR)
	CGO_ENABLED=${CGO_ENABLED} GOARCH=${GOARCH} GOOS=${GOOS} go build -o ${BUILD_DIR}/${BINARY_NAME} ${SRC_DIR}

.PHONY: run
run:
	${BUILD_DIR}/${BINARY_NAME}

.PHONY: clean
clean:
	rm -fr ${BUILD_DIR}
	go clean -testcache

.PHONY: migrations/create
migrations/create:
	migrate create -ext sql -dir ${MIGRATIONS_DIR} -seq ${MIGRATION_NAME}

.PHONY: migrations/up
migrations/up:
	migrate -database ${DATABASE_DSN} -path ${MIGRATIONS_DIR} up

.PHONY: migrations/down
migrations/down:
	migrate -database ${DATABASE_DSN} -path ${MIGRATIONS_DIR} down -all

API_PATH := bin/api
CLI_PATH := bin/cli

.PHONY: build
build:
	go build -o ${API_PATH} cmd/recite/main.go
	go build -o ${CLI_PATH} cmd/cli/main.go

.PHONY: run
run:
	${API_PATH}

.PHONY: api
api: build run

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	go fmt ./...
	go vet ./...
	golangci-lint run

.PHONY: clean
clean:
	go mod verify
	go mod download
	go mod tidy

.PHONY: ci
ci: clean lint test

.DEFAULT_GOAL := build

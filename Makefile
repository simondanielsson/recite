.PHONY: build
build:
	go build -o recite cmd/recite/main.go
	go build -o cli cmd/cli/main.go

.PHONY: run-api
run-api:
	./recite

.PHONY: api
api: build run-api

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

.DEFAULT_GOAL := cli

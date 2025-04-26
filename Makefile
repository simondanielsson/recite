.PHONY: build
build:
	go build ./...

.PHONY: run
run:
	./recite

.PHONY: build-and-run
build-and-run: build run

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	go fmt ./...
	go vet ./...
	golangci-lint run

.DEFAULT_GOAL := build-and-run

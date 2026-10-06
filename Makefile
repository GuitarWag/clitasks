BIN := bin/tasks
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build install run test lint cover tidy clean

build:
	go build -ldflags '$(LDFLAGS)' -o $(BIN) ./cmd/tasks

install:
	go install -ldflags '$(LDFLAGS)' ./cmd/tasks

run:
	go run ./cmd/tasks $(ARGS)

test:
	go test ./... -race -count=1

cover:
	go test ./... -coverprofile=cover.out && go tool cover -html=cover.out

lint:
	golangci-lint run

tidy:
	go mod tidy

clean:
	rm -rf bin/ cover.out

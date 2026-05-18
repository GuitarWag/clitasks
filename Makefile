BIN := bin/tasks
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build install run test lint cover tidy sync-skill clean

sync-skill:
	cp SKILL.md internal/cli/SKILL.md

build: sync-skill
	go build -ldflags '$(LDFLAGS)' -o $(BIN) ./cmd/tasks

install: sync-skill
	go install -ldflags '$(LDFLAGS)' ./cmd/tasks

run: sync-skill
	go run ./cmd/tasks $(ARGS)

test: sync-skill
	go test ./... -race -count=1

cover: sync-skill
	go test ./... -coverprofile=cover.out && go tool cover -html=cover.out

lint:
	golangci-lint run

tidy:
	go mod tidy

clean:
	rm -rf bin/ cover.out internal/cli/SKILL.md

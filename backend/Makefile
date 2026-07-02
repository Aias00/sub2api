.PHONY: build generate test test-unit test-integration test-e2e lint-fmt lint-check

VERSION ?= $(shell ./scripts/resolve-version.sh)
LDFLAGS ?= -s -w -X main.Version=$(VERSION)
GOLANGCI_LINT ?= golangci-lint

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -trimpath -o bin/server ./cmd/server

generate:
	go generate ./ent
	go generate ./cmd/server

test:
	go test ./...
	$(MAKE) lint-check

test-unit:
	go test -tags=unit ./...

test-integration:
	go test -tags=integration ./...

test-e2e:
	./scripts/e2e-test.sh

test-e2e-local:
	go test -tags=e2e -v -timeout=300s ./internal/integration/...

lint-fmt:
	$(GOLANGCI_LINT) fmt

lint-check:
	$(GOLANGCI_LINT) run --path-mode=abs --timeout=30m

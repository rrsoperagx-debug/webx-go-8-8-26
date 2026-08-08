
.PHONY: all build test lint fmt audit bench clean dev release docker

BIN := webx-metrics-pro
PORT ?= 8080
GO := go
LDFLAGS := -s -w -X main.version=2.2.1 -X main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo dev)

all: fmt lint test build

dev:
	$(GO) run ./cmd/server --dev

build:
	CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o bin/$(BIN) ./cmd/server

test:
	$(GO) test ./... -v -race -cover

lint:
	$(GO) vet ./...
	golangci-lint run ./... || true

fmt:
	$(GO) fmt ./...
	@echo "fmt ok"

audit:
	$(GO) list -json -m all | nancy sleuth || true
	govulncheck ./... || true
	@echo "Checking weak crypto (openssl banned, stdlib only)..."
	@! grep -r "openssl" --include="*.go" . || echo "WARN: openssl found"

bench:
	$(GO) test -bench=. -benchmem ./internal/metrics

clean:
	rm -rf bin/ data/*.db data/*.db-journal tmp/
	$(GO) clean -cache

release:
	@echo "Building multi-arch..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-linux-arm64 ./cmd/server
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-darwin-amd64 ./cmd/server
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-darwin-arm64 ./cmd/server
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-windows-amd64.exe ./cmd/server
	GOOS=linux GOARCH=arm64 GOOS=android CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o dist/$(BIN)-android-arm64 ./cmd/server || echo "android skip"

docker:
	docker build -t ghcr.io/webx/metrics-pro:go-latest .

docker-multi:
	docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/webx/metrics-pro:go-latest .

migrate:
	$(GO) run ./cmd/server --migrate-only

install-tools:
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install golang.org/x/vuln/cmd/govulncheck@latest
	$(GO) install github.com/sonatype-nexus-community/nancy@latest

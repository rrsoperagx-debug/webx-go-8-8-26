#!/usr/bin/env bash
set -e
echo "=== WebX Metrics Pro Go Setup ==="
if ! command -v go &> /dev/null; then
  echo "Go not found, install https://go.dev/dl/"
  exit 1
fi
echo "Go $(go version)"
mkdir -p data bin
go mod tidy
go build -o bin/webx-metrics-pro ./cmd/server
echo "✅ Build OK bin/webx-metrics-pro"
echo "Run: ./bin/webx-metrics-pro --dev"

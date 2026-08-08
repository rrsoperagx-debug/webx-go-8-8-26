
# WebX Metrics Pro Go v2.2.1 LTS - Go + Chi + SQLite pure + SRE Hardening

Przepisana wersja Rust -> Go 1.22. 1 binarka ~7MB, CGO_ENABLED=0, działa wszędzie: watch 320px -> TV 8K, 30 przeglądarek.

**Stack:** Go 1.22, Chi v5, Zerolog JSON, Viper, modernc.org/sqlite (pure Go, bez CGO), Prometheus client_golang, gofpdf, JWT v5, bcrypt, distroless static.

**Patch v2.2.1.1:** Backpressure 512 semaphore -> 503, Timeout 5s, in-flight gauge, CI bench, Helm tmpfs 256Mi DLP, SLO p99<10ms, PrometheusRule, Runbook, ADR-014 PQC Q4 2027.

## Szybki start Linux (Twój przypadek)

```bash
unzip webx-metrics-pro-go.zip && cd webx-metrics-pro-go
chmod +x setup.sh && ./setup.sh
./bin/webx-metrics-pro --dev
# lub go run ./cmd/server --dev

curl http://localhost:8080/health
curl http://localhost:8080/api/metrics/all
curl -X POST http://localhost:8080/api/metrics/ingest -H "Content-Type: application/json" -d '{"name":"cpu","value":42.5}'
# http://localhost:8080/
```

## GitHub Codespaces 1-klik

```bash
./scripts/init_github.sh https://github.com/YOU/webx-metrics-pro-go.git
# GitHub -> Code -> Codespaces -> Create codespace on main
# postCreateCommand: go mod tidy + golangci-lint + govulncheck
# go run ./cmd/server --dev -> Open in Browser na 8080
```

Devcontainer: `mcr.microsoft.com/devcontainers/go:1-1.22-bullseye` + docker-in-docker + gh-cli

## API

- GET / -> dashboard html/template responsive
- GET /health, /healthz
- GET /metrics - prometheus
- GET /api/metrics/all, /list
- POST /api/metrics/ingest {name, value, labels}
- POST /api/export/pdf|md|txt|json

## Build

```bash
make all      # fmt + vet + test + build
make test     # race + cover
make bench    # bench hot path
make release  # 6 platform: linux amd64/arm64, darwin amd64/arm64, windows, android
make docker   # distroless static nonroot
```

CI: ci.yml (fmt, vet, lint, test, vulncheck, build), perf.yml, release.yml (cross + docker multi-arch ghcr.io)

## Bezpieczeństwo Go

- SQLite pure Go WAL, foreign_keys, tmpfs /tmp 256Mi DLP
- CSP nosniff DENY, CORS locked prod, BodyLimit 5MB, Timeout 5s, Recoverer
- Bcrypt + JWT HS256, zerolog JSON
- govulncheck, nancy, no openssl (stdlib only)
- PQC ADR-014 Q4 2027 ML-KEM/ML-DSA

## Kompatybilność

9 urządzeń: Watch, Mobile, Tablet, Laptop, 4K, TV, Auto - CSS grid media queries
30 przeglądarek: 100% bo HTML5+CSS3 bez JS frameworków

## Deploy K8s

```bash
kubectl apply -f k8s/prometheus-rules.yaml
helm upgrade --install webx-go ./deploy --values deploy/values.yaml
```

SLO: p99<10ms, 99.9% uptime, <0.1% 5xx, in-flight <450 warn

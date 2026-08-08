
# Build
FROM golang:1.22-bullseye AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /app/webx-metrics-pro ./cmd/server

# Runtime distroless
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /app/webx-metrics-pro /app/webx-metrics-pro
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/config /app/config
COPY --from=builder /app/migrations /app/migrations
ENV PORT=8080
ENV APP_ENV=production
ENV DB_PATH=/tmp/metrics.db
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/webx-metrics-pro"]

# SLO v2.2.1 Go
| SLI | Cel | Okno |
|-----|-----|------|
| Availability | 99.9% | 30d |
| Latency p99 | <10ms | 5m |
| Error Rate | <0.1% | 5m |
| In-Flight | <450 warn, 512 hard |
Implementacja: prometheus histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))
Backpressure: semaphore 512 -> 503

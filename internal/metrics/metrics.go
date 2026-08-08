
package metrics

import (
	"math/rand"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	Registry = prometheus.NewRegistry()

	RequestsTotal = promauto.With(Registry).NewCounter(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests",
	})

	RequestDuration = promauto.With(Registry).NewHistogram(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration",
		Buckets: prometheus.DefBuckets,
	})

	InFlight = promauto.With(Registry).NewGauge(prometheus.GaugeOpts{
		Name: "http_in_flight_requests",
		Help: "Current in-flight requests",
	})

	CPUUsage = promauto.With(Registry).NewGauge(prometheus.GaugeOpts{
		Name: "cpu_usage_percent",
		Help: "CPU usage %",
	})

	MemUsage = promauto.With(Registry).NewGauge(prometheus.GaugeOpts{
		Name: "mem_usage_mb",
		Help: "Memory usage MB",
	})

	ActiveUsers = promauto.With(Registry).NewGauge(prometheus.GaugeOpts{
		Name: "active_users",
		Help: "Active users",
	})

	ErrorsTotal = promauto.With(Registry).NewCounter(prometheus.CounterOpts{
		Name: "errors_total",
		Help: "Total errors",
	})

	MetricsIngested = promauto.With(Registry).NewCounter(prometheus.CounterOpts{
		Name: "metrics_ingested_total",
		Help: "Metrics ingested",
	})

	startTime = time.Now()
)

func Init() {
	// Background fake metrics for demo (replace with gopsutil in prod)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			CPUUsage.Set(rand.Float64()*30 + 10)
			MemUsage.Set(rand.Float64()*200 + 100)
			ActiveUsers.Set(rand.Float64()*500 + 50)
		}
	}()
}

func Uptime() time.Duration {
	return time.Since(startTime)
}

func UptimeSeconds() float64 {
	return time.Since(startTime).Seconds()
}


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

	// Tracked values for reading metrics
	cpuValue    float64
	memValue    float64
	usersValue  float64
	requestsVal float64
	errorsVal   float64
	metricsVal  float64
	inflightVal float64
)

func Init() {
	// Background fake metrics for demo (replace with gopsutil in prod)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			cpuValue = rand.Float64()*30 + 10
			CPUUsage.Set(cpuValue)
			
			memValue = rand.Float64()*200 + 100
			MemUsage.Set(memValue)
			
			usersValue = rand.Float64()*500 + 50
			ActiveUsers.Set(usersValue)
		}
	}()
}

func Uptime() time.Duration {
	return time.Since(startTime)
}

func UptimeSeconds() float64 {
	return time.Since(startTime).Seconds()
}

// Helper functions to get metric values
func GetCPUUsage() float64 {
	return cpuValue
}

func GetMemUsage() float64 {
	return memValue
}

func GetActiveUsers() float64 {
	return usersValue
}

func GetRequestsTotal() float64 {
	return requestsVal
}

func GetErrorsTotal() float64 {
	return errorsVal
}

func GetMetricsIngested() float64 {
	return metricsVal
}

func GetInFlight() float64 {
	return inflightVal
}

func SetRequestsTotal(v float64) {
	requestsVal = v
}

func SetErrorsTotal(v float64) {
	errorsVal = v
}

func SetMetricsIngested(v float64) {
	metricsVal = v
}

func SetInFlight(v float64) {
	inflightVal = v
}

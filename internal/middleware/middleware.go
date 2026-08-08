
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"

	"github.com/webx/metrics-pro/internal/config"
	"github.com/webx/metrics-pro/internal/metrics"
)

// Concurrency limiter - backpressure 512
type ConcurrencyLimiter struct {
	sem chan struct{}
}

func NewConcurrencyLimiter(limit int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		sem: make(chan struct{}, limit),
	}
}

func (c *ConcurrencyLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case c.sem <- struct{}{}:
			defer func() { <-c.sem }()
			next.ServeHTTP(w, r)
		default:
			// 503 when full - backpressure
			metrics.ErrorsTotal.Inc()
			w.Header().Set("Retry-After", "1")
			http.Error(w, "Service Unavailable - too many concurrent requests (limit 512)", http.StatusServiceUnavailable)
		}
	})
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; script-src 'self' 'unsafe-inline'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		metrics.InFlight.Inc()
		metrics.RequestsTotal.Inc()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			metrics.InFlight.Dec()
			duration := time.Since(start).Seconds()
			metrics.RequestDuration.Observe(duration)

			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", ww.Status()).
				Dur("latency", time.Since(start)).
				Msg("request")
		}()

		next.ServeHTTP(ww, r)
	})
}

func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			select {
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					w.WriteHeader(http.StatusGatewayTimeout)
					_, _ = w.Write([]byte("Request timeout after 5s"))
				}
				return
			case <-done:
				return
			}
		})
	}
}

func BodyLimitMiddleware(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

func CorsMiddleware(cfg *config.Config) *cors.Cors {
	allowedOrigins := []string{"*"}
	if cfg.Security.CorsOrigin != "*" {
		allowedOrigins = []string{cfg.Security.CorsOrigin}
	}
	return cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	})
}

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				log.Error().Interface("panic", rvr).Msg("panic recovered")
				metrics.ErrorsTotal.Inc()
				http.Error(w, fmt.Sprintf("Internal Server Error: %v", rvr), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

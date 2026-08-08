
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/webx/metrics-pro/internal/config"
	"github.com/webx/metrics-pro/internal/db"
	"github.com/webx/metrics-pro/internal/handlers"
	mw "github.com/webx/metrics-pro/internal/middleware"
	"github.com/webx/metrics-pro/internal/metrics"
)

var (
	version = "2.2.1"
	commit  = "dev"
)

func main() {
	// flags
	dev := flag.Bool("dev", false, "dev mode")
	healthcheck := flag.Bool("healthcheck", false, "healthcheck")
	migrateOnly := flag.Bool("migrate-only", false, "run migrations only")
	flag.Parse()

	if *healthcheck {
		fmt.Println("OK")
		os.Exit(0)
	}

	// logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if os.Getenv("LOG_LEVEL") == "" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	if *dev {
		cfg.App.Env = "development"
		log.Info().Msg("Running in DEV mode")
	}

	log.Info().
		Str("version", cfg.App.Version).
		Str("env", cfg.App.Env).
		Int("port", cfg.Server.Port).
		Str("db", cfg.DB.Path).
		Str("commit", commit).
		Msg("Starting WebX Metrics Pro Go")

	// metrics
	metrics.Init()

	// db
	database, err := db.Init(cfg.DB.Path)
	if err != nil {
		log.Fatal().Err(err).Msg("DB init failed")
	}
	defer database.Close()

	if *migrateOnly {
		log.Info().Msg("Migrations done")
		os.Exit(0)
	}

	// handlers
	h := handlers.New(database, cfg)

	// router
	r := chi.NewRouter()

	// global middleware
	limiter := mw.NewConcurrencyLimiter(cfg.Server.ConcurrencyLimit)
	r.Use(limiter.Middleware)
	r.Use(mw.Recoverer)
	r.Use(mw.SecurityHeaders)
	r.Use(mw.MetricsMiddleware)
	r.Use(mw.TimeoutMiddleware(time.Duration(cfg.Server.TimeoutSecs) * time.Second))
	r.Use(mw.BodyLimitMiddleware(int64(cfg.Server.BodyLimitMB * 1024 * 1024)))

	// CORS
	corsMw := mw.CorsMiddleware(cfg)
	r.Use(corsMw.Handler)

	// routes
	r.Get("/", h.Dashboard)
	r.Get("/health", h.Health)
	r.Get("/healthz", h.Health)
	r.Get("/api/metrics/all", h.AllMetrics)
	r.Get("/api/metrics/list", h.ListMetrics)
	r.Post("/api/metrics/ingest", h.IngestMetric)
	r.Post("/api/export/pdf", h.ExportPDF)
	r.Post("/api/export/md", h.ExportMD)
	r.Post("/api/export/txt", h.ExportTXT)
	r.Post("/api/export/json", h.ExportJSON)
	r.Post("/api/auth/login", h.Login)

	// Prometheus metrics
	r.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}))

	// pprof optional
	if cfg.Features.EnablePprof {
		r.Mount("/debug", chi.NewRouter())
	}

	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSecs) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSecs) * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// start
	go func() {
		log.Info().Str("addr", srv.Addr).Msg("Listening")
		if *dev || cfg.IsDev() {
			fmt.Printf("\n🚀 WebX Metrics Pro Go v%s (%s) dev\n", cfg.App.Version, commit)
			fmt.Printf("   Local:   http://localhost:%d/\n", cfg.Server.Port)
			fmt.Printf("   Health:  http://localhost:%d/health\n", cfg.Server.Port)
			fmt.Printf("   Metrics: http://localhost:%d/metrics\n", cfg.Server.Port)
			fmt.Printf("   API:     http://localhost:%d/api/metrics/all\n\n", cfg.Server.Port)
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Shutdown error")
	}
	log.Info().Msg("Bye")
}

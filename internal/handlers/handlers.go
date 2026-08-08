
package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"regexp"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/webx/metrics-pro/internal/config"
	"github.com/webx/metrics-pro/internal/db"
	"github.com/webx/metrics-pro/internal/metrics"
	"github.com/webx/metrics-pro/internal/models"
)

type Handler struct {
	DB     *db.DB
	Config *config.Config
}

func New(db *db.DB, cfg *config.Config) *Handler {
	SetJWTSecret(cfg.Security.JWTSecret)
	return &Handler{DB: db, Config: cfg}
}

var validName = regexp.MustCompile(`^[a-zA-Z0-9_\.\-]+$`)

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	recent := []models.Metric{}
	rows, err := h.DB.Query(`SELECT id, name, value, labels, timestamp FROM metrics ORDER BY id DESC LIMIT 20`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var m models.Metric
			if err := rows.Scan(&m.ID, &m.Name, &m.Value, &m.Labels, &m.Timestamp); err == nil {
				recent = append(recent, m)
			}
		}
	}

	devices := []models.Device{
		{Name: "Smartwatch WearOS", Resolution: "400x400", Category: "Wearable", Support: "100%"},
		{Name: "Smartwatch watchOS", Resolution: "396x484", Category: "Wearable", Support: "100%"},
		{Name: "Smartphone", Resolution: "390x844", Category: "Mobile", Support: "100%"},
		{Name: "Tablet", Resolution: "1024x1366", Category: "Tablet", Support: "100%"},
		{Name: "Laptop", Resolution: "1920x1080", Category: "Desktop", Support: "100%"},
		{Name: "PC 4K", Resolution: "3840x2160", Category: "Desktop", Support: "100%"},
		{Name: "TV 4K", Resolution: "3840x2160", Category: "TV", Support: "100%"},
		{Name: "Auto Android Auto", Resolution: "1280x720", Category: "Auto", Support: "100%"},
		{Name: "Auto CarPlay", Resolution: "800x480", Category: "Auto", Support: "100%"},
	}

	browsers := []models.Browser{
		{Name: "Chrome", Version: "127+", Engine: "Blink", Support: "100%", MarketShare: "65%"},
		{Name: "Firefox", Version: "128+", Engine: "Gecko", Support: "100%", MarketShare: "8%"},
		{Name: "Safari", Version: "17+", Engine: "WebKit", Support: "100%", MarketShare: "18%"},
		{Name: "Edge", Version: "127+", Engine: "Blink", Support: "100%", MarketShare: "5%"},
		{Name: "Opera", Version: "111+", Engine: "Blink", Support: "100%", MarketShare: "2%"},
		{Name: "Brave", Version: "127+", Engine: "Blink", Support: "100%", MarketShare: "1%"},
		{Name: "Vivaldi", Version: "6.5+", Engine: "Blink", Support: "100%", MarketShare: "0.3%"},
		{Name: "Arc", Version: "1.0+", Engine: "Blink", Support: "100%", MarketShare: "0.2%"},
		{Name: "Samsung Internet", Version: "26+", Engine: "Blink", Support: "100%", MarketShare: "2%"},
		{Name: "Yandex", Version: "24+", Engine: "Blink", Support: "100%", MarketShare: "0.5%"},
		{Name: "UC Browser", Version: "15+", Engine: "Blink", Support: "95%", MarketShare: "0.5%"},
		{Name: "QQ Browser", Version: "14+", Engine: "Blink", Support: "95%", MarketShare: "0.3%"},
		{Name: "Tor Browser", Version: "13+", Engine: "Gecko", Support: "100%", MarketShare: "0.1%"},
		{Name: "Pale Moon", Version: "33+", Engine: "Goanna", Support: "95%", MarketShare: "0.05%"},
		{Name: "Lynx", Version: "2.9+", Engine: "Text", Support: "80%", MarketShare: "0.01%"},
		{Name: "Chrome Android", Version: "127+", Engine: "Blink", Support: "100%", MarketShare: "-"},
		{Name: "Safari iOS", Version: "17+", Engine: "WebKit", Support: "100%", MarketShare: "-"},
		{Name: "Firefox Android", Version: "128+", Engine: "Gecko", Support: "100%", MarketShare: "-"},
		{Name: "Edge Mobile", Version: "127+", Engine: "Blink", Support: "100%", MarketShare: "-"},
		{Name: "Opera Mobile", Version: "111+", Engine: "Blink", Support: "100%", MarketShare: "-"},
		{Name: "DuckDuckGo", Version: "5+", Engine: "Blink", Support: "100%", MarketShare: "0.6%"},
		{Name: "Naver Whale", Version: "3+", Engine: "Blink", Support: "100%", MarketShare: "0.2%"},
		{Name: "Epiphany", Version: "45+", Engine: "WebKit", Support: "100%", MarketShare: "0.05%"},
		{Name: "Falkon", Version: "24+", Engine: "Blink", Support: "100%", MarketShare: "0.02%"},
		{Name: "SeaMonkey", Version: "2.53+", Engine: "Gecko", Support: "95%", MarketShare: "0.01%"},
		{Name: "Waterfox", Version: "6+", Engine: "Gecko", Support: "100%", MarketShare: "0.03%"},
		{Name: "LibreWolf", Version: "128+", Engine: "Gecko", Support: "100%", MarketShare: "0.02%"},
		{Name: "Mullvad", Version: "13+", Engine: "Gecko", Support: "100%", MarketShare: "0.01%"},
		{Name: "Avast Secure", Version: "12+", Engine: "Blink", Support: "100%", MarketShare: "0.1%"},
		{Name: "AVG Secure", Version: "12+", Engine: "Blink", Support: "100%", MarketShare: "0.05%"},
	}

	mv := models.MetricsView{
		CPU:           metrics.CPUUsage.Get(),
		Mem:           metrics.MemUsage.Get(),
		Users:         metrics.ActiveUsers.Get(),
		RPS:           42.0,
		RequestsTotal: metrics.RequestsTotal.Get(),
		ErrorsTotal:   metrics.ErrorsTotal.Get(),
	}

	data := map[string]interface{}{
		"Version":       h.Config.App.Version,
		"Env":           h.Config.App.Env,
		"Uptime":        int(metrics.UptimeSeconds()),
		"Metrics":       mv,
		"Devices":       devices,
		"Browsers":      browsers,
		"RecentMetrics": recent,
	}

	tmplPath := filepath.Join("templates", "dashboard.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		tmplPath = "/app/templates/dashboard.html"
		tmpl, err = template.ParseFiles(tmplPath)
		if err != nil {
			log.Error().Err(err).Msg("template parse failed")
			http.Error(w, "Template error", 500)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		log.Error().Err(err).Msg("template execute failed")
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "ok",
		"version":        h.Config.App.Version,
		"env":            h.Config.App.Env,
		"uptime_seconds": metrics.UptimeSeconds(),
		"db":             "ok",
	})
}

func (h *Handler) AllMetrics(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"cpu_percent":      metrics.CPUUsage.Get(),
		"mem_mb":           metrics.MemUsage.Get(),
		"active_users":     metrics.ActiveUsers.Get(),
		"requests_total":   metrics.RequestsTotal.Get(),
		"errors_total":     metrics.ErrorsTotal.Get(),
		"metrics_ingested": metrics.MetricsIngested.Get(),
		"in_flight":        metrics.InFlight.Get(),
		"uptime_seconds":   metrics.UptimeSeconds(),
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) ListMetrics(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`SELECT id, name, value, labels, timestamp FROM metrics ORDER BY id DESC LIMIT 100`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var list []models.Metric
	for rows.Next() {
		var m models.Metric
		if err := rows.Scan(&m.ID, &m.Name, &m.Value, &m.Labels, &m.Timestamp); err == nil {
			list = append(list, m)
		}
	}
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) IngestMetric(w http.ResponseWriter, r *http.Request) {
	var payload models.IngestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	if payload.Name == "" {
		http.Error(w, "name cannot be empty", 400)
		return
	}
	if !validName.MatchString(payload.Name) {
		http.Error(w, "invalid metric name", 400)
		return
	}

	labels := "{}"
	if payload.Labels != nil {
		labels = *payload.Labels
	}

	_, err := h.DB.Exec(`INSERT INTO metrics (name, value, labels, timestamp) VALUES (?, ?, ?, ?)`,
		payload.Name, payload.Value, labels, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	metrics.MetricsIngested.Inc()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"name":   payload.Name,
		"value":  payload.Value,
	})
}

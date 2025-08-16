package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterHealthHandler registers a health endpoint on the provided router.
// @Summary Health check
// @Description Returns basic health information (status, time, uptime)
// @Tags monitoring
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Router /healthz [get]
func RegisterHealthHandler(r chi.Router, startTime time.Time) {
	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		resp := map[string]interface{}{
			"status":         "ok",
			"time_unix":      time.Now().Unix(),
			"uptime_seconds": int(time.Since(startTime).Seconds()),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}

// NewPrometheusRegistry creates a new prometheus.Registry and registers
// standard Go and Process collectors. This avoids using the global registry.
func NewPrometheusRegistry() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	// Register standard collectors (process + go)
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	reg.MustRegister(collectors.NewGoCollector())
	return reg
}

// RegisterMetricsHandler registers a Prometheus metrics endpoint using the provided registry.
// Uses promhttp.HandlerFor to avoid global handler usage.
// @Summary Metrics
// @Description Exposes Prometheus metrics
// @Tags monitoring
// @Produce  text/plain
// @Success 200 {string} string "Prometheus metrics"
// @Router /metrics [get]
func RegisterMetricsHandler(r chi.Router, registry *prometheus.Registry) {
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	r.Handle("/metrics", h)
}

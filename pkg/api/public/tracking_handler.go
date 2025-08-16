// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package public

import (
	"encoding/base64"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/headmail/headmail/pkg/config"
	"github.com/headmail/headmail/pkg/service"
)

// Default 1x1 transparent PNG
var transparentPNG, _ = base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR4nGNgYAAAAAMAAWgmWQ0AAAAASUVORK5CYII=")

// TrackingHandler handles HTTP requests for deliveries.
type TrackingHandler struct {
	cfg     *config.TrackingConfig
	service service.TrackingServiceProvider
}

// NewTrackingHandler creates a new TrackingHandler.
func NewTrackingHandler(cfg *config.TrackingConfig, service service.TrackingServiceProvider) *TrackingHandler {
	return &TrackingHandler{
		cfg:     cfg,
		service: service,
	}
}

// RegisterRoutes registers tracking routes on the provided router.
func (h *TrackingHandler) RegisterRoutes(r chi.Router) {
	r.Get("/r/{deliveryID}/o", h.openHandler)
	r.Get("/r/{deliveryID}/c", h.clickHandler)
}

// @Summary Track open (1x1 pixel)
// @Description Records an open event for a delivery and returns a 1x1 transparent PNG (or configured image).
// @Tags tracking
// @Param deliveryID path string true "Delivery ID"
// @Produce image/png
// @Success 200 {file} binary image
// @Failure 400 {object} map[string]string
// @Router /r/{deliveryID}/o [get]
func (h *TrackingHandler) openHandler(w http.ResponseWriter, r *http.Request) {
	deliveryID := chi.URLParam(r, "deliveryID")
	if deliveryID == "" {
		http.Error(w, "missing delivery id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ua := r.UserAgent()
	ip := extractRemoteIP(r)

	if err := h.service.LogOpenEvent(ctx, deliveryID, &ua, &ip); err != nil {
		log.Printf("log open event failed: %+v", err)
	}

	// Serve image
	if h.cfg.ImagePath != "" {
		// If ImagePath is a URL (starts with http/https) just redirect to it
		if strings.HasPrefix(h.cfg.ImagePath, "http://") || strings.HasPrefix(h.cfg.ImagePath, "https://") {
			http.Redirect(w, r, h.cfg.ImagePath, http.StatusFound)
			return
		}
		// Try to serve local file
		if f, err := os.Open(h.cfg.ImagePath); err == nil {
			defer f.Close()
			stat, err := f.Stat()
			if err == nil {
				http.ServeContent(w, r, filepath.Base(h.cfg.ImagePath), stat.ModTime(), f)
				return
			}
		}
		// fallback to default if file not readable
	}

	// default: return 1x1 transparent PNG (no-cache)
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(transparentPNG)
}

// @Summary Track click and redirect
// @Description Records a click event and redirects to the original URL.
// @Tags tracking
// @Param deliveryID path string true "Delivery ID"
// @Param u query string true "URL encoded target"
// @Success 302 "Redirect"
// @Failure 400 {object} map[string]string
// @Router /r/{deliveryID}/c [get]
func (h *TrackingHandler) clickHandler(w http.ResponseWriter, r *http.Request) {
	deliveryID := chi.URLParam(r, "deliveryID")
	if deliveryID == "" {
		http.Error(w, "missing delivery id", http.StatusBadRequest)
		return
	}
	u := r.URL.Query().Get("u")
	if u == "" {
		http.Error(w, "missing url param 'u'", http.StatusBadRequest)
		return
	}
	decoded, err := url.QueryUnescape(u)
	if err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}
	target, err := url.Parse(decoded)
	if err != nil || !isAllowedScheme(target.Scheme) {
		http.Error(w, "invalid or unsupported url scheme", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ua := r.UserAgent()
	ip := extractRemoteIP(r)

	if err := h.service.LogClickEvent(ctx, deliveryID, &ua, &ip, decoded); err != nil {
		log.Printf("log click event failed: %+v", err)
	}

	// Redirect
	http.Redirect(w, r, decoded, http.StatusFound)
}

func isAllowedScheme(s string) bool {
	ls := strings.ToLower(s)
	return ls == "http" || ls == "https"
}

// extractRemoteIP extracts IP from X-Forwarded-For or RemoteAddr.
func extractRemoteIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		ip := strings.TrimSpace(parts[0])
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

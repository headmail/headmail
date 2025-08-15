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
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/headmail/headmail/pkg/config"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
)

// Default 1x1 transparent PNG
var transparentPNG, _ = base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR4nGNgYAAAAAMAAWgmWQ0AAAAASUVORK5CYII=")

// RegisterRoutes registers tracking routes on the provided router.
func RegisterRoutes(r chi.Router, db repository.DB, cfg *config.Config) {
	r.Get("/r/{delivery_id}/o", openHandler(db, cfg))
	r.Get("/r/{delivery_id}/c", clickHandler(db, cfg))
}

func openHandler(db repository.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deliveryID := chi.URLParam(r, "delivery_id")
		if deliveryID == "" {
			http.Error(w, "missing delivery id", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		now := time.Now().Unix()

		ua := r.UserAgent()
		ip := extractRemoteIP(r)

		ev := &domain.DeliveryEvent{
			ID:         "", // repo may set ID, but domain expects it; repository implementation can fill or DB auto-generate
			DeliveryID: deliveryID,
			EventType:  domain.EventTypeOpened,
			EventData:  map[string]interface{}{},
			UserAgent:  &ua,
			IPAddress:  &ip,
			CreatedAt:  now,
		}

		// Atomically increment open count (repository will set OpenedAt on first open)
		if err := db.DeliveryRepository().IncrementCount(ctx, deliveryID, domain.EventTypeOpened); err != nil {
			log.Printf("tracking: failed to increment open count for %s: %v", deliveryID, err)
		}

		// store event synchronously
		_ = db.EventRepository().Create(ctx, ev)

		// Serve image
		if cfg != nil && cfg.Tracking.ImagePath != "" {
			// If ImagePath is a URL (starts with http/https) just redirect to it
			if strings.HasPrefix(cfg.Tracking.ImagePath, "http://") || strings.HasPrefix(cfg.Tracking.ImagePath, "https://") {
				http.Redirect(w, r, cfg.Tracking.ImagePath, http.StatusFound)
				return
			}
			// Try to serve local file
			if f, err := os.Open(cfg.Tracking.ImagePath); err == nil {
				defer f.Close()
				stat, err := f.Stat()
				if err == nil {
					http.ServeContent(w, r, filepath.Base(cfg.Tracking.ImagePath), stat.ModTime(), f)
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
}

func clickHandler(db repository.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deliveryID := chi.URLParam(r, "delivery_id")
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
		now := time.Now().Unix()
		ua := r.UserAgent()
		ip := extractRemoteIP(r)

		ev := &domain.DeliveryEvent{
			ID:         "",
			DeliveryID: deliveryID,
			EventType:  domain.EventTypeClicked,
			EventData:  map[string]interface{}{"url": decoded},
			UserAgent:  &ua,
			IPAddress:  &ip,
			URL:        &decoded,
			CreatedAt:  now,
		}

		// Atomically increment click count
		if err := db.DeliveryRepository().IncrementCount(ctx, deliveryID, domain.EventTypeClicked); err != nil {
			log.Printf("tracking: failed to increment click count for %s: %v", deliveryID, err)
		}

		// record click synchronously
		_ = db.EventRepository().Create(ctx, ev)

		// Redirect
		http.Redirect(w, r, decoded, http.StatusFound)
	}
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

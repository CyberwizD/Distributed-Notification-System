package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/CyberwizD/Distributed-Notification-System/services/push_service/pkg/metrics"
)

// NewRouter wires lightweight health/metrics endpoints so the service can be monitored.
func NewRouter(metrics *metrics.Metrics, started time.Time) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "push service healthy",
			"meta": map[string]interface{}{
				"uptime_seconds": int(time.Since(started).Seconds()),
				"timestamp":      time.Now().UTC(),
			},
		})
	})
	mux.Handle("/metrics", metrics.Handler())
	return mux
}

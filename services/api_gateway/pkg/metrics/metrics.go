package metrics

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// Collector tracks a few basic gateway metrics without external deps.
type Collector struct {
	totalRequests   atomic.Int64
	failedRequests  atomic.Int64
	totalLatencyMic atomic.Int64
	startedAt       time.Time
}

func New() *Collector {
	return &Collector{
		startedAt: time.Now(),
	}
}

// GinMiddleware records request count, failures, and aggregate latency.
func (c *Collector) GinMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()

		c.totalRequests.Add(1)
		if ctx.Writer.Status() >= http.StatusInternalServerError {
			c.failedRequests.Add(1)
		}
		c.totalLatencyMic.Add(time.Since(start).Microseconds())
	}
}

// Handler exposes the metrics in a simple JSON form.
func (c *Collector) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqs := c.totalRequests.Load()
		latency := c.totalLatencyMic.Load()
		var avgMicros int64
		if reqs > 0 {
			avgMicros = latency / reqs
		}

		payload := map[string]interface{}{
			"requests_total":     reqs,
			"requests_failed":    c.failedRequests.Load(),
			"avg_latency_micros": avgMicros,
			"uptime_seconds":     int64(time.Since(c.startedAt).Seconds()),
			"timestamp":          time.Now().UTC(),
			"success":            true,
			"message":            "api-gateway metrics snapshot",
			"meta":               map[string]interface{}{},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})
}

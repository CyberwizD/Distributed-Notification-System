package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// Metrics exposes a tiny in-memory counter set for the push service.
type Metrics struct {
	consumed  atomic.Int64
	delivered atomic.Int64
	failed    atomic.Int64
	retried   atomic.Int64
}

// New returns a zeroed Metrics collector.
func New() *Metrics {
	return &Metrics{}
}

func (m *Metrics) IncConsumed()  { m.consumed.Add(1) }
func (m *Metrics) IncDelivered() { m.delivered.Add(1) }
func (m *Metrics) IncFailed()    { m.failed.Add(1) }
func (m *Metrics) IncRetried()   { m.retried.Add(1) }

// Handler exposes the counters via a very small JSON response so we do not
// need to pull in a heavy metrics dependency for the assignment.
func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "consumed": ` + itoa(m.consumed.Load()) + `,
  "delivered": ` + itoa(m.delivered.Load()) + `,
  "failed": ` + itoa(m.failed.Load()) + `,
  "retried": ` + itoa(m.retried.Load()) + `
}`))
	})
}

func itoa(v int64) string {
	return fmt.Sprintf("%d", v)
}

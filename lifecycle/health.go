package lifecycle

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// HealthStatus represents one Component's health verdict at a point in time.
// OK=true means healthy; Message and Details are free-form diagnostics.
type HealthStatus struct {
	OK      bool           `json:"ok"`
	Message string         `json:"message,omitempty"`
	Details map[string]any `json:"details,omitempty"`
	// Latency records how long the underlying Check call took (set by the
	// aggregator, not the checker).
	Latency time.Duration `json:"latency,omitempty"`
}

// HealthChecker is the optional interface a Component may implement to
// participate in active health probing. Components that do NOT implement
// HealthChecker are reported by Health() based on their ComponentState only.
type HealthChecker interface {
	// Check returns the current health of this component. Implementations
	// should respect ctx for cancellation/deadline and return quickly.
	Check(ctx context.Context) HealthStatus
}

// HealthReport is the aggregated map returned by Health(): one entry per
// registered component, keyed by component Id.
type HealthReport map[string]HealthStatus

// AllOK returns true iff every entry in the report is OK.
func (r HealthReport) AllOK() bool {
	for _, s := range r {
		if !s.OK {
			return false
		}
	}
	return true
}

// Health aggregates the health of every registered component on the manager.
// Components that implement HealthChecker have Check() invoked in parallel
// (each respecting ctx). Components that do not are mapped from
// ComponentState: Running → OK; Starting/Stopping → not OK with the state in
// the message; Error → not OK; everything else → not OK.
//
// The implementation lives on *SimpleComponentManager so it is opt-in via
// type assertion at call sites:
//
//	if mgr, ok := componentManager.(*lifecycle.SimpleComponentManager); ok {
//	    report := mgr.Health(ctx)
//	}
func (scm *SimpleComponentManager) Health(ctx context.Context) HealthReport {
	// Snapshot under the read lock then run probes outside it so a slow
	// HealthChecker can't block Register/Unregister/StartAll.
	scm.cMutex.RLock()
	snap := make([]Component, 0, len(scm.components))
	for _, c := range scm.components {
		snap = append(snap, c)
	}
	scm.cMutex.RUnlock()

	if len(snap) == 0 {
		return HealthReport{}
	}

	report := make(HealthReport, len(snap))
	type result struct {
		id     string
		status HealthStatus
	}
	resCh := make(chan result, len(snap))

	var wg sync.WaitGroup
	for _, c := range snap {
		wg.Add(1)
		go func(c Component) {
			defer wg.Done()
			resCh <- result{id: c.Id(), status: checkComponent(ctx, c)}
		}(c)
	}
	wg.Wait()
	close(resCh)
	for r := range resCh {
		report[r.id] = r.status
	}
	return report
}

// checkComponent invokes the HealthChecker if implemented, otherwise derives
// a HealthStatus from ComponentState. Records latency for the Check call.
func checkComponent(ctx context.Context, c Component) HealthStatus {
	if hc, ok := c.(HealthChecker); ok {
		start := time.Now()
		s := hc.Check(ctx)
		s.Latency = time.Since(start)
		return s
	}
	return stateToHealth(c.State())
}

// stateToHealth maps a coarse ComponentState into a HealthStatus.
func stateToHealth(state ComponentState) HealthStatus {
	switch state {
	case Running:
		return HealthStatus{OK: true}
	case Starting:
		return HealthStatus{OK: false, Message: "starting"}
	case Stopping:
		return HealthStatus{OK: false, Message: "stopping"}
	case Stopped:
		return HealthStatus{OK: false, Message: "stopped"}
	case Error:
		return HealthStatus{OK: false, Message: "error"}
	default:
		return HealthStatus{OK: false, Message: "unknown state"}
	}
}

// HealthHandler returns an http.Handler that writes the aggregated report as
// JSON with status 200 if AllOK and 503 otherwise. Suitable for
// /healthz, /readyz, etc.
//
//	http.Handle("/healthz", lifecycle.HealthHandler(componentMgr))
func HealthHandler(mgr ComponentManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scm, ok := mgr.(*SimpleComponentManager)
		if !ok {
			http.Error(w, "lifecycle: unsupported ComponentManager", http.StatusInternalServerError)
			return
		}
		report := scm.Health(r.Context())
		w.Header().Set("Content-Type", "application/json")
		if !report.AllOK() {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(report)
	})
}

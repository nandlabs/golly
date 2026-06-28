package lifecycle

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// healthyComp is a SimpleComponent that also implements HealthChecker.
type healthyComp struct {
	*SimpleComponent
	checkFn func(ctx context.Context) HealthStatus
}

func (h *healthyComp) Check(ctx context.Context) HealthStatus { return h.checkFn(ctx) }

func newHealthyComp(id string, fn func(ctx context.Context) HealthStatus) *healthyComp {
	return &healthyComp{
		SimpleComponent: &SimpleComponent{
			CompId:    id,
			StartFunc: func() error { return nil },
			StopFunc:  func() error { return nil },
		},
		checkFn: fn,
	}
}

func TestHealth_HealthCheckerInvoked(t *testing.T) {
	mgr := NewSimpleComponentManager().(*SimpleComponentManager)
	a := newHealthyComp("a", func(_ context.Context) HealthStatus { return HealthStatus{OK: true} })
	mgr.Register(a)

	report := mgr.Health(context.Background())
	if len(report) != 1 {
		t.Fatalf("expected 1 entry; got %d", len(report))
	}
	if !report["a"].OK {
		t.Errorf("a should be OK; got %+v", report["a"])
	}
}

func TestHealth_NonCheckerFallsBackToState(t *testing.T) {
	mgr := NewSimpleComponentManager().(*SimpleComponentManager)
	// Plain SimpleComponent — no HealthChecker.
	mgr.Register(&SimpleComponent{
		CompId:    "plain",
		StartFunc: func() error { return nil },
		StopFunc:  func() error { return nil },
	})
	// State is Unknown by default — should map to OK=false.
	report := mgr.Health(context.Background())
	if report["plain"].OK {
		t.Errorf("non-running plain component should be unhealthy; got %+v", report["plain"])
	}
}

func TestHealth_AllOK(t *testing.T) {
	mgr := NewSimpleComponentManager().(*SimpleComponentManager)
	mgr.Register(newHealthyComp("a", func(_ context.Context) HealthStatus { return HealthStatus{OK: true} }))
	mgr.Register(newHealthyComp("b", func(_ context.Context) HealthStatus { return HealthStatus{OK: true} }))

	report := mgr.Health(context.Background())
	if !report.AllOK() {
		t.Errorf("expected AllOK; got %+v", report)
	}
}

func TestHealth_OneBadFailsAllOK(t *testing.T) {
	mgr := NewSimpleComponentManager().(*SimpleComponentManager)
	mgr.Register(newHealthyComp("good", func(_ context.Context) HealthStatus { return HealthStatus{OK: true} }))
	mgr.Register(newHealthyComp("bad", func(_ context.Context) HealthStatus {
		return HealthStatus{OK: false, Message: "db unreachable"}
	}))

	report := mgr.Health(context.Background())
	if report.AllOK() {
		t.Errorf("expected !AllOK")
	}
	if report["bad"].Message != "db unreachable" {
		t.Errorf("bad.Message = %q; want db unreachable", report["bad"].Message)
	}
}

func TestHealth_LatencyRecorded(t *testing.T) {
	mgr := NewSimpleComponentManager().(*SimpleComponentManager)
	mgr.Register(newHealthyComp("a", func(_ context.Context) HealthStatus { return HealthStatus{OK: true} }))
	report := mgr.Health(context.Background())
	if report["a"].Latency == 0 {
		t.Errorf("expected non-zero Latency; got %v", report["a"].Latency)
	}
}

func TestHealthHandler_StatusCodes(t *testing.T) {
	mgr := NewSimpleComponentManager().(*SimpleComponentManager)
	mgr.Register(newHealthyComp("ok", func(_ context.Context) HealthStatus { return HealthStatus{OK: true} }))
	srv := httptest.NewServer(HealthHandler(mgr))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("healthy → status = %d, want 200", resp.StatusCode)
	}
	var report HealthReport
	_ = json.NewDecoder(resp.Body).Decode(&report)
	if !report["ok"].OK {
		t.Errorf("decoded report missing ok component: %+v", report)
	}

	// Add an unhealthy one and re-poll.
	mgr.Register(newHealthyComp("bad", func(_ context.Context) HealthStatus { return HealthStatus{OK: false} }))
	resp2, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("GET 2: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("unhealthy → status = %d, want 503", resp2.StatusCode)
	}
}

func TestHealth_EmptyManager(t *testing.T) {
	mgr := NewSimpleComponentManager().(*SimpleComponentManager)
	report := mgr.Health(context.Background())
	if len(report) != 0 {
		t.Errorf("empty manager → report should be empty; got %+v", report)
	}
	if !report.AllOK() {
		t.Errorf("empty report should be AllOK (vacuous true)")
	}
}

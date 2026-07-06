package genai

import (
	"context"
	"testing"
)

func TestResolve_MatchesExecutedPlanNoSideEffects(t *testing.T) {
	primary, fallback, _, opts := twoRoute()
	r, _ := NewRouter(opts...)

	route, decision, err := r.Resolve(chatTask(), nil)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	// Resolve must not touch providers.
	if primary.callCount() != 0 || fallback.callCount() != 0 {
		t.Fatalf("Resolve had side effects: primary=%d fallback=%d", primary.callCount(), fallback.callCount())
	}
	// Plan is populated and self-consistent.
	if route.Primary.Provider != "pp" || route.Primary.Model != "m-primary" {
		t.Errorf("primary = %+v, want pp/m-primary", route.Primary)
	}
	if len(route.Fallbacks) != 1 || route.Fallbacks[0].Provider != "fp" {
		t.Errorf("fallbacks = %+v, want [fp/m-fallback]", route.Fallbacks)
	}
	if decision == nil || decision.Strategy != "capability" {
		t.Fatalf("decision not populated: %+v", decision)
	}
	if decision.Chosen != route.Primary {
		t.Errorf("decision.Chosen %+v != route.Primary %+v", decision.Chosen, route.Primary)
	}
	if len(decision.Considered) != 2 {
		t.Errorf("considered = %d, want 2", len(decision.Considered))
	}

	// The executed path picks the same primary.
	primary.queue(okResult())
	if _, err := r.Route(context.Background(), chatTask(), NewTextMessage(RoleUser, "hi"), nil); err != nil {
		t.Fatalf("route: %v", err)
	}
	if primary.callCount() != 1 {
		t.Errorf("executed path should hit resolved primary once, got %d", primary.callCount())
	}
}

func TestResolve_PinnedDecision(t *testing.T) {
	_, _, _, opts := twoRoute()
	r, _ := NewRouter(opts...)
	task := chatTask()
	task.Pin = &Pin{Provider: "fp", Model: "m-fallback"}

	route, decision, err := r.Resolve(task, nil)
	if err != nil {
		t.Fatalf("resolve pinned: %v", err)
	}
	if !decision.Pinned || decision.Strategy != "pin" {
		t.Errorf("expected pinned decision, got %+v", decision)
	}
	if route.Primary.Provider != "fp" || len(route.Fallbacks) != 0 {
		t.Errorf("pinned route = %+v, want fp with no fallbacks", route)
	}
}

func TestResolve_NoRouteError(t *testing.T) {
	_, _, _, opts := twoRoute()
	r, _ := NewRouter(opts...)
	task := Task{Type: TaskChat, RequiredCapabilities: []Capability{CapAudio}}
	if _, _, err := r.Resolve(task, nil); err == nil {
		t.Fatal("expected ErrNoRoute")
	}
}

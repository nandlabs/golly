package genai

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// twoRoute builds two single-model providers where "pp/m-primary" is cheaper
// (so it ranks first under PriorityCost) and "fp/m-fallback" is the fallback.
func twoRoute() (primary, fallback *stubProvider, router ModelRouter, opts []RouterOption) {
	primary = newStubProvider("pp",
		ModelInfo{Name: "m-primary", Provider: "pp", Capabilities: []Capability{CapText, CapChat},
			MaxInputTokens: 100000, MaxOutputTokens: 4096, InputCostPerMTok: 1, OutputCostPerMTok: 1})
	fallback = newStubProvider("fp",
		ModelInfo{Name: "m-fallback", Provider: "fp", Capabilities: []Capability{CapText, CapChat},
			MaxInputTokens: 100000, MaxOutputTokens: 4096, InputCostPerMTok: 5, OutputCostPerMTok: 5})
	set := NewProviderSet(primary, fallback)
	reg, _ := NewModelRegistry(set, RoutingConfig{})
	opts = []RouterOption{WithProviders(set), WithRegistry(reg), WithStrategy(NewCapabilityStrategy())}
	return
}

func chatTask() Task {
	return Task{Type: TaskChat, EstInputTokens: 1000, Priority: PriorityCost,
		RequiredCapabilities: []Capability{CapText}}
}

func TestRoute_TransientPrimaryFailsFallbackSucceeds(t *testing.T) {
	primary, fallback, _, opts := twoRoute()
	primary.queue(errResult(transientErr()))
	fallback.queue(okResult())
	r, err := NewRouter(opts...)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := r.Route(context.Background(), chatTask(), NewTextMessage(RoleUser, "hi"), nil)
	if err != nil {
		t.Fatalf("expected success via fallback, got %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if primary.callCount() != 1 {
		t.Errorf("primary should be attempted once, got %d", primary.callCount())
	}
	if fallback.callCount() != 1 {
		t.Errorf("fallback should be used once, got %d", fallback.callCount())
	}
}

type transientSentinel struct{ msg string }

func (e transientSentinel) Error() string   { return e.msg }
func (e transientSentinel) Retryable() bool { return true }

func TestRoute_AllFailWrapsError(t *testing.T) {
	primary, fallback, _, opts := twoRoute()
	primary.queue(errResult(transientSentinel{"primary down"}))
	last := transientSentinel{"fallback down"}
	fallback.queue(errResult(last))
	r, _ := NewRouter(opts...)

	_, err := r.Route(context.Background(), chatTask(), NewTextMessage(RoleUser, "hi"), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "routes failed") {
		t.Errorf("error not wrapped as aggregate: %v", err)
	}
	if !errors.Is(err, last) {
		t.Errorf("aggregate error should wrap last failure via %%w: %v", err)
	}
}

func TestRoute_PermanentPrimaryNoFallback(t *testing.T) {
	primary, fallback, _, opts := twoRoute()
	primary.queue(errResult(permanentErr()))
	fallback.queue(okResult())
	r, _ := NewRouter(opts...)

	_, err := r.Route(context.Background(), chatTask(), NewTextMessage(RoleUser, "hi"), nil)
	if err == nil {
		t.Fatal("expected permanent error")
	}
	if primary.callCount() != 1 {
		t.Errorf("primary attempts = %d, want 1", primary.callCount())
	}
	if fallback.callCount() != 0 {
		t.Errorf("fallback must NOT be attempted on permanent error, got %d calls", fallback.callCount())
	}
}

func TestRoute_ContextCanceledMidFallback(t *testing.T) {
	primary, fallback, _, opts := twoRoute()
	ctx, cancel := context.WithCancel(context.Background())
	// Primary transient-fails and cancels the context, simulating a deadline
	// hit mid-route. The router must abort before trying the fallback.
	primary.genHook = func(string) { cancel() }
	primary.queue(errResult(transientErr()))
	fallback.queue(okResult())
	r, _ := NewRouter(opts...)

	_, err := r.Route(ctx, chatTask(), NewTextMessage(RoleUser, "hi"), nil)
	if err == nil {
		t.Fatal("expected error after cancellation")
	}
	if fallback.callCount() != 0 {
		t.Errorf("fallback must NOT be attempted after cancellation, got %d", fallback.callCount())
	}
}

func TestRoute_BreakerOpenSkipsToNextModel(t *testing.T) {
	primary, fallback, _, opts := twoRoute()
	primary.queue(okResult()) // would succeed if attempted
	fallback.queue(okResult())
	// Install controllable per-model breakers.
	opts = append(opts, withBreakerFactory(func() circuitBreaker { return &fakeBreaker{} }))
	r, _ := NewRouter(opts...)
	dr := r.(*defaultRouter)

	// Open the primary's breaker before routing.
	pb := dr.breakerFor(regKey("pp", "m-primary")).(*fakeBreaker)
	pb.open = true

	resp, err := r.Route(context.Background(), chatTask(), NewTextMessage(RoleUser, "hi"), nil)
	if err != nil {
		t.Fatalf("expected success via fallback, got %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if primary.callCount() != 0 {
		t.Errorf("primary breaker open: primary must be skipped, got %d calls", primary.callCount())
	}
	if fallback.callCount() != 1 {
		t.Errorf("fallback should be used, got %d", fallback.callCount())
	}
}

func TestClampOptions_MaxTokensClampedAndJSONDefault(t *testing.T) {
	info := ModelInfo{Provider: "p", Name: "m", MaxOutputTokens: 1000}
	opts := NewOptionsBuilder().SetMaxTokens(5000).Build()
	task := Task{RequiredCapabilities: []Capability{CapJSONMode}}

	clamped := clampOptions(opts, info, task)
	if got := clamped.GetInt(OptionMaxTokens); got != 1000 {
		t.Errorf("max_tokens = %d, want clamped to 1000", got)
	}
	// The caller's original options are not mutated.
	if got := opts.GetInt(OptionMaxTokens); got != 5000 {
		t.Errorf("caller options mutated: max_tokens = %d, want 5000", got)
	}
	// JSON-mode task sets a sane output mime default.
	if got := clamped.GetString(OptionOutputMime); got != "application/json" {
		t.Errorf("output_mime = %q, want application/json", got)
	}
}

func TestClampOptions_UnderLimitUntouched(t *testing.T) {
	info := ModelInfo{Provider: "p", Name: "m", MaxOutputTokens: 8000}
	opts := NewOptionsBuilder().SetMaxTokens(2000).Build()
	clamped := clampOptions(opts, info, Task{})
	if got := clamped.GetInt(OptionMaxTokens); got != 2000 {
		t.Errorf("max_tokens = %d, want unchanged 2000", got)
	}
}

func TestValidateOptions_ToolCallingRequiresTools(t *testing.T) {
	task := Task{RequiredCapabilities: []Capability{CapToolCalling}}
	if err := validateOptions(task, NewOptionsBuilder().Build()); err == nil {
		t.Fatal("expected error when tools missing")
	}
	withTools := NewOptionsBuilder().SetTools(Tool{}).Build()
	if err := validateOptions(task, withTools); err != nil {
		t.Errorf("unexpected error when tools present: %v", err)
	}
}

func TestRoute_Pinbypass(t *testing.T) {
	primary, fallback, _, opts := twoRoute()
	fallback.queue(okResult())
	r, _ := NewRouter(opts...)

	task := chatTask()
	task.Pin = &Pin{Provider: "fp", Model: "m-fallback"}
	resp, err := r.Route(context.Background(), task, NewTextMessage(RoleUser, "hi"), nil)
	if err != nil {
		t.Fatalf("pin route failed: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if primary.callCount() != 0 {
		t.Errorf("pin should bypass primary, got %d calls", primary.callCount())
	}
	if fallback.callCount() != 1 {
		t.Errorf("pinned provider should be called once, got %d", fallback.callCount())
	}
}

func TestRouteStream_FallsOverPreFirstToken(t *testing.T) {
	primary, fallback, _, opts := twoRoute()
	primary.queue(errResult(transientErr()))
	fallback.queue(okResult())
	r, _ := NewRouter(opts...)

	respCh, errCh := r.RouteStream(context.Background(), chatTask(), NewTextMessage(RoleUser, "hi"), nil)
	var got int
	for range respCh {
		got++
	}
	if err := <-errCh; err != nil {
		t.Fatalf("stream error: %v", err)
	}
	if got == 0 {
		t.Error("expected at least one streamed response from fallback")
	}
	if fallback.callCount() != 1 {
		t.Errorf("fallback should stream, got %d calls", fallback.callCount())
	}
}

func TestNewRouter_RequiresProvidersAndRegistry(t *testing.T) {
	if _, err := NewRouter(); err == nil {
		t.Error("expected error without providers/registry")
	}
	set := NewProviderSet(baseProvider())
	if _, err := NewRouter(WithProviders(set)); err == nil {
		t.Error("expected error without registry")
	}
}

package genai

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"oss.nandlabs.io/golly/clients"
)

// ModelRouter selects a (provider, model) for a task, reconciles options with
// the chosen model, and executes with error-classified, ctx-aware fallback. It
// is a selection layer above the unchanged Provider abstraction.
type ModelRouter interface {
	// Route selects a model, clamps options, executes with fallback, and
	// returns the response.
	Route(ctx context.Context, task Task, msg *Message, opts *Options) (*GenResponse, error)
	// RouteStream is the streaming variant. It falls over only on
	// pre-first-token transient errors; once tokens are emitted a mid-stream
	// failure surfaces to the caller.
	RouteStream(ctx context.Context, task Task, msg *Message, opts *Options) (<-chan *GenResponse, <-chan error)
	// Resolve returns the plan and full decision trace WITHOUT executing.
	Resolve(task Task, opts *Options) (*Route, *RouteDecision, error)
}

// regHolder wraps a ModelRegistry so it can be swapped atomically.
type regHolder struct{ reg ModelRegistry }

// circuitBreaker is the minimal breaker behaviour the router composes with.
// *clients.CircuitBreaker satisfies it; the seam also keeps the router's
// breaker-skip logic testable without depending on breaker internals.
type circuitBreaker interface {
	CanExecute() error
	OnExecution(success bool)
}

// routerCore holds the shared configuration and machinery used by both the
// generate-shaped ModelRouter and the EmbeddingRouter.
type routerCore struct {
	providers     *ProviderSet
	strategy      RoutingStrategy
	retry         *clients.RetryPolicy
	fallbackDepth int

	// newBreaker constructs a fresh per-(provider,model) breaker, or is nil when
	// no breaker is configured.
	newBreaker func() circuitBreaker

	registry atomic.Pointer[regHolder]

	mu       sync.Mutex
	breakers map[string]circuitBreaker
}

// RouterOption configures a router at construction time.
type RouterOption func(*routerCore)

// WithProviders sets the provider set the router dispatches to.
func WithProviders(set *ProviderSet) RouterOption {
	return func(c *routerCore) { c.providers = set }
}

// WithRegistry sets the (immutable) model registry the router routes over.
func WithRegistry(reg ModelRegistry) RouterOption {
	return func(c *routerCore) { c.registry.Store(&regHolder{reg: reg}) }
}

// WithStrategy sets the routing strategy (default: CapabilityStrategy).
func WithStrategy(s RoutingStrategy) RouterOption {
	return func(c *routerCore) { c.strategy = s }
}

// WithRetryPolicy sets the same-model transient-retry backoff policy.
func WithRetryPolicy(p *clients.RetryPolicy) RouterOption {
	return func(c *routerCore) { c.retry = p }
}

// WithCircuitBreaker sets a circuit-breaker template. Per-(provider,model)
// breakers are cloned from its BreakerInfo so one failing model does not trip
// routing for all of them.
func WithCircuitBreaker(cb *clients.CircuitBreaker) RouterOption {
	return func(c *routerCore) {
		if cb == nil {
			return
		}
		info := *cb.BreakerInfo
		c.newBreaker = func() circuitBreaker {
			cp := info
			return clients.NewCircuitBreaker(&cp)
		}
	}
}

// WithFallbackDepth sets how many ranked tail candidates become fallbacks.
func WithFallbackDepth(n int) RouterOption {
	return func(c *routerCore) { c.fallbackDepth = n }
}

func newRouterCore(opts ...RouterOption) (*routerCore, error) {
	c := &routerCore{
		strategy:      NewCapabilityStrategy(),
		fallbackDepth: 2,
		breakers:      map[string]circuitBreaker{},
	}
	for _, o := range opts {
		o(c)
	}
	if c.providers == nil {
		return nil, errors.New("genai: router requires a provider set (WithProviders)")
	}
	if c.registry.Load() == nil {
		return nil, errors.New("genai: router requires a registry (WithRegistry)")
	}
	if c.fallbackDepth < 0 {
		c.fallbackDepth = 0
	}
	return c, nil
}

// defaultRouter is the standard ModelRouter implementation.
type defaultRouter struct {
	*routerCore
}

// NewRouter constructs a ModelRouter from the given options. WithProviders and
// WithRegistry are required.
func NewRouter(opts ...RouterOption) (ModelRouter, error) {
	core, err := newRouterCore(opts...)
	if err != nil {
		return nil, err
	}
	return &defaultRouter{routerCore: core}, nil
}

// Reload atomically swaps the registry for hot-reload. Readers on the hot path
// never see a torn state and never take a lock.
func (c *routerCore) Reload(reg ModelRegistry) {
	c.registry.Store(&regHolder{reg: reg})
}

func (c *routerCore) loadRegistry() ModelRegistry {
	return c.registry.Load().reg
}

// breakerFor returns the per-(provider,model) circuit breaker, lazily created
// from the configured factory. Returns nil when no breaker is configured.
func (c *routerCore) breakerFor(key string) circuitBreaker {
	if c.newBreaker == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if b, ok := c.breakers[key]; ok {
		return b
	}
	b := c.newBreaker()
	c.breakers[key] = b
	return b
}

// resolve turns a task into an executable Route plus a decision trace, without
// any side effects. Pin bypasses strategy evaluation.
func (c *routerCore) resolve(task Task, reg ModelRegistry) (*Route, *RouteDecision, error) {
	if task.Pin != nil {
		chosen := RouteCandidate{Provider: task.Pin.Provider, Model: task.Pin.Model}
		if info, ok := reg.Get(task.Pin.Provider, task.Pin.Model); ok {
			chosen.EstCost = info.EstCost(task.EstInputTokens, defaultOutputEstimate(task))
		}
		route := &Route{Primary: chosen}
		decision := &RouteDecision{
			Strategy:   "pin",
			Chosen:     chosen,
			Considered: []RouteCandidate{chosen},
			Pinned:     true,
		}
		return route, decision, nil
	}

	var ranked []RouteCandidate
	var excluded []Exclusion
	if ex, ok := c.strategy.(explainer); ok {
		ranked, excluded = ex.rankExplain(task, reg)
	} else {
		ranked = c.strategy.Rank(task, reg)
	}
	if len(ranked) == 0 {
		return nil, nil, fmt.Errorf("%w: type=%q required=%v", ErrNoRoute, task.Type, task.RequiredCapabilities)
	}

	primary := ranked[0]
	var fallbacks []RouteCandidate
	if len(ranked) > 1 {
		end := 1 + c.fallbackDepth
		if end > len(ranked) {
			end = len(ranked)
		}
		fallbacks = append(fallbacks, ranked[1:end]...)
	}

	route := &Route{Primary: primary, Fallbacks: fallbacks}
	decision := &RouteDecision{
		Strategy:   c.strategy.Name(),
		Chosen:     primary,
		Fallbacks:  fallbacks,
		Considered: ranked,
		Excluded:   excluded,
	}
	return route, decision, nil
}

// Resolve implements ModelRouter.
func (r *defaultRouter) Resolve(task Task, _ *Options) (*Route, *RouteDecision, error) {
	return r.resolve(task, r.loadRegistry())
}

// validateOptions surfaces caller errors early (before any API call).
func validateOptions(task Task, opts *Options) error {
	if task.hasCapability(CapToolCalling) {
		if opts == nil || len(opts.GetTools()) == 0 {
			return errors.New("genai: task requires tool_calling but no tools set in options")
		}
	}
	return nil
}

// clampOptions reconciles caller options with the chosen model's limits and the
// task's capability requirements, returning a new Options (the caller's is
// never mutated).
func clampOptions(opts *Options, info ModelInfo, task Task) *Options {
	b := NewOptionsBuilder()
	if opts != nil {
		for k, v := range opts.All() {
			b.Add(k, v)
		}
	}
	out := b.Build()

	if info.MaxOutputTokens > 0 && out.Has(OptionMaxTokens) {
		if mt := out.GetInt(OptionMaxTokens); mt > info.MaxOutputTokens {
			LOGGER.DebugF("genai: clamping max_tokens %d -> %d for %s/%s", mt, info.MaxOutputTokens, info.Provider, info.Name)
			out.Set(OptionMaxTokens, info.MaxOutputTokens)
		}
	}

	if task.hasCapability(CapJSONMode) {
		if !out.Has(OptionOutputMime) && !out.Has(OptionSchema) {
			out.Set(OptionOutputMime, "application/json")
		}
	}
	return out
}

// Route implements ModelRouter.
func (r *defaultRouter) Route(ctx context.Context, task Task, msg *Message, opts *Options) (*GenResponse, error) {
	reg := r.loadRegistry()
	route, decision, err := r.resolve(task, reg)
	if err != nil {
		return nil, err
	}
	if err := validateOptions(task, opts); err != nil {
		return nil, err
	}
	r.logDecision(decision)

	attempts := append([]RouteCandidate{route.Primary}, route.Fallbacks...)
	var lastErr error
	for _, c := range attempts {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("genai: routing aborted: %w", err)
		}
		p, ok := r.providers.Get(c.Provider)
		if !ok {
			lastErr = fmt.Errorf("genai: provider %q not registered", c.Provider)
			continue
		}
		key := regKey(c.Provider, c.Model)
		breaker := r.breakerFor(key)
		if breaker != nil {
			if err := breaker.CanExecute(); err != nil {
				lastErr = fmt.Errorf("genai: breaker open for %s: %w", key, err)
				continue
			}
		}

		info, ok := reg.Get(c.Provider, c.Model)
		if !ok {
			// Registry miss (e.g. a pinned model absent from the merged view):
			// fall back to the provider's own metadata for clamping.
			info, _ = modelInfoFor(p, c.Model)
		}
		callOpts := clampOptions(opts, info, task)

		resp, err := r.callWithRetry(ctx, p, c.Model, msg, callOpts)
		if breaker != nil {
			breaker.OnExecution(err == nil)
		}
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !retryable(err, resp) {
			return nil, fmt.Errorf("genai: route %s failed permanently: %w", key, err)
		}
		// transient → fall over to the next candidate
	}
	return nil, fmt.Errorf("genai: all %d routes failed: %w", len(attempts), lastErr)
}

// callWithRetry runs a single model's Generate, retrying same-model transient
// failures using the configured clients.RetryPolicy backoff. Escalation to the
// next model is the router's separate fallback axis.
func (r *defaultRouter) callWithRetry(ctx context.Context, p Provider, model string, msg *Message, opts *Options) (*GenResponse, error) {
	maxAttempts := 1
	if r.retry != nil && r.retry.MaxRetries > 0 {
		maxAttempts = r.retry.MaxRetries + 1
	}
	var resp *GenResponse
	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			wait := r.retry.WaitTime(attempt)
			if wait > 0 {
				select {
				case <-ctx.Done():
					return resp, ctx.Err()
				case <-time.After(wait):
				}
			}
		}
		resp, err = p.Generate(ctx, model, msg, opts)
		if err == nil {
			return resp, nil
		}
		if !retryable(err, resp) {
			return resp, err
		}
		if ctx.Err() != nil {
			return resp, err
		}
	}
	return resp, err
}

// logDecision emits a compact decision trace via l3.
func (c *routerCore) logDecision(d *RouteDecision) {
	if d == nil {
		return
	}
	LOGGER.InfoF("genai: routed via %s -> %s/%s (fallbacks=%d, considered=%d, pinned=%t)",
		d.Strategy, d.Chosen.Provider, d.Chosen.Model, len(d.Fallbacks), len(d.Considered), d.Pinned)
}

// RouteStream implements ModelRouter.
func (r *defaultRouter) RouteStream(ctx context.Context, task Task, msg *Message, opts *Options) (<-chan *GenResponse, <-chan error) {
	outResp := make(chan *GenResponse)
	outErr := make(chan error, 1)

	go func() {
		defer close(outResp)
		defer close(outErr)

		reg := r.loadRegistry()
		route, decision, err := r.resolve(task, reg)
		if err != nil {
			outErr <- err
			return
		}
		if err := validateOptions(task, opts); err != nil {
			outErr <- err
			return
		}
		r.logDecision(decision)

		attempts := append([]RouteCandidate{route.Primary}, route.Fallbacks...)
		var lastErr error
		for _, c := range attempts {
			if err := ctx.Err(); err != nil {
				outErr <- fmt.Errorf("genai: routing aborted: %w", err)
				return
			}
			p, ok := r.providers.Get(c.Provider)
			if !ok {
				lastErr = fmt.Errorf("genai: provider %q not registered", c.Provider)
				continue
			}
			key := regKey(c.Provider, c.Model)
			breaker := r.breakerFor(key)
			if breaker != nil {
				if err := breaker.CanExecute(); err != nil {
					lastErr = fmt.Errorf("genai: breaker open for %s: %w", key, err)
					continue
				}
			}
			info, _ := reg.Get(c.Provider, c.Model)
			callOpts := clampOptions(opts, info, task)

			respCh, errCh := p.GenerateStream(ctx, c.Model, msg, callOpts)
			emitted, streamErr := forwardStream(ctx, respCh, errCh, outResp)
			if breaker != nil {
				breaker.OnExecution(streamErr == nil)
			}
			if streamErr == nil {
				return
			}
			lastErr = streamErr
			// Once tokens are emitted we cannot cleanly fail over; and permanent
			// errors never fall over.
			if emitted || !retryable(streamErr, nil) {
				outErr <- streamErr
				return
			}
			// pre-first-token transient → try the next candidate
		}
		outErr <- fmt.Errorf("genai: all %d routes failed: %w", len(attempts), lastErr)
	}()

	return outResp, outErr
}

// forwardStream drains a provider's stream into out, reporting whether any
// response was emitted and the terminal error (nil on clean completion).
func forwardStream(ctx context.Context, respCh <-chan *GenResponse, errCh <-chan error, out chan<- *GenResponse) (bool, error) {
	emitted := false
	for {
		select {
		case <-ctx.Done():
			return emitted, ctx.Err()
		case resp, ok := <-respCh:
			if !ok {
				respCh = nil
				if errCh == nil {
					return emitted, nil
				}
				continue
			}
			emitted = true
			select {
			case out <- resp:
			case <-ctx.Done():
				return emitted, ctx.Err()
			}
		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				if respCh == nil {
					return emitted, nil
				}
				continue
			}
			if err != nil {
				return emitted, err
			}
			errCh = nil
			if respCh == nil {
				return emitted, nil
			}
		}
	}
}

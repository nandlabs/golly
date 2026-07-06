package genai

import (
	"context"
	"fmt"
)

// EmbeddingRouter routes embedding requests. Embeddings have a different call
// shape than generation (Embedder.Embed, not Generate), so they get a parallel
// router that shares the same registry, strategies, and fallback engine —
// filtered to models advertising CapEmbeddings.
type EmbeddingRouter interface {
	// Embed selects an embedding model, executes with fallback, and returns the
	// vectors.
	Embed(ctx context.Context, task Task, req *EmbedRequest) (*EmbedResponse, error)
	// Resolve returns the plan and decision trace WITHOUT executing.
	Resolve(task Task) (*Route, *RouteDecision, error)
}

// defaultEmbeddingRouter is the standard EmbeddingRouter implementation.
type defaultEmbeddingRouter struct {
	*routerCore
}

// NewEmbeddingRouter constructs an EmbeddingRouter from the given options.
// WithProviders and WithRegistry are required; the default strategy is a
// CapabilityStrategy.
func NewEmbeddingRouter(opts ...RouterOption) (EmbeddingRouter, error) {
	core, err := newRouterCore(opts...)
	if err != nil {
		return nil, err
	}
	return &defaultEmbeddingRouter{routerCore: core}, nil
}

// withEmbeddings returns a copy of task with CapEmbeddings required, so routing
// only ever considers embedding-capable models.
func withEmbeddings(task Task) Task {
	if task.hasCapability(CapEmbeddings) {
		return task
	}
	caps := make([]Capability, 0, len(task.RequiredCapabilities)+1)
	caps = append(caps, task.RequiredCapabilities...)
	caps = append(caps, CapEmbeddings)
	task.RequiredCapabilities = caps
	return task
}

// Resolve implements EmbeddingRouter.
func (r *defaultEmbeddingRouter) Resolve(task Task) (*Route, *RouteDecision, error) {
	return r.resolve(withEmbeddings(task), r.loadRegistry())
}

// Embed implements EmbeddingRouter.
func (r *defaultEmbeddingRouter) Embed(ctx context.Context, task Task, req *EmbedRequest) (*EmbedResponse, error) {
	reg := r.loadRegistry()
	route, decision, err := r.resolve(withEmbeddings(task), reg)
	if err != nil {
		return nil, err
	}
	r.logDecision(decision)

	attempts := append([]RouteCandidate{route.Primary}, route.Fallbacks...)
	var lastErr error
	for _, c := range attempts {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("genai: embedding routing aborted: %w", err)
		}
		p, ok := r.providers.Get(c.Provider)
		if !ok {
			lastErr = fmt.Errorf("genai: provider %q not registered", c.Provider)
			continue
		}
		embedder, ok := p.(Embedder)
		if !ok {
			lastErr = fmt.Errorf("genai: provider %q does not implement Embedder", c.Provider)
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

		callReq := *req
		callReq.Model = c.Model
		resp, err := embedder.Embed(ctx, &callReq)
		if breaker != nil {
			breaker.OnExecution(err == nil)
		}
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !retryable(err, nil) {
			return nil, fmt.Errorf("genai: embedding route %s failed permanently: %w", key, err)
		}
	}
	return nil, fmt.Errorf("genai: all %d embedding routes failed: %w", len(attempts), lastErr)
}

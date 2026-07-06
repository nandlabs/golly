package genai

import (
	"context"
	"fmt"
	"sync"
)

// stubResult is one queued outcome for a stubProvider.Generate call.
type stubResult struct {
	resp *GenResponse
	err  error
}

// stubProvider implements both Provider and CapabilityProvider. Generate and
// Embed return queued results in order; when the queue is exhausted the last
// result repeats. It records every model it was called with, so tests can
// assert which routes were exercised.
type stubProvider struct {
	name    string
	catalog []ModelInfo

	mu       sync.Mutex
	results  []stubResult
	idx      int
	calls    []string // models passed to Generate/GenerateStream/Embed
	embedRes *EmbedResponse
	genHook  func(model string) // invoked at the start of Generate, if set
}

func newStubProvider(name string, catalog ...ModelInfo) *stubProvider {
	return &stubProvider{name: name, catalog: catalog}
}

// queue appends outcomes returned by successive Generate calls.
func (s *stubProvider) queue(results ...stubResult) *stubProvider {
	s.results = append(s.results, results...)
	return s
}

func okResult() stubResult {
	return stubResult{resp: &GenResponse{Candidates: []Candidate{{FinishReason: FinishReasonStop}}}}
}

func errResult(err error) stubResult { return stubResult{err: err} }

// --- Provider ---

func (s *stubProvider) Name() string        { return s.name }
func (s *stubProvider) Description() string { return "stub provider " + s.name }
func (s *stubProvider) Version() string     { return "test" }
func (s *stubProvider) Close() error        { return nil }

func (s *stubProvider) Models() []string {
	out := make([]string, 0, len(s.catalog))
	for _, m := range s.catalog {
		out = append(out, m.Name)
	}
	return out
}

func (s *stubProvider) next(model string) stubResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, model)
	if len(s.results) == 0 {
		return okResult()
	}
	i := s.idx
	if i >= len(s.results) {
		i = len(s.results) - 1
	} else {
		s.idx++
	}
	return s.results[i]
}

func (s *stubProvider) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

func (s *stubProvider) Generate(ctx context.Context, model string, _ *Message, _ *Options) (*GenResponse, error) {
	if s.genHook != nil {
		s.genHook(model)
	}
	r := s.next(model)
	if r.err != nil {
		// An explicitly queued error wins over ctx state so tests can drive
		// transient-then-cancel sequences deterministically.
		return nil, r.err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return r.resp, nil
}

func (s *stubProvider) GenerateStream(ctx context.Context, model string, _ *Message, _ *Options) (<-chan *GenResponse, <-chan error) {
	respCh := make(chan *GenResponse, 1)
	errCh := make(chan error, 1)
	r := s.next(model)
	go func() {
		defer close(respCh)
		defer close(errCh)
		if r.err != nil {
			errCh <- r.err
			return
		}
		if r.resp != nil {
			respCh <- r.resp
		}
	}()
	return respCh, errCh
}

// --- CapabilityProvider ---

func (s *stubProvider) ModelCatalog() []ModelInfo { return s.catalog }

func (s *stubProvider) ModelInfoFor(model string) (ModelInfo, bool) {
	for _, m := range s.catalog {
		if m.Name == model {
			return m, true
		}
	}
	return ModelInfo{}, false
}

// --- Embedder ---

func (s *stubProvider) Embed(ctx context.Context, req *EmbedRequest) (*EmbedResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r := s.next(req.Model)
	if r.err != nil {
		return nil, r.err
	}
	if s.embedRes != nil {
		return s.embedRes, nil
	}
	return &EmbedResponse{Vectors: [][]float32{{0.1, 0.2}}}, nil
}

// transientErr / permanentErr are canonical classified errors for tests.
func transientErr() error { return fmt.Errorf("openai API error: status 503 unavailable") }
func permanentErr() error { return fmt.Errorf("openai API error: status 400 bad request") }

// fakeBreaker is a controllable circuitBreaker for router tests (the real
// clients.CircuitBreaker is not substitutable and its state machine is exercised
// in its own package).
type fakeBreaker struct {
	open  bool
	execs []bool
}

func (f *fakeBreaker) CanExecute() error {
	if f.open {
		return errBreakerOpen
	}
	return nil
}
func (f *fakeBreaker) OnExecution(success bool) { f.execs = append(f.execs, success) }

var errBreakerOpen = errTest("breaker open")

type errTest string

func (e errTest) Error() string { return string(e) }

// withBreakerFactory is a test-only RouterOption that installs a custom breaker
// factory, so tests can inject controllable breakers.
func withBreakerFactory(f func() circuitBreaker) RouterOption {
	return func(c *routerCore) { c.newBreaker = f }
}

// mustRegistry builds a registry from a provider set, failing the test on error.
func mustRegistry(t interface{ Fatalf(string, ...any) }, set *ProviderSet, cfg RoutingConfig) ModelRegistry {
	reg, err := NewModelRegistry(set, cfg)
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	return reg
}

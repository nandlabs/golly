package genai

import "sort"

// ProviderSet is an immutable, name-indexed set of providers. The router takes
// an explicit ProviderSet rather than consulting a mutable package-global
// registry: this keeps routing testable (inject stub providers) and free of
// hidden global state.
type ProviderSet struct {
	byName map[string]Provider
}

// NewProviderSet builds a ProviderSet from the given providers, indexed by
// each provider's Name(). If two providers share a name, the last one wins.
func NewProviderSet(providers ...Provider) *ProviderSet {
	m := make(map[string]Provider, len(providers))
	for _, p := range providers {
		if p == nil {
			continue
		}
		m[p.Name()] = p
	}
	return &ProviderSet{byName: m}
}

// Get returns the provider registered under name, if present.
func (s *ProviderSet) Get(name string) (Provider, bool) {
	p, ok := s.byName[name]
	return p, ok
}

// All returns every provider in the set, stable-sorted by name for
// deterministic iteration (registry builds, decision traces).
func (s *ProviderSet) All() []Provider {
	names := make([]string, 0, len(s.byName))
	for n := range s.byName {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]Provider, 0, len(names))
	for _, n := range names {
		out = append(out, s.byName[n])
	}
	return out
}

// Len returns the number of providers in the set.
func (s *ProviderSet) Len() int { return len(s.byName) }

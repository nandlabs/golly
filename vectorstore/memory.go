package vectorstore

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// MemoryStore is a goroutine-safe in-memory Store. It uses cosine similarity
// for ranking and a naïve full scan for search — fine for tests and small
// datasets (a few thousand vectors); use a real backend for production
// workloads.
//
// Filter semantics: equality on each metadata key. Pass an empty filter
// (or nil) to skip filtering.
type MemoryStore struct {
	mu   sync.RWMutex
	dim  int
	docs map[string]Doc
}

// NewMemoryStore returns an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{docs: make(map[string]Doc)}
}

// Upsert inserts or replaces docs by ID. The first non-empty vector seen
// fixes the collection dimension; subsequent vectors must match.
func (m *MemoryStore) Upsert(_ context.Context, docs ...Doc) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range docs {
		if err := ValidateDoc(d); err != nil {
			return err
		}
		if m.dim == 0 {
			m.dim = len(d.Vector)
		} else if len(d.Vector) != m.dim {
			return fmt.Errorf("%w: collection dim=%d, doc=%q dim=%d", ErrDimMismatch, m.dim, d.ID, len(d.Vector))
		}
		m.docs[d.ID] = d
	}
	return nil
}

// Search runs a full scan and returns the top-K hits by cosine similarity.
func (m *MemoryStore) Search(_ context.Context, q Query) ([]Hit, error) {
	if len(q.Vector) == 0 {
		return nil, ErrEmptyVector
	}
	if q.TopK <= 0 {
		q.TopK = 10
	}
	m.mu.RLock()
	if m.dim != 0 && len(q.Vector) != m.dim {
		m.mu.RUnlock()
		return nil, fmt.Errorf("%w: query dim=%d, collection dim=%d", ErrDimMismatch, len(q.Vector), m.dim)
	}
	hits := make([]Hit, 0, len(m.docs))
	for _, d := range m.docs {
		if !matchesFilter(d.Metadata, q.Filter) {
			continue
		}
		score := CosineSimilarity(d.Vector, q.Vector)
		if score < q.MinScore {
			continue
		}
		hits = append(hits, Hit{ID: d.ID, Score: score, Metadata: d.Metadata, Content: d.Content})
	}
	m.mu.RUnlock()

	sort.SliceStable(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	if len(hits) > q.TopK {
		hits = hits[:q.TopK]
	}
	return hits, nil
}

// Delete removes the named ids (idempotent).
func (m *MemoryStore) Delete(_ context.Context, ids ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range ids {
		delete(m.docs, id)
	}
	return nil
}

// Close is a no-op for the in-memory store.
func (m *MemoryStore) Close() error { return nil }

// Len returns the number of stored documents.
func (m *MemoryStore) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.docs)
}

// matchesFilter returns true when every key in filter has an equal value in
// md. Empty/nil filter matches everything.
func matchesFilter(md, filter map[string]any) bool {
	if len(filter) == 0 {
		return true
	}
	for k, want := range filter {
		got, ok := md[k]
		if !ok || got != want {
			return false
		}
	}
	return true
}

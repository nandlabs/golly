// Package vectorstore is the in-tree interface for vector databases used for
// embedding-based retrieval. Pair it with genai.Embedder to build RAG and
// semantic-search pipelines without committing to a specific backend at the
// call site.
//
// The core package is stdlib-only — interface, types, and helpers. Backend
// implementations (pgvector, libSQL, Qdrant, Pinecone, …) live in separate
// satellite modules so the database driver dep stays out of any consumer that
// doesn't use that backend. See the integration guide for the conventions.
//
// Minimal usage:
//
//	import "oss.nandlabs.io/golly/vectorstore"
//
//	// upsert
//	store.Upsert(ctx, vectorstore.Doc{
//	    ID:       "doc-1",
//	    Vector:   embedding,                  // []float32 from genai.Embedder
//	    Metadata: map[string]any{"source": "wiki"},
//	    Content:  "raw text...",              // optional original content
//	})
//
//	// search
//	hits, _ := store.Search(ctx, vectorstore.Query{
//	    Vector: queryEmbedding,
//	    TopK:   5,
//	    Filter: map[string]any{"source": "wiki"},
//	})
package vectorstore

import (
	"context"
	"errors"
	"fmt"
	"math"
)

// Vector is a dense embedding produced by genai.Embedder (or any other
// source). Length is the embedding dimension — implementations should
// validate that all vectors in a collection share the same dim.
type Vector []float32

// Doc is one stored embedding plus optional metadata and source content.
// ID is the caller-owned identifier; backends MUST honor upsert semantics
// (insert if absent, replace if present) keyed by ID.
type Doc struct {
	ID       string         `json:"id"`
	Vector   Vector         `json:"vector"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Content  string         `json:"content,omitempty"`
}

// Query is a similarity search request.
type Query struct {
	// Vector is the query embedding. Must match the collection's dimension.
	Vector Vector `json:"vector"`
	// TopK is the maximum number of hits to return (default: 10).
	TopK int `json:"top_k"`
	// Filter is a backend-specific metadata predicate. Pass-through.
	// Implementations should document the dialect they accept (e.g. simple
	// equality vs. boolean expression strings).
	Filter map[string]any `json:"filter,omitempty"`
	// MinScore optionally drops hits whose score is below the threshold.
	// Score range depends on the metric; for cosine it's typically -1..1.
	MinScore float32 `json:"min_score,omitempty"`
}

// Hit is one similarity-search result.
type Hit struct {
	ID       string         `json:"id"`
	Score    float32        `json:"score"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Content  string         `json:"content,omitempty"`
}

// Store is the backend interface every implementation provides. All methods
// are context-aware and MUST honor cancellation.
type Store interface {
	// Upsert inserts or replaces docs by ID. Vectors must share the
	// collection's dimension; mismatched dims should return ErrDimMismatch.
	Upsert(ctx context.Context, docs ...Doc) error

	// Search returns up to q.TopK hits ordered by descending Score
	// (most similar first).
	Search(ctx context.Context, q Query) ([]Hit, error)

	// Delete removes docs by id. Unknown ids are silently skipped (the
	// operation is idempotent).
	Delete(ctx context.Context, ids ...string) error

	// Close releases backend resources (connections, etc.).
	Close() error
}

// Standard errors. Backends may wrap these with %w so callers can match
// with errors.Is.
var (
	ErrDimMismatch  = errors.New("vectorstore: vector dimension mismatch")
	ErrEmptyVector  = errors.New("vectorstore: empty vector")
	ErrEmptyID      = errors.New("vectorstore: empty document id")
	ErrInvalidTopK  = errors.New("vectorstore: TopK must be > 0")
	ErrNotSupported = errors.New("vectorstore: operation not supported by this backend")
)

// --- Similarity helpers (in-memory backends use these; satellites may too) ---

// CosineSimilarity returns the cosine similarity of a and b in [-1, 1].
// Returns 0 (and no panic) when either vector is empty or has zero norm.
func CosineSimilarity(a, b Vector) float32 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		fa, fb := float64(a[i]), float64(b[i])
		dot += fa * fb
		normA += fa * fa
		normB += fb * fb
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float32(dot / (math.Sqrt(normA) * math.Sqrt(normB)))
}

// EuclideanDistance returns the L2 distance between a and b.
// Returns +Inf for mismatched / empty inputs.
func EuclideanDistance(a, b Vector) float32 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return float32(math.Inf(1))
	}
	var sum float64
	for i := range a {
		d := float64(a[i] - b[i])
		sum += d * d
	}
	return float32(math.Sqrt(sum))
}

// DotProduct returns the inner product of a and b. Returns 0 for mismatched
// inputs.
func DotProduct(a, b Vector) float32 {
	if len(a) != len(b) {
		return 0
	}
	var sum float64
	for i := range a {
		sum += float64(a[i]) * float64(b[i])
	}
	return float32(sum)
}

// ValidateDoc returns nil if d is well-formed for storage; otherwise wraps
// one of the standard error sentinels.
func ValidateDoc(d Doc) error {
	if d.ID == "" {
		return ErrEmptyID
	}
	if len(d.Vector) == 0 {
		return fmt.Errorf("%w: id=%q", ErrEmptyVector, d.ID)
	}
	return nil
}

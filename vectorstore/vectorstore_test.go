package vectorstore

import (
	"context"
	"errors"
	"math"
	"testing"
)

// ---- helpers ----

func TestCosineSimilarity(t *testing.T) {
	a := Vector{1, 0, 0}
	b := Vector{1, 0, 0}
	if got := CosineSimilarity(a, b); math.Abs(float64(got-1)) > 1e-6 {
		t.Errorf("parallel = %v, want 1", got)
	}
	c := Vector{0, 1, 0}
	if got := CosineSimilarity(a, c); math.Abs(float64(got)) > 1e-6 {
		t.Errorf("orthogonal = %v, want 0", got)
	}
	d := Vector{-1, 0, 0}
	if got := CosineSimilarity(a, d); math.Abs(float64(got+1)) > 1e-6 {
		t.Errorf("antiparallel = %v, want -1", got)
	}
	// Mismatched / empty → 0
	if CosineSimilarity(a, Vector{1, 0}) != 0 {
		t.Error("dim mismatch should return 0")
	}
	if CosineSimilarity(nil, b) != 0 {
		t.Error("empty should return 0")
	}
}

func TestEuclideanAndDotProduct(t *testing.T) {
	if d := EuclideanDistance(Vector{0, 0}, Vector{3, 4}); math.Abs(float64(d-5)) > 1e-6 {
		t.Errorf("euclidean = %v, want 5", d)
	}
	if dp := DotProduct(Vector{1, 2, 3}, Vector{4, 5, 6}); dp != 32 {
		t.Errorf("dot product = %v, want 32", dp)
	}
}

func TestValidateDoc(t *testing.T) {
	if err := ValidateDoc(Doc{Vector: Vector{1}}); !errors.Is(err, ErrEmptyID) {
		t.Errorf("empty id should fail: %v", err)
	}
	if err := ValidateDoc(Doc{ID: "x"}); !errors.Is(err, ErrEmptyVector) {
		t.Errorf("empty vector should fail: %v", err)
	}
	if err := ValidateDoc(Doc{ID: "x", Vector: Vector{1}}); err != nil {
		t.Errorf("valid doc should pass: %v", err)
	}
}

// ---- MemoryStore behavior ----

func TestMemoryStore_UpsertAndSearch(t *testing.T) {
	store := NewMemoryStore()
	docs := []Doc{
		{ID: "a", Vector: Vector{1, 0, 0}, Metadata: map[string]any{"tag": "fruit"}, Content: "apple"},
		{ID: "b", Vector: Vector{0.9, 0.1, 0}, Metadata: map[string]any{"tag": "fruit"}, Content: "banana"},
		{ID: "c", Vector: Vector{0, 1, 0}, Metadata: map[string]any{"tag": "veg"}, Content: "carrot"},
	}
	if err := store.Upsert(context.Background(), docs...); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	hits, err := store.Search(context.Background(), Query{
		Vector: Vector{1, 0, 0},
		TopK:   2,
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) != 2 {
		t.Fatalf("hits = %d, want 2", len(hits))
	}
	if hits[0].ID != "a" || hits[1].ID != "b" {
		t.Errorf("ranking wrong: %v", hits)
	}
	if hits[0].Score < hits[1].Score {
		t.Errorf("scores not descending: %v vs %v", hits[0].Score, hits[1].Score)
	}
}

func TestMemoryStore_Filter(t *testing.T) {
	store := NewMemoryStore()
	_ = store.Upsert(context.Background(),
		Doc{ID: "a", Vector: Vector{1, 0}, Metadata: map[string]any{"tag": "x"}},
		Doc{ID: "b", Vector: Vector{1, 0}, Metadata: map[string]any{"tag": "y"}},
	)
	hits, _ := store.Search(context.Background(), Query{
		Vector: Vector{1, 0},
		TopK:   10,
		Filter: map[string]any{"tag": "y"},
	})
	if len(hits) != 1 || hits[0].ID != "b" {
		t.Errorf("filter didn't apply: %v", hits)
	}
}

func TestMemoryStore_MinScore(t *testing.T) {
	store := NewMemoryStore()
	_ = store.Upsert(context.Background(),
		Doc{ID: "near", Vector: Vector{1, 0}},
		Doc{ID: "far", Vector: Vector{0, 1}}, // orthogonal → score 0
	)
	hits, _ := store.Search(context.Background(), Query{
		Vector:   Vector{1, 0},
		TopK:     10,
		MinScore: 0.5,
	})
	if len(hits) != 1 || hits[0].ID != "near" {
		t.Errorf("MinScore should drop the far hit; got %v", hits)
	}
}

func TestMemoryStore_DimMismatch(t *testing.T) {
	store := NewMemoryStore()
	_ = store.Upsert(context.Background(), Doc{ID: "a", Vector: Vector{1, 0, 0}})
	err := store.Upsert(context.Background(), Doc{ID: "b", Vector: Vector{1, 0}})
	if !errors.Is(err, ErrDimMismatch) {
		t.Errorf("expected ErrDimMismatch on second upsert; got %v", err)
	}
	_, err = store.Search(context.Background(), Query{Vector: Vector{1, 0}})
	if !errors.Is(err, ErrDimMismatch) {
		t.Errorf("expected ErrDimMismatch on query with wrong dim; got %v", err)
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	store := NewMemoryStore()
	_ = store.Upsert(context.Background(),
		Doc{ID: "a", Vector: Vector{1}},
		Doc{ID: "b", Vector: Vector{1}},
	)
	if got := store.Len(); got != 2 {
		t.Fatalf("len = %d", got)
	}
	if err := store.Delete(context.Background(), "a", "missing"); err != nil {
		t.Errorf("delete should be idempotent: %v", err)
	}
	if got := store.Len(); got != 1 {
		t.Errorf("len after delete = %d, want 1", got)
	}
}

func TestMemoryStore_ValidationOnUpsert(t *testing.T) {
	store := NewMemoryStore()
	if err := store.Upsert(context.Background(), Doc{ID: "", Vector: Vector{1}}); !errors.Is(err, ErrEmptyID) {
		t.Errorf("expected ErrEmptyID; got %v", err)
	}
	if err := store.Upsert(context.Background(), Doc{ID: "x"}); !errors.Is(err, ErrEmptyVector) {
		t.Errorf("expected ErrEmptyVector; got %v", err)
	}
}

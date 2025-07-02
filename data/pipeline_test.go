package data

import (
	"errors"
	"testing"
)

func TestNewPipeline(t *testing.T) {
	p := NewPipeline("test-id")
	if p.Id() != "test-id" {
		t.Errorf("expected id 'test-id', got '%s'", p.Id())
	}
}

func TestNewPipelineFrom(t *testing.T) {
	values := map[string]any{"foo": 42, "bar": "baz"}
	p := NewPipelineFrom(values)
	for k, v := range values {
		val, err := p.Get(k)
		if err != nil || val != v {
			t.Errorf("expected %v for key %s, got %v, err: %v", v, k, val, err)
		}
	}
}

func TestSetGetHasDelete(t *testing.T) {
	p := NewPipeline("")
	if p.Has("missing") {
		t.Error("expected Has to be false for missing key")
	}
	if _, err := p.Get("missing"); !errors.Is(err, ErrKeyNotFound) {
		t.Error("expected ErrKeyNotFound for missing key")
	}
	if err := p.Set("foo", 123); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !p.Has("foo") {
		t.Error("expected Has to be true after Set")
	}
	val, err := p.Get("foo")
	if err != nil || val != 123 {
		t.Errorf("expected 123, got %v, err: %v", val, err)
	}
	if err := p.Delete("foo"); err != nil {
		t.Errorf("unexpected error on Delete: %v", err)
	}
	if p.Has("foo") {
		t.Error("expected Has to be false after Delete")
	}
}

func TestKeysAndMap(t *testing.T) {
	p := NewPipeline("")
	p.Set("a", 1)
	p.Set("b", 2)
	keys := p.Keys()
	m := p.Map()
	if len(keys) != 2 || len(m) != 2 {
		t.Errorf("expected 2 keys and 2 map entries, got %d and %d", len(keys), len(m))
	}
	if m["a"] != 1 || m["b"] != 2 {
		t.Error("map values incorrect")
	}
}

func TestErrorHandling(t *testing.T) {
	p := NewPipeline("")
	if p.HasError() {
		t.Error("expected no error initially")
	}
	err := errors.New("fail")
	p.SetError(err)
	if !p.HasError() {
		t.Error("expected HasError true after SetError")
	}
	if p.GetError() != err {
		t.Error("GetError did not return set error")
	}
}

func TestMergeFrom(t *testing.T) {
	p := NewPipeline("")
	p.Set("a", 1)
	other := map[string]any{"b": 2, "a": 3}
	p.MergeFrom(other)
	if v, _ := p.Get("a"); v != 3 {
		t.Error("expected 'a' to be overwritten to 3")
	}
	if v, _ := p.Get("b"); v != 2 {
		t.Error("expected 'b' to be 2")
	}
}

func TestMerge(t *testing.T) {
	p1 := NewPipeline("")
	p1.Set("x", 1)
	p2 := NewPipeline("")
	p2.Set("x", 2)
	p2.Set("y", 3)
	p1.Merge(p2)
	if v, _ := p1.Get("x"); v != 2 {
		t.Error("expected 'x' to be overwritten to 2")
	}
	if v, _ := p1.Get("y"); v != 3 {
		t.Error("expected 'y' to be 3")
	}
}

func TestClone(t *testing.T) {
	p := NewPipeline("id1")
	p.Set("foo", 42)
	p.SetError(errors.New("err"))
	clone := p.Clone()
	if clone.Id() != "id1" {
		t.Error("clone id mismatch")
	}
	if v, _ := clone.Get("foo"); v != 42 {
		t.Error("clone data mismatch")
	}
	if !clone.HasError() {
		t.Error("clone error mismatch")
	}
	// Mutate clone and check original is unchanged
	clone.Set("foo", 100)
	v, _ := p.Get("foo")
	if v != 42 {
		t.Error("original should not be affected by clone mutation")
	}
}

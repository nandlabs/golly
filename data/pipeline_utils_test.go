package data

import (
	"errors"
	"testing"
)

type mockPipeline map[string]any

func (m mockPipeline) Get(key string) (any, error) {
	v, ok := m[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}

// Add Clone method to satisfy Pipeline interface
func (m mockPipeline) Clone() Pipeline {
	return m
}

// Add Delete method to satisfy Pipeline interface
func (m mockPipeline) Delete(key string) error {
	return nil
}

// Add GetError method to satisfy Pipeline interface
func (m mockPipeline) GetError() error {
	return nil
}

// Add Has method to satisfy Pipeline interface
func (m mockPipeline) Has(key string) bool {
	_, ok := m[key]
	return ok
}

// Add HasError method to satisfy Pipeline interface
func (m mockPipeline) HasError() bool {
	return false
}

// Add Id method to satisfy Pipeline interface
func (m mockPipeline) Id() string {
	return "mock"
}

// Add Keys method to satisfy Pipeline interface
func (m mockPipeline) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Add Map method to satisfy Pipeline interface
func (m mockPipeline) Map() map[string]any {
	return m
}

// Correct Merge method to match Pipeline interface
func (m mockPipeline) Merge(other Pipeline) error {
	return nil
}

// Add MergeFrom method to satisfy Pipeline interface
func (m mockPipeline) MergeFrom(other map[string]any) error {
	return nil
}

// Add Set method to satisfy Pipeline interface
func (m mockPipeline) Set(key string, value any) error {
	m[key] = value
	return nil
}

// Add SetError method to satisfy Pipeline interface
func (m mockPipeline) SetError(err error) {}

func TestExtractValue_SimpleKey(t *testing.T) {
	p := mockPipeline{"foo": 42}
	v, err := ExtractValue[int](p, "foo")
	if err != nil || v != 42 {
		t.Errorf("expected 42, got %v, err=%v", v, err)
	}
}

func TestExtractValue_DotNotation(t *testing.T) {
	p := mockPipeline{"user": map[string]any{"city": "delhi"}}
	v, err := ExtractValue[string](p, "user.city")
	if err != nil || v != "delhi" {
		t.Errorf("expected delhi, got %v, err=%v", v, err)
	}
}

func TestExtractValue_ArrayIndex(t *testing.T) {
	users := []any{
		map[string]any{"name": "nanda", "address": map[string]any{"city": "blr"}},
		map[string]any{"name": "foo", "address": map[string]any{"city": "nyc"}},
	}
	p := mockPipeline{"users": users}
	v, err := ExtractValue[string](p, "users[1].address.city")
	if err != nil || v != "nyc" {
		t.Errorf("expected blr, got %v, err=%v", v, err)
	}
}

func TestExtractValue_ArrayFilter(t *testing.T) {
	users := []any{
		map[string]any{"name": "nanda", "address": map[string]any{"city": "blr"}},
		map[string]any{"name": "foo", "address": map[string]any{"city": "nyc"}},
	}
	p := mockPipeline{"users": users}
	v, err := ExtractValue[string](p, "users[name==\"nanda\"].address.city")
	if err != nil || v != "blr" {
		t.Errorf("expected blr, got %v, err=%v", v, err)
	}
}

func TestExtractValue_MapKey(t *testing.T) {
	users := map[string]any{
		"nanda": map[string]any{"address": map[string]any{"city": "blr"}},
		"foo":   map[string]any{"address": map[string]any{"city": "nyc"}},
	}
	p := mockPipeline{"users": users}
	v, err := ExtractValue[string](p, "users[nanda].address.city")
	if err != nil || v != "blr" {
		t.Errorf("expected blr, got %v, err=%v", v, err)
	}
}

func TestExtractValue_NotFound(t *testing.T) {
	p := mockPipeline{"foo": 42}
	_, err := ExtractValue[int](p, "bar")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestExtractValue_InvalidPath(t *testing.T) {
	p := mockPipeline{"foo": 42}
	_, err := ExtractValue[int](p, "foo.bar.baz")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestExtractValue_TypeMismatch(t *testing.T) {
	p := mockPipeline{"foo": "bar"}
	_, err := ExtractValue[int](p, "foo")
	if err == nil {
		t.Error("expected error for type mismatch")
	}
}

func TestExtractValue_StructArrayFilter(t *testing.T) {
	type Address struct{ City string }
	type User struct {
		Name    string
		Address Address
	}
	users := []User{{Name: "nanda", Address: Address{"blr"}}, {Name: "foo", Address: Address{"nyc"}}}
	p := mockPipeline{"users": users}
	v, err := ExtractValue[string](p, "users[Name==\"nanda\"].Address.City")
	if err != nil || v != "blr" {
		t.Errorf("expected blr, got %v, err=%v", v, err)
	}
}

func TestExtractValue_StructArrayIndex(t *testing.T) {
	type Address struct{ City string }
	type User struct {
		Name    string
		Address Address
	}
	users := []User{{Name: "nanda", Address: Address{"blr"}}, {Name: "foo", Address: Address{"nyc"}}}
	p := mockPipeline{"users": users}
	v, err := ExtractValue[string](p, "users[1].Address.City")
	if err != nil || v != "nyc" {
		t.Errorf("expected nyc, got %v, err=%v", v, err)
	}
}

package data

import "testing"

func TestSetValue_SimpleKey(t *testing.T) {
	p := mockPipeline{}
	err := SetValue(p, "foo", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, _ := p.Get("foo")
	if v != 42 {
		t.Errorf("expected 42, got %v", v)
	}
}

func TestSetValue_DotNotation(t *testing.T) {
	user := mockPipeline{"city": "delhi"}
	p := mockPipeline{"user": user}
	err := SetValue(p, "user.city", "blr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, _ := user.Get("city")
	if v != "blr" {
		t.Errorf("expected blr, got %v", v)
	}
}

func TestSetValue_ArrayFilter(t *testing.T) {
	users := []any{
		mockPipeline{"name": "nanda", "city": "blr"},
		mockPipeline{"name": "foo", "city": "nyc"},
	}
	p := mockPipeline{"users": users}
	err := SetValue(p, "users[name==\"foo\"].city", "sfo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if users[1].(mockPipeline)["city"] != "sfo" {
		t.Errorf("expected sfo, got %v", users[1].(mockPipeline)["city"])
	}
}

func TestSetValue_NestedPipeline(t *testing.T) {
	address := mockPipeline{"city": "blr"}
	user := mockPipeline{"address": address}
	p := mockPipeline{"user": user}
	err := SetValue(p, "user.address.city", "nyc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, _ := address.Get("city")
	if v != "nyc" {
		t.Errorf("expected nyc, got %v", v)
	}
}

func TestSetValue_NotPipelineError(t *testing.T) {
	p := mockPipeline{"foo": 123}
	err := SetValue(p, "foo.bar", 42)
	if err == nil {
		t.Error("expected error for non-pipeline final item")
	}
}

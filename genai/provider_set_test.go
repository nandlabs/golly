package genai

import (
	"reflect"
	"testing"
)

func TestProviderSet_IndexingAndStableOrder(t *testing.T) {
	pa := newStubProvider("alpha")
	pb := newStubProvider("bravo")
	pc := newStubProvider("charlie")

	// Construct out of order; All() must be stable-sorted by name.
	set := NewProviderSet(pc, pa, pb)

	if got := set.Len(); got != 3 {
		t.Fatalf("Len = %d, want 3", got)
	}
	if _, ok := set.Get("bravo"); !ok {
		t.Fatalf("Get(bravo) not found")
	}
	if _, ok := set.Get("missing"); ok {
		t.Fatalf("Get(missing) unexpectedly found")
	}

	var names []string
	for _, p := range set.All() {
		names = append(names, p.Name())
	}
	want := []string{"alpha", "bravo", "charlie"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("All() order = %v, want %v", names, want)
	}
}

func TestProviderSet_DuplicateNameLastWins(t *testing.T) {
	p1 := newStubProvider("dup", ModelInfo{Name: "m1", Provider: "dup"})
	p2 := newStubProvider("dup", ModelInfo{Name: "m2", Provider: "dup"})
	set := NewProviderSet(p1, p2)
	if set.Len() != 1 {
		t.Fatalf("Len = %d, want 1", set.Len())
	}
	got, _ := set.Get("dup")
	if got != Provider(p2) {
		t.Fatalf("duplicate name: last provider did not win")
	}
}

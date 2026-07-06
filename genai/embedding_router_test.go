package genai

import (
	"context"
	"testing"
)

func embedProviders() (*stubProvider, *ProviderSet, ModelRegistry) {
	p := newStubProvider("emb",
		ModelInfo{Name: "embed-small", Provider: "emb", Capabilities: []Capability{CapEmbeddings},
			MaxInputTokens: 8192, InputCostPerMTok: 0.02},
		ModelInfo{Name: "chat-only", Provider: "emb", Capabilities: []Capability{CapText, CapChat},
			MaxInputTokens: 8192},
	)
	set := NewProviderSet(p)
	reg, _ := NewModelRegistry(set, RoutingConfig{})
	return p, set, reg
}

func TestEmbeddingRouter_RoutesToEmbeddingModel(t *testing.T) {
	p, set, reg := embedProviders()
	er, err := NewEmbeddingRouter(WithProviders(set), WithRegistry(reg))
	if err != nil {
		t.Fatal(err)
	}

	// Resolve must only ever consider embedding-capable models.
	route, _, err := er.Resolve(Task{Type: TaskChat})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if route.Primary.Model != "embed-small" {
		t.Errorf("embedding route = %q, want embed-small", route.Primary.Model)
	}

	resp, err := er.Embed(context.Background(), Task{}, &EmbedRequest{Inputs: []Part{{Name: "t"}}})
	if err != nil {
		t.Fatalf("embed: %v", err)
	}
	if resp == nil || len(resp.Vectors) == 0 {
		t.Fatal("empty embed response")
	}
	if p.callCount() != 1 {
		t.Errorf("embed calls = %d, want 1", p.callCount())
	}
}

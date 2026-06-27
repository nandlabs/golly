package impl

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"oss.nandlabs.io/golly/genai"
)

const ollamaEmbedRespBody = `{
  "model":"nomic-embed-text",
  "embeddings":[[0.1,0.2,0.3],[0.4,0.5,0.6]],
  "total_duration":1500000,
  "load_duration":500000,
  "prompt_eval_count":12
}`

func TestOllamaEmbed_HitsNativeEndpoint(t *testing.T) {
	srv, rec := newRecorder(t, ollamaEmbedRespBody)
	// Ollama defaults to a /v1 base URL because of the OpenAI-compat path.
	// Embed must strip /v1 and target the native /api/embed root.
	p := NewOllamaProviderWithConfig(&OllamaProviderConfig{BaseURL: srv.URL + "/v1"}, nil)

	req := &genai.EmbedRequest{
		Model: "nomic-embed-text",
		Inputs: []genai.Part{
			{Text: &genai.TextPart{Content: "hello"}},
			{Text: &genai.TextPart{Content: "world"}},
		},
	}
	resp, err := p.Embed(context.Background(), req)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	if rec.path != "/api/embed" {
		t.Errorf("expected /api/embed (not /v1/...), got %s", rec.path)
	}
	var sent map[string]any
	if err := json.Unmarshal(rec.mu, &sent); err != nil {
		t.Fatalf("decode sent: %v", err)
	}
	if sent["model"] != "nomic-embed-text" {
		t.Errorf("model = %v", sent["model"])
	}
	inputs := sent["input"].([]any)
	if len(inputs) != 2 || inputs[0] != "hello" || inputs[1] != "world" {
		t.Errorf("input = %v", inputs)
	}
	if len(resp.Vectors) != 2 || resp.Vectors[0][0] != 0.1 || resp.Vectors[1][2] != 0.6 {
		t.Errorf("vectors decoded wrong: %v", resp.Vectors)
	}
	if resp.Meta == nil || resp.Meta.TotalTime == 0 {
		t.Errorf("meta missing: %+v", resp.Meta)
	}
}

func TestOllamaEmbed_NoTextInputs(t *testing.T) {
	srv, _ := newRecorder(t, ollamaEmbedRespBody)
	p := NewOllamaProviderWithConfig(&OllamaProviderConfig{BaseURL: srv.URL + "/v1"}, nil)
	req := &genai.EmbedRequest{
		Model:  "nomic-embed-text",
		Inputs: []genai.Part{{Bin: &genai.BinPart{Data: []byte{1, 2}}}},
	}
	if _, err := p.Embed(context.Background(), req); err == nil {
		t.Fatalf("expected error for embed with no text inputs")
	}
}

func TestOllamaRootURL_StripsV1(t *testing.T) {
	cases := []struct{ in, want string }{
		{"http://localhost:11434/v1", "http://localhost:11434"},
		{"http://localhost:11434/v1/", "http://localhost:11434"},
		{"http://localhost:11434", "http://localhost:11434"},
		{"https://ollama.example.com:443/v1", "https://ollama.example.com:443"},
	}
	for _, c := range cases {
		got := ollamaRootURL(c.in)
		if got != c.want {
			t.Errorf("ollamaRootURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// Sanity: the Embedder interface assertion compiles for both providers.
func TestProviders_ImplementEmbedder(t *testing.T) {
	var _ genai.Embedder = (*OpenAIProvider)(nil)
	var _ genai.Embedder = (*OllamaProvider)(nil)
	_ = strings.NewReader // silence unused-import linter
}

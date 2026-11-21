package impl

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"oss.nandlabs.io/golly/genai"
)

// customOpts returns options mixing standard options with custom ones:
// a nested map, a list, a scalar, a large integer, and names that conflict
// with fields the providers set themselves.
func customOpts() *genai.Options {
	return genai.NewOptionsBuilder().
		SetTemperature(0).
		SetMaxTokens(256).
		Add("reasoning", map[string]any{"effort": "high", "budget": 1024}).
		Add("modalities", []string{"text", "audio"}).
		Add("parallel_tool_calls", false).
		Add("big_id", int64(9007199254740993)). // > 2^53: must not lose precision
		Add("model", "evil-model").
		Add("messages", "evil-messages").
		Build()
}

func decodeBody(t *testing.T, b []byte) map[string]json.RawMessage {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("decode sent body: %v\nbody=%s", err, b)
	}
	return m
}

func assertForwarded(t *testing.T, body map[string]json.RawMessage, wantModel string) {
	t.Helper()

	var reasoning map[string]any
	if err := json.Unmarshal(body["reasoning"], &reasoning); err != nil {
		t.Fatalf("reasoning not forwarded as object: %s", body["reasoning"])
	}
	if !reflect.DeepEqual(reasoning, map[string]any{"effort": "high", "budget": float64(1024)}) {
		t.Errorf("reasoning = %v", reasoning)
	}

	var modalities []string
	if err := json.Unmarshal(body["modalities"], &modalities); err != nil || !reflect.DeepEqual(modalities, []string{"text", "audio"}) {
		t.Errorf("modalities = %s", body["modalities"])
	}

	if string(body["parallel_tool_calls"]) != "false" {
		t.Errorf("parallel_tool_calls = %s", body["parallel_tool_calls"])
	}
	if string(body["big_id"]) != "9007199254740993" {
		t.Errorf("big_id = %s", body["big_id"])
	}

	// Conflicting names never override provider-set fields.
	var model string
	_ = json.Unmarshal(body["model"], &model)
	if model != wantModel {
		t.Errorf("model overridden: got %q, want %q", model, wantModel)
	}
	if body["messages"][0] != '[' {
		t.Errorf("messages overridden: %s", body["messages"])
	}

	// Standard options keep their typed mapping and are not duplicated.
	if string(body["temperature"]) != "0" {
		t.Errorf("temperature = %s", body["temperature"])
	}
	if string(body["max_tokens"]) != "256" {
		t.Errorf("max_tokens = %s", body["max_tokens"])
	}
}

func assertNotForwarded(t *testing.T, body map[string]json.RawMessage) {
	t.Helper()
	for _, k := range []string{"reasoning", "modalities", "parallel_tool_calls", "big_id"} {
		if _, ok := body[k]; ok {
			t.Errorf("custom option %q forwarded although forwarding is disabled", k)
		}
	}
}

func drain(respCh <-chan *genai.GenResponse, errCh <-chan error) {
	for range respCh {
	}
	for range errCh {
	}
}

func TestOpenAI_ForwardsCustomOptions(t *testing.T) {
	srv, rec := newRecorder(t, openAIChatResp)
	p := NewOpenAIProvider("k", nil)
	p.baseURL = srv.URL

	msg := genai.NewTextMessage(genai.RoleUser, "hi")
	if _, err := p.Generate(context.Background(), "gpt-4o", msg, customOpts()); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	assertForwarded(t, decodeBody(t, rec.mu), "gpt-4o")
}

func TestOpenAI_Stream_ForwardsCustomOptions(t *testing.T) {
	srv, rec := newRecorder(t, "data: [DONE]\n\n")
	p := NewOpenAIProvider("k", nil)
	p.baseURL = srv.URL

	msg := genai.NewTextMessage(genai.RoleUser, "hi")
	drain(p.GenerateStream(context.Background(), "gpt-4o", msg, customOpts()))
	body := decodeBody(t, rec.mu)
	assertForwarded(t, body, "gpt-4o")
	if string(body["stream"]) != "true" {
		t.Errorf("stream = %s", body["stream"])
	}
}

func TestOpenAI_ForwardingDisabled(t *testing.T) {
	srv, rec := newRecorder(t, openAIChatResp)
	off := false
	p := NewOpenAIProviderWithConfig(&OpenAIProviderConfig{BaseURL: srv.URL, ForwardUnknownOptions: &off}, nil)

	msg := genai.NewTextMessage(genai.RoleUser, "hi")
	if _, err := p.Generate(context.Background(), "gpt-4o", msg, customOpts()); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	assertNotForwarded(t, decodeBody(t, rec.mu))
}

func TestOpenAI_UnsupportedStandardOptionNotForwarded(t *testing.T) {
	srv, rec := newRecorder(t, openAIChatResp)
	p := NewOpenAIProvider("k", nil)
	p.baseURL = srv.URL

	// top_k is a genai standard option that OpenAI does not support; it must
	// not leak into the body under its genai name.
	opts := genai.NewOptionsBuilder().SetTopK(5).SetSystemInstructions("sys").Build()
	if _, err := p.Generate(context.Background(), "gpt-4o", genai.NewTextMessage(genai.RoleUser, "hi"), opts); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	body := decodeBody(t, rec.mu)
	for _, k := range []string{genai.OptionTopK, genai.OptionSystemInstructions} {
		if _, ok := body[k]; ok {
			t.Errorf("standard option %q leaked into request body", k)
		}
	}
}

func TestOllama_ForwardsCustomOptions(t *testing.T) {
	srv, rec := newRecorder(t, openAIChatResp)
	p := NewOllamaProviderWithConfig(&OllamaProviderConfig{BaseURL: srv.URL}, nil)

	msg := genai.NewTextMessage(genai.RoleUser, "hi")
	if _, err := p.Generate(context.Background(), "llama3", msg, customOpts()); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	assertForwarded(t, decodeBody(t, rec.mu), "llama3")
}

func TestOllama_ForwardingDisabled(t *testing.T) {
	srv, rec := newRecorder(t, openAIChatResp)
	off := false
	p := NewOllamaProviderWithConfig(&OllamaProviderConfig{BaseURL: srv.URL, ForwardUnknownOptions: &off}, nil)

	msg := genai.NewTextMessage(genai.RoleUser, "hi")
	if _, err := p.Generate(context.Background(), "llama3", msg, customOpts()); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	assertNotForwarded(t, decodeBody(t, rec.mu))
}

func TestClaude_ForwardsCustomOptions(t *testing.T) {
	srv, rec := newRecorder(t, claudeResp)
	p := NewClaudeProvider("k", nil)
	p.baseURL = srv.URL

	opts := customOpts()
	opts.Set("thinking", map[string]any{"type": "enabled", "budget_tokens": 2048})
	opts.Set("metadata", map[string]any{"user_id": "u-1"}) // typed field left empty by the provider

	msg := genai.NewTextMessage(genai.RoleUser, "hi")
	if _, err := p.Generate(context.Background(), "claude-sonnet-5-5", msg, opts); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	body := decodeBody(t, rec.mu)
	assertForwarded(t, body, "claude-sonnet-5-5")
	if string(body["thinking"]) != `{"budget_tokens":2048,"type":"enabled"}` {
		t.Errorf("thinking = %s", body["thinking"])
	}
	if string(body["metadata"]) != `{"user_id":"u-1"}` {
		t.Errorf("metadata = %s", body["metadata"])
	}
}

func TestClaude_Stream_ForwardsCustomOptions(t *testing.T) {
	srv, rec := newRecorder(t, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
	p := NewClaudeProvider("k", nil)
	p.baseURL = srv.URL

	msg := genai.NewTextMessage(genai.RoleUser, "hi")
	drain(p.GenerateStream(context.Background(), "claude-sonnet-5-5", msg, customOpts()))
	body := decodeBody(t, rec.mu)
	assertForwarded(t, body, "claude-sonnet-5-5")
	if string(body["stream"]) != "true" {
		t.Errorf("stream = %s", body["stream"])
	}
}

func TestClaude_ForwardingDisabled(t *testing.T) {
	srv, rec := newRecorder(t, claudeResp)
	off := false
	p := NewClaudeProviderWithConfig(&ClaudeProviderConfig{BaseURL: srv.URL, ForwardUnknownOptions: &off}, nil)

	msg := genai.NewTextMessage(genai.RoleUser, "hi")
	if _, err := p.Generate(context.Background(), "claude-sonnet-5-5", msg, customOpts()); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	assertNotForwarded(t, decodeBody(t, rec.mu))
}

func TestMarshalRequestBody_NilOptionsAndUnmarshalable(t *testing.T) {
	type body struct {
		Model string `json:"model"`
	}
	b, err := marshalRequestBody(body{Model: "m"}, nil, true, "test")
	if err != nil || string(b) != `{"model":"m"}` {
		t.Fatalf("nil options: %s, %v", b, err)
	}

	opts := genai.NewOptionsBuilder().Add("bad", func() {}).Build()
	if _, err := marshalRequestBody(body{Model: "m"}, opts, true, "test"); err == nil {
		t.Fatal("expected error for unmarshalable custom option")
	}
}

func TestIsStandardOption(t *testing.T) {
	if !genai.IsStandardOption(genai.OptionTemperature) || !genai.IsStandardOption(genai.OptionTools) {
		t.Error("expected standard options to be recognised")
	}
	if genai.IsStandardOption("reasoning") {
		t.Error("custom option reported as standard")
	}
}

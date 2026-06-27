package impl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"oss.nandlabs.io/golly/data"
	"oss.nandlabs.io/golly/genai"
)

// recorder captures the last request body and path sent to the test server.
type recorder struct {
	mu       []byte
	path     string
	respBody string
}

func newRecorder(t *testing.T, resp string) (*httptest.Server, *recorder) {
	t.Helper()
	r := &recorder{respBody: resp}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		b, _ := io.ReadAll(req.Body)
		r.mu = b
		r.path = req.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, r.respBody)
	}))
	t.Cleanup(srv.Close)
	return srv, r
}

const openAIChatResp = `{
  "id":"x","object":"chat.completion","created":1,"model":"gpt-4o",
  "choices":[{"index":0,"message":{"role":"assistant","content":"hi"},"finish_reason":"stop"}],
  "usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}
}`

func TestOpenAIBuildRequest_ToolsAndChoice(t *testing.T) {
	srv, rec := newRecorder(t, openAIChatResp)
	p := NewOpenAIProvider("test-key", nil)
	p.baseURL = srv.URL

	schema := &data.Schema{
		Type:       "object",
		Properties: map[string]*data.Schema{"city": {Type: "string"}},
		Required:   []string{"city"},
	}
	opts := genai.NewOptionsBuilder().
		SetTools(genai.Tool{Function: &genai.FunctionDecl{
			Name:        "get_weather",
			Description: "Get the weather for a city.",
			Parameters:  schema,
		}}).
		SetToolChoice(genai.NewNamedToolChoice("get_weather")).
		Build()

	msg := genai.NewTextMessage(genai.RoleUser, "weather in Sydney?")
	if _, err := p.Generate(context.Background(), "gpt-4o", msg, opts); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.HasSuffix(rec.path, "/chat/completions") {
		t.Errorf("expected /chat/completions path, got %s", rec.path)
	}

	var sent map[string]any
	if err := json.Unmarshal(rec.mu, &sent); err != nil {
		t.Fatalf("decode sent body: %v\nbody=%s", err, rec.mu)
	}
	tools, _ := sent["tools"].([]any)
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool in request; got %d", len(tools))
	}
	tool := tools[0].(map[string]any)
	if tool["type"] != "function" {
		t.Errorf("tool.type = %v, want function", tool["type"])
	}
	fn := tool["function"].(map[string]any)
	if fn["name"] != "get_weather" {
		t.Errorf("tool.function.name = %v", fn["name"])
	}
	if fn["parameters"] == nil {
		t.Errorf("tool.function.parameters missing")
	}

	tc := sent["tool_choice"].(map[string]any)
	if tc["type"] != "function" {
		t.Errorf("tool_choice.type = %v", tc["type"])
	}
	tcFn := tc["function"].(map[string]any)
	if tcFn["name"] != "get_weather" {
		t.Errorf("tool_choice.function.name = %v", tcFn["name"])
	}
}

func TestOpenAIBuildRequest_StructuredOutputSchema(t *testing.T) {
	srv, rec := newRecorder(t, openAIChatResp)
	p := NewOpenAIProvider("test-key", nil)
	p.baseURL = srv.URL

	schema := &data.Schema{
		Title:       "Weather",
		Description: "Forecast for one city",
		Type:        "object",
		Properties: map[string]*data.Schema{
			"temp_c":  {Type: "number"},
			"summary": {Type: "string"},
		},
		Required: []string{"temp_c", "summary"},
	}
	opts := genai.NewOptionsBuilder().SetSchema(schema).Build()

	msg := genai.NewTextMessage(genai.RoleUser, "weather?")
	if _, err := p.Generate(context.Background(), "gpt-4o", msg, opts); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var sent map[string]any
	if err := json.Unmarshal(rec.mu, &sent); err != nil {
		t.Fatalf("decode sent body: %v", err)
	}
	rf := sent["response_format"].(map[string]any)
	if rf["type"] != "json_schema" {
		t.Errorf("response_format.type = %v, want json_schema", rf["type"])
	}
	js := rf["json_schema"].(map[string]any)
	if js["name"] != "Weather" {
		t.Errorf("json_schema.name = %v, want Weather (from schema.Title)", js["name"])
	}
	if js["strict"] != true {
		t.Errorf("json_schema.strict = %v, want true", js["strict"])
	}
	if js["schema"] == nil {
		t.Errorf("json_schema.schema missing")
	}
}

const openAIEmbedResp = `{
  "object":"list",
  "data":[
    {"object":"embedding","index":0,"embedding":[0.1,0.2,0.3]},
    {"object":"embedding","index":1,"embedding":[0.4,0.5,0.6]}
  ],
  "model":"text-embedding-3-small",
  "usage":{"prompt_tokens":7,"total_tokens":7}
}`

func TestOpenAIEmbed_RoundTrip(t *testing.T) {
	srv, rec := newRecorder(t, openAIEmbedResp)
	p := NewOpenAIProvider("test-key", nil)
	p.baseURL = srv.URL

	req := &genai.EmbedRequest{
		Model: "text-embedding-3-small",
		Inputs: []genai.Part{
			{Text: &genai.TextPart{Content: "hello"}},
			{Text: &genai.TextPart{Content: "world"}},
		},
	}
	resp, err := p.Embed(context.Background(), req)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	if !strings.HasSuffix(rec.path, "/embeddings") {
		t.Errorf("expected /embeddings path, got %s", rec.path)
	}
	var sent map[string]any
	if err := json.Unmarshal(rec.mu, &sent); err != nil {
		t.Fatalf("decode sent: %v", err)
	}
	if sent["model"] != "text-embedding-3-small" {
		t.Errorf("model = %v", sent["model"])
	}
	inputs := sent["input"].([]any)
	if len(inputs) != 2 || inputs[0] != "hello" || inputs[1] != "world" {
		t.Errorf("input = %v", inputs)
	}
	if len(resp.Vectors) != 2 {
		t.Fatalf("expected 2 vectors; got %d", len(resp.Vectors))
	}
	if resp.Vectors[0][0] != 0.1 || resp.Vectors[1][2] != 0.6 {
		t.Errorf("vector contents wrong: %v", resp.Vectors)
	}
	if resp.Meta == nil || resp.Meta.TotalTokens != 7 {
		t.Errorf("meta usage missing/wrong: %+v", resp.Meta)
	}
}

func TestOpenAIEmbed_NoTextInputs(t *testing.T) {
	srv, _ := newRecorder(t, openAIEmbedResp)
	p := NewOpenAIProvider("test-key", nil)
	p.baseURL = srv.URL

	req := &genai.EmbedRequest{
		Model:  "text-embedding-3-small",
		Inputs: []genai.Part{{Bin: &genai.BinPart{Data: []byte{1, 2, 3}}}},
	}
	if _, err := p.Embed(context.Background(), req); err == nil {
		t.Fatalf("expected error for embed with no text inputs")
	}
}

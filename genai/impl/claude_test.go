package impl

import (
	"context"
	"encoding/json"
	"testing"

	"oss.nandlabs.io/golly/data"
	"oss.nandlabs.io/golly/genai"
)

const claudeResp = `{
  "id":"msg_1","type":"message","role":"assistant","model":"claude-sonnet-4",
  "content":[{"type":"text","text":"hi"}],
  "stop_reason":"end_turn",
  "usage":{"input_tokens":1,"output_tokens":1}
}`

func TestClaudeBuildRequest_ToolsAndChoice(t *testing.T) {
	srv, rec := newRecorder(t, claudeResp)
	p := NewClaudeProvider("test-key", nil)
	p.baseURL = srv.URL

	schema := &data.Schema{
		Type: "object",
		Properties: map[string]*data.Schema{
			"city": {Type: "string"},
		},
		Required: []string{"city"},
	}
	opts := genai.NewOptionsBuilder().
		SetMaxTokens(256).
		SetTools(genai.Tool{Function: &genai.FunctionDecl{
			Name:        "get_weather",
			Description: "Weather lookup.",
			Parameters:  schema,
		}}).
		SetToolChoice(genai.NewNamedToolChoice("get_weather")).
		Build()

	msg := genai.NewTextMessage(genai.RoleUser, "weather?")
	if _, err := p.Generate(context.Background(), "claude-sonnet-4", msg, opts); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var sent map[string]any
	if err := json.Unmarshal(rec.mu, &sent); err != nil {
		t.Fatalf("decode sent: %v", err)
	}
	tools := sent["tools"].([]any)
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool; got %d", len(tools))
	}
	tool := tools[0].(map[string]any)
	if tool["name"] != "get_weather" {
		t.Errorf("tool.name = %v", tool["name"])
	}
	if tool["input_schema"] == nil {
		t.Errorf("tool.input_schema missing")
	}

	tc := sent["tool_choice"].(map[string]any)
	if tc["type"] != "tool" || tc["name"] != "get_weather" {
		t.Errorf("tool_choice = %v", tc)
	}
}

func TestClaudeBuildRequest_SchemaAsTool(t *testing.T) {
	srv, rec := newRecorder(t, claudeResp)
	p := NewClaudeProvider("test-key", nil)
	p.baseURL = srv.URL

	schema := &data.Schema{
		Title:       "Weather",
		Description: "Forecast for one city",
		Type:        "object",
		Properties:  map[string]*data.Schema{"temp_c": {Type: "number"}},
		Required:    []string{"temp_c"},
	}
	opts := genai.NewOptionsBuilder().SetMaxTokens(256).SetSchema(schema).Build()

	msg := genai.NewTextMessage(genai.RoleUser, "weather?")
	if _, err := p.Generate(context.Background(), "claude-sonnet-4", msg, opts); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var sent map[string]any
	if err := json.Unmarshal(rec.mu, &sent); err != nil {
		t.Fatalf("decode sent: %v", err)
	}
	tools, _ := sent["tools"].([]any)
	if len(tools) != 1 {
		t.Fatalf("expected synthetic tool for schema; got %v", tools)
	}
	tool := tools[0].(map[string]any)
	if tool["name"] != "Weather" {
		t.Errorf("synthetic tool name = %v, want Weather (schema.Title)", tool["name"])
	}
	tc, _ := sent["tool_choice"].(map[string]any)
	if tc == nil || tc["type"] != "tool" || tc["name"] != "Weather" {
		t.Errorf("tool_choice should pin to synthetic tool: %v", tc)
	}
}

func TestClaudeToolChoice_Mapping(t *testing.T) {
	cases := []struct {
		mode genai.ToolChoiceMode
		want string
	}{
		{genai.ToolChoiceAuto, "auto"},
		{genai.ToolChoiceNone, "none"},
		{genai.ToolChoiceRequired, "any"},
	}
	for _, c := range cases {
		got := claudeToolChoice(genai.NewToolChoice(c.mode)).(map[string]any)
		if got["type"] != c.want {
			t.Errorf("%s → type = %v, want %v", c.mode, got["type"], c.want)
		}
	}
}

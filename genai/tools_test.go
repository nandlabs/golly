package genai

import (
	"testing"

	"oss.nandlabs.io/golly/data"
)

func TestOptionsBuilder_ToolsRoundTrip(t *testing.T) {
	schema := &data.Schema{
		Type: "object",
		Properties: map[string]*data.Schema{
			"city": {Type: "string"},
		},
		Required: []string{"city"},
	}
	tools := []Tool{{Function: &FunctionDecl{
		Name:        "get_weather",
		Description: "Look up the current weather for a city.",
		Parameters:  schema,
	}}}
	choice := NewNamedToolChoice("get_weather")

	opts := NewOptionsBuilder().SetTools(tools...).SetToolChoice(choice).Build()

	got := opts.GetTools()
	if len(got) != 1 || got[0].Function == nil || got[0].Function.Name != "get_weather" {
		t.Fatalf("GetTools roundtrip: got %+v", got)
	}
	if got[0].Function.Parameters != schema {
		t.Fatalf("schema pointer not preserved on roundtrip")
	}

	gotChoice := opts.GetToolChoice()
	if gotChoice == nil || gotChoice.Mode != ToolChoiceNamed || gotChoice.Name != "get_weather" {
		t.Fatalf("GetToolChoice roundtrip: got %+v", gotChoice)
	}
}

func TestOptions_NoToolsAbsent(t *testing.T) {
	opts := NewOptionsBuilder().Build()
	if got := opts.GetTools(); got != nil {
		t.Fatalf("expected nil tools when unset; got %v", got)
	}
	if got := opts.GetToolChoice(); got != nil {
		t.Fatalf("expected nil tool-choice when unset; got %v", got)
	}
}

func TestToolChoice_Constructors(t *testing.T) {
	cases := []struct {
		name string
		got  *ToolChoice
		want ToolChoiceMode
	}{
		{"auto", NewToolChoice(ToolChoiceAuto), ToolChoiceAuto},
		{"none", NewToolChoice(ToolChoiceNone), ToolChoiceNone},
		{"required", NewToolChoice(ToolChoiceRequired), ToolChoiceRequired},
		{"named", NewNamedToolChoice("foo"), ToolChoiceNamed},
	}
	for _, c := range cases {
		if c.got.Mode != c.want {
			t.Errorf("%s: mode = %q, want %q", c.name, c.got.Mode, c.want)
		}
	}
}

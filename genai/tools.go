package genai

import "oss.nandlabs.io/golly/data"

// FunctionDecl declares a callable function (tool) that the model may invoke.
// Parameters is the JSON-Schema-shaped argument schema reused from golly/data.
type FunctionDecl struct {
	Name        string       `json:"name" yaml:"name" toml:"name"`
	Description string       `json:"description,omitempty" yaml:"description,omitempty" toml:"description,omitempty"`
	Parameters  *data.Schema `json:"parameters,omitempty" yaml:"parameters,omitempty" toml:"parameters,omitempty"`
}

// Tool is a single tool exposed to the model. Today only function tools exist,
// but the wrapper keeps room for future tool kinds (code interpreter, retrieval, etc.).
type Tool struct {
	Function *FunctionDecl `json:"function,omitempty" yaml:"function,omitempty" toml:"function,omitempty"`
}

// ToolChoiceMode controls how the model selects among the declared tools.
type ToolChoiceMode string

const (
	// ToolChoiceAuto lets the model decide whether to call a tool or respond directly.
	ToolChoiceAuto ToolChoiceMode = "auto"
	// ToolChoiceNone forbids tool calls; the model must respond with text only.
	ToolChoiceNone ToolChoiceMode = "none"
	// ToolChoiceRequired forces the model to call at least one tool.
	ToolChoiceRequired ToolChoiceMode = "required"
	// ToolChoiceNamed pins the model to a specific named function.
	// Pair with ToolChoice.Name.
	ToolChoiceNamed ToolChoiceMode = "named"
)

// ToolChoice expresses a tool-selection policy for a single request.
// Construct via NewToolChoice / NewNamedToolChoice.
type ToolChoice struct {
	Mode ToolChoiceMode `json:"mode" yaml:"mode" toml:"mode"`
	Name string         `json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty"`
}

// NewToolChoice returns a non-named tool-choice (auto / none / required).
// Use NewNamedToolChoice for a specific function.
func NewToolChoice(mode ToolChoiceMode) *ToolChoice {
	return &ToolChoice{Mode: mode}
}

// NewNamedToolChoice returns a tool-choice pinned to a named function.
func NewNamedToolChoice(name string) *ToolChoice {
	return &ToolChoice{Mode: ToolChoiceNamed, Name: name}
}

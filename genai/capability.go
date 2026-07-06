package genai

// Capability is a fact a model either supports or does not. Capabilities are
// authoritative: they are declared by the provider (or, for custom models, by
// operator configuration) and describe what a model can do — not what a caller
// would prefer. Capability-based routing treats them as a hard filter.
type Capability string

const (
	// CapText indicates plain text generation.
	CapText Capability = "text"
	// CapChat indicates multi-turn chat / conversation support.
	CapChat Capability = "chat"
	// CapStreaming indicates token-by-token streaming via GenerateStream.
	CapStreaming Capability = "streaming"
	// CapVision indicates image / visual input support.
	CapVision Capability = "vision"
	// CapAudio indicates audio input support.
	CapAudio Capability = "audio"
	// CapToolCalling indicates function / tool-calling support.
	CapToolCalling Capability = "tool_calling"
	// CapJSONMode indicates structured JSON output support.
	CapJSONMode Capability = "json_mode"
	// CapEmbeddings indicates the model produces embedding vectors (routed via
	// EmbeddingRouter, not the generate-shaped ModelRouter).
	CapEmbeddings Capability = "embeddings"
	// CapReasoning indicates extended reasoning / thinking support.
	CapReasoning Capability = "reasoning"
)

// ModelInfo describes a single model offered by a provider. Capabilities and
// token limits are authoritative (provider-owned); cost is a compiled-in
// default that operator configuration may override (see RoutingConfig).
type ModelInfo struct {
	// Name is the model id passed to Provider.Generate.
	Name string `json:"name" yaml:"name" toml:"name"`
	// Provider matches the name the serving provider is registered under.
	Provider string `json:"provider" yaml:"provider" toml:"provider"`
	// DisplayName is an optional human-friendly label.
	DisplayName string `json:"display_name,omitempty" yaml:"display_name,omitempty" toml:"display_name,omitempty"`
	// Capabilities is the set of facts this model supports.
	Capabilities []Capability `json:"capabilities,omitempty" yaml:"capabilities,omitempty" toml:"capabilities,omitempty"`
	// MaxInputTokens is the largest input context the model accepts.
	MaxInputTokens int `json:"max_input_tokens,omitempty" yaml:"max_input_tokens,omitempty" toml:"max_input_tokens,omitempty"`
	// MaxOutputTokens is the largest output the model can produce.
	MaxOutputTokens int `json:"max_output_tokens,omitempty" yaml:"max_output_tokens,omitempty" toml:"max_output_tokens,omitempty"`

	// InputCostPerMTok is the cost in USD per million input tokens.
	InputCostPerMTok float64 `json:"input_cost_per_mtok,omitempty" yaml:"input_cost_per_mtok,omitempty" toml:"input_cost_per_mtok,omitempty"`
	// OutputCostPerMTok is the cost in USD per million output tokens.
	OutputCostPerMTok float64 `json:"output_cost_per_mtok,omitempty" yaml:"output_cost_per_mtok,omitempty" toml:"output_cost_per_mtok,omitempty"`

	// Metadata carries provider-specific extras (region, tier, etc.).
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty" toml:"metadata,omitempty"`
}

// Has reports whether the model advertises the given capability.
func (m ModelInfo) Has(c Capability) bool {
	for _, x := range m.Capabilities {
		if x == c {
			return true
		}
	}
	return false
}

// EstCost returns an estimated USD cost for a single call with the given token
// counts. Cost-aware strategies rank on this blended input+output figure —
// output pricing usually dominates, so ranking on input price alone is wrong.
func (m ModelInfo) EstCost(inTokens, outTokens int) float64 {
	return float64(inTokens)/1e6*m.InputCostPerMTok +
		float64(outTokens)/1e6*m.OutputCostPerMTok
}

// CapabilityProvider is an optional interface a Provider may implement to
// advertise structured model metadata. It is discovered by type assertion; a
// provider that does not implement it is treated as "capabilities unknown" and
// is invisible to capability-based strategies (it can still be targeted by
// explicit name via a rule or a Pin).
//
// Note: this deliberately does NOT reuse the existing Provider.Models()
// []string method. ModelCatalog returns structured ModelInfo without colliding
// with that signature, keeping Provider fully backward compatible.
type CapabilityProvider interface {
	// ModelCatalog returns the authoritative metadata for every model the
	// provider serves.
	ModelCatalog() []ModelInfo
	// ModelInfoFor returns metadata for a single model, if known.
	ModelInfoFor(model string) (ModelInfo, bool)
}

// modelInfoFor is a helper that looks up a model's metadata from a provider if
// it implements CapabilityProvider.
func modelInfoFor(p Provider, model string) (ModelInfo, bool) {
	if cp, ok := p.(CapabilityProvider); ok {
		return cp.ModelInfoFor(model)
	}
	return ModelInfo{}, false
}

package genai

import (
	"fmt"
	"path"

	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/ioutils"
	"oss.nandlabs.io/golly/vfs"
)

// RoutingConfig is the operator policy overlay applied on top of authoritative
// provider capability metadata when building a ModelRegistry. It is typically
// loaded from YAML (see LoadRoutingConfig) under the genai.routing key.
type RoutingConfig struct {
	// DefaultStrategy names the strategy to use: rule_based | capability | composite.
	DefaultStrategy string `json:"default_strategy,omitempty" yaml:"default_strategy,omitempty" toml:"default_strategy,omitempty"`
	// FallbackDepth is how many ranked tail entries become fallbacks.
	FallbackDepth int `json:"fallback_depth,omitempty" yaml:"fallback_depth,omitempty" toml:"fallback_depth,omitempty"`

	// Rules are ordered (task -> provider/model) entries for RuleBasedStrategy.
	Rules []RoutingRule `json:"rules,omitempty" yaml:"rules,omitempty" toml:"rules,omitempty"`

	// CostOverrides replace the compiled-in prices for named models.
	CostOverrides []CostOverride `json:"cost_overrides,omitempty" yaml:"cost_overrides,omitempty" toml:"cost_overrides,omitempty"`
	// DisabledModels drop models entirely from the registry.
	DisabledModels []DisabledModel `json:"disabled_models,omitempty" yaml:"disabled_models,omitempty" toml:"disabled_models,omitempty"`
	// DisabledCapabilities strip specific capabilities from named models.
	DisabledCapabilities []DisabledCapability `json:"disabled_capabilities,omitempty" yaml:"disabled_capabilities,omitempty" toml:"disabled_capabilities,omitempty"`
	// CustomModels declare self-hosted / fine-tuned models wholly in config;
	// their provider must be present in the ProviderSet.
	CustomModels []CustomModel `json:"custom_models,omitempty" yaml:"custom_models,omitempty" toml:"custom_models,omitempty"`
}

// RoutingRule maps a TaskType to a concrete (provider, model).
type RoutingRule struct {
	Task     string `json:"task" yaml:"task" toml:"task"`
	Provider string `json:"provider" yaml:"provider" toml:"provider"`
	Model    string `json:"model" yaml:"model" toml:"model"`
}

// CostOverride replaces the input/output price of a named model.
type CostOverride struct {
	Provider          string  `json:"provider" yaml:"provider" toml:"provider"`
	Model             string  `json:"model" yaml:"model" toml:"model"`
	InputCostPerMTok  float64 `json:"input_cost_per_mtok" yaml:"input_cost_per_mtok" toml:"input_cost_per_mtok"`
	OutputCostPerMTok float64 `json:"output_cost_per_mtok" yaml:"output_cost_per_mtok" toml:"output_cost_per_mtok"`
}

// DisabledModel drops a (provider, model) from the registry.
type DisabledModel struct {
	Provider string `json:"provider" yaml:"provider" toml:"provider"`
	Model    string `json:"model" yaml:"model" toml:"model"`
}

// DisabledCapability strips capabilities from a named model.
type DisabledCapability struct {
	Provider     string       `json:"provider" yaml:"provider" toml:"provider"`
	Model        string       `json:"model" yaml:"model" toml:"model"`
	Capabilities []Capability `json:"capabilities" yaml:"capabilities" toml:"capabilities"`
}

// CustomModel declares a model entirely in config. Capabilities are allowed
// here (and only here) because the library has no compiled knowledge of
// self-hosted / fine-tuned models.
type CustomModel struct {
	Provider          string            `json:"provider" yaml:"provider" toml:"provider"`
	Name              string            `json:"name" yaml:"name" toml:"name"`
	DisplayName       string            `json:"display_name,omitempty" yaml:"display_name,omitempty" toml:"display_name,omitempty"`
	Capabilities      []Capability      `json:"capabilities,omitempty" yaml:"capabilities,omitempty" toml:"capabilities,omitempty"`
	MaxInputTokens    int               `json:"max_input_tokens,omitempty" yaml:"max_input_tokens,omitempty" toml:"max_input_tokens,omitempty"`
	MaxOutputTokens   int               `json:"max_output_tokens,omitempty" yaml:"max_output_tokens,omitempty" toml:"max_output_tokens,omitempty"`
	InputCostPerMTok  float64           `json:"input_cost_per_mtok,omitempty" yaml:"input_cost_per_mtok,omitempty" toml:"input_cost_per_mtok,omitempty"`
	OutputCostPerMTok float64           `json:"output_cost_per_mtok,omitempty" yaml:"output_cost_per_mtok,omitempty" toml:"output_cost_per_mtok,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty" toml:"metadata,omitempty"`
}

// ToModelInfo converts a config-declared custom model into a ModelInfo.
func (c CustomModel) ToModelInfo() ModelInfo {
	return ModelInfo{
		Name:              c.Name,
		Provider:          c.Provider,
		DisplayName:       c.DisplayName,
		Capabilities:      c.Capabilities,
		MaxInputTokens:    c.MaxInputTokens,
		MaxOutputTokens:   c.MaxOutputTokens,
		InputCostPerMTok:  c.InputCostPerMTok,
		OutputCostPerMTok: c.OutputCostPerMTok,
		Metadata:          c.Metadata,
	}
}

// routingConfigFile is the on-disk envelope: genai.routing.
type routingConfigFile struct {
	Genai struct {
		Routing RoutingConfig `json:"routing" yaml:"routing" toml:"routing"`
	} `json:"genai" yaml:"genai" toml:"genai"`
}

// LoadRoutingConfig loads a RoutingConfig from a file (YAML/JSON, by extension)
// using the in-tree codec package. The routing block is expected under the
// genai.routing key.
func LoadRoutingConfig(filePath string) (RoutingConfig, error) {
	var cfg routingConfigFile
	f, err := vfs.GetManager().OpenRaw(filePath)
	if err != nil {
		return RoutingConfig{}, fmt.Errorf("genai: opening routing config %q: %w", filePath, err)
	}
	defer func() { _ = f.Close() }()

	mimeType := ioutils.GetMimeFromExt(path.Ext(filePath))
	c, err := codec.GetDefault(mimeType)
	if err != nil {
		return RoutingConfig{}, fmt.Errorf("genai: no codec for routing config %q: %w", filePath, err)
	}
	if err := c.Read(f, &cfg); err != nil {
		return RoutingConfig{}, fmt.Errorf("genai: decoding routing config %q: %w", filePath, err)
	}
	return cfg.Genai.Routing, nil
}

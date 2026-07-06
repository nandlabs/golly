package genai

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleYAML = `genai:
  routing:
    default_strategy: capability
    fallback_depth: 2
    rules:
      - task: codegen
        provider: bedrock
        model: sonnet
    cost_overrides:
      - provider: bedrock
        model: opus
        input_cost_per_mtok: 12.00
        output_cost_per_mtok: 60.00
    disabled_models:
      - provider: vertex
        model: gemini-1.0-pro
    disabled_capabilities:
      - provider: bedrock
        model: haiku
        capabilities: [vision]
    custom_models:
      - provider: bedrock
        name: my-imported-finetune
        capabilities: [text, chat, json_mode]
        max_input_tokens: 32000
        max_output_tokens: 4096
        input_cost_per_mtok: 0.5
        output_cost_per_mtok: 1.5
`

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	return p
}

func configProviders() *ProviderSet {
	bedrock := newStubProvider("bedrock",
		ModelInfo{Name: "sonnet", Provider: "bedrock", Capabilities: []Capability{CapText, CapChat}, MaxInputTokens: 200000, MaxOutputTokens: 8192, InputCostPerMTok: 3, OutputCostPerMTok: 15},
		ModelInfo{Name: "opus", Provider: "bedrock", Capabilities: []Capability{CapText, CapChat}, MaxInputTokens: 200000, MaxOutputTokens: 4096, InputCostPerMTok: 15, OutputCostPerMTok: 75},
		ModelInfo{Name: "haiku", Provider: "bedrock", Capabilities: []Capability{CapText, CapChat, CapVision}, MaxInputTokens: 200000, MaxOutputTokens: 8192, InputCostPerMTok: 0.8, OutputCostPerMTok: 4},
	)
	vertex := newStubProvider("vertex",
		ModelInfo{Name: "gemini-1.0-pro", Provider: "vertex", Capabilities: []Capability{CapText, CapChat}, MaxInputTokens: 32000, MaxOutputTokens: 8192},
	)
	return NewProviderSet(bedrock, vertex)
}

func TestLoadRoutingConfig_RoundTripToRegistry(t *testing.T) {
	path := writeTemp(t, "genai.yaml", sampleYAML)
	cfg, err := LoadRoutingConfig(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.DefaultStrategy != "capability" || cfg.FallbackDepth != 2 {
		t.Errorf("config header wrong: %+v", cfg)
	}
	if len(cfg.Rules) != 1 || cfg.Rules[0].Task != "codegen" {
		t.Errorf("rules not parsed: %+v", cfg.Rules)
	}

	reg, err := NewModelRegistry(configProviders(), cfg)
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}

	// cost override applied to bedrock/opus.
	opus, _ := reg.Get("bedrock", "opus")
	if opus.InputCostPerMTok != 12.0 || opus.OutputCostPerMTok != 60.0 {
		t.Errorf("cost override not applied: %+v", opus)
	}
	// vision stripped from bedrock/haiku.
	haiku, _ := reg.Get("bedrock", "haiku")
	if haiku.Has(CapVision) {
		t.Errorf("vision should be disabled on haiku")
	}
	// vertex/gemini-1.0-pro dropped.
	if _, ok := reg.Get("vertex", "gemini-1.0-pro"); ok {
		t.Errorf("gemini-1.0-pro should be disabled")
	}
	// custom model added.
	custom, ok := reg.Get("bedrock", "my-imported-finetune")
	if !ok || !custom.Has(CapJSONMode) {
		t.Errorf("custom model missing or wrong: %+v", custom)
	}
}

func TestLoadRoutingConfig_PhantomCustomModel(t *testing.T) {
	yaml := `genai:
  routing:
    custom_models:
      - provider: ghost
        name: phantom
        capabilities: [text]
`
	path := writeTemp(t, "phantom.yaml", yaml)
	cfg, err := LoadRoutingConfig(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if _, err := NewModelRegistry(configProviders(), cfg); err == nil {
		t.Fatal("expected build error for phantom custom model")
	}
}

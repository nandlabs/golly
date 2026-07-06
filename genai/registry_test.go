package genai

import (
	"strings"
	"testing"
)

// baseProvider returns a stub provider "base" with a couple of models.
func baseProvider() *stubProvider {
	return newStubProvider("base",
		ModelInfo{
			Name: "big", Provider: "base",
			Capabilities:      []Capability{CapText, CapChat, CapVision},
			MaxInputTokens:    100000,
			MaxOutputTokens:   4096,
			InputCostPerMTok:  3.0,
			OutputCostPerMTok: 15.0,
		},
		ModelInfo{
			Name: "small", Provider: "base",
			Capabilities:      []Capability{CapText, CapChat},
			MaxInputTokens:    16000,
			MaxOutputTokens:   2048,
			InputCostPerMTok:  0.5,
			OutputCostPerMTok: 1.5,
		},
	)
}

func TestBuildRegistry_FiveStepMerge(t *testing.T) {
	set := NewProviderSet(baseProvider())

	cfg := RoutingConfig{
		// Step 2: cost override on base/big.
		CostOverrides: []CostOverride{
			{Provider: "base", Model: "big", InputCostPerMTok: 12.0, OutputCostPerMTok: 60.0},
		},
		// Step 3: strip vision from base/big.
		DisabledCapabilities: []DisabledCapability{
			{Provider: "base", Model: "big", Capabilities: []Capability{CapVision}},
		},
		// Step 4: drop base/small entirely.
		DisabledModels: []DisabledModel{
			{Provider: "base", Model: "small"},
		},
		// Step 5: add a custom model on the registered provider.
		CustomModels: []CustomModel{
			{
				Provider: "base", Name: "custom-ft",
				Capabilities:    []Capability{CapText, CapChat, CapJSONMode},
				MaxInputTokens:  32000,
				MaxOutputTokens: 4096,
			},
		},
	}

	reg, err := NewModelRegistry(set, cfg)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	// Step 1 base present; step 2 cost override applied.
	big, ok := reg.Get("base", "big")
	if !ok {
		t.Fatalf("base/big missing")
	}
	if big.InputCostPerMTok != 12.0 || big.OutputCostPerMTok != 60.0 {
		t.Errorf("cost override not applied: got in=%v out=%v", big.InputCostPerMTok, big.OutputCostPerMTok)
	}
	// Step 3 capability disabled.
	if big.Has(CapVision) {
		t.Errorf("vision capability should have been stripped from base/big")
	}
	if !big.Has(CapText) {
		t.Errorf("text capability should remain on base/big")
	}
	// Step 4 model disabled.
	if _, ok := reg.Get("base", "small"); ok {
		t.Errorf("base/small should have been dropped")
	}
	// Step 5 custom model added.
	custom, ok := reg.Get("base", "custom-ft")
	if !ok {
		t.Fatalf("custom-ft missing")
	}
	if !custom.Has(CapJSONMode) {
		t.Errorf("custom-ft should advertise json_mode")
	}

	if len(reg.All()) != 2 { // big + custom-ft (small dropped)
		t.Errorf("All() len = %d, want 2", len(reg.All()))
	}
}

func TestBuildRegistry_CustomModelUnregisteredProvider(t *testing.T) {
	set := NewProviderSet(baseProvider())
	cfg := RoutingConfig{
		CustomModels: []CustomModel{
			{Provider: "ghost", Name: "phantom", Capabilities: []Capability{CapText}},
		},
	}
	_, err := NewModelRegistry(set, cfg)
	if err == nil {
		t.Fatalf("expected error for custom model on unregistered provider")
	}
	if !strings.Contains(err.Error(), "phantom") || !strings.Contains(err.Error(), "ghost") {
		t.Errorf("error should name the model and provider: %v", err)
	}
}

func TestRegistry_AllStableSorted(t *testing.T) {
	set := NewProviderSet(baseProvider())
	reg := mustRegistry(t, set, RoutingConfig{})
	all := reg.All()
	if len(all) != 2 {
		t.Fatalf("want 2 models, got %d", len(all))
	}
	if all[0].Name != "big" || all[1].Name != "small" {
		t.Errorf("All() not stable-sorted by name: %v", []string{all[0].Name, all[1].Name})
	}
}

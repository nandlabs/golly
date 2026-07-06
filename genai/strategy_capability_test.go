package genai

import "testing"

// costModels builds a registry exercising the blended-cost regression.
func costModels() ModelRegistry {
	p := newStubProvider("p",
		// balanced: cheaper *total* for a balanced in/out workload.
		ModelInfo{Name: "balanced", Provider: "p", Capabilities: []Capability{CapText, CapChat},
			MaxInputTokens: 100000, MaxOutputTokens: 4096,
			InputCostPerMTok: 5.0, OutputCostPerMTok: 5.0},
		// cheap input, expensive output: looks cheap if you rank on input only.
		ModelInfo{Name: "cheapin", Provider: "p", Capabilities: []Capability{CapText, CapChat},
			MaxInputTokens: 100000, MaxOutputTokens: 4096,
			InputCostPerMTok: 1.0, OutputCostPerMTok: 30.0},
		// vision model, for capability filtering.
		ModelInfo{Name: "seer", Provider: "p", Capabilities: []Capability{CapText, CapChat, CapVision},
			MaxInputTokens: 100000, MaxOutputTokens: 4096,
			InputCostPerMTok: 10.0, OutputCostPerMTok: 40.0},
	)
	reg, _ := NewModelRegistry(NewProviderSet(p), RoutingConfig{})
	return reg
}

func TestCapabilityStrategy_BlendedCostRanking(t *testing.T) {
	reg := costModels()
	s := NewCapabilityStrategy()
	task := Task{
		Type:                 TaskChat,
		EstInputTokens:       1000,
		EstOutputTokens:      1000,
		Priority:             PriorityCost,
		RequiredCapabilities: []Capability{CapText},
	}
	ranked := s.Rank(task, reg)
	if len(ranked) != 3 {
		t.Fatalf("want 3 candidates, got %d", len(ranked))
	}
	// Regression: cheap-input/expensive-output must NOT outrank the balanced
	// cheaper-total model.
	if ranked[0].Model != "balanced" {
		t.Errorf("cost ranking: want balanced first, got %q (costs: %v)", ranked[0].Model, ranked)
	}
	if ranked[1].Model != "cheapin" {
		t.Errorf("cost ranking: want cheapin second, got %q", ranked[1].Model)
	}
}

func TestCapabilityStrategy_CapabilityFilter(t *testing.T) {
	reg := costModels()
	s := NewCapabilityStrategy()
	task := Task{
		Type:                 TaskChat,
		EstInputTokens:       1000,
		Priority:             PriorityCost,
		RequiredCapabilities: []Capability{CapVision},
	}
	ranked, excluded := s.rankExplain(task, reg)
	if len(ranked) != 1 || ranked[0].Model != "seer" {
		t.Fatalf("vision filter: want only seer, got %v", ranked)
	}
	if len(excluded) != 2 {
		t.Errorf("want 2 exclusions, got %d (%v)", len(excluded), excluded)
	}
	for _, e := range excluded {
		if e.Reason == "" {
			t.Errorf("exclusion missing reason: %+v", e)
		}
	}
}

func TestCapabilityStrategy_MaxCostFiltering(t *testing.T) {
	reg := costModels()
	s := NewCapabilityStrategy()
	// balanced est cost @ in=1000,out=1000 = 0.01; cheapin = 0.031; seer = 0.05.
	task := Task{
		Type:                 TaskChat,
		EstInputTokens:       1000,
		EstOutputTokens:      1000,
		Priority:             PriorityCost,
		RequiredCapabilities: []Capability{CapText},
		MaxCostUSD:           0.02, // only balanced fits.
	}
	ranked := s.Rank(task, reg)
	if len(ranked) != 1 || ranked[0].Model != "balanced" {
		t.Fatalf("budget filter: want only balanced, got %v", ranked)
	}
}

func TestCapabilityStrategy_EmptyOnNoMatch(t *testing.T) {
	reg := costModels()
	s := NewCapabilityStrategy()
	task := Task{
		Type:                 TaskChat,
		RequiredCapabilities: []Capability{CapAudio}, // no model has it
	}
	if ranked := s.Rank(task, reg); len(ranked) != 0 {
		t.Fatalf("want empty ranking, got %v", ranked)
	}
}

func TestCapabilityStrategy_QualityAndSpeed(t *testing.T) {
	p := newStubProvider("p",
		ModelInfo{Name: "tiny", Provider: "p", Capabilities: []Capability{CapText}, MaxInputTokens: 8000, MaxOutputTokens: 1024},
		ModelInfo{Name: "huge", Provider: "p", Capabilities: []Capability{CapText}, MaxInputTokens: 200000, MaxOutputTokens: 8192},
	)
	reg, _ := NewModelRegistry(NewProviderSet(p), RoutingConfig{})
	s := NewCapabilityStrategy()

	q := s.Rank(Task{RequiredCapabilities: []Capability{CapText}, Priority: PriorityQuality}, reg)
	if q[0].Model != "huge" {
		t.Errorf("quality: want huge first, got %q", q[0].Model)
	}
	sp := s.Rank(Task{RequiredCapabilities: []Capability{CapText}, Priority: PrioritySpeed}, reg)
	if sp[0].Model != "tiny" {
		t.Errorf("speed: want tiny first, got %q", sp[0].Model)
	}
}

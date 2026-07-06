package genai

import "testing"

func ruleReg() ModelRegistry {
	p := newStubProvider("bedrock",
		ModelInfo{Name: "sonnet", Provider: "bedrock", Capabilities: []Capability{CapText, CapChat}, MaxInputTokens: 200000, MaxOutputTokens: 8192, InputCostPerMTok: 3, OutputCostPerMTok: 15},
		ModelInfo{Name: "haiku", Provider: "bedrock", Capabilities: []Capability{CapText, CapChat}, MaxInputTokens: 200000, MaxOutputTokens: 8192, InputCostPerMTok: 0.8, OutputCostPerMTok: 4},
	)
	reg, _ := NewModelRegistry(NewProviderSet(p), RoutingConfig{})
	return reg
}

func TestRuleBasedStrategy_MatchAndDefault(t *testing.T) {
	reg := ruleReg()
	s := NewRuleBasedStrategy().
		When(TaskCodeGen).Use("bedrock", "sonnet").
		When(TaskSummarize).Use("bedrock", "haiku").
		Default("bedrock", "haiku")

	// Matching task type.
	got := s.Rank(Task{Type: TaskCodeGen}, reg)
	if len(got) != 1 || got[0].Model != "sonnet" {
		t.Fatalf("codegen rule: got %v", got)
	}
	if got[0].EstCost == 0 {
		t.Errorf("expected EstCost populated from registry")
	}

	// Unmatched task type falls through to default.
	def := s.Rank(Task{Type: TaskClassify}, reg)
	if len(def) != 1 || def[0].Model != "haiku" {
		t.Fatalf("default rule: got %v", def)
	}
}

func TestRuleBasedStrategy_MultipleRulesSameTaskAreOrdered(t *testing.T) {
	reg := ruleReg()
	s := NewRuleBasedStrategy().
		When(TaskChat).Use("bedrock", "sonnet").
		When(TaskChat).Use("bedrock", "haiku")
	got := s.Rank(Task{Type: TaskChat}, reg)
	if len(got) != 2 || got[0].Model != "sonnet" || got[1].Model != "haiku" {
		t.Fatalf("ordered rules: got %v", got)
	}
	if got[0].Score <= got[1].Score {
		t.Errorf("first rule should score higher: %v", got)
	}
}

func TestRuleBasedStrategy_NoMatchNoDefaultEmpty(t *testing.T) {
	reg := ruleReg()
	s := NewRuleBasedStrategy().When(TaskCodeGen).Use("bedrock", "sonnet")
	if got := s.Rank(Task{Type: TaskChat}, reg); len(got) != 0 {
		t.Fatalf("want empty, got %v", got)
	}
}

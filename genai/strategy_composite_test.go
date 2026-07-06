package genai

import "testing"

// abstainStrategy always returns an empty ranking.
type abstainStrategy struct{}

func (abstainStrategy) Name() string                              { return "abstain" }
func (abstainStrategy) Rank(Task, ModelRegistry) []RouteCandidate { return nil }

func TestCompositeStrategy_FirstNonEmptyWins(t *testing.T) {
	reg := ruleReg()
	rules := NewRuleBasedStrategy().When(TaskChat).Use("bedrock", "haiku")
	comp := NewCompositeStrategy(abstainStrategy{}, rules, NewCapabilityStrategy())

	got := comp.Rank(Task{Type: TaskChat}, reg)
	if len(got) != 1 || got[0].Model != "haiku" {
		t.Fatalf("composite: want rule result (haiku), got %v", got)
	}
}

func TestCompositeStrategy_FallsThroughToCapability(t *testing.T) {
	reg := ruleReg()
	// Rule strategy abstains for TaskReason (no rule, no default); capability
	// strategy then ranks.
	rules := NewRuleBasedStrategy().When(TaskChat).Use("bedrock", "haiku")
	comp := NewCompositeStrategy(rules, NewCapabilityStrategy())

	got := comp.Rank(Task{Type: TaskReason, RequiredCapabilities: []Capability{CapText}}, reg)
	if len(got) == 0 {
		t.Fatalf("composite: expected capability fallthrough to produce candidates")
	}
}

func TestCompositeStrategy_AllAbstainEmpty(t *testing.T) {
	reg := ruleReg()
	comp := NewCompositeStrategy(abstainStrategy{}, abstainStrategy{})
	if got := comp.Rank(Task{Type: TaskChat}, reg); len(got) != 0 {
		t.Fatalf("want empty, got %v", got)
	}
}

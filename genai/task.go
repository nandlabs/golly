package genai

// TaskType is a fixed-vocabulary rule-lookup key. It drives RuleBasedStrategy
// lookups and carries no deterministic model mapping on its own; teams that
// route purely on capabilities or tags may leave Task.Type empty.
//
// Note: there is intentionally no TaskEmbed — embeddings are routed through
// EmbeddingRouter, which has a different terminal call shape.
type TaskType string

const (
	TaskChat      TaskType = "chat"
	TaskCodeGen   TaskType = "codegen"
	TaskSummarize TaskType = "summarize"
	TaskExtract   TaskType = "extract"
	TaskClassify  TaskType = "classify"
	TaskReason    TaskType = "reason"
)

// RoutingPriority expresses the caller's optimisation goal, used by
// capability-based ranking to break ties between eligible models.
type RoutingPriority string

const (
	// PrioritySpeed prefers smaller / faster models.
	PrioritySpeed RoutingPriority = "speed"
	// PriorityCost prefers the cheapest estimated per-call cost.
	PriorityCost RoutingPriority = "cost"
	// PriorityQuality prefers larger-context / higher-tier models.
	PriorityQuality RoutingPriority = "quality"
)

// Pin bypasses routing entirely, dispatching directly to a named
// (Provider, Model). It is an escape hatch for A/B tests and for reproducing a
// specific model's behavior while debugging.
type Pin struct {
	Provider string
	Model    string
}

// Task describes what a call needs so the router can select a model, rather
// than the caller hard-coding a model name.
type Task struct {
	// Type is the rule-lookup key (see TaskType).
	Type TaskType
	// Tags are free-form policy hooks, e.g. ["sensitive","long-context"].
	Tags []string
	// EstInputTokens is the estimated input size, used for limit and cost checks.
	EstInputTokens int
	// EstOutputTokens is the estimated output size, used for cost ranking.
	// Zero falls back to a per-TaskType default.
	EstOutputTokens int
	// MaxCostUSD is an optional per-call budget in USD (0 = unbounded). It is
	// compared against the estimated per-call cost, not a per-Mtok rate.
	MaxCostUSD float64
	// Priority is the optimisation goal for tie-breaking.
	Priority RoutingPriority
	// RequiredCapabilities is a hard filter — a model must advertise all of them.
	RequiredCapabilities []Capability
	// Pin, if non-nil, bypasses strategy evaluation entirely.
	Pin *Pin
}

// hasCapability reports whether the task requires the given capability.
func (t Task) hasCapability(c Capability) bool {
	for _, x := range t.RequiredCapabilities {
		if x == c {
			return true
		}
	}
	return false
}

// defaultOutputEstimate returns a per-TaskType output-token estimate used for
// cost ranking when Task.EstOutputTokens is unset.
func defaultOutputEstimate(t Task) int {
	if t.EstOutputTokens > 0 {
		return t.EstOutputTokens
	}
	switch t.Type {
	case TaskCodeGen, TaskReason:
		return 1024
	case TaskChat, TaskSummarize:
		return 512
	case TaskExtract:
		return 256
	case TaskClassify:
		return 64
	default:
		return 512
	}
}

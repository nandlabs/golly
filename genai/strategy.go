package genai

// RoutingStrategy ranks the models in a registry for a given task. Returning a
// ranked list (not a single pick) is what lets the router derive fallbacks
// automatically: ranked[0] becomes the primary and the tail becomes the
// ordered fallbacks.
type RoutingStrategy interface {
	// Name identifies the strategy in decision traces / logs.
	Name() string
	// Rank returns candidates best-first. An empty slice means the strategy
	// abstains (a CompositeStrategy then moves on to the next strategy).
	Rank(task Task, reg ModelRegistry) []RouteCandidate
}

// explainer is an optional internal interface a strategy may implement to also
// report why models were excluded, populating RouteDecision.Excluded.
type explainer interface {
	rankExplain(task Task, reg ModelRegistry) (ranked []RouteCandidate, excluded []Exclusion)
}

// ruleEntry is a single ordered rule.
type ruleEntry struct {
	task     TaskType
	provider string
	model    string
}

// RuleBasedStrategy maps TaskType (in declaration order) to concrete
// (provider, model) pairs. The first rule matching the task becomes the
// primary; remaining rules for the same task form the fallback tail. A default
// applies when no rule matches.
//
//	genai.NewRuleBasedStrategy().
//	    When(genai.TaskCodeGen).Use("bedrock", "claude-3-5-sonnet").
//	    When(genai.TaskSummarize).Use("vertex", "gemini-2.0-flash").
//	    Default("bedrock", "claude-3-5-haiku")
type RuleBasedStrategy struct {
	rules       []ruleEntry
	def         *ruleEntry
	pendingTask *TaskType
}

// NewRuleBasedStrategy creates an empty rule-based strategy.
func NewRuleBasedStrategy() *RuleBasedStrategy {
	return &RuleBasedStrategy{}
}

// When begins a rule for the given task type. Complete it with Use.
func (s *RuleBasedStrategy) When(t TaskType) *RuleBasedStrategy {
	tt := t
	s.pendingTask = &tt
	return s
}

// Use completes the pending When rule, targeting (provider, model). Calling Use
// without a preceding When is a no-op.
func (s *RuleBasedStrategy) Use(provider, model string) *RuleBasedStrategy {
	if s.pendingTask == nil {
		return s
	}
	s.rules = append(s.rules, ruleEntry{task: *s.pendingTask, provider: provider, model: model})
	s.pendingTask = nil
	return s
}

// Default sets the fallback rule used when no task-specific rule matches.
func (s *RuleBasedStrategy) Default(provider, model string) *RuleBasedStrategy {
	s.def = &ruleEntry{provider: provider, model: model}
	return s
}

// Name implements RoutingStrategy.
func (s *RuleBasedStrategy) Name() string { return "rule_based" }

// Rank implements RoutingStrategy.
func (s *RuleBasedStrategy) Rank(task Task, reg ModelRegistry) []RouteCandidate {
	var matches []ruleEntry
	for _, r := range s.rules {
		if r.task == task.Type {
			matches = append(matches, r)
		}
	}
	if len(matches) == 0 && s.def != nil {
		matches = append(matches, *s.def)
	}

	est := defaultOutputEstimate(task)
	out := make([]RouteCandidate, 0, len(matches))
	for i, r := range matches {
		c := RouteCandidate{
			Provider: r.provider,
			Model:    r.model,
			// Decreasing score preserves declaration order after any stable sort.
			Score: float64(len(matches) - i),
		}
		if info, ok := reg.Get(r.provider, r.model); ok {
			c.EstCost = info.EstCost(task.EstInputTokens, est)
		}
		out = append(out, c)
	}
	return out
}

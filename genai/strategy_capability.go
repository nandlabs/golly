package genai

import (
	"fmt"
	"sort"
)

// CapabilityStrategy filters the registry to models that satisfy a task's
// required capabilities, input-size limit, and (if set) cost budget, then ranks
// the survivors by the task's priority. It returns the whole ranked list, so
// the router can use the tail as fallbacks — this is what makes capability
// routing actually resilient.
type CapabilityStrategy struct{}

// NewCapabilityStrategy creates a CapabilityStrategy.
func NewCapabilityStrategy() *CapabilityStrategy { return &CapabilityStrategy{} }

// Name implements RoutingStrategy.
func (s *CapabilityStrategy) Name() string { return "capability" }

// Rank implements RoutingStrategy.
func (s *CapabilityStrategy) Rank(task Task, reg ModelRegistry) []RouteCandidate {
	ranked, _ := s.rankExplain(task, reg)
	return ranked
}

// rankExplain implements the internal explainer interface, returning both the
// ranked candidates and the exclusions with reasons.
func (s *CapabilityStrategy) rankExplain(task Task, reg ModelRegistry) ([]RouteCandidate, []Exclusion) {
	est := defaultOutputEstimate(task)
	var excluded []Exclusion
	scored := make([]RouteCandidate, 0)

	for _, m := range reg.All() {
		if reason, ok := s.exclude(m, task, est); ok {
			excluded = append(excluded, Exclusion{Provider: m.Provider, Model: m.Name, Reason: reason})
			continue
		}
		scored = append(scored, RouteCandidate{
			Provider: m.Provider,
			Model:    m.Name,
			EstCost:  m.EstCost(task.EstInputTokens, est),
			Score:    score(m, task.Priority, task.EstInputTokens, est),
		})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].Score != scored[j].Score {
			return scored[i].Score > scored[j].Score
		}
		// Deterministic tie-break: cheaper first, then name.
		if scored[i].EstCost != scored[j].EstCost {
			return scored[i].EstCost < scored[j].EstCost
		}
		return scored[i].Provider+"/"+scored[i].Model < scored[j].Provider+"/"+scored[j].Model
	})
	return scored, excluded
}

// exclude reports whether a model fails the task's hard filters, with a reason.
func (s *CapabilityStrategy) exclude(m ModelInfo, task Task, estOut int) (string, bool) {
	for _, req := range task.RequiredCapabilities {
		if !m.Has(req) {
			return fmt.Sprintf("missing cap: %s", req), true
		}
	}
	if task.EstInputTokens > 0 && m.MaxInputTokens > 0 && task.EstInputTokens > m.MaxInputTokens {
		return "ctx too small", true
	}
	if task.MaxCostUSD > 0 && m.EstCost(task.EstInputTokens, estOut) > task.MaxCostUSD {
		return "over budget", true
	}
	return "", false
}

// score computes a priority-driven ranking score (higher = better). Cost ranks
// on the blended EstCost so a cheap-input/expensive-output model cannot outrank
// a balanced cheaper-total model.
func score(m ModelInfo, priority RoutingPriority, in, out int) float64 {
	switch priority {
	case PrioritySpeed:
		// Smaller context ~ faster: negate size so smaller ranks higher.
		return -float64(m.MaxInputTokens)
	case PriorityQuality:
		// Larger context ~ higher tier.
		return float64(m.MaxInputTokens)
	case PriorityCost:
		fallthrough
	default:
		// Cheaper total cost ranks higher (negate so lower cost => higher score).
		return -m.EstCost(in, out)
	}
}

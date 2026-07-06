package genai

// CompositeStrategy tries its sub-strategies in order; the first that returns a
// non-empty ranking wins. This lets operators layer policy, e.g. a
// sensitive-tag strategy, then capability ranking, then rule fallbacks.
//
//	genai.NewCompositeStrategy(sensitiveTagStrategy, capabilityStrategy, ruleStrategy)
type CompositeStrategy struct {
	strategies []RoutingStrategy
}

// NewCompositeStrategy creates a CompositeStrategy over the given strategies,
// tried in order.
func NewCompositeStrategy(strategies ...RoutingStrategy) *CompositeStrategy {
	return &CompositeStrategy{strategies: strategies}
}

// Name implements RoutingStrategy.
func (s *CompositeStrategy) Name() string { return "composite" }

// Rank implements RoutingStrategy, returning the first non-empty sub-ranking.
func (s *CompositeStrategy) Rank(task Task, reg ModelRegistry) []RouteCandidate {
	for _, sub := range s.strategies {
		if ranked := sub.Rank(task, reg); len(ranked) > 0 {
			return ranked
		}
	}
	return nil
}

// rankExplain implements the internal explainer interface, delegating to the
// firing sub-strategy so exclusions are preserved when possible.
func (s *CompositeStrategy) rankExplain(task Task, reg ModelRegistry) ([]RouteCandidate, []Exclusion) {
	var lastExcluded []Exclusion
	for _, sub := range s.strategies {
		if ex, ok := sub.(explainer); ok {
			ranked, excluded := ex.rankExplain(task, reg)
			if len(ranked) > 0 {
				return ranked, excluded
			}
			lastExcluded = excluded
			continue
		}
		if ranked := sub.Rank(task, reg); len(ranked) > 0 {
			return ranked, nil
		}
	}
	return nil, lastExcluded
}

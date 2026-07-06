package genai

// RouteCandidate is a scored routing option produced by a strategy. Higher Score is
// better; the router turns the best-ranked candidate into the primary and the
// tail into ordered fallbacks.
type RouteCandidate struct {
	Provider string
	Model    string
	Score    float64 // strategy-defined; higher = better
	EstCost  float64 // estimated USD per-call cost, for the decision trace
}

// Route is the executable plan: a primary plus ordered fallbacks derived from a
// strategy's ranked candidate list.
type Route struct {
	Primary   RouteCandidate
	Fallbacks []RouteCandidate
}

// Exclusion records a model a strategy considered but filtered out, with the
// reason, for the decision trace.
type Exclusion struct {
	Provider string
	Model    string
	Reason   string // e.g. "missing cap: vision", "over budget", "ctx too small"
}

// RouteDecision is the full, side-effect-free trace of a routing decision:
// which strategy fired, what it chose, what it would fall over to, everything
// it ranked, and what it excluded and why. Resolve returns it; Route logs a
// compact form via l3.
type RouteDecision struct {
	Strategy   string           // which strategy fired
	Chosen     RouteCandidate   // the primary
	Fallbacks  []RouteCandidate // ordered fallbacks
	Considered []RouteCandidate // everything the strategy ranked
	Excluded   []Exclusion      // models filtered out, with reasons
	Pinned     bool             // true when Task.Pin bypassed strategy evaluation
}

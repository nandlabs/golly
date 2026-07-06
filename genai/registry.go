package genai

import (
	"fmt"
	"sort"
)

// ModelRegistry is a read-only, queryable view of the models available to the
// router. It is the seam where authoritative provider capability metadata is
// merged with operator policy (RoutingConfig). Implementations are immutable
// after construction; hot-reload swaps a whole new registry atomically.
type ModelRegistry interface {
	// All returns every model, stable-sorted by provider then name.
	All() []ModelInfo
	// Get returns a single model by (provider, model), if present.
	Get(provider, model string) (ModelInfo, bool)
	// Filter returns every model matching the predicate, in All() order.
	Filter(pred func(ModelInfo) bool) []ModelInfo
}

// regKey builds the composite registry key for a (provider, model) pair.
func regKey(provider, model string) string { return provider + "/" + model }

// immutableRegistry is the concrete, immutable ModelRegistry.
type immutableRegistry struct {
	byKey  map[string]ModelInfo
	sorted []ModelInfo // stable-sorted snapshot for All()/Filter()
}

func newImmutableRegistry(merged map[string]ModelInfo) *immutableRegistry {
	sorted := make([]ModelInfo, 0, len(merged))
	for _, m := range merged {
		sorted = append(sorted, m)
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Provider != sorted[j].Provider {
			return sorted[i].Provider < sorted[j].Provider
		}
		return sorted[i].Name < sorted[j].Name
	})
	// Rebuild the map so byKey and sorted share identical values.
	byKey := make(map[string]ModelInfo, len(sorted))
	for _, m := range sorted {
		byKey[regKey(m.Provider, m.Name)] = m
	}
	return &immutableRegistry{byKey: byKey, sorted: sorted}
}

func (r *immutableRegistry) All() []ModelInfo {
	out := make([]ModelInfo, len(r.sorted))
	copy(out, r.sorted)
	return out
}

func (r *immutableRegistry) Get(provider, model string) (ModelInfo, bool) {
	m, ok := r.byKey[regKey(provider, model)]
	return m, ok
}

func (r *immutableRegistry) Filter(pred func(ModelInfo) bool) []ModelInfo {
	out := make([]ModelInfo, 0, len(r.sorted))
	for _, m := range r.sorted {
		if pred(m) {
			out = append(out, m)
		}
	}
	return out
}

// NewModelRegistry composes the provider base layer (authoritative capability
// metadata) with the operator config overlay (policy) into one immutable
// registry. It fails if a custom model names a provider that is not present in
// the set (a phantom model that could never be served).
func NewModelRegistry(set *ProviderSet, cfg RoutingConfig) (ModelRegistry, error) {
	return buildRegistry(set, cfg)
}

// buildRegistry applies the deterministic five-step merge described in the
// design doc §5.2:
//
//  1. Base       — collect ModelCatalog() from every CapabilityProvider.
//  2. Cost       — override Input/OutputCostPerMTok for named models.
//  3. Cap disable — strip listed capabilities (org declines a claimed cap).
//  4. Model disable — drop the (provider, model) entirely.
//  5. Custom     — append config-defined models iff their provider is present,
//     else fail the build.
func buildRegistry(set *ProviderSet, cfg RoutingConfig) (ModelRegistry, error) {
	if set == nil {
		set = NewProviderSet()
	}
	merged := map[string]ModelInfo{}

	// 1. Base layer.
	for _, p := range set.All() {
		if cp, ok := p.(CapabilityProvider); ok {
			for _, m := range cp.ModelCatalog() {
				merged[regKey(m.Provider, m.Name)] = m
			}
		}
	}

	// 2. Cost overrides.
	for _, ov := range cfg.CostOverrides {
		k := regKey(ov.Provider, ov.Model)
		if m, ok := merged[k]; ok {
			m.InputCostPerMTok = ov.InputCostPerMTok
			m.OutputCostPerMTok = ov.OutputCostPerMTok
			merged[k] = m
		}
	}

	// 3. Capability disable.
	for _, dc := range cfg.DisabledCapabilities {
		k := regKey(dc.Provider, dc.Model)
		if m, ok := merged[k]; ok {
			m.Capabilities = removeCaps(m.Capabilities, dc.Capabilities)
			merged[k] = m
		}
	}

	// 4. Model disable.
	for _, dm := range cfg.DisabledModels {
		delete(merged, regKey(dm.Provider, dm.Model))
	}

	// 5. Custom models (capabilities allowed here only; provider must exist).
	for _, cm := range cfg.CustomModels {
		if _, ok := set.Get(cm.Provider); !ok {
			return nil, fmt.Errorf("genai: custom model %q names unregistered provider %q", cm.Name, cm.Provider)
		}
		merged[regKey(cm.Provider, cm.Name)] = cm.ToModelInfo()
	}

	return newImmutableRegistry(merged), nil
}

// removeCaps returns caps with every capability in drop removed.
func removeCaps(caps, drop []Capability) []Capability {
	if len(drop) == 0 {
		return caps
	}
	dropSet := make(map[Capability]struct{}, len(drop))
	for _, d := range drop {
		dropSet[d] = struct{}{}
	}
	out := make([]Capability, 0, len(caps))
	for _, c := range caps {
		if _, gone := dropSet[c]; !gone {
			out = append(out, c)
		}
	}
	return out
}

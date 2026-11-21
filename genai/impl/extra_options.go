package impl

import (
	"encoding/json"
	"fmt"
	"sort"

	"oss.nandlabs.io/golly/genai"
)

// forwardEnabled resolves a provider's ForwardUnknownOptions setting.
// A nil value means the default, which is to forward.
func forwardEnabled(v *bool) bool {
	return v == nil || *v
}

// marshalRequestBody encodes the typed request body and, when forward is true,
// adds every custom option (any name that is not a genai standard option) as an
// extra top-level field of the JSON object.
//
// An extra field never overrides a field the provider has already set in the
// typed body (model, messages, stream, ...); such a conflicting option is
// skipped and logged. Fields the provider left empty (omitted by omitempty)
// can be supplied by a custom option, e.g. Claude's "metadata".
//
// The typed body's fields are kept as raw JSON, so numbers are not re-encoded.
func marshalRequestBody(body any, options *genai.Options, forward bool, provider string) ([]byte, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	if !forward || options == nil {
		return raw, nil
	}

	var extraNames []string
	for name := range options.All() {
		if !genai.IsStandardOption(name) {
			extraNames = append(extraNames, name)
		}
	}
	if len(extraNames) == 0 {
		return raw, nil
	}
	sort.Strings(extraNames)

	fields := make(map[string]json.RawMessage)
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, fmt.Errorf("failed to prepare request for custom options: %w", err)
	}

	for _, name := range extraNames {
		if _, exists := fields[name]; exists {
			genai.LOGGER.WarnF("genai/%s: custom option %q conflicts with a request field set by the provider; ignored", provider, name)
			continue
		}
		v, err := json.Marshal(options.Get(name))
		if err != nil {
			return nil, fmt.Errorf("failed to marshal custom option %q: %w", name, err)
		}
		fields[name] = v
	}

	out, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	return out, nil
}

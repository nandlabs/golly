package genai

import "context"

// EmbedRequest is a request to generate one or more embedding vectors.
// Inputs reuse the multimodal Part shape from genai messages — most providers
// accept text parts today; image/audio embeddings are a follow-on.
type EmbedRequest struct {
	// Model is the embedding-model id (e.g. "text-embedding-3-small",
	// "nomic-embed-text"). Empty selects the provider's default.
	Model string `json:"model,omitempty" yaml:"model,omitempty" toml:"model,omitempty"`
	// Inputs is the list of inputs to embed; one vector per input is returned.
	Inputs []Part `json:"inputs" yaml:"inputs" toml:"inputs"`
}

// EmbedResponse holds the resulting vectors aligned with EmbedRequest.Inputs.
type EmbedResponse struct {
	Vectors [][]float32   `json:"vectors" yaml:"vectors" toml:"vectors"`
	Meta    *ResponseMeta `json:"meta,omitempty" yaml:"meta,omitempty" toml:"meta,omitempty"`
}

// Embedder is the optional interface a Provider may implement to expose
// embedding-model access. Callers should type-assert on Provider.
//
//	if e, ok := provider.(genai.Embedder); ok {
//	    resp, err := e.Embed(ctx, &genai.EmbedRequest{...})
//	}
type Embedder interface {
	Embed(ctx context.Context, req *EmbedRequest) (*EmbedResponse, error)
}

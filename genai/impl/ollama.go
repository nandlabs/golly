package impl

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"oss.nandlabs.io/golly/clients"
	"oss.nandlabs.io/golly/genai"
	"oss.nandlabs.io/golly/ioutils"
	"oss.nandlabs.io/golly/rest"
)

const (
	// OllamaProviderName is the name of the Ollama provider
	OllamaProviderName = "ollama"
	// OllamaProviderVersion is the version of the Ollama provider
	OllamaProviderVersion = "1.0.0"
	// OllamaProviderDescription is the description of the Ollama provider
	OllamaProviderDescription = "Ollama provider for local model inference (OpenAI-compatible API)"
	// OllamaDefaultBaseURL is the default base URL for the Ollama OpenAI-compatible API
	OllamaDefaultBaseURL = "http://localhost:11434/v1"
)

// OllamaProvider wraps OpenAIProvider to talk to Ollama's OpenAI-compatible endpoint.
// Ollama (v0.1.14+) exposes /v1/chat/completions with the same request/response
// format as OpenAI, so the full OpenAI implementation is reused.
type OllamaProvider struct {
	*OpenAIProvider
}

// OllamaProviderConfig contains configuration for the Ollama provider
type OllamaProviderConfig struct {
	// Auth is an optional authentication provider. Set this when Ollama is
	// deployed behind an authenticated reverse proxy or gateway.
	// For example, use clients.NewBasicAuth(user, pass) for HTTP Basic Auth,
	// clients.NewBearerAuth(token) for Bearer tokens, or
	// clients.NewAPIKeyAuth("X-API-Key", key) for custom header auth.
	// Leave nil for unauthenticated local instances.
	Auth clients.AuthProvider
	// BaseURL is the Ollama server URL (default: http://localhost:11434/v1)
	BaseURL string
	// Models is the list of available models
	Models []string
	// Description is a custom description
	Description string
	// Version is a custom version
	Version string
	// ExtraHeaders are additional HTTP headers to include with every request.
	ExtraHeaders map[string]string
}

// NewOllamaProvider creates a new Ollama provider with default settings and the given REST client options.
// No API key is required for a local Ollama instance.
func NewOllamaProvider(opts *rest.ClientOpts) *OllamaProvider {
	openai := NewOllamaProviderWithConfig(&OllamaProviderConfig{}, opts)
	return openai
}

// NewOllamaProviderWithConfig creates a new Ollama provider with custom configuration.
func NewOllamaProviderWithConfig(config *OllamaProviderConfig, opts *rest.ClientOpts) *OllamaProvider {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = OllamaDefaultBaseURL
	}

	description := config.Description
	if description == "" {
		description = OllamaProviderDescription
	}

	version := config.Version
	if version == "" {
		version = OllamaProviderVersion
	}

	openai := NewOpenAIProviderWithConfig(&OpenAIProviderConfig{
		Auth:         config.Auth, // nil for unauthenticated local instances
		BaseURL:      baseURL,
		Models:       config.Models,
		Description:  description,
		Version:      version,
		ExtraHeaders: config.ExtraHeaders,
	}, opts)

	return &OllamaProvider{OpenAIProvider: openai}
}

// Name returns the name of the provider, overriding OpenAIProvider.
func (o *OllamaProvider) Name() string {
	return OllamaProviderName
}

// Generate delegates to the embedded OpenAIProvider.
func (o *OllamaProvider) Generate(ctx context.Context, model string, message *genai.Message, options *genai.Options) (*genai.GenResponse, error) {
	return o.OpenAIProvider.Generate(ctx, model, message, options)
}

// GenerateStream delegates to the embedded OpenAIProvider.
func (o *OllamaProvider) GenerateStream(ctx context.Context, model string, message *genai.Message, options *genai.Options) (<-chan *genai.GenResponse, <-chan error) {
	return o.OpenAIProvider.GenerateStream(ctx, model, message, options)
}

// --- Embeddings (native Ollama /api/embed) ---

// ollamaEmbedRequest is the native Ollama batch-embedding request shape.
type ollamaEmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// ollamaEmbedResponse is the native Ollama embedding response shape.
type ollamaEmbedResponse struct {
	Model           string      `json:"model"`
	Embeddings      [][]float32 `json:"embeddings"`
	TotalDuration   int64       `json:"total_duration,omitempty"` // nanoseconds
	LoadDuration    int64       `json:"load_duration,omitempty"`  // nanoseconds
	PromptEvalCount int         `json:"prompt_eval_count,omitempty"`
}

// Embed implements genai.Embedder using Ollama's native batch endpoint
// (POST /api/embed). The OpenAI-compatible /v1 endpoint does not surface
// the same batching guarantees, so the native path is preferred.
func (o *OllamaProvider) Embed(ctx context.Context, req *genai.EmbedRequest) (*genai.EmbedResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("embed request is nil")
	}
	inputs := partsToTextInputs(req.Inputs)
	if len(inputs) == 0 {
		return nil, fmt.Errorf("embed request has no text inputs")
	}

	url := fmt.Sprintf("%s/api/embed", ollamaRootURL(o.baseURL))
	httpReq, err := o.client.NewRequest(url, http.MethodPost)
	if err != nil {
		return nil, fmt.Errorf("failed to create embed request: %w", err)
	}
	if _, err = httpReq.WithContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to set context: %w", err)
	}
	httpReq.SetBody(&ollamaEmbedRequest{Model: req.Model, Input: inputs})
	httpReq.SetContentType(ioutils.MimeApplicationJSON)

	resp, err := o.client.Execute(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama embeddings request failed: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("ollama embeddings API error: status %d", resp.StatusCode())
	}

	var embedResp ollamaEmbedResponse
	if err := resp.Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode embed response: %w", err)
	}

	out := &genai.EmbedResponse{
		Vectors: embedResp.Embeddings,
	}
	if embedResp.PromptEvalCount > 0 || embedResp.TotalDuration > 0 {
		out.Meta = &genai.ResponseMeta{
			InputTokens: embedResp.PromptEvalCount,
			TotalTokens: embedResp.PromptEvalCount,
			TotalTime:   embedResp.TotalDuration / 1_000_000, // ns → ms
		}
	}
	return out, nil
}

// ollamaRootURL strips a trailing "/v1" (or "/v1/") from the configured base
// URL so the native Ollama endpoints (which live at the root, not under /v1)
// can be addressed.
func ollamaRootURL(baseURL string) string {
	trimmed := strings.TrimRight(baseURL, "/")
	return strings.TrimSuffix(trimmed, "/v1")
}

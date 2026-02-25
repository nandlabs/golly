package impl

import (
	"context"

	"oss.nandlabs.io/golly/clients"
	"oss.nandlabs.io/golly/genai"
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

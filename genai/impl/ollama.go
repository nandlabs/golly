package impl

import (
	"context"

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
	// BaseURL is the Ollama server URL (default: http://localhost:11434/v1)
	BaseURL string
	// Models is the list of available models
	Models []string
	// Description is a custom description
	Description string
	// Version is a custom version
	Version string
}

// NewOllamaProvider creates a new Ollama provider with default settings and the given REST client options.
// No API key is required for a local Ollama instance.
func NewOllamaProvider(opts *rest.ClientOpts) *OllamaProvider {
	openai := NewOpenAIProviderWithConfig(&OpenAIProviderConfig{
		APIKey:      "", // Ollama does not require an API key
		BaseURL:     OllamaDefaultBaseURL,
		Models:      []string{},
		Description: OllamaProviderDescription,
		Version:     OllamaProviderVersion,
	}, opts)

	return &OllamaProvider{OpenAIProvider: openai}
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
		APIKey:      "", // Ollama does not require an API key
		BaseURL:     baseURL,
		Models:      config.Models,
		Description: description,
		Version:     version,
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

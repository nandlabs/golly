# Ollama Provider for GenAI

This package implements the Ollama provider for the GenAI framework, enabling local model inference through the Ollama API.

## Features

- **Full Ollama API Support**: Implements chat completions with streaming support
- **Multi-modal Support**: Handle text, images (base64), and function calls
- **Flexible Configuration**: Customizable base URL, models list, and metadata
- **Advanced Options**: Support for temperature, top-p, top-k, seed, and more
- **Streaming Responses**: Real-time token generation via Go channels
- **JSON Mode**: Structured output support for JSON responses

## Installation

```bash
go get oss.nandlabs.io/golly/genai
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "oss.nandlabs.io/golly/genai"
    "oss.nandlabs.io/golly/genai/impl"
    "oss.nandlabs.io/golly/rest"
)

func main() {
    // Create Ollama provider with default settings (http://localhost:11434)
    opts := rest.CliOptsBuilder().Build()
    provider := impl.NewOllamaProvider(opts)
    defer provider.Close()

    // Create a simple user message
    message := genai.NewTextMessage(genai.RoleUser, "Why is the sky blue?")

    // Generate response
    response, err := provider.Generate("llama3.2", message, nil)
    if err != nil {
        panic(err)
    }

    // Print the response
    if len(response.Candidates) > 0 && response.Candidates[0].Message != nil {
        for _, part := range response.Candidates[0].Message.Parts {
            if part.Text != nil {
                fmt.Println(part.Text.Text)
            }
        }
    }
}
```

### Custom Configuration

```go
// Create provider with custom configuration
config := &impl.OllamaProviderConfig{
    BaseURL:     "http://custom-host:11434",
    Models:      []string{"llama3.2", "mistral", "codellama"},
    Description: "Production Ollama Instance",
    Version:     "1.0.0",
}

opts := rest.CliOptsBuilder().Build()
provider := impl.NewOllamaProviderWithConfig(config, opts)
defer provider.Close()
```

### Streaming Responses

```go
message := genai.NewTextMessage(genai.RoleUser, "Write a short story")

responseChan, errorChan := provider.GenerateStream("llama3.2", message, nil)

for {
    select {
    case resp, ok := <-responseChan:
        if !ok {
            return // Stream completed
        }
        if len(resp.Candidates) > 0 && resp.Candidates[0].Message != nil {
            for _, part := range resp.Candidates[0].Message.Parts {
                if part.Text != nil {
                    fmt.Print(part.Text.Text)
                }
            }
        }
    case err := <-errorChan:
        if err != nil {
            fmt.Println("Error:", err)
            return
        }
    }
}
```

### Advanced Options

```go
// Configure generation parameters
options := &genai.Options{}
options.Set(genai.OptionTemperature, 0.7)      // Creativity (0.0 - 1.0)
options.Set(genai.OptionTopP, 0.9)             // Nucleus sampling
options.Set(genai.OptionTopK, 40)              // Top-k sampling
options.Set(genai.OptionMaxTokens, 500)        // Maximum response length
options.Set(genai.OptionSeed, 42)              // Reproducible outputs
options.Set(genai.OptionRepetitionPenalty, 1.2) // Reduce repetition
options.Set(genai.OptionFrequencyPenalty, 0.5) // Penalize frequent tokens
options.Set(genai.OptionPresencePenalty, 0.3)  // Encourage diversity
options.Set(genai.OptionStopWords, []string{"\n\n", "END"}) // Stop sequences

message := genai.NewTextMessage(genai.RoleUser, "Tell me a joke")
response, err := provider.Generate("llama3.2", message, options)
```

### JSON Mode

```go
// Request structured JSON output
options := &genai.Options{}
options.Set(genai.OptionOutputMime, "application/json")

message := genai.NewTextMessage(
    genai.RoleUser,
    "List 3 programming languages with their release years in JSON format",
)

response, err := provider.Generate("llama3.2", message, options)
```

### Multi-modal Messages (Images)

```go
// Create a message with text and image
message := &genai.Message{
    Role: genai.RoleUser,
    Parts: []genai.Part{
        {
            Text: &genai.TextPart{Text: "What's in this image?"},
        },
        {
            MimeType: "image/png",
            Bin: &genai.BinPart{
                Data: imageBytes, // Your base64-encoded image
            },
        },
    },
}

response, err := provider.Generate("llava", message, nil)
```

### Conversation History

```go
// Build a conversation with multiple turns
messages := []genai.Message{
    {
        Role: genai.RoleUser,
        Parts: []genai.Part{
            {Text: &genai.TextPart{Text: "What is Rayleigh scattering?"}},
        },
    },
    {
        Role: genai.RoleAssistant,
        Parts: []genai.Part{
            {Text: &genai.TextPart{Text: "Rayleigh scattering is the scattering of light by particles much smaller than the wavelength of light..."}},
        },
    },
    {
        Role: genai.RoleUser,
        Parts: []genai.Part{
            {Text: &genai.TextPart{Text: "How is that different from Mie scattering?"}},
        },
    },
}

// Send the latest message with context
response, err := provider.Generate("llama3.2", &messages[len(messages)-1], nil)
```

## API Reference

### Types

#### OllamaProvider
```go
type OllamaProvider struct {
    // Implementation details
}
```

Implements the `genai.Provider` interface with Ollama-specific functionality.

#### OllamaProviderConfig
```go
type OllamaProviderConfig struct {
    BaseURL     string   // Base URL for Ollama API (default: http://localhost:11434)
    Models      []string // List of available models
    Description string   // Custom description
    Version     string   // Custom version
}
```

### Functions

#### NewOllamaProvider
```go
func NewOllamaProvider(opts *rest.ClientOpts) *OllamaProvider
```
Creates a new Ollama provider with default settings.

#### NewOllamaProviderWithConfig
```go
func NewOllamaProviderWithConfig(config *OllamaProviderConfig, opts *rest.ClientOpts) *OllamaProvider
```
Creates a new Ollama provider with custom configuration.

### Interface Methods

#### Name
```go
func (o *OllamaProvider) Name() string
```
Returns "ollama".

#### Description
```go
func (o *OllamaProvider) Description() string
```
Returns the provider description.

#### Version
```go
func (o *OllamaProvider) Version() string
```
Returns the provider version.

#### Models
```go
func (o *OllamaProvider) Models() []string
```
Returns the list of supported model IDs.

#### Generate
```go
func (o *OllamaProvider) Generate(model string, message *genai.Message, options *genai.Options) (*genai.GenResponse, error)
```
Generates a single response for the given model and message.

#### GenerateStream
```go
func (o *OllamaProvider) GenerateStream(model string, message *genai.Message, options *genai.Options) (<-chan *genai.GenResponse, <-chan error)
```
Generates a streaming response via channels.

#### Close
```go
func (o *OllamaProvider) Close() error
```
Closes the provider and releases resources.

## Supported Options

| Option | Type | Description |
|--------|------|-------------|
| `OptionTemperature` | float64 | Sampling temperature (0.0-1.0) |
| `OptionTopP` | float64 | Nucleus sampling threshold |
| `OptionTopK` | int | Top-k sampling parameter |
| `OptionMaxTokens` | int | Maximum tokens to generate |
| `OptionSeed` | int | Random seed for reproducibility |
| `OptionRepetitionPenalty` | float64 | Penalty for repeated tokens |
| `OptionFrequencyPenalty` | float64 | Penalty based on token frequency |
| `OptionPresencePenalty` | float64 | Penalty for token presence |
| `OptionStopWords` | []string | Stop sequences |
| `OptionOutputMime` | string | Output format ("application/json") |

## Response Metadata

The `GenResponse.Meta` field contains:

- `InputTokens`: Number of tokens in the prompt
- `OutputTokens`: Number of tokens generated
- `TotalTokens`: Total tokens (input + output)
- `TotalTime`: Total generation time in milliseconds

## Error Handling

```go
response, err := provider.Generate("llama3.2", message, nil)
if err != nil {
    // Handle errors:
    // - Connection errors (Ollama not running)
    // - Model not found
    // - Invalid parameters
    // - API errors
    fmt.Println("Error:", err)
    return
}
```

## Best Practices

1. **Close the Provider**: Always defer `provider.Close()` to clean up resources
2. **Handle Streaming Errors**: Monitor the error channel when streaming
3. **Model Selection**: Ensure the model is pulled/available in Ollama before use
4. **Temperature**: Use lower values (0.1-0.3) for deterministic tasks, higher (0.7-1.0) for creative tasks
5. **Token Limits**: Set appropriate `OptionMaxTokens` to control response length and costs
6. **Reproducibility**: Set `OptionSeed` for consistent outputs across runs

## Ollama Setup

To use this provider, you need Ollama installed and running:

```bash
# Install Ollama (macOS)
brew install ollama

# Start Ollama service
ollama serve

# Pull a model
ollama pull llama3.2
ollama pull mistral
ollama pull llava  # For multi-modal support
```

For other platforms, see: https://ollama.ai/download

## Examples

See the `examples/` directory for complete working examples:
- Basic chat
- Streaming responses
- JSON mode
- Multi-modal inputs
- Conversation history

## Contributing

Contributions are welcome! Please see CONTRIBUTING.md for guidelines.

## License

See LICENSE file for details.

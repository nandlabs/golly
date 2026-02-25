# OpenAI Provider

The OpenAI provider implements the `genai.Provider` interface for the [OpenAI Chat Completions API](https://platform.openai.com/docs/api-reference/chat). It works with any OpenAI-compatible endpoint, including Azure OpenAI, vLLM, LiteLLM, and other compatible services.

---

- [Quick Start](#quick-start)
- [Configuration](#configuration)
  - [Simple Constructor](#simple-constructor)
  - [Full Configuration](#full-configuration)
  - [Config Reference](#config-reference)
- [Authentication](#authentication)
  - [Standard OpenAI](#standard-openai)
  - [Azure OpenAI](#azure-openai)
  - [Custom Auth Provider](#custom-auth-provider)
- [Supported Options](#supported-options)
- [Generating Responses](#generating-responses)
  - [Basic Generation](#basic-generation)
  - [System Instructions](#system-instructions)
  - [Streaming](#streaming)
  - [JSON Output](#json-output)
  - [Multi-Modal (Vision)](#multi-modal-vision)
  - [Tool / Function Calling](#tool--function-calling)
- [Response Structure](#response-structure)
- [Error Handling](#error-handling)

---

## Quick Start

```go
import (
    "context"
    "fmt"

    "oss.nandlabs.io/golly/genai"
    "oss.nandlabs.io/golly/genai/impl"
)

provider := impl.NewOpenAIProvider("sk-...", nil)
defer provider.Close()

msg := genai.NewTextMessage(genai.RoleUser, "Hello, what can you do?")
resp, err := provider.Generate(context.Background(), "gpt-4o", msg, nil)
if err != nil {
    panic(err)
}
for _, c := range resp.Candidates {
    for _, p := range c.Message.Parts {
        if p.Text != nil {
            fmt.Println(p.Text.Text)
        }
    }
}
```

## Configuration

### Simple Constructor

```go
// Uses Bearer token auth (Authorization: Bearer sk-...)
provider := impl.NewOpenAIProvider("sk-...", nil)
```

The simple constructor wraps the API key with `clients.NewBearerAuth(apiKey)` and uses default settings.

### Full Configuration

```go
import (
    "oss.nandlabs.io/golly/clients"
    "oss.nandlabs.io/golly/genai/impl"
    "oss.nandlabs.io/golly/rest"
)

provider := impl.NewOpenAIProviderWithConfig(&impl.OpenAIProviderConfig{
    Auth:        clients.NewBearerAuth("sk-..."),
    OrgID:       "org-...",
    BaseURL:     "https://api.openai.com/v1",
    Models:      []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo"},
    Description: "Production OpenAI",
    Version:     "2.0.0",
    ExtraHeaders: map[string]string{
        "X-Request-Source": "my-app",
    },
}, &rest.ClientOpts{
    // Optional: configure retry, circuit breaker, timeouts, etc.
})
```

### Config Reference

| Field          | Type                   | Default                                         | Description                                                   |
| -------------- | ---------------------- | ----------------------------------------------- | ------------------------------------------------------------- |
| `Auth`         | `clients.AuthProvider` | —                                               | Authentication provider (required)                            |
| `OrgID`        | `string`               | `""`                                            | OpenAI organization ID (sent as `OpenAI-Organization` header) |
| `BaseURL`      | `string`               | `https://api.openai.com/v1`                     | API base URL                                                  |
| `Models`       | `[]string`             | `nil`                                           | List of available model IDs (informational)                   |
| `Description`  | `string`               | `"OpenAI API provider for GPT model inference"` | Provider description                                          |
| `Version`      | `string`               | `"1.0.0"`                                       | Provider version                                              |
| `ExtraHeaders` | `map[string]string`    | `nil`                                           | Additional HTTP headers on every request                      |

## Authentication

### Standard OpenAI

```go
// Convenience — wraps key as Bearer token automatically
provider := impl.NewOpenAIProvider("sk-...", nil)

// Equivalent explicit config
provider = impl.NewOpenAIProviderWithConfig(&impl.OpenAIProviderConfig{
    Auth: clients.NewBearerAuth("sk-..."),
}, nil)
```

### Azure OpenAI

Azure OpenAI uses an `api-key` header instead of Bearer token, and a custom base URL:

```go
provider := impl.NewOpenAIProviderWithConfig(&impl.OpenAIProviderConfig{
    Auth:    clients.NewAPIKeyAuth("api-key", "your-azure-key"),
    BaseURL: "https://your-resource.openai.azure.com/openai/deployments/gpt-4o/v1",
    ExtraHeaders: map[string]string{
        "api-version": "2024-02-01",
    },
}, nil)
```

### Custom Auth Provider

For keys stored in Vault, AWS Secrets Manager, or other external stores:

```go
type VaultAuth struct {
    client *vault.Client
    path   string
}

func (v *VaultAuth) Type() clients.AuthType { return clients.AuthTypeBearer }
func (v *VaultAuth) User() (string, error)  { return "", nil }
func (v *VaultAuth) Pass() (string, error)  { return "", nil }
func (v *VaultAuth) Token() (string, error) {
    secret, err := v.client.Logical().Read(v.path)
    if err != nil {
        return "", err
    }
    return secret.Data["api_key"].(string), nil
}

provider := impl.NewOpenAIProviderWithConfig(&impl.OpenAIProviderConfig{
    Auth: &VaultAuth{client: vc, path: "secret/openai"},
}, nil)
```

## Supported Options

| GenAI Option               | OpenAI Parameter       | Type       | Description                                |
| -------------------------- | ---------------------- | ---------- | ------------------------------------------ |
| `OptionMaxTokens`          | `max_tokens`           | `int`      | Maximum tokens in the response             |
| `OptionTemperature`        | `temperature`          | `float32`  | Sampling temperature (0.0–2.0)             |
| `OptionTopP`               | `top_p`                | `float32`  | Nucleus sampling threshold                 |
| `OptionCandidateCount`     | `n`                    | `int`      | Number of completions to generate          |
| `OptionSeed`               | `seed`                 | `int`      | Random seed for deterministic output       |
| `OptionFrequencyPenalty`   | `frequency_penalty`    | `float64`  | Penalise frequent tokens (-2.0–2.0)        |
| `OptionPresencePenalty`    | `presence_penalty`     | `float64`  | Penalise already-present tokens (-2.0–2.0) |
| `OptionStopWords`          | `stop`                 | `[]string` | Stop sequences                             |
| `OptionOutputMime`         | `response_format`      | `string`   | Set to `"application/json"` for JSON mode  |
| `OptionSystemInstructions` | `messages[0]` (system) | `string`   | Prepended as a system message              |

## Generating Responses

### Basic Generation

```go
msg := genai.NewTextMessage(genai.RoleUser, "What is the capital of France?")
opts := genai.NewOptionsBuilder().
    SetMaxTokens(256).
    SetTemperature(0.7).
    Build()

resp, err := provider.Generate(ctx, "gpt-4o", msg, opts)
```

### System Instructions

```go
opts := genai.NewOptionsBuilder().
    SetMaxTokens(512).
    Build()
opts.Set(genai.OptionSystemInstructions, "You are a helpful coding assistant. Always include code examples.")

msg := genai.NewTextMessage(genai.RoleUser, "How do I read a file in Go?")
resp, err := provider.Generate(ctx, "gpt-4o", msg, opts)
```

### Streaming

```go
msg := genai.NewTextMessage(genai.RoleUser, "Write a short story about a robot.")
opts := genai.NewOptionsBuilder().SetMaxTokens(1024).Build()

respCh, errCh := provider.GenerateStream(ctx, "gpt-4o", msg, opts)
for resp := range respCh {
    for _, c := range resp.Candidates {
        for _, p := range c.Message.Parts {
            if p.Text != nil {
                fmt.Print(p.Text.Text) // prints token by token
            }
        }
    }
}
if err := <-errCh; err != nil {
    // handle streaming error
}
```

### JSON Output

```go
opts := genai.NewOptionsBuilder().
    SetMaxTokens(512).
    Build()
opts.Set(genai.OptionOutputMime, "application/json")
opts.Set(genai.OptionSystemInstructions, "Respond with a JSON object containing 'name' and 'capital' fields.")

msg := genai.NewTextMessage(genai.RoleUser, "Tell me about France")
resp, err := provider.Generate(ctx, "gpt-4o", msg, opts)
```

### Multi-Modal (Vision)

```go
// Image from URL
msg := genai.NewTextMessage(genai.RoleUser, "What's in this image?")
genai.AddFilePart(msg, "photo", "https://example.com/photo.jpg", "image/jpeg")

// Image from bytes
msg2 := genai.NewTextMessage(genai.RoleUser, "Describe this diagram.")
genai.AddBinPart(msg2, "diagram", pngBytes, "image/png")

resp, err := provider.Generate(ctx, "gpt-4o", msg, nil)
```

### Tool / Function Calling

The provider automatically maps `FuncCallPart` and `FuncResponsePart` to OpenAI's tool calling format:

```go
// The model may return a response with FuncCall parts
resp, _ := provider.Generate(ctx, "gpt-4o", msg, opts)
for _, c := range resp.Candidates {
    for _, p := range c.Message.Parts {
        if p.FuncCall != nil {
            fmt.Printf("Tool call: %s(%v)\n", p.FuncCall.FunctionName, p.FuncCall.Arguments)
        }
    }
}
```

## Response Structure

```go
resp, _ := provider.Generate(ctx, "gpt-4o", msg, opts)

// Candidates (one per `n` value)
for _, candidate := range resp.Candidates {
    candidate.Index        // 0, 1, ...
    candidate.FinishReason // genai.FinishReasonStop, FinishReasonLength, etc.
    candidate.Message      // *genai.Message with Parts
}

// Token usage metadata
resp.Meta.InputTokens   // prompt tokens
resp.Meta.OutputTokens  // completion tokens
resp.Meta.TotalTokens   // input + output
```

## Error Handling

The provider returns structured errors from the API:

```go
resp, err := provider.Generate(ctx, "gpt-4o", msg, opts)
if err != nil {
    // Error format: "openai API error [error_type]: error message"
    // e.g., "openai API error [invalid_request_error]: model 'gpt-5' does not exist"
    log.Fatal(err)
}
```

For HTTP-level errors (network failures, timeouts), the underlying `rest.Client` error is wrapped with context.

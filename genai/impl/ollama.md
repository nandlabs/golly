# Ollama Provider

The Ollama provider implements the `genai.Provider` interface for [Ollama](https://ollama.com/), an open-source tool for running LLMs locally. It works by wrapping the OpenAI provider, since Ollama (v0.1.14+) exposes an [OpenAI-compatible API](https://ollama.com/blog/openai-compatibility) at `/v1/chat/completions`.

---

- [Quick Start](#quick-start)
- [Prerequisites](#prerequisites)
- [Configuration](#configuration)
  - [Simple Constructor](#simple-constructor)
  - [Full Configuration](#full-configuration)
  - [Config Reference](#config-reference)
- [Authentication](#authentication)
  - [Local (No Auth)](#local-no-auth)
  - [Behind a Reverse Proxy](#behind-a-reverse-proxy)
  - [Behind an API Gateway](#behind-an-api-gateway)
- [Supported Options](#supported-options)
- [Generating Responses](#generating-responses)
  - [Basic Generation](#basic-generation)
  - [System Instructions](#system-instructions)
  - [Streaming](#streaming)
  - [Multi-Modal (Vision)](#multi-modal-vision)
- [Response Structure](#response-structure)
- [Architecture](#architecture)
- [Popular Models](#popular-models)
- [Error Handling](#error-handling)
- [Deployment Patterns](#deployment-patterns)

---

## Quick Start

```go
import (
    "context"
    "fmt"

    "oss.nandlabs.io/golly/genai"
    "oss.nandlabs.io/golly/genai/impl"
)

provider := impl.NewOllamaProvider(nil)
defer provider.Close()

msg := genai.NewTextMessage(genai.RoleUser, "Hello! What is Go?")
resp, err := provider.Generate(context.Background(), "llama3", msg, nil)
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

## Prerequisites

1. **Install Ollama**: https://ollama.com/download
2. **Pull a model**:
   ```sh
   ollama pull llama3
   ```
3. **Start Ollama** (if not running as a service):
   ```sh
   ollama serve
   ```

The default endpoint is `http://localhost:11434`.

## Configuration

### Simple Constructor

```go
// Connects to http://localhost:11434/v1 with no authentication
provider := impl.NewOllamaProvider(nil)
```

### Full Configuration

```go
import (
    "oss.nandlabs.io/golly/clients"
    "oss.nandlabs.io/golly/genai/impl"
    "oss.nandlabs.io/golly/rest"
)

provider := impl.NewOllamaProviderWithConfig(&impl.OllamaProviderConfig{
    Auth:        nil, // no auth for local
    BaseURL:     "http://localhost:11434/v1",
    Models:      []string{"llama3", "mistral", "codellama", "llava"},
    Description: "Local Ollama instance",
    Version:     "1.0.0",
    ExtraHeaders: map[string]string{
        "X-Request-Source": "dev-machine",
    },
}, &rest.ClientOpts{
    // Optional: configure timeouts for long-running local inference
})
```

### Config Reference

| Field          | Type                   | Default                                          | Description                                              |
| -------------- | ---------------------- | ------------------------------------------------ | -------------------------------------------------------- |
| `Auth`         | `clients.AuthProvider` | `nil`                                            | Authentication provider (nil for local, set for proxied) |
| `BaseURL`      | `string`               | `http://localhost:11434/v1`                      | Ollama OpenAI-compatible API URL                         |
| `Models`       | `[]string`             | `nil`                                            | List of available model IDs (informational)              |
| `Description`  | `string`               | `"Ollama provider for local model inference..."` | Provider description                                     |
| `Version`      | `string`               | `"1.0.0"`                                        | Provider version                                         |
| `ExtraHeaders` | `map[string]string`    | `nil`                                            | Additional HTTP headers on every request                 |

## Authentication

### Local (No Auth)

The default — no credentials needed:

```go
provider := impl.NewOllamaProvider(nil)
```

### Behind a Reverse Proxy

When Ollama is deployed behind nginx, Caddy, or another proxy with authentication:

```go
// HTTP Basic Auth proxy
provider := impl.NewOllamaProviderWithConfig(&impl.OllamaProviderConfig{
    Auth:    clients.NewBasicAuth("username", "password"),
    BaseURL: "https://ollama.internal.company.com/v1",
}, nil)

// Bearer token proxy
provider = impl.NewOllamaProviderWithConfig(&impl.OllamaProviderConfig{
    Auth:    clients.NewBearerAuth("my-proxy-token"),
    BaseURL: "https://ollama.internal.company.com/v1",
}, nil)
```

### Behind an API Gateway

When Ollama is fronted by Kong, AWS API Gateway, or similar:

```go
// API key gateway
provider := impl.NewOllamaProviderWithConfig(&impl.OllamaProviderConfig{
    Auth:    clients.NewAPIKeyAuth("X-API-Key", "my-gateway-key"),
    BaseURL: "https://api.company.com/ollama/v1",
}, nil)

// OAuth2 gateway
oauth := rest.NewOAuth2Provider(
    "https://auth.company.com/oauth/token",
    "client-id", "client-secret",
    "openid", // scopes
)
provider = impl.NewOllamaProviderWithConfig(&impl.OllamaProviderConfig{
    Auth:    oauth,
    BaseURL: "https://api.company.com/ollama/v1",
}, nil)
```

## Supported Options

Since Ollama uses the OpenAI-compatible API, it supports the same options as the OpenAI provider. However, actual support depends on the model:

| GenAI Option               | Parameter              | Type       | Notes                          |
| -------------------------- | ---------------------- | ---------- | ------------------------------ |
| `OptionMaxTokens`          | `max_tokens`           | `int`      | Maximum tokens in the response |
| `OptionTemperature`        | `temperature`          | `float32`  | Sampling temperature           |
| `OptionTopP`               | `top_p`                | `float32`  | Nucleus sampling               |
| `OptionSeed`               | `seed`                 | `int`      | Deterministic output           |
| `OptionStopWords`          | `stop`                 | `[]string` | Stop sequences                 |
| `OptionFrequencyPenalty`   | `frequency_penalty`    | `float64`  | Penalise frequent tokens       |
| `OptionPresencePenalty`    | `presence_penalty`     | `float64`  | Penalise repeated tokens       |
| `OptionSystemInstructions` | `messages[0]` (system) | `string`   | Prepended as system message    |

> **Note:** Not all models support all options. For example, vision models like `llava` support image inputs, while text-only models like `llama3` do not.

## Generating Responses

### Basic Generation

```go
msg := genai.NewTextMessage(genai.RoleUser, "Write a haiku about Go programming.")
opts := genai.NewOptionsBuilder().
    SetMaxTokens(256).
    SetTemperature(0.8).
    Build()

resp, err := provider.Generate(ctx, "llama3", msg, opts)
```

### System Instructions

```go
opts := genai.NewOptionsBuilder().
    SetMaxTokens(1024).
    Build()
opts.Set(genai.OptionSystemInstructions, "You are a Linux sysadmin. Respond with shell commands and brief explanations.")

msg := genai.NewTextMessage(genai.RoleUser, "How do I find large files on disk?")
resp, err := provider.Generate(ctx, "llama3", msg, opts)
```

### Streaming

```go
msg := genai.NewTextMessage(genai.RoleUser, "Explain the differences between goroutines and threads.")
opts := genai.NewOptionsBuilder().SetMaxTokens(2048).Build()

respCh, errCh := provider.GenerateStream(ctx, "llama3", msg, opts)
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

### Multi-Modal (Vision)

Use a vision-capable model like `llava`:

```go
msg := genai.NewTextMessage(genai.RoleUser, "What's in this image?")
genai.AddBinPart(msg, "photo", jpegBytes, "image/jpeg")

resp, err := provider.Generate(ctx, "llava", msg, nil)
```

## Response Structure

Responses follow the same structure as the OpenAI provider:

```go
resp, _ := provider.Generate(ctx, "llama3", msg, opts)

candidate := resp.Candidates[0]
candidate.FinishReason // genai.FinishReasonStop, FinishReasonLength, etc.
candidate.Message      // *genai.Message with text Parts

// Token usage (if reported by Ollama)
resp.Meta.InputTokens
resp.Meta.OutputTokens
resp.Meta.TotalTokens
```

## Architecture

```
┌──────────────────────┐
│   OllamaProvider     │  wraps
│                      │────────►  OpenAIProvider
│  Name() = "ollama"   │           (full implementation)
└──────────────────────┘
         │
         │  HTTP POST to /v1/chat/completions
         ▼
┌──────────────────────┐
│   Ollama Server      │
│  localhost:11434      │
│                      │
│  ┌────────────────┐  │
│  │  llama3        │  │
│  │  mistral       │  │
│  │  codellama     │  │
│  │  llava         │  │
│  └────────────────┘  │
└──────────────────────┘
```

The `OllamaProvider` embeds `OpenAIProvider` and overrides only `Name()`. All `Generate` and `GenerateStream` calls are delegated to the OpenAI implementation, which makes requests to Ollama's OpenAI-compatible endpoint.

## Popular Models

| Model            | Size     | Use Case             | Vision |
| ---------------- | -------- | -------------------- | ------ |
| `llama3`         | 8B       | General purpose      | No     |
| `llama3:70b`     | 70B      | High quality, slower | No     |
| `mistral`        | 7B       | Fast, good quality   | No     |
| `codellama`      | 7B–34B   | Code generation      | No     |
| `llava`          | 7B–13B   | Vision + text        | Yes    |
| `phi3`           | 3.8B     | Small, fast          | No     |
| `gemma2`         | 9B/27B   | Google's open model  | No     |
| `deepseek-coder` | 6.7B–33B | Code generation      | No     |
| `qwen2`          | 7B–72B   | Multilingual         | No     |

Pull models with:

```sh
ollama pull llama3
ollama pull llava
ollama pull codellama
```

## Error Handling

```go
resp, err := provider.Generate(ctx, "llama3", msg, opts)
if err != nil {
    // Common errors:
    // - "openai API request failed: ..." — Ollama not running or unreachable
    // - "openai API error [...]" — model not found or invalid request
    log.Fatal(err)
}
```

> **Tip:** Errors use the "openai" prefix because the Ollama provider delegates to the OpenAI implementation. The error messages still accurately describe the issue.

## Deployment Patterns

### Local Development

```go
provider := impl.NewOllamaProvider(nil)
```

### Docker

```yaml
# docker-compose.yml
services:
  ollama:
    image: ollama/ollama
    ports:
      - "11434:11434"
    volumes:
      - ollama_data:/root/.ollama
```

```go
provider := impl.NewOllamaProviderWithConfig(&impl.OllamaProviderConfig{
    BaseURL: "http://ollama:11434/v1", // Docker service name
}, nil)
```

### Kubernetes with Auth Sidecar

```go
provider := impl.NewOllamaProviderWithConfig(&impl.OllamaProviderConfig{
    Auth:    clients.NewBearerAuth(os.Getenv("OLLAMA_TOKEN")),
    BaseURL: os.Getenv("OLLAMA_URL"), // e.g., https://ollama.cluster.local/v1
}, nil)
```

### Custom Port

```go
provider := impl.NewOllamaProviderWithConfig(&impl.OllamaProviderConfig{
    BaseURL: "http://localhost:8080/v1",
}, nil)
```

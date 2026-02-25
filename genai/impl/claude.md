# Claude (Anthropic) Provider

The Claude provider implements the `genai.Provider` interface for the [Anthropic Messages API](https://docs.anthropic.com/en/api/messages). It supports all Claude models including Claude Opus, Sonnet, and Haiku.

---

- [Quick Start](#quick-start)
- [Configuration](#configuration)
  - [Simple Constructor](#simple-constructor)
  - [Full Configuration](#full-configuration)
  - [Config Reference](#config-reference)
- [Authentication](#authentication)
  - [Standard Anthropic API](#standard-anthropic-api)
  - [Custom Auth Provider](#custom-auth-provider)
- [Supported Options](#supported-options)
- [Generating Responses](#generating-responses)
  - [Basic Generation](#basic-generation)
  - [System Instructions](#system-instructions)
  - [Streaming](#streaming)
  - [Multi-Modal (Vision)](#multi-modal-vision)
  - [Tool / Function Calling](#tool--function-calling)
- [Response Structure](#response-structure)
- [API Differences from OpenAI](#api-differences-from-openai)
- [Error Handling](#error-handling)
- [Beta Features](#beta-features)

---

## Quick Start

```go
import (
    "context"
    "fmt"

    "oss.nandlabs.io/golly/genai"
    "oss.nandlabs.io/golly/genai/impl"
)

provider := impl.NewClaudeProvider("sk-ant-...", nil)
defer provider.Close()

msg := genai.NewTextMessage(genai.RoleUser, "Hello, what can you do?")
resp, err := provider.Generate(context.Background(), "claude-sonnet-4-20250514", msg, nil)
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
// Uses x-api-key header auth (Anthropic's standard mechanism)
provider := impl.NewClaudeProvider("sk-ant-...", nil)
```

The simple constructor wraps the API key with `clients.NewAPIKeyAuth("x-api-key", apiKey)`.

### Full Configuration

```go
import (
    "oss.nandlabs.io/golly/clients"
    "oss.nandlabs.io/golly/genai/impl"
    "oss.nandlabs.io/golly/rest"
)

provider := impl.NewClaudeProviderWithConfig(&impl.ClaudeProviderConfig{
    Auth:       clients.NewAPIKeyAuth("x-api-key", "sk-ant-..."),
    APIVersion: "2023-06-01",
    BaseURL:    "https://api.anthropic.com",
    Models:     []string{"claude-sonnet-4-20250514", "claude-haiku-4-20250414", "claude-opus-4-20250514"},
    Description: "Production Claude",
    Version:     "2.0.0",
    ExtraHeaders: map[string]string{
        "anthropic-beta": "prompt-caching-2024-07-31",
    },
}, &rest.ClientOpts{
    // Optional: configure retry, circuit breaker, timeouts, etc.
})
```

### Config Reference

| Field          | Type                   | Default                              | Description                                                |
| -------------- | ---------------------- | ------------------------------------ | ---------------------------------------------------------- |
| `Auth`         | `clients.AuthProvider` | —                                    | Authentication provider (required)                         |
| `APIVersion`   | `string`               | `"2023-06-01"`                       | Anthropic API version (sent as `anthropic-version` header) |
| `BaseURL`      | `string`               | `https://api.anthropic.com`          | API base URL                                               |
| `Models`       | `[]string`             | `nil`                                | List of available model IDs (informational)                |
| `Description`  | `string`               | `"Anthropic Claude API provider..."` | Provider description                                       |
| `Version`      | `string`               | `"1.0.0"`                            | Provider version                                           |
| `ExtraHeaders` | `map[string]string`    | `nil`                                | Additional HTTP headers on every request                   |

## Authentication

### Standard Anthropic API

Anthropic uses an `x-api-key` header (not Bearer tokens):

```go
// Convenience — wraps key as x-api-key header automatically
provider := impl.NewClaudeProvider("sk-ant-...", nil)

// Equivalent explicit config
provider = impl.NewClaudeProviderWithConfig(&impl.ClaudeProviderConfig{
    Auth: clients.NewAPIKeyAuth("x-api-key", "sk-ant-..."),
}, nil)
```

### Custom Auth Provider

For keys stored in external secret stores:

```go
type SecretManagerAuth struct {
    client     *secretmanager.Client
    secretName string
}

func (s *SecretManagerAuth) Type() clients.AuthType { return clients.AuthTypeAPIKey }
func (s *SecretManagerAuth) User() (string, error)  { return "", nil }
func (s *SecretManagerAuth) Pass() (string, error)  { return "", nil }
func (s *SecretManagerAuth) Token() (string, error) {
    result, err := s.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
        Name: s.secretName,
    })
    if err != nil {
        return "", err
    }
    return string(result.Payload.Data), nil
}

provider := impl.NewClaudeProviderWithConfig(&impl.ClaudeProviderConfig{
    Auth: &SecretManagerAuth{client: smClient, secretName: "projects/my-proj/secrets/claude-key/versions/latest"},
}, nil)
```

## Supported Options

| GenAI Option               | Claude Parameter | Type       | Description                                                         |
| -------------------------- | ---------------- | ---------- | ------------------------------------------------------------------- |
| `OptionMaxTokens`          | `max_tokens`     | `int`      | Maximum tokens in the response (default: 4096, **required by API**) |
| `OptionTemperature`        | `temperature`    | `float32`  | Sampling temperature (0.0–1.0)                                      |
| `OptionTopP`               | `top_p`          | `float32`  | Nucleus sampling threshold                                          |
| `OptionTopK`               | `top_k`          | `int`      | Top-K sampling (Claude-specific, not available in OpenAI)           |
| `OptionStopWords`          | `stop_sequences` | `[]string` | Stop sequences (up to 4)                                            |
| `OptionSystemInstructions` | `system`         | `string`   | System prompt (top-level field, not a message)                      |

> **Note:** Unlike OpenAI, Claude does not support `frequency_penalty`, `presence_penalty`, `seed`, `candidate_count (n)`, or `response_format`. These options are silently ignored.

## Generating Responses

### Basic Generation

```go
msg := genai.NewTextMessage(genai.RoleUser, "Explain quantum entanglement.")
opts := genai.NewOptionsBuilder().
    SetMaxTokens(1024).
    SetTemperature(0.5).
    Build()

resp, err := provider.Generate(ctx, "claude-sonnet-4-20250514", msg, opts)
```

### System Instructions

Claude handles system instructions as a top-level `system` field (not as a message in the conversation). The provider handles this mapping automatically:

```go
opts := genai.NewOptionsBuilder().
    SetMaxTokens(2048).
    Build()
opts.Set(genai.OptionSystemInstructions,
    "You are an expert Go programmer. Always provide idiomatic Go code with error handling.")

msg := genai.NewTextMessage(genai.RoleUser, "How do I implement a worker pool?")
resp, err := provider.Generate(ctx, "claude-sonnet-4-20250514", msg, opts)
```

### Streaming

Claude's streaming uses typed SSE events (`content_block_delta`, `message_delta`, `message_start`, `message_stop`) which are all mapped to `genai.GenResponse` automatically:

```go
msg := genai.NewTextMessage(genai.RoleUser, "Write a detailed analysis of Go's concurrency model.")
opts := genai.NewOptionsBuilder().SetMaxTokens(2048).Build()

respCh, errCh := provider.GenerateStream(ctx, "claude-sonnet-4-20250514", msg, opts)
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

**Streaming events emitted:**

| Claude SSE Event      | GenResponse Behaviour                          |
| --------------------- | ---------------------------------------------- |
| `message_start`       | Emits `Meta.InputTokens`                       |
| `content_block_delta` | Emits text chunks as `Candidate.Message.Parts` |
| `message_delta`       | Emits `FinishReason` and `Meta.OutputTokens`   |
| `message_stop`        | Stream ends                                    |
| `error`               | Error sent to error channel                    |

### Multi-Modal (Vision)

Claude supports inline base64 images (JPEG, PNG, GIF, WebP). Unlike OpenAI, Claude does **not** support image URLs directly:

```go
// Image from bytes (supported)
msg := genai.NewTextMessage(genai.RoleUser, "What's in this image?")
genai.AddBinPart(msg, "photo", jpegBytes, "image/jpeg")

resp, err := provider.Generate(ctx, "claude-sonnet-4-20250514", msg, nil)
```

> **Note:** If you pass a `FilePart` with an image URL, the provider will include a text placeholder noting that Claude doesn't support image URLs directly. Download the image first and use `BinPart` instead.

### Tool / Function Calling

Tool calls are mapped between genai's `FuncCallPart`/`FuncResponsePart` and Claude's `tool_use`/`tool_result` content block types:

```go
// The model may return a response with FuncCall parts
resp, _ := provider.Generate(ctx, "claude-sonnet-4-20250514", msg, opts)
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
resp, _ := provider.Generate(ctx, "claude-sonnet-4-20250514", msg, opts)

// Claude returns a single candidate (no multi-candidate support)
candidate := resp.Candidates[0]
candidate.FinishReason // genai.FinishReasonEndTurn, FinishReasonLength, FinishReasonStop, FinishReasonToolCall
candidate.Message      // *genai.Message with Parts (text, tool_use)

// Token usage metadata
resp.Meta.InputTokens   // prompt tokens
resp.Meta.OutputTokens  // completion tokens
resp.Meta.TotalTokens   // input + output
resp.Meta.CachedTokens  // cache_read_input_tokens (if prompt caching is enabled)
```

**Finish reason mapping:**

| Claude `stop_reason` | GenAI `FinishReason`     |
| -------------------- | ------------------------ |
| `end_turn`           | `FinishReasonEndTurn`    |
| `max_tokens`         | `FinishReasonLength`     |
| `stop_sequence`      | `FinishReasonStop`       |
| `tool_use`           | `FinishReasonToolCall`   |
| `null` (streaming)   | `FinishReasonInProgress` |

## API Differences from OpenAI

The provider abstracts away these differences, but they're useful to know:

| Aspect              | OpenAI                                | Claude                              |
| ------------------- | ------------------------------------- | ----------------------------------- |
| Auth header         | `Authorization: Bearer <key>`         | `x-api-key: <key>`                  |
| System messages     | In `messages[]` with `role: "system"` | Top-level `system` field            |
| Content format      | String or array of content parts      | Always array of content blocks      |
| Message roles       | `system`, `user`, `assistant`, `tool` | `user`, `assistant` only            |
| `max_tokens`        | Optional                              | **Required** (default: 4096)        |
| `top_k`             | Not supported                         | Supported                           |
| `n` (candidates)    | Supported                             | Not supported                       |
| `seed`              | Supported                             | Not supported                       |
| `frequency_penalty` | Supported                             | Not supported                       |
| Image URLs          | Supported directly                    | Not supported (use base64)          |
| Streaming format    | `data: {...}` / `data: [DONE]`        | `event:` + `data:` typed SSE        |
| API version         | Not required                          | `anthropic-version` header required |

## Error Handling

```go
resp, err := provider.Generate(ctx, "claude-sonnet-4-20250514", msg, opts)
if err != nil {
    // Error format: "claude API error [error_type]: error message"
    // e.g., "claude API error [authentication_error]: invalid x-api-key"
    // e.g., "claude API error [invalid_request_error]: max_tokens must be greater than 0"
    log.Fatal(err)
}
```

**Common error types:**

| Error Type              | Cause                      |
| ----------------------- | -------------------------- |
| `authentication_error`  | Invalid or missing API key |
| `invalid_request_error` | Bad request parameters     |
| `rate_limit_error`      | Rate limit exceeded        |
| `overloaded_error`      | API temporarily overloaded |
| `not_found_error`       | Model not found            |

## Beta Features

Enable Anthropic beta features using `ExtraHeaders`:

```go
// Prompt caching
provider := impl.NewClaudeProviderWithConfig(&impl.ClaudeProviderConfig{
    Auth: clients.NewAPIKeyAuth("x-api-key", "sk-ant-..."),
    ExtraHeaders: map[string]string{
        "anthropic-beta": "prompt-caching-2024-07-31",
    },
}, nil)

// Multiple beta features (comma-separated)
provider = impl.NewClaudeProviderWithConfig(&impl.ClaudeProviderConfig{
    Auth: clients.NewAPIKeyAuth("x-api-key", "sk-ant-..."),
    ExtraHeaders: map[string]string{
        "anthropic-beta": "prompt-caching-2024-07-31,max-tokens-3-5-sonnet-2024-07-15",
    },
}, nil)
```

# GenAI Package

The `genai` package defines the provider abstraction for Generative AI services. It provides interfaces and types for interacting with large language models (LLMs) in a provider-agnostic way.

---

- [Installation](#installation)
- [Features](#features)
- [Usage](#usage)
  - [Creating Messages](#creating-messages)
  - [Multi-Part Messages](#multi-part-messages)
  - [Prompt Templates](#prompt-templates)
  - [Provider Interface](#provider-interface)
  - [Options](#options)
- [Providers](#providers)
- [Authentication](#authentication)

---

## Installation

```sh
go get oss.nandlabs.io/golly
```

## Features

- **Provider-agnostic**: Uniform interface for OpenAI, Claude, Ollama, and other LLM providers
- **Message types**: Text, binary, file reference, JSON, and YAML messages
- **Multi-part messages**: Combine text, images, and files in a single message
- **Prompt templates**: In-memory prompt store with Go template variable substitution
- **Streaming support**: `GenerateStream` for token-by-token response streaming
- **Configurable options**: Max tokens, temperature, top-p, penalties, and more
- **Flexible authentication**: Bearer tokens, API keys, Basic auth, OAuth2, and custom providers via the `clients.AuthProvider` interface

## Usage

### Creating Messages

```go
import "oss.nandlabs.io/golly/genai"

// Simple text message
msg := genai.NewMsgFromPrompt(genai.RoleUser, "greeting", "Hello!")

// System instruction
sys := genai.NewMsgFromPrompt(genai.RoleSystem, "system", "You are a helpful assistant.")

// Binary message (e.g., image)
bin := genai.NewBinMessage(genai.RoleUser, "photo", imageBytes, "image/jpeg")

// File reference
file := genai.NewFileMessage(genai.RoleUser, "doc", "gs://bucket/doc.pdf", "application/pdf")

// JSON message
jsonMsg, _ := genai.NewJsonMessage(genai.RoleUser, "data", myStruct)
```

### Multi-Part Messages

```go
msg := genai.NewMsgFromPrompt(genai.RoleUser, "analysis", "Analyze this:")
genai.AddBinPart(msg, "image", imageBytes, "image/png")
genai.AddTextPart(msg, "followup", "What do you see?")
```

### Prompt Templates

```go
store := genai.NewInMemoryPromptStore()
pt, _ := genai.NewPromptTemplate("greet", "greeting", "Hello {{.name}}!")
store.Add(pt)

msg, _ := genai.NewMsgFromPromptId(genai.RoleAssistant, store, "greet",
    map[string]any{"name": "Alice"})
```

### Provider Interface

```go
type Provider interface {
    Name() string
    Description() string
    Version() string
    Models() []string
    Generate(ctx context.Context, model string, message *Message, options *Options) (*GenResponse, error)
    GenerateStream(ctx context.Context, model string, message *Message, options *Options) (<-chan *GenResponse, <-chan error)
    Close() error
}
```

### Options

```go
opts := genai.NewOptionsBuilder().
    SetMaxTokens(1024).
    SetTemperature(0.7).
    SetTopP(0.9).
    Build()

// Or set individually
opts2 := genai.NewOptionsBuilder().Build()
opts2.Set(genai.OptionMaxTokens, 2048)
opts2.Set(genai.OptionSystemInstructions, "You are a coding assistant.")
```

## Providers

Three provider implementations are available in the `genai/impl` sub-package:

| Provider               | Auth Mechanism                               | Documentation                    |
| ---------------------- | -------------------------------------------- | -------------------------------- |
| **OpenAI**             | Bearer token (`Authorization: Bearer <key>`) | [impl/openai.md](impl/openai.md) |
| **Claude (Anthropic)** | API key header (`x-api-key: <key>`)          | [impl/claude.md](impl/claude.md) |
| **Ollama**             | None (local) or configurable for proxied     | [impl/ollama.md](impl/ollama.md) |

**Quick start for each:**

```go
import "oss.nandlabs.io/golly/genai/impl"

// OpenAI
openai := impl.NewOpenAIProvider("sk-...", nil)

// Claude
claude := impl.NewClaudeProvider("sk-ant-...", nil)

// Ollama (local, no auth)
ollama := impl.NewOllamaProvider(nil)
```

All three implement `genai.Provider`, so you can swap providers without changing your application logic:

```go
func summarize(provider genai.Provider, text string) (string, error) {
    msg := genai.NewTextMessage(genai.RoleUser, "Summarize: "+text)
    opts := genai.NewOptionsBuilder().SetMaxTokens(256).Build()
    resp, err := provider.Generate(context.Background(), "model-id", msg, opts)
    if err != nil {
        return "", err
    }
    if len(resp.Candidates) > 0 && len(resp.Candidates[0].Message.Parts) > 0 {
        if p := resp.Candidates[0].Message.Parts[0].Text; p != nil {
            return p.Text, nil
        }
    }
    return "", fmt.Errorf("no response")
}
```

## Authentication

All providers use the `clients.AuthProvider` interface for authentication, which supports multiple mechanisms:

| Auth Type      | Constructor                          | Usage                                          |
| -------------- | ------------------------------------ | ---------------------------------------------- |
| Bearer Token   | `clients.NewBearerAuth(token)`       | OpenAI, custom providers                       |
| API Key Header | `clients.NewAPIKeyAuth(header, key)` | Claude (`x-api-key`), Azure OpenAI (`api-key`) |
| Basic Auth     | `clients.NewBasicAuth(user, pass)`   | Authenticated proxies                          |
| OAuth2         | `rest.NewOAuth2Provider(...)`        | Enterprise SSO/OAuth2 flows                    |
| Custom         | Implement `clients.AuthProvider`     | Vault, AWS Secrets Manager, etc.               |

**Custom auth provider example (e.g., fetching keys from a secrets store):**

```go
type VaultAuth struct {
    vaultClient *vault.Client
    path        string
}

func (v *VaultAuth) Type() clients.AuthType { return clients.AuthTypeAPIKey }
func (v *VaultAuth) User() (string, error)  { return "", nil }
func (v *VaultAuth) Pass() (string, error)  { return "", nil }
func (v *VaultAuth) Token() (string, error) {
    secret, err := v.vaultClient.Logical().Read(v.path)
    if err != nil {
        return "", err
    }
    return secret.Data["api_key"].(string), nil
}

// Use with any provider
provider := impl.NewClaudeProviderWithConfig(&impl.ClaudeProviderConfig{
    Auth: &VaultAuth{vaultClient: vc, path: "secret/claude"},
}, nil)
```

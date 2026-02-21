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

---

## Installation

```sh
go get oss.nandlabs.io/golly
```

## Features

- **Provider-agnostic**: Uniform interface for OpenAI, Ollama, and other LLM providers
- **Message types**: Text, binary, file reference, JSON, and YAML messages
- **Multi-part messages**: Combine text, images, and files in a single message
- **Prompt templates**: In-memory prompt store with Go template variable substitution
- **Streaming support**: `GenerateStream` for token-by-token response streaming
- **Configurable options**: Max tokens, temperature, top-p, penalties, and more

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

Concrete implementations are available in the `genai/impl` sub-package (OpenAI, Ollama).

### Options

```go
opts := &genai.Options{}
opts.Set(genai.OptionMaxTokens, 1024)
opts.Set("temperature", 0.7)
opts.Set("top_p", 0.9)

maxTokens := opts.GetMaxTokens(512)       // 1024
temp := opts.GetTemperature(0.5)           // 0.7
```

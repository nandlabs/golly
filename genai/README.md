# Golly

Golly is a generative AI framework that allows you to create and manage AI models.

## Index

- [Installation](#installation)
- [Usage](#usage)
  - [Creating a Provider](#creating-a-provider)
  - [Using Templates](#using-templates)
  - [Managing Options](#managing-options)
  - [Memory Management](#memory-management)
  - [Messages](#messages)
- [Components](#components)
  - [Provider](#provider)
  - [Template](#template)
  - [Options](#options)
  - [Memory](#memory)
  - [Message](#message)

## Installation

To install Golly, use the following command:

```sh
go get -u oss.nandlabs.io/golly
```

## Usage

### Creating a Provider

To create a new provider, implement the `Provider` interface:

```go
package main

import (
    "oss.nandlabs.io/golly/genai"
    "fmt"
)

type MyProvider struct {
    // ...existing code...
}

func (p *MyProvider) Name() string {
    return "MyProvider"
}

func (p *MyProvider) Description() string {
    return "A custom provider"
}

func (p *MyProvider) Version() string {
    return "1.0.0"
}

func (p *MyProvider) Author() string {
    return "Author Name"
}

func (p *MyProvider) License() string {
    return "MIT"
}

func (p *MyProvider) Supports(model, mime string) (consumer bool, provider bool) {
    // ...existing code...
}

func (p *MyProvider) Accepts(model string) []string {
    // ...existing code...
}

func (p *MyProvider) Produces(model string) []string {
    // ...existing code...
}

func (p *MyProvider) Generate(model string, exchange genai.Exchange, options *genai.Options) error {
    // ...existing code...
}

func (p *MyProvider) GenerateStream(model string, exchange genai.Exchange, handler func(reader io.Reader), options genai.Options) error {
    // ...existing code...
}

func main() {
    provider := &MyProvider{}
    genai.Providers.Register(provider)
    fmt.Println("Provider registered:", provider.Name())
}
```

### Using Templates

To create and use templates, use the `PromptTemplate` interface and related functions:

```go
package main

import (
    "oss.nandlabs.io/golly/genai"
    "fmt"
)

func main() {
    templateContent := "Hello, {{.Name}}!"
    templateID := "greeting"

    tmpl, err := genai.NewGoTemplate(templateID, templateContent)
    if err != nil {
        fmt.Println("Error creating template:", err)
        return
    }

    data := map[string]any{
        "Name": "World",
    }

    result, err := tmpl.FormatAsText(data)
    if err != nil {
        fmt.Println("Error formatting template:", err)
        return
    }

    fmt.Println(result)
}
```

### Managing Options

To manage options for the provider, use the `Options` and `OptionsBuilder` structs:

```go
package main

import (
    "oss.nandlabs.io/golly/genai"
    "fmt"
)

func main() {
    options := genai.NewOptionsBuilder().
        SetMaxTokens(100).
        SetTemperature(0.7).
        Build()

    fmt.Println("Max Tokens:", options.GetMaxTokens(0))
    fmt.Println("Temperature:", options.GetTemperature(0))
}
```

### Memory Management

To manage memory, use the `Memory` interface and related functions:

```go
package main

import (
    "oss.nandlabs.io/golly/genai"
    "fmt"
)

func main() {
    memory := genai.NewRamMemory()
    sessionID := "session1"
    exchange := genai.Exchange{
        // ...existing code...
    }

    err := memory.Add(sessionID, exchange)
    if err != nil {
        fmt.Println("Error adding to memory:", err)
        return
    }

    exchanges, err := memory.Fetch(sessionID, "")
    if err != nil {
        fmt.Println("Error fetching from memory:", err)
        return
    }

    fmt.Println("Exchanges:", exchanges)
}
```

### Messages

To work with messages, use the `Message` struct:

```go
package main

import (
    "oss.nandlabs.io/golly/genai"
    "bytes"
    "fmt"
)

func main() {
    message := &genai.Message{
        rwer:     bytes.NewBufferString("Hello, World!"),
        mimeType: "text/plain",
    }

    fmt.Println("Message MIME type:", message.Mime())
    fmt.Println("Message content:", message.String())
}
```

## Components

### Provider

The `Provider` interface represents a generative AI model. It includes methods for generating responses and handling input and output MIME types.

### Template

The `PromptTemplate` interface represents a template for formatting prompts. The `goTemplate` struct provides an implementation using Go templates.

### Options

The `Options` struct represents the options for the provider. The `OptionsBuilder` struct provides a builder for creating options.

### Memory

The `Memory` interface represents a memory for storing exchanges. The `RamMemory` struct provides an in-memory implementation.

### Message

The `Message` struct represents a message with MIME type and content.

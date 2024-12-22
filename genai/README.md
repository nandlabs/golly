# GenAI Package

The `genai` package provides functionality for interacting with generative AI models. This package includes support for managing sessions, exchanges, and models, as well as handling memory and templates.

## Installation

To install the package, use the following command:

```sh
go get github.com/nandlabs/golly/genai
```

## Index

- [Installation](#installation)
- [Usage](#usage)
  - [Creating a Model](#creating-a-model)
  - [Creating a Session](#creating-a-session)
  - [Adding Exchanges](#adding-exchanges)
  - [Contextualizing Queries](#contextualizing-queries)
- [Components](#components)
  - [Model](#model)
  - [Session](#session)
  - [Exchange](#exchange)
  - [Memory](#memory)
  - [Template](#template)
- [License](#license)

## Usage

### Creating a Model

To create a model, you need to implement the `Model` interface. Here is an example of a simple model implementation:

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/genai"
)

type SimpleModel struct {
    genai.AbstractModel
}

func (m *SimpleModel) Generate(exchange genai.Exchange) error {
    // Implement the generation logic here
    return nil
}

func main() {
    model := &SimpleModel{
        AbstractModel: genai.AbstractModel{
            name:        "SimpleModel",
            description: "A simple generative AI model",
            version:     "1.0",
            author:      "Author Name",
            license:     "MIT",
        },
    }
    fmt.Println("Model created:", model.Name())
}
```

### Creating a Session

To create a session, you need to use the `LocalSession` struct. Here is an example:

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/genai"
)

func main() {
    model := &SimpleModel{
        AbstractModel: genai.AbstractModel{
            name:        "SimpleModel",
            description: "A simple generative AI model",
            version:     "1.0",
            author:      "Author Name",
            license:     "MIT",
        },
    }

    session := &genai.LocalSession{
        id:                    "session1",
        model:                 model,
        attributes:            make(map[string]interface{}),
        memory:                genai.NewRamMemory(),
        contextualiseTemplate: genai.NewGoTemplate(""),
    }

    fmt.Println("Session created:", session.Id())
}
```

### Adding Exchanges

To add exchanges to a session, you can use the `Add` method. Here is an example:

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/genai"
)

func main() {
    // ...existing code...

    exchange := genai.NewExchange("exchange1")
    message, _ := exchange.AddTxtMsg("Hello, AI!", genai.UserActor)
    session.SaveExchange(exchange)

    fmt.Println("Exchange added:", message.String())
}
```

### Contextualizing Queries

To contextualize queries based on previous exchanges, you can use the `Contextualise` method. Here is an example:

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/genai"
)

func main() {
    // ...existing code...

    newQuestion, err := session.Contextualise("What is the weather like?", 5)
    if err != nil {
        fmt.Println("Error contextualizing query:", err)
        return
    }

    fmt.Println("Contextualized query:", newQuestion)
}
```

## Components

### Model

The `Model` interface represents a generative AI model. It includes methods for generating responses and handling input and output MIME types.

### Session

The `Session` interface represents a session with a generative AI model. It includes methods for managing exchanges and contextualizing queries.

### Exchange

The `Exchange` interface represents an exchange between users and the AI. It includes methods for adding and retrieving messages.

### Memory

The `Memory` interface represents a memory for storing exchanges. The `RamMemory` struct provides an in-memory implementation.

### Template

The `PromptTemplate` interface represents a template for formatting prompts. The `goTemplate` struct provides an implementation using Go templates.

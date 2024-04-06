## Codec

Codec provides a unified interface for interacting with multiple formats along with validation

---

- [Installation](#installation)
- [Supported Formats](#supported--formats)
- [Codec](#codec-usage)
    - [Supported Formats](#supported--formats)
- [Examples](#examples)
  - [Basic Example](#basic-example)
  - [Advanced Example](#advanced-example)
  - [Validation Example](#validation-example)

---

### Installation

```bash
go get oss.nandlabs.io/golly/codec
```

### Usage

It comes with a simple usage as explained below, just import the package, and you are good to go.

### Codec Usage

#### Supported  Formats

| Format |  Status   |
|:-------|:---------:|
| JSON   | Completed |
| YAML   | Completed |
| XML    | Completed |

### Examples

#### Advanced Example

1. JSON Codec - Encode struct
```go
package main

import (
  "bytes"
  "fmt"
  codec "oss.nandlabs.io/golly/codec"
)

type Message struct {
  Name string `json:"name"`
  Body string `json:"body"`
  Time int64  `json:"time"`
}

func main() {
  m := Message{"TestUser", "Hello", 123124124}
  
  cd, _ := codec.Get("application/json", nil)
  buf := new(bytes.Buffer)
  if err := cd.Write(m, buf); err != nil {
    fmt.Errorf("error in write: %d", err)
  }
  // use buf in the application
}
```

#### Validation Example

```go
package main

import(
  "bytes"
  "fmt"
  codec "oss.nandlabs.io/golly/codec"
)

//Message - add validations for the fields, codec internally validates the struct
type Message struct {
  Name string `json:"name" constraints:"min-length=5"`
  Body string `json:"body" constraints:"max-length=50"`
  Time int64  `json:"time" constraints:"min=10"`
}

func main() {
  m := Message{"TestUser", "Hello", 123124124}
  c, _ := codec.Get("application/json", nil)
  buf := new(bytes.Buffer)
  if err := c.Write(m, buf); err != nil {
    fmt.Errorf("error in write: %d", err)
  }
}
```

# UUID Package

The `uuid` package provides functionality for generating and handling universally unique identifiers (UUIDs) in Go. This package supports multiple versions of UUIDs, including version 1, 2, 3, and 4.

## Installation

To install the package, use the following command:

```sh
go get github.com/nandlabs/golly/uuid
```

---

- [UUID Package](#uuid-package)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Generating UUIDs](#generating-uuids)
      - [Version 1 UUID](#version-1-uuid)
      - [Version 2 UUID](#version-2-uuid)
      - [Version 3 UUID](#version-3-uuid)
      - [Version 4 UUID](#version-4-uuid)
    - [Parsing UUIDs](#parsing-uuids)
  - [License](#license)

---

## Usage

### Generating UUIDs

The `uuid` package provides functions to generate different versions of UUIDs.

#### Version 1 UUID

Version 1 UUIDs are based on the current timestamp and the MAC address of the machine.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/uuid"
)

func main() {
    u, err := uuid.V1()
    if err != nil {
        fmt.Println("Error generating UUID:", err)
        return
    }
    fmt.Println("Version 1 UUID:", u.String())
}
```

#### Version 2 UUID

Version 2 UUIDs are based on the MAC address, process ID, and current timestamp.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/uuid"
)

func main() {
    u, err := uuid.V2()
    if err != nil {
        fmt.Println("Error generating UUID:", err)
        return
    }
    fmt.Println("Version 2 UUID:", u.String())
}
```

#### Version 3 UUID

Version 3 UUIDs are based on a namespace and a name, using MD5 hashing.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/uuid"
)

func main() {
    namespace := "example.com"
    name := "example"
    u, err := uuid.V3(namespace, name)
    if err != nil {
        fmt.Println("Error generating UUID:", err)
        return
    }
    fmt.Println("Version 3 UUID:", u.String())
}
```

#### Version 4 UUID

Version 4 UUIDs are randomly generated.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/uuid"
)

func main() {
    u, err := uuid.V4()
    if err != nil {
        fmt.Println("Error generating UUID:", err)
        return
    }
    fmt.Println("Version 4 UUID:", u.String())
}
```

### Parsing UUIDs

The `ParseUUID` function allows you to parse a UUID string into a `UUID` object.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/uuid"
)

func main() {
    uuidStr := "123e4567-e89b-12d3-a456-426655440000"
    u, err := uuid.ParseUUID(uuidStr)
    if err != nil {
        fmt.Println("Error parsing UUID:", err)
        return
    }
    fmt.Println("Parsed UUID:", u.String())
}
```

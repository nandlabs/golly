# Assertion Package

The `assertion` package provides a set of utility functions for making assertions in Go. These functions are useful for writing tests and ensuring that certain conditions hold true during the execution of your code.

---

- [Installation](#installation)
- [Usage](#usage)
  - [Equal](#equal)
  - [NotEqual](#notequal)
  - [MapContains](#mapcontains)
  - [MapMissing](#mapmissing)
  - [HasValue](#hasvalue)
  - [ListHas](#listhas)
  - [ListMissing](#listmissing)
  - [Empty](#empty)
  - [NotEmpty](#notempty)
  - [Len](#len)
  - [ElementsMatch](#elementsmatch)
  - [True](#true)
  - [False](#false)
  - [Nil](#nil)
  - [NotNil](#notnil)

---

## Installation

To install the package, use the following command:

```sh
go get github.com/nandlabs/golly/assertion
```

## Usage

The `assertion` package provides various functions to assert conditions. If an assertion fails, the function will return `false` with a descriptive error message.

### Equal

The `Equal` function checks if two values are equal.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/assertion"
)

func main() {
    a := 5
    b := 5
    if assertion.Equal(a, b) {
        fmt.Println("Values are equal")
    } else {
        fmt.Println("Values are not equal")
    }

    c := 10
    if assertion.Equal(a, c) {
        fmt.Println("Values are equal")
    } else {
        fmt.Println("Values are not equal")
    }
}
```

### NotEqual

The `NotEqual` function checks if two values are not equal.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/assertion"
)

func main() {
    a := 5
    b := 10
    if assertion.NotEqual(a, b) {
        fmt.Println("Values are not equal")
    } else {
        fmt.Println("Values are equal")
    }

    c := 5
    if assertion.NotEqual(a, c) {
        fmt.Println("Values are not equal")
    } else {
        fmt.Println("Values are equal")
    }
}
```

### MapContains

The `MapContains` function checks if a map contains a key-value pair.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/assertion"
)

func main() {
    m := map[string]interface{}{"key1": "value1", "key2": "value2"}
    if assertion.MapContains(m, "key1", "value1") {
        fmt.Println("Map contains the key-value pair")
    } else {
        fmt.Println("Map does not contain the key-value pair")
    }
}
```

### MapMissing

The `MapMissing` function checks if a map does not contain a key-value pair.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/assertion"
)

func main() {
    m := map[string]interface{}{"key1": "value1", "key2": "value2"}
    if assertion.MapMissing(m, "key3", "value3") {
        fmt.Println("Map does not contain the key-value pair")
    } else {
        fmt.Println("Map contains the key-value pair")
    }
}
```

### HasValue

The `HasValue` function checks if a map contains a value.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/assertion"
)

func main() {
    m := map[string]interface{}{"key1": "value1", "key2": "value2"}
    if assertion.HasValue(m, "value1") {
        fmt.Println("Map contains the value")
    } else {
        fmt.Println("Map does not contain the value")
    }
}
```

### ListHas

The `ListHas` function checks if a list contains a value.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/assertion"
)

func main() {
    list := []int{1, 2, 3, 4, 5}
    if assertion.ListHas(3, list) {
        fmt.Println("List contains the value")
    } else {
        fmt.Println("List does not contain the value")
    }
}
```

### ListMissing

The `ListMissing` function checks if a list does not contain a value.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/assertion"
)

func main() {
    list := []int{1, 2, 3, 4, 5}
    if assertion.ListMissing(6, list) {
        fmt.Println("List does not contain the value")
    } else {
        fmt.Println("List contains the value")
    }
}
```

### Empty

The `Empty` function checks if an object is empty.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/assertion"
)

func main() {
    var ptr *int
    if assertion.Empty(ptr) {
        fmt.Println("Value is empty")
    } else {
        fmt.Println("Value is not empty")
    }
}
```

### NotEmpty

The `NotEmpty` function checks if an object is not empty.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/assertion"
)

func main() {
    ptr := new(int)
    if assertion.NotEmpty(ptr) {
        fmt.Println("Value is not empty")
    } else {
        fmt.Println("Value is empty")
    }
}
```

### Len

The `Len` function checks if the length of an object is equal to the expected length.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/assertion"
)

func main() {
    list := []int{1, 2, 3, 4, 5}
    if assertion.Len(list, 5) {
        fmt.Println("Length is equal to expected length")
    } else {
        fmt.Println("Length is not equal to expected length")
    }
}
```

### ElementsMatch

The `ElementsMatch` function checks if the elements of a list match the expected elements.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/assertion"
)

func main() {
    list := []int{1, 2, 3, 4, 5}
    expected := []int{1, 2, 3, 4, 5}
    if assertion.ElementsMatch(list, expected...) {
        fmt.Println("Elements match")
    } else {
        fmt.Println("Elements do not match")
    }
}
```

### True

The `True` function checks if a condition is true.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/assertion"
)

func main() {
    condition := (5 > 3)
    if assertion.True(condition) {
        fmt.Println("Condition is true")
    } else {
        fmt.Println("Condition is false")
    }

    condition = (5 < 3)
    if assertion.True(condition) {
        fmt.Println("Condition is true")
    } else {
        fmt.Println("Condition is false")
    }
}
```

### False

The `False` function checks if a condition is false.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/assertion"
)

func main() {
    condition := (5 < 3)
    if assertion.False(condition) {
        fmt.Println("Condition is false")
    } else {
        fmt.Println("Condition is true")
    }

    condition = (5 > 3)
    if assertion.False(condition) {
        fmt.Println("Condition is false")
    } else {
        fmt.Println("Condition is true")
    }
}
```

### Nil

The `Nil` function checks if a value is nil.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/assertion"
)

func main() {
    var ptr *int
    if assertion.Nil(ptr) {
        fmt.Println("Value is nil")
    } else {
        fmt.Println("Value is not nil")
    }

    ptr = new(int)
    if assertion.Nil(ptr) {
        fmt.Println("Value is nil")
    } else {
        fmt.Println("Value is not nil")
    }
}
```

### NotNil

The `NotNil` function checks if a value is not nil.

```go
package main

import (
    "fmt"
    "github.com/nandlabs/golly/assertion"
)

func main() {
    ptr := new(int)
    if assertion.NotNil(ptr) {
        fmt.Println("Value is not nil")
    } else {
        fmt.Println("Value is nil")
    }

    ptr = nil
    if assertion.NotNil(ptr) {
        fmt.Println("Value is not nil")
    } else {
        fmt.Println("Value is nil")
    }
}
```

# Data Package

The `data` package provides utilities for data schema generation, pipeline-based key-value data handling, and type-safe value extraction.

---

- [Installation](#installation)
- [Features](#features)
- [Usage](#usage)
  - [Pipeline](#pipeline)
  - [Schema Generation](#schema-generation)
  - [Type-Safe Extraction](#type-safe-extraction)

---

## Installation

```sh
go get oss.nandlabs.io/golly
```

## Features

- **Pipeline**: A key-value data container with `Set`, `Get`, `Has`, `Keys`, and nested path support
- **Schema Generation**: Generate JSON Schema from Go struct types via reflection
- **Type-Safe Extraction**: Extract values from a pipeline with generics (`ExtractValue[T]`)
- **Type Conversion**: Convert between types with `Convert[T]`
- **Condition Evaluation**: Evaluate filter expressions against pipeline data

## Usage

### Pipeline

```go
import "oss.nandlabs.io/golly/data"

pipeline := data.NewPipeline("my-pipeline")
_ = pipeline.Set("name", "Alice")
_ = pipeline.Set("age", 30)

name, _ := pipeline.Get("name")
fmt.Println(name)                    // Alice
fmt.Println(pipeline.Has("age"))     // true
fmt.Println(pipeline.Keys())         // [name age]
```

### Schema Generation

```go
import (
    "reflect"
    "oss.nandlabs.io/golly/data"
)

type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

schema, _ := data.GenerateSchema(reflect.TypeOf(User{}))
```

### Type-Safe Extraction

```go
age, err := data.ExtractValue[int](pipeline, "age")
// age is typed as int
```

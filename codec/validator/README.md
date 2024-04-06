# go-struct-validator

The validator is heavily inspired by the OAS Specification approach leading to the creation of structs in a generic manner.

The validator covers the specifications, and its respective validations according to OAS.

---

- [Installation](#installation)
- [Quick Start Guide](#quick-start-guide)
- [Features](#features)
    - [Validations Supported](#validations-supported)

---

### Installation

```bash
go get oss.nandlabs.io/golly/codec/validator
```

### Quick Start Guide

It comes with a simple usage as explained below, just import the package, and you are good to go.

To add check for validations, add the `constraints` tag in the struct fields.

#### Basic Validations Example

```go
type TestStruct struct {
  Name          string  `json:"name" constraints:"min-length=5"`
  Age           int     `json:"age" constraints:"min=21"`
  Description   string  `json:"description" constraints:"max-length=50"`
  Cost          float64 `json:"cost" constraints:"exclusiveMin=200"`
  ItemCount     int     `json:"itemCount" constraints:"multipleOf=5"`
  MobileNumber  int    `json:"mobile"` // skip validation by not providing any constraints
}
```

#### Basic Example
```go
package main

import (
    "fmt"
    validator "oss.nandlabs.io/golly/codec/validator"
)

type TestStruct struct {
    Name        string  `json:"name" constraints:"min-length=5"`
    Age         int     `json:"age" constraints:"min=21"`
    Description string  `json:"description" constraints:"max-length=50"`
    Cost        float64 `json:"cost" constraints:"exclusiveMin=200"`
    ItemCount   int     `json:"itemCount" constraints:"multipleOf=5"`
}

func main() {
    var sv = validator.NewStructValidator()
    msg := TestStruct{
        Name:        "Test",
        Age:         25,
        Description: "this is bench testing",
        Cost:        299.9,
        ItemCount:   2000,
    }
    
    if err := sv.Validate(msg); err != nil {
        fmt.Errorf(err)
    }
}
```

### Features

#### Validations Supported

| S.No. |     Name     | Data Type Supported | Status |
|:------|:------------:|---------------------|--------|
| 4     |     min      | numeric             | ✅      |
| 5     |     max      | numeric             | ✅      |
| 6     | exclusiveMin | numeric             | ✅      |
| 7     | exclusiveMax | numeric             | ✅      |
| 8     |  multipleOf  | numeric             | ✅      |
| 9     |  max-length  | string              | ✅      |
| 10    |  min-length  | string              | ✅      |
| 11    |   pattern    | string              | ✅      |
| 11    |   notnull    | string              | ✅      |
| 12    |     enum     | all                 | ✅      |

# TextUtils Package

The `textutils` package provides named ASCII character constants and text utilities for Go.

---

- [Installation](#installation)
- [Features](#features)
- [Usage](#usage)

---

## Installation

```sh
go get oss.nandlabs.io/golly
```

## Features

- Named constants for all ASCII letters (`AUpperChar`, `ALowerChar`, ..., `ZUpperChar`, `ZLowerChar`)
- Enables readable, self-documenting code instead of magic character literals

## Usage

```go
import "oss.nandlabs.io/golly/textutils"

// Use named constants for readability
if ch >= textutils.AUpperChar && ch <= textutils.ZUpperChar {
    fmt.Println("Uppercase letter")
}

// Convert uppercase to lowercase
lower := ch + (textutils.ALowerChar - textutils.AUpperChar)
```

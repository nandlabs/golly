# Assertion

This is a flexible and extensible assertion library designed to provide a unified interface for asserting conditions in Go tests. It allows developers to seamlessly integrate assertion functionality into their tests without being tied to a specific assertion library.

---

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)

---

## Features

- General assertion interface for asserting conditions in tests.
- Supports various assertion functions for different types of conditions.
- Easy-to-use assertion functions for consistent handling of assertions in tests.

## Installation

To install the assertion library, use the following command:

```bash
go get oss.nandlabs.io/golly/testing/assert
```

## Usage

1. Import the library into your Go test file:
   ```go
   import "oss.nandlabs.io/golly/testing/assert"
   ```
2. Use the assertion functions in your test cases. For example:
   ```go
   func TestAdd(t *testing.T) {
       result := add(1, 2)
       assert.Equal(t, result, 3)
   }
   ```
3. Run your tests using the `go test` command:
   ```bash
   go test
   ```
4. View the test results and assertions in the test output.

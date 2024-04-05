# Virtual File System (VFS) Package
The VFS package provides a unified api for accessing multiple file system. It is extensible and new implementation can be
easily plugged in.

The default package has methods for local file system.

---
- [Installation](#installation)
- [Usage](#usage)
---

### Installation
```bash
go get go.nandlabs.io/golly/vfs
```

### Usage
A simple usage of the library to create a directory in the OS.

```go
package main

import (
    "fmt"
    "go.nandlabs.io/golly/vfs"
)

var (
    testManager = GetManager()
)

func GetRawPath(input string) (output string) {
    currentPath, _ := os.Getwd()
    u, _ := url.Parse(input)
    path := currentPath + u.Path
    output = u.Scheme + "://" + path
    return
}

func main() {
    u := GetRawPath("file:///test-data")
    _, err := testManager.MkdirRaw(u)
    if err != nil {
       fmt.Errorf("MkdirRaw() error = %v", err)
    }
}
```
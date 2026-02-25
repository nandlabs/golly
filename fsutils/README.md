# FsUtils Package

The `fsutils` package provides filesystem utility functions for checking paths, detecting content types, and working with files in Go.

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

- Check whether a path, file, or directory exists
- Detect content type of a file by reading its bytes
- Look up content type by file extension

## Usage

```go
import "oss.nandlabs.io/golly/fsutils"

// Check path existence
fmt.Println(fsutils.PathExists("/tmp"))       // true
fmt.Println(fsutils.FileExists("/tmp"))       // false (it's a directory)
fmt.Println(fsutils.DirExists("/tmp"))        // true

// Detect content type from file contents
contentType, _ := fsutils.DetectContentType("/path/to/image.png")
fmt.Println(contentType)  // image/png

// Look up content type by extension
mime := fsutils.LookupContentType("report.pdf")
fmt.Println(mime)  // application/pdf
```

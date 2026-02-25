# IOUtils Package

The `ioutils` package provides I/O utility functions for MIME type handling, channel management, and checksum calculation in Go.

---

- [Installation](#installation)
- [Features](#features)
- [Usage](#usage)
  - [MIME Types](#mime-types)
  - [Channel Utilities](#channel-utilities)
  - [Checksum Calculation](#checksum-calculation)

---

## Installation

```sh
go get oss.nandlabs.io/golly
```

## Features

- Look up MIME types from file extensions and vice versa
- Check MIME type categories (image, audio, video)
- Safely close channels and detect closed channels
- Calculate and verify SHA-256 checksums for strings, files, and readers

## Usage

### MIME Types

```go
import "oss.nandlabs.io/golly/ioutils"

mime := ioutils.GetMimeFromExt(".json")  // "application/json"
exts := ioutils.GetExtsFromMime("image/png")

fmt.Println(ioutils.IsImageMime("image/png"))  // true
fmt.Println(ioutils.IsAudioMime("audio/mp3"))  // true
fmt.Println(ioutils.IsVideoMime("video/mp4"))  // true
```

### Channel Utilities

```go
ch := make(chan int, 1)
fmt.Println(ioutils.IsChanClosed(ch))  // false

ioutils.CloseChannel(ch)
fmt.Println(ioutils.IsChanClosed(ch))  // true
```

### Checksum Calculation

```go
calc := ioutils.NewChkSumCalc(ioutils.SHA256)

// Calculate checksum of a string
sum, _ := calc.Calculate("Hello, World!")

// Verify a checksum
ok, _ := calc.Verify("Hello, World!", sum)

// Calculate checksum of a file
fileSum, _ := calc.CalculateFile("/path/to/file.txt")
```

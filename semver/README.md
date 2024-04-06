## Semver

A SemVer Package is a package that implements the Semantic Versioning (SemVer) specification. SemVer is a widely adopted standard for versioning software libraries and applications, making it easier for developers to understand when and what changes have been made to a package.

---

- [Installation](#installation)
- [Features](#features)
- [Usage](#usage)
- [Documentation](#documentation)
- [Contributing](#contributing)

---

### Installation

```bash
go get oss.nandlabs.io/golly/semver
```

### Features

* Adheres to the [SemVer 2.0.0](https://semver.org/spec/v2.0.0.html) specification
* Easy to use API for parsing, comparing and generating SemVer versions
* Supports pre-release and build metadata
* Written in modern Golang and follows best practices

### Usage

Here is an example of how to use the SemVer package in a Golang project:

```go
package main

import (
	"fmt"
	"oss.nandlabs.io/golly/semver"
)

func main() {
	version, err := semver.Parse("v1.2.3")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Major :: %d", version.major)
	fmt.Printf("Minor :: %d", version.minor)
	fmt.Printf("Patch :: %d", version.patch)
	
	metadataVersion, err := semver.Parse("v1.2.3-SNAPSHOT")
	if err != nil {
		fmt.Println(err)
    }
	fmt.Printf("Major :: %d", metadataVersion.major)
	fmt.Printf("Minor :: %d", metadataVersion.minor)
	fmt.Printf("Patch :: %d", metadataVersion.patch)
	fmt.Printf("Pre-Release :: %s", metadataVersion.preRelease)
}
```



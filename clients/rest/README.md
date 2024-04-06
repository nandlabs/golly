# REST Client
This is a README file for a REST client built using the Go programming language. The client provides a simple and efficient way to interact with RESTful APIs in Go.

---
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
---

## Features
* HTTP methods: GET, POST, PUT, DELETE
* Query parameters 
* Request headers 
* Retry
* CircuitBreaker Configuration
* Proxy Configuration
* TLS Configuration
* Transport Layer Configuration
  * MaxIdle Connections
  * Connection Timeout
  * TLS Handshake Timeout
* SSL Verification and Configuration
* CA Certs Configuration
* Error handling
  * ErrorOnHttpStatus : sets the list of status codes that can be considered failures

## Installation
To install the REST client, use the following command:
```bash
go get oss.nandlabs.io/golly/clients/rest
```

## Usage

To use the REST client in your Go project, you first need to import the package:
```go
import "oss.nandlabs.io/golly/clients/rest"
```

#### HTTP Methods : Sending a GET Request
```go
package main

import (
  "fmt"
  "oss.nandlabs.io/golly/clients/rest"
)

func main() {
  client := rest.NewClient()
  req := client.NewRequest("http://localhost:8080/api/v1/getData", "GET")
  res, err := client.Execute(req)
  if err != nil {
    // handle error
    fmt.Errorf("error executing request: %v", err)
  }
  // handle response
  fmt.Println(res)
}
```

#### Retry Configuration
```go
package main

import (
  "fmt"
  "oss.nandlabs.io/golly/clients/rest"
)

func main() {
  client := rest.NewClient()
  req := client.NewRequest("http://localhost:8080/api/v1/getData", "GET")
  // maxRetries -> 3, wait -> 5 seconds
  client.Retry(3, 5)
  res, err := client.Execute(req)
  if err != nil {
    // handle error
    fmt.Errorf("error executing request: %v", err)
  }
  // handle response
  fmt.Println(res)
}
```

#### CircuitBreaker Configuration
```go
package main

import (
  "fmt"
  "oss.nandlabs.io/golly/clients/rest"
)

func main() {
  client := rest.NewClient()
  req := client.NewRequest("http://localhost:8080/api/v1/getData", "GET")
  client.UseCircuitBreaker(1, 2, 1, 3)
  res, err := client.Execute(req)
  if err != nil {
    // handle error
    fmt.Errorf("error executing request: %v", err)
  }
  // handle response
  fmt.Println(res)
}
```

#### Proxy Configuration
```go
package main

import (
  "fmt"
  "oss.nandlabs.io/golly/clients/rest"
)

func main() {
  client := rest.NewClient()
  req := client.NewRequest("http://localhost:8080/api/v1/getData", "GET")
  err := client.SetProxy("proxy:url", "proxy_user", "proxy_pass")
  if err != nil {
	  fmt.Errorf("unable to set proxy: %v", err)
  }
  res, err := client.Execute(req)
  if err != nil {
    // handle error
    fmt.Errorf("error executing request: %v", err)
  }
  // handle response
  fmt.Println(res)
}
```

#### TLS Configuration
```go
package main

import (
  "crypto/tls"
  "fmt"
  "oss.nandlabs.io/golly/clients/rest"
)

func main() {
  client := rest.NewClient()
  req := client.NewRequest("http://localhost:8080/api/v1/getData", "GET")
  client, err := client.SetTLSCerts(tls.Certificate{})
  if err != nil {
    fmt.Errorf("error adding tls certificates: %v", err)
  }
  res, err := client.Execute(req)
  if err != nil {
    // handle error
    fmt.Errorf("error executing request: %v", err)
  }
  // handle response
  fmt.Println(res)
}
```

#### SSL Verification and CA Certs Configuration
```go
package main

import (
  "fmt"
  "oss.nandlabs.io/golly/clients/rest"
)

func main() {
  client := rest.NewClient()
  req := client.NewRequest("http://localhost:8080/api/v1/getData", "GET")
  client, err := client.SSlVerify(true)
  if err != nil {
	  fmt.Errorf("unable to set ssl verification, %v", err)
  }
  client, err = client.SetCACerts("./test-cert.pem", "./test-cert-2.pem")
  if err != nil {
    fmt.Errorf("error adding ca certificates: %v", err)
  }
  res, err := client.Execute(req)
  if err != nil {
    // handle error
    fmt.Errorf("error executing request: %v", err)
  }
  // handle response
  fmt.Println(res)
}
```

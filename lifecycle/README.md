# golly/lifecycle

[![Go Reference](https://pkg.go.dev/badge/oss.nandlabs.io/golly/lifecycle.svg)](https://pkg.go.dev/oss.nandlabs.io/golly/lifecycle)

## Overview

The `golly/lifecycle` package provides a lifecycle management library for Go
applications. It allows you to easily manage the lifecycle of your application
components, such as starting and stopping them in a controlled manner.

## Installation

To install the package, use the `go get` command:

```sh
go get oss.nandlabs.io/golly/lifecycle
```

## Usage

The `golly/lifecycle` package provides a simple and flexible way to manage the lifecycle of your application components. It allows you to start and stop components in a controlled manner, ensuring that they are properly initialized and cleaned up.

### Simple Component

The `simple_component.go` file contains a simple implementation of a component that can be used with the `golly/lifecycle` package. This component demonstrates the basic structure and behavior of a component.

To use the `SimpleComponent`, follow these steps:

1. Import the `golly/lifecycle` package and the `simple_component.go` file.

```go
import (
    "oss.nandlabs.io/golly/lifecycle"
    "oss.nandlabs.io/golly/lifecycle/examples/simple_component"
)
```

2. Create an instance of the `SimpleComponent` struct.

```go
simpleComponent := &simple_component.SimpleComponent{}
```

3. Add the `SimpleComponent` to the `Lifecycle` struct.

```go
lifecycle.AddComponent(simpleComponent)
```

4. Start and stop the components using the `Start` and `Stop` methods of the `Lifecycle` struct.

```go
err := lifecycle.Start()
if err != nil {
    // handle start error
}

// ...

err = lifecycle.Stop()
if err != nil {
    // handle stop error
}
```

By following these steps, you can use the `SimpleComponent` in your application and manage its lifecycle along with other components in the `golly/lifecycle` package.

## Cusom Components

The `component.go` file contains the interfaces that define the behavior of components in the `golly/lifecycle` package. These interfaces allow you to create custom components and integrate them into the lifecycle management system.

To use the `component.go` interfaces, follow these steps:

1. Implement the `Component` interface in your custom component struct. This interface defines the `Start` and `Stop` methods that will be called when the component is started or stopped.

```go
type MyComponent struct {
    // component fields
}

func (c *MyComponent) Start() error {
    // implementation of start logic
    return nil
}

func (c *MyComponent) Stop() error {
    // implementation of stop logic
    return nil
}
```

2. Create an instance of your custom component and add it to the `Lifecycle` struct. The `Lifecycle` struct manages the lifecycle of all registered components.

```go
lifecycle := &lifecycle.Lifecycle{}

myComponent := &MyComponent{}
lifecycle.AddComponent(myComponent)
```

3. Start and stop the components using the `Start` and `Stop` methods of the `Lifecycle` struct.

```go
err := lifecycle.Start()
if err != nil {
    // handle start error
}

// ...

err = lifecycle.Stop()
if err != nil {
    // handle stop error
}
```

By following these steps, you can integrate your custom components into the `golly/lifecycle` package and manage their lifecycle in a controlled manner.

For more information, refer to the [GoDoc](https://pkg.go.dev/oss.nandlabs.io/golly/lifecycle) documentation.

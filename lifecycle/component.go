package lifecycle

import (
	"errors"
	"time"
)

type ComponentState int

const (
	// Unknown is the state of the component when it is not known.
	Unknown ComponentState = iota
	// Error is the state of the component when it is in error.
	Error
	// Stopped is the state of the component when it is stopped.
	Stopped
	// Stopping is the state of the component when it is stopping.
	Stopping
	// Running is the state of the component when it is running.
	Running
	// Starting is the state of the component when it is starting.
	Starting
)

var ErrCompNotFound = errors.New("component not found")

var ErrCompAlreadyStarted = errors.New("component already started")

var ErrCompAlreadyStopped = errors.New("component already stopped")

var ErrInvalidComponentState = errors.New("invalid component state")

var ErrCyclicDependency = errors.New("cyclic dependency")

// ErrTimeout is returned when a lifecycle operation exceeds its timeout.
var ErrTimeout = errors.New("lifecycle operation timed out")

// Component is the interface that wraps the basic Start and Stop methods.
type Component interface {
	// Id is the unique identifier for the component.
	Id() string
	// OnChange is the function that will be called when the component state changes.
	// It will be called with the previous state and the new state.
	OnChange(f func(prevState, newState ComponentState))
	// Start will starting the LifeCycle.
	Start() error
	// Stop will stop the LifeCycle.
	Stop() error
	// State will return the current state of the LifeCycle.
	State() ComponentState
}

// ComponentManager is the interface that manages multiple components.
type ComponentManager interface {
	// AddDependency will add a dependency between the two components.
	AddDependency(id, dependsOn string) error
	// GetState will return the current state of the LifeCycle for the component with the given id.
	GetState(id string) ComponentState
	// List will return a list of all the Components.
	List() []Component
	// OnChange is the function that will be called when the component state changes.
	OnChange(id string, f func(prevState, newState ComponentState))
	// Register will register a new Components.
	Register(component Component) Component
	// StartAll will start all the Components. Returns the number of components started
	StartAll() error
	// StartAllWithTimeout will start all the Components with a timeout.
	// Returns ErrTimeout if the operation exceeds the specified duration.
	StartAllWithTimeout(timeout time.Duration) error
	// StartAndWait will start all the Components and wait for them to finish.
	StartAndWait()
	// Start will start the LifeCycle for the component with the given id.
	// It returns an error if the component was not found or if the component failed to start.
	Start(id string) error
	// StartWithTimeout will start the component with the given id, subject to a timeout.
	// Returns ErrTimeout if the operation exceeds the specified duration.
	StartWithTimeout(id string, timeout time.Duration) error
	// StopAll will stop all the Components.
	StopAll() error
	// StopAllWithTimeout will stop all the Components with a timeout.
	// Returns ErrTimeout if the operation exceeds the specified duration.
	StopAllWithTimeout(timeout time.Duration) error
	// Stop will stop the LifeCycle for the component with the given id. It returns if the component was stopped.
	Stop(id string) error
	// StopWithTimeout will stop the component with the given id, subject to a timeout.
	// Returns ErrTimeout if the operation exceeds the specified duration.
	StopWithTimeout(id string, timeout time.Duration) error
	// Unregister will unregister a Component.
	Unregister(id string)
	// Wait will wait for all the Components to finish.
	Wait()
}

package lifecycle

import (
	"sync"
)

// SimpleComponent is the struct that implements the Component interface.
type SimpleComponent struct {
	CompId string
	// AfterStart is the function that will be called after the component is started
	// The function will be called with the error returned by the StartFunc.
	AfterStart func(err error)
	// BeforeStart is the function that will be called before the component is started
	BeforeStart func()
	// AfterStop is the function that will be called after the component is stopped
	// The function will be called with the error returned by the StopFunc.
	AfterStop func(err error)
	// BeforeStop is the function that will be called before the component is stopped.
	BeforeStop func()
	// CompState is the current state of the component.
	CompState ComponentState
	// OnStateChange is the function that will be called when the component state changes.
	OnStateChange func(prevState, newState ComponentState)
	//StartFunc is the function that will be called when the component is started.
	// It returns an error if the component failed to start.
	// This can be a blocking call with respect to the component.
	// In onrder to make it non-blocking, use a go routine inside the function.
	StartFunc func() error
	// StopFunc is the function that will be called when the component is stopped.
	// It returns an error if the component failed to stop.
	// This will always be a blocking call with respect to the component.
	StopFunc func() error
}

// ComponentId is the unique identifier for the component.
func (sc *SimpleComponent) Id() string {
	return sc.CompId
}

// OnChange is the function that will be called when the component state changes.
func (sc *SimpleComponent) OnChange(prevState, newState ComponentState) {
	if sc.OnStateChange != nil {
		sc.OnStateChange(prevState, newState)
	}
	if newState == Starting && sc.BeforeStart != nil {
		sc.BeforeStart()
	}

	if newState == Stopping && sc.BeforeStop != nil {
		sc.BeforeStop()
	}
}

// Start will starting the LifeCycle.
func (sc *SimpleComponent) Start() (err error) {
	if sc.StartFunc != nil {
		if sc.BeforeStart != nil {
			sc.BeforeStart()
		}
		sc.CompState = Starting
		err = sc.StartFunc()
		if sc.AfterStart != nil {
			sc.AfterStart(err)
		}
		if err != nil {
			sc.CompState = Error
		} else {
			sc.CompState = Running
		}
	}
	return
}

// Stop will stop the LifeCycle.
func (sc *SimpleComponent) Stop() (err error) {
	if sc.StopFunc != nil {
		sc.OnChange(sc.CompState, Stopping)
		sc.CompState = Stopping
		err = sc.StopFunc()
		if err != nil {
			sc.CompState = Error
		} else {
			sc.CompState = Stopped
		}
	}
	return
}

// State will return the current state of the LifeCycle.
func (sc *SimpleComponent) State() ComponentState {
	return sc.CompState
}

// SimpleComponentManager is the struct that manages the component.
type SimpleComponentManager struct {
	components map[string]Component
	status     ComponentState
	cMutex     *sync.RWMutex
	cwg        *sync.WaitGroup
}

// GetState will return the current state of the LifeCycle for the component with the given id.
func (scm *SimpleComponentManager) GetState(id string) ComponentState {
	scm.cMutex.RLock()
	defer scm.cMutex.RUnlock()
	component, exists := scm.components[id]
	if exists {
		return component.State()
	}
	return Unknown
}

// List will return a list of all the Components.
func (scm *SimpleComponentManager) List() []Component {
	scm.cMutex.RLock()
	defer scm.cMutex.RUnlock()
	// Create a slice of Component and iterate over the components map and append the components to the slice.
	components := make([]Component, 0, len(scm.components))
	for _, component := range scm.components {
		components = append(components, component)
	}
	return components
}

// Register will register a new Components.
// if the component is already registered, get the old component.
func (scm *SimpleComponentManager) Register(component Component) Component {
	scm.cMutex.Lock()
	defer scm.cMutex.Unlock()
	//if the component is already registered, get the old component and stop it
	oldComponent, exists := scm.components[component.Id()]
	if !exists {
		scm.components[component.Id()] = component
	}
	return oldComponent
}

// StartAll will start all the Components. Returns the number of components started
func (scm *SimpleComponentManager) StartAll() {
	scm.cMutex.Lock()
	defer scm.cMutex.Unlock()
	count := 0
	for _, component := range scm.components {
		scm.cwg.Add(1)
		go func(c Component) {
			err := c.Start()
			if err != nil {
				c.Stop()
				scm.cwg.Done()
			} else {
				count++
			}
		}(component)
	}
	scm.status = Running

}

// StartAndWait will start all the Components. And will wait for them to be stopped.
func (scm *SimpleComponentManager) StartAndWait() {
	scm.StartAll() // Start all the components
	// Wait for all the components to finish. This will block until all components are stopped.
	scm.cwg.Wait()
}

// Start will start the LifeCycle for the component with the given id. It returns if the component was started.
func (scm *SimpleComponentManager) Start(id string) (err error) {
	scm.cMutex.Lock()
	defer scm.cMutex.Unlock()
	component, exists := scm.components[id]
	if exists {
		if component.State() != Running {
			scm.cwg.Add(1)
			var err error = nil
			go func(c Component) {
				err = component.Start()
				if err != nil {
					component.Stop()
					scm.cwg.Done()
				}
			}(component)
			return err
		} else {
			return ErrCompAlreadyStarted
		}
	}
	return ErrCompNotFound
}

// StopAll will stop all the Components.
func (scm *SimpleComponentManager) StopAll() {
	scm.cMutex.Lock()
	defer scm.cMutex.Unlock()
	for _, component := range scm.components {
		if component.State() == Running {
			err := component.Stop()
			if err == nil {
				scm.cwg.Done()
			}
		}
	}
	scm.status = Stopped
}

// Stop will stop the LifeCycle for the component with the given id. It returns if the component was stopped.
func (scm *SimpleComponentManager) Stop(id string) error {
	scm.cMutex.Lock()
	defer scm.cMutex.Unlock()
	component, exists := scm.components[id]
	if exists {
		if component.State() == Running {
			err := component.Stop()
			scm.cwg.Done()
			return err
		} else if component.State() == Stopped {
			return ErrCompAlreadyStopped
		} else {
			return ErrInvalidComponentState
		}

	}
	return ErrCompNotFound
}

// Unregister will unregister a Component.
func (scm *SimpleComponentManager) Unregister(id string) {
	scm.cMutex.Lock()
	defer scm.cMutex.Unlock()
	// If component is not registered, return
	if component, exists := scm.components[id]; exists {
		if component.State() == Running {
			component.Stop()
			scm.cwg.Done()
		}
		delete(scm.components, id)
	}
}

// Wait will wait for all the Components to finish.
func (scm *SimpleComponentManager) Wait() {
	scm.cwg.Wait()
}

// NewSimpleComponentManager will return a new SimpleComponentManager.
func NewSimpleComponentManager() ComponentManager {
	return &SimpleComponentManager{
		components: make(map[string]Component),
		status:     Stopped,
		cMutex:     &sync.RWMutex{},
		cwg:        &sync.WaitGroup{},
	}
}

package lifecycle

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"oss.nandlabs.io/golly/errutils"
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
	StartFunc func() error
	// StopFunc is the function that will be called when the component is stopped.
	// It returns an error if the component failed to stop.
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
	} else if newState == Stopping && sc.BeforeStop != nil {
		sc.BeforeStop()
	}
}

// Start will starting the LifeCycle.
func (sc *SimpleComponent) Start() (err error) {
	if sc.StartFunc != nil {
		sc.OnChange(sc.CompState, Starting)
		sc.CompState = Starting
		err = sc.StartFunc()
		if err != nil {
			sc.CompState = Error
		} else {
			sc.CompState = Running
		}
		if sc.OnStateChange != nil {
			sc.OnStateChange(Starting, sc.CompState)
		}
		if sc.AfterStart != nil {
			sc.AfterStart(err)
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
		if sc.OnStateChange != nil {
			sc.OnStateChange(Stopping, sc.CompState)
		}
		if sc.AfterStop != nil {
			sc.AfterStop(err)

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
	cMutex     *sync.RWMutex
	waitChan   chan struct{}
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
func (scm *SimpleComponentManager) StartAll() error {
	var err *errutils.MultiError = errutils.NewMultiErr(nil)
	for id := range scm.components {
		e := scm.Start(id)
		if e != nil {
			err.Add(e)
		}
	}
	if err.HasErrors() {
		return err
	} else {
		return nil
	}
}

// StartAndWait will start all the Components. And will wait for them to be stopped.
func (scm *SimpleComponentManager) StartAndWait() {
	scm.StartAll() // Start all the components
	scm.Wait()     // Wait for all the components to finish

}

// Start will start the LifeCycle for the component with the given id. It returns if the component was started.
func (scm *SimpleComponentManager) Start(id string) (err error) {
	scm.cMutex.Lock()
	defer scm.cMutex.Unlock()
	component, exists := scm.components[id]
	if exists {
		if component.State() != Running {
			var err error = nil
			go func(c Component, scm *SimpleComponentManager) {
				err = component.Start()
				if err != nil {
					logger.ErrorF("Error starting component: %v", err)
				}
			}(component, scm)
			return err
		} else {
			return ErrCompAlreadyStarted
		}
	}
	return ErrCompNotFound
}

// StopAll will stop all the Components.
func (scm *SimpleComponentManager) StopAll() error {
	logger.InfoF("Stopping all components")
	err := errutils.NewMultiErr(nil)
	scm.cMutex.Lock()
	defer scm.cMutex.Unlock()
	wg := &sync.WaitGroup{}
	for _, component := range scm.components {
		if component.State() == Running {
			wg.Add(1)
			go func(c Component, wg *sync.WaitGroup) {
				e := component.Stop()
				if e != nil {
					logger.ErrorF("Error stopping component: %v", err)
					err.Add(e)
				}

				wg.Done()
			}(component, wg)
		}
	}
	wg.Wait()
	close(scm.waitChan)
	if err.HasErrors() {
		return err
	} else {
		return nil
	}
}

// Stop will stop the LifeCycle for the component with the given id. It returns if the component was stopped.
func (scm *SimpleComponentManager) Stop(id string) error {
	scm.cMutex.Lock()
	defer scm.cMutex.Unlock()
	component, exists := scm.components[id]
	if exists {
		if component.State() == Running {
			err := component.Stop()
			if err != nil {
				logger.ErrorF("Error stopping component: %v", err)
			}
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
		}
		delete(scm.components, id)
	}
}

// Wait will wait for all the Components to finish.
func (scm *SimpleComponentManager) Wait() {
	go func() {
		// Wait for a signal to stop the components.
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
		<-signalChan
		scm.StopAll()
	}()
	<-scm.waitChan

}

// NewSimpleComponentManager will return a new SimpleComponentManager.
func NewSimpleComponentManager() ComponentManager {
	manager := &SimpleComponentManager{
		components: make(map[string]Component),
		cMutex:     &sync.RWMutex{},
		waitChan:   make(chan struct{}),
	}
	return manager
}

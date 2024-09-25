package lifecycle

import (
	"fmt"
	"testing"
	"time"
)

// TestSimpleComponent_Start tests the Start method of the SimpleComponent struct.
// It verifies the behavior of the Start method under different scenarios.
// The test cases include starting with no StartFunc, starting with a StartFunc,
// and starting with a StartFunc that returns an error.
// For each test case, it checks if the Start method returns the expected error
// and if the component state is updated correctly.
func TestSimpleComponent_Start(t *testing.T) {
	tests := []struct {
		name      string
		component *SimpleComponent
		wantErr   bool
		wantState ComponentState
	}{
		{
			name: "Start with no StartFunc",
			component: &SimpleComponent{
				CompId: "test1",
			},
			wantErr:   false,
			wantState: Unknown,
		},
		{
			name: "Start with StartFunc",
			component: &SimpleComponent{
				CompId: "test2",
				StartFunc: func() error {
					return nil
				},
			},
			wantErr:   false,
			wantState: Running,
		},
		{
			name: "Start with StartFunc returning error",
			component: &SimpleComponent{
				CompId: "test3",
				StartFunc: func() error {
					return fmt.Errorf("start error")
				},
			},
			wantErr:   true,
			wantState: Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.component.Start()
			if (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.component.CompState != tt.wantState {
				t.Errorf("Start() state = %v, wantState %v", tt.component.CompState, tt.wantState)
			}
		})
	}
}

// TestSimpleComponent_Stop tests the Stop method of the SimpleComponent struct.
// It verifies the behavior of the Stop method under different scenarios.
// The test cases include stopping with no StopFunc, stopping with a StopFunc,
// and stopping with a StopFunc that returns an error.
// For each test case, it checks if the Stop method returns the expected error
// and if the component state is updated correctly.

func TestSimpleComponent_Stop(t *testing.T) {
	tests := []struct {
		name      string
		component *SimpleComponent
		wantErr   bool
		wantState ComponentState
	}{
		{
			name: "Stop with no StopFunc",
			component: &SimpleComponent{
				CompId: "test1",
			},
			wantErr:   false,
			wantState: Unknown,
		},
		{
			name: "Stop with StopFunc",
			component: &SimpleComponent{
				CompId: "test2",
				StopFunc: func() error {
					return nil
				},
			},
			wantErr:   false,
			wantState: Stopped,
		},
		{
			name: "Stop with StopFunc returning error",
			component: &SimpleComponent{
				CompId: "test3",
				StopFunc: func() error {
					return fmt.Errorf("stop error")
				},
			},
			wantErr:   true,
			wantState: Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.component.Stop()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stop() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.component.CompState != tt.wantState {
				t.Errorf("Stop() state = %v, wantState %v", tt.component.CompState, tt.wantState)
			}
		})
	}
}

// TestSimpleComponentManager_Register tests the Register method of the SimpleComponentManager struct.
// It verifies the behavior of the Register method by checking if the component is registered correctly.
func TestSimpleComponentManager_Register(t *testing.T) {
	manager := NewSimpleComponentManager().(*SimpleComponentManager)
	component := &SimpleComponent{CompId: "test"}

	manager.Register(component)
	if _, exists := manager.components["test"]; !exists {
		t.Errorf("Register() component not registered")
	}
}

// TestSimpleComponentManager_StartAll tests the StartAll method of the SimpleComponentManager struct.
// It verifies the behavior of the StartAll method by checking if the components are started correctly.
// The test case includes starting a component with a StartFunc that returns no error.
// It checks if the StartAll method returns the expected number of components started.

func TestSimpleComponentManager_StartAll(t *testing.T) {
	manager := NewSimpleComponentManager().(*SimpleComponentManager)
	component := &SimpleComponent{
		CompId: "test",
		StartFunc: func() error {
			return nil
		},
		StopFunc: func() error {
			return nil
		},
	}
	manager.Register(component)

	manager.StartAll()
	time.Sleep(1 * time.Second)
	if component.State() != Running {
		t.Errorf("Component State = %v, want %v", component.State(), Running)
	}
}

// TestSimpleComponentManager_StartAndWait tests the StartAndWait method of the SimpleComponentManager struct.
// It verifies the behavior of the StartAndWait method by checking if the components are started correctly.
// The test case includes starting a component with a StartFunc that returns no error.
// It checks if the StartAndWait method waits for the components to finish.
func TestSimpleComponentManager_StopAll(t *testing.T) {
	manager := NewSimpleComponentManager()
	component := &SimpleComponent{
		CompId: "test",
		StopFunc: func() error {
			return nil
		},
		StartFunc: func() error {
			return nil
		},
	}
	manager.Register(component)
	manager.StartAll()
	time.Sleep(500 * time.Millisecond)

	manager.StopAll()
	time.Sleep(500 * time.Millisecond)
	if component.State() != Stopped {
		t.Errorf("StopAll() state = %v, want %v", component.CompState, Stopped)
	}
}

// TestSimpleComponentManager_Unregister tests the Unregister method of the SimpleComponentManager struct.
// It verifies the behavior of the Unregister method by checking if the component is unregistered correctly.
func TestSimpleComponentManager_Unregister(t *testing.T) {
	manager := NewSimpleComponentManager().(*SimpleComponentManager)
	component := &SimpleComponent{
		CompId: "test",
		StartFunc: func() error {
			return nil
		}, StopFunc: func() error {
			return nil
		},
	}

	manager.Register(component)
	manager.Unregister("test")
	if _, exists := manager.components["test"]; exists {
		t.Errorf("Unregister() component not unregistered")
	}
}

// TestSimpleComponentManager_Wait tests the Wait method of the SimpleComponentManager struct.
// It verifies the behavior of the Wait method by checking if the components are stopped correctly.
// The test case includes stopping a component with a StopFunc that returns no error.
// It checks if the Wait method waits for the components to finish.
func TestSimpleComponentManager_Wait(t *testing.T) {
	manager := NewSimpleComponentManager()
	component := &SimpleComponent{
		CompId: "test",
		StartFunc: func() error {
			return nil
		},
		StopFunc: func() error {
			return nil
		},
	}
	manager.Register(component)
	go func(manager ComponentManager) {
		time.Sleep(1 * time.Second)
		manager.StopAll()

	}(manager)
	manager.StartAndWait()

	if component.State() != Stopped {
		t.Errorf("Wait() state = %v, want %v", component.CompState, Stopped)
	}
}

func TestSimpleComponentManager_GetState(t *testing.T) {
	manager := NewSimpleComponentManager().(*SimpleComponentManager)
	component := &SimpleComponent{CompId: "test"}

	manager.Register(component)
	state := manager.GetState("test")
	if state != Unknown {
		t.Errorf("GetState() state = %v, want %v", state, Unknown)
	}
}

func TestSimpleComponentManager_List(t *testing.T) {
	manager := NewSimpleComponentManager().(*SimpleComponentManager)
	component := &SimpleComponent{CompId: "test"}

	manager.Register(component)
	components := manager.List()
	if len(components) != 1 {
		t.Errorf("List() len = %v, want %v", len(components), 1)
	}
}

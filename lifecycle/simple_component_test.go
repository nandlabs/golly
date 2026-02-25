package lifecycle

import (
	"errors"
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
	componentA := &SimpleComponent{
		CompId: "comp-a",
		StopFunc: func() error {
			logger.InfoF("Stopping component %s", "comp-a")
			return nil
		},
		StartFunc: func() error {
			logger.InfoF("Starting component %s", "comp-a")
			return nil
		},
	}
	componentB := &SimpleComponent{
		CompId: "comp-b",
		StopFunc: func() error {
			logger.InfoF("Stopping component %s", "comp-b")
			return nil
		},
		StartFunc: func() error {
			logger.InfoF("Starting component %s", "comp-b")
			return nil
		},
	}
	manager.Register(componentA)
	manager.Register(componentB)
	manager.StartAll()
	time.Sleep(500 * time.Millisecond)

	manager.StopAll()
	time.Sleep(500 * time.Millisecond)
	if componentA.State() != Stopped {
		t.Errorf("StopAll() state = %v, want %v", componentA.CompState, Stopped)
	}

	if componentB.State() != Stopped {
		t.Errorf("StopAll() state = %v, want %v", componentB.CompState, Stopped)
	}

}

// TestSimpleComponentManager_StartAndWait tests the StartAndWait method of the SimpleComponentManager struct.
// It verifies the behavior of the StartAndWait method by checking if the components are started correctly.
// The test case includes starting a component with a StartFunc that returns no error.
// It checks if the StartAndWait method waits for the components to finish.
func TestSimpleComponentManager_DependencyCheck(t *testing.T) {
	manager := NewSimpleComponentManager()
	componentA := &SimpleComponent{
		CompId: "comp-a",
		StopFunc: func() error {
			logger.InfoF("Stopping component %s", "comp-a")
			return nil
		},
		StartFunc: func() error {
			logger.InfoF("Starting component %s", "comp-a")
			return nil
		},
	}
	componentB := &SimpleComponent{
		CompId: "comp-b",
		StopFunc: func() error {
			logger.InfoF("Stopping component %s", "comp-b")
			return nil
		},
		StartFunc: func() error {
			logger.InfoF("Starting component %s", "comp-b")
			return nil
		},
	}
	manager.Register(componentA)
	manager.Register(componentB)
	manager.AddDependency("comp-a", "comp-b")
	manager.StartAll()
	time.Sleep(500 * time.Millisecond)

	manager.StopAll()
	time.Sleep(500 * time.Millisecond)
	if componentA.State() != Stopped {
		t.Errorf("StopAll() state = %v, want %v", componentA.CompState, Stopped)
	}

	if componentB.State() != Stopped {
		t.Errorf("StopAll() state = %v, want %v", componentB.CompState, Stopped)
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

// TestSimpleComponentManager_StartWithTimeout tests StartWithTimeout.
func TestSimpleComponentManager_StartWithTimeout(t *testing.T) {
	t.Run("completes before timeout", func(t *testing.T) {
		manager := NewSimpleComponentManager()
		component := &SimpleComponent{
			CompId: "fast",
			StartFunc: func() error {
				return nil
			},
			StopFunc: func() error {
				return nil
			},
		}
		manager.Register(component)
		err := manager.StartWithTimeout("fast", 2*time.Second)
		if err != nil {
			t.Errorf("StartWithTimeout() unexpected error: %v", err)
		}
		if component.State() != Running {
			t.Errorf("StartWithTimeout() state = %v, want %v", component.State(), Running)
		}
	})

	t.Run("exceeds timeout", func(t *testing.T) {
		manager := NewSimpleComponentManager()
		component := &SimpleComponent{
			CompId: "slow",
			StartFunc: func() error {
				time.Sleep(5 * time.Second)
				return nil
			},
			StopFunc: func() error {
				return nil
			},
		}
		manager.Register(component)
		err := manager.StartWithTimeout("slow", 100*time.Millisecond)
		if err == nil {
			t.Errorf("StartWithTimeout() expected timeout error, got nil")
		}
		if !errors.Is(err, ErrTimeout) {
			t.Errorf("StartWithTimeout() error = %v, want ErrTimeout", err)
		}
	})

	t.Run("component not found", func(t *testing.T) {
		manager := NewSimpleComponentManager()
		err := manager.StartWithTimeout("nonexistent", 1*time.Second)
		if err == nil {
			t.Errorf("StartWithTimeout() expected error, got nil")
		}
	})
}

// TestSimpleComponentManager_StopWithTimeout tests StopWithTimeout.
func TestSimpleComponentManager_StopWithTimeout(t *testing.T) {
	t.Run("completes before timeout", func(t *testing.T) {
		manager := NewSimpleComponentManager()
		component := &SimpleComponent{
			CompId: "fast",
			StartFunc: func() error {
				return nil
			},
			StopFunc: func() error {
				return nil
			},
		}
		manager.Register(component)
		manager.Start("fast")
		time.Sleep(100 * time.Millisecond)

		err := manager.StopWithTimeout("fast", 2*time.Second)
		if err != nil {
			t.Errorf("StopWithTimeout() unexpected error: %v", err)
		}
		if component.State() != Stopped {
			t.Errorf("StopWithTimeout() state = %v, want %v", component.State(), Stopped)
		}
	})

	t.Run("exceeds timeout", func(t *testing.T) {
		manager := NewSimpleComponentManager()
		component := &SimpleComponent{
			CompId: "slow",
			StartFunc: func() error {
				return nil
			},
			StopFunc: func() error {
				time.Sleep(5 * time.Second)
				return nil
			},
		}
		manager.Register(component)
		manager.Start("slow")
		time.Sleep(100 * time.Millisecond)

		err := manager.StopWithTimeout("slow", 100*time.Millisecond)
		if err == nil {
			t.Errorf("StopWithTimeout() expected timeout error, got nil")
		}
		if !errors.Is(err, ErrTimeout) {
			t.Errorf("StopWithTimeout() error = %v, want ErrTimeout", err)
		}
	})
}

// TestSimpleComponentManager_StartAllWithTimeout tests StartAllWithTimeout.
func TestSimpleComponentManager_StartAllWithTimeout(t *testing.T) {
	t.Run("completes before timeout", func(t *testing.T) {
		manager := NewSimpleComponentManager()
		compA := &SimpleComponent{
			CompId:    "a",
			StartFunc: func() error { return nil },
			StopFunc:  func() error { return nil },
		}
		compB := &SimpleComponent{
			CompId:    "b",
			StartFunc: func() error { return nil },
			StopFunc:  func() error { return nil },
		}
		manager.Register(compA)
		manager.Register(compB)

		err := manager.StartAllWithTimeout(2 * time.Second)
		if err != nil {
			t.Errorf("StartAllWithTimeout() unexpected error: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
		if compA.State() != Running {
			t.Errorf("StartAllWithTimeout() compA state = %v, want %v", compA.State(), Running)
		}
		if compB.State() != Running {
			t.Errorf("StartAllWithTimeout() compB state = %v, want %v", compB.State(), Running)
		}
	})

	t.Run("exceeds timeout", func(t *testing.T) {
		manager := NewSimpleComponentManager()
		comp := &SimpleComponent{
			CompId: "slow",
			StartFunc: func() error {
				time.Sleep(5 * time.Second)
				return nil
			},
			StopFunc: func() error { return nil },
		}
		manager.Register(comp)

		err := manager.StartAllWithTimeout(100 * time.Millisecond)
		if err == nil {
			t.Errorf("StartAllWithTimeout() expected timeout error, got nil")
		}
		if !errors.Is(err, ErrTimeout) {
			t.Errorf("StartAllWithTimeout() error = %v, want ErrTimeout", err)
		}
	})
}

// TestSimpleComponentManager_StopAllWithTimeout tests StopAllWithTimeout.
func TestSimpleComponentManager_StopAllWithTimeout(t *testing.T) {
	t.Run("completes before timeout", func(t *testing.T) {
		manager := NewSimpleComponentManager()
		compA := &SimpleComponent{
			CompId:    "a",
			StartFunc: func() error { return nil },
			StopFunc:  func() error { return nil },
		}
		compB := &SimpleComponent{
			CompId:    "b",
			StartFunc: func() error { return nil },
			StopFunc:  func() error { return nil },
		}
		manager.Register(compA)
		manager.Register(compB)
		manager.StartAll()
		time.Sleep(200 * time.Millisecond)

		err := manager.StopAllWithTimeout(2 * time.Second)
		if err != nil {
			t.Errorf("StopAllWithTimeout() unexpected error: %v", err)
		}
		if compA.State() != Stopped {
			t.Errorf("StopAllWithTimeout() compA state = %v, want %v", compA.State(), Stopped)
		}
		if compB.State() != Stopped {
			t.Errorf("StopAllWithTimeout() compB state = %v, want %v", compB.State(), Stopped)
		}
	})

	t.Run("exceeds timeout", func(t *testing.T) {
		manager := NewSimpleComponentManager()
		comp := &SimpleComponent{
			CompId:    "slow",
			StartFunc: func() error { return nil },
			StopFunc: func() error {
				time.Sleep(5 * time.Second)
				return nil
			},
		}
		manager.Register(comp)
		manager.Start("slow")
		time.Sleep(100 * time.Millisecond)

		err := manager.StopAllWithTimeout(100 * time.Millisecond)
		if err == nil {
			t.Errorf("StopAllWithTimeout() expected timeout error, got nil")
		}
		if !errors.Is(err, ErrTimeout) {
			t.Errorf("StopAllWithTimeout() error = %v, want ErrTimeout", err)
		}
	})
}

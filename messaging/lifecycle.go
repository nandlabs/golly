package messaging

import (
	"fmt"
	"time"

	"oss.nandlabs.io/golly/lifecycle"
)

// LifecycleProvider is an optional capability a Provider may implement
// when it also satisfies lifecycle.Component. Providers that don't
// implement this interface directly can be adapted with
// AsLifecycleComponent.
type LifecycleProvider interface {
	Provider
	lifecycle.Component
}

// Drainer is an optional capability providers implement when they can
// wait for in-flight work to complete before closing. AsLifecycleComponent's
// Stop func calls Drain (when available) before Close.
type Drainer interface {
	// Drain blocks until every in-flight dispatch has finished or the
	// timeout expires. A zero or negative timeout means "wait forever".
	// Providers that cannot track in-flight work return nil immediately.
	Drain(timeout time.Duration) error
}

// DefaultDrainTimeout is the graceful-shutdown budget used by
// AsLifecycleComponent when the caller does not specify one.
const DefaultDrainTimeout = 5 * time.Second

// AsLifecycleComponent adapts any Provider into a lifecycle.Component
// so it can be registered with a lifecycle.ComponentManager.
//
// The returned component:
//
//   - Start() calls p.Setup(). Providers that have already been Setup
//     externally may still be started — Setup implementations are
//     expected to be idempotent (LocalProvider re-initializes maps and
//     the closed flag).
//
//   - Stop() calls p.Drain(DefaultDrainTimeout) when p implements
//     Drainer, then p.Close(). Errors from Drain are logged but do not
//     block Close — a graceful attempt is preferable to leaking the
//     provider, but a hard close guarantees release of resources.
//
// The component's Id is derived from p.Id(); when that is empty a
// synthesized "messaging-provider-<schemes>" fallback is used so the
// component is still registrable.
func AsLifecycleComponent(p Provider) lifecycle.Component {
	return AsLifecycleComponentWithTimeout(p, DefaultDrainTimeout)
}

// AsLifecycleComponentWithTimeout is AsLifecycleComponent with an
// explicit drain budget for the Stop phase. A zero or negative timeout
// disables the drain step and Stop calls Close directly.
func AsLifecycleComponentWithTimeout(p Provider, drainTimeout time.Duration) lifecycle.Component {
	id := p.Id()
	if id == "" {
		id = fmt.Sprintf("messaging-provider-%v", p.Schemes())
	}
	return &lifecycle.SimpleComponent{
		CompId: id,
		StartFunc: func() error {
			return p.Setup()
		},
		StopFunc: func() error {
			if drainTimeout > 0 {
				if d, ok := p.(Drainer); ok {
					if err := d.Drain(drainTimeout); err != nil {
						logger.WarnF("messaging: drain of %s reported error before close: %v", id, err)
					}
				}
			}
			return p.Close()
		},
	}
}

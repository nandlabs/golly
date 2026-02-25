package pool

import (
	"sync"
	"time"

	"oss.nandlabs.io/golly/assertion"
)

// objectCache is a generic, thread-safe object pool.
type objectCache[T any] struct {
	// creator is a function to create new objects of type T.
	creator ObjectCreator[T]
	// destroyer is a function to destroy objects of type T.
	destroyer ObjectHandler[T]
	// min is the minimum number of objects to keep in the pool.
	min int
	// max is the maximum number of objects to keep in the pool.
	max int
	// current is the total number of live objects (idle + in-use).
	current int
	// nextID is a monotonically increasing counter for assigning unique IDs.
	nextID int
	// highWaterMark is the peak number of concurrent live objects observed.
	highWaterMark int
	// lowWaterMark is the minimum number of live objects observed after Start.
	lowWaterMark int
	// idleTimeout is the duration in seconds after which an idle object can be evicted.
	idleTimeout int
	// maxWait is the maximum wait time in seconds for Checkout when pool is exhausted.
	maxWait int
	// closed indicates whether the pool has been closed.
	closed bool
	// mutex synchronizes access to mutable state.
	mutex sync.Mutex
	// pool is a buffered channel holding idle objects.
	pool chan *PooledObject[T]
	// inUse tracks objects currently checked out, keyed by PooledObject.id.
	inUse map[int]*PooledObject[T]
}

// NewPool creates a new Pool with the given configuration.
// The pool is not started until Start() is called.
//
// Parameters:
//   - creator: function to create new objects
//   - destroyer: function to destroy objects (may be nil)
//   - min: minimum number of pre-created objects (>= 0)
//   - max: maximum number of objects (>= min, > 0)
//   - maxWait: maximum wait time in seconds when pool is exhausted
func NewPool[T any](creator ObjectCreator[T], destroyer ObjectHandler[T], min, max, maxWait int) (Pool[T], error) {
	if creator == nil {
		return nil, ErrInvalidConfig
	}
	if max <= 0 {
		return nil, ErrInvalidConfig
	}
	if min < 0 {
		min = 0
	}
	if min > max {
		min = max
	}
	if maxWait < 0 {
		maxWait = 0
	}

	return &objectCache[T]{
		creator:   creator,
		destroyer: destroyer,
		min:       min,
		max:       max,
		maxWait:   maxWait,
		pool:      make(chan *PooledObject[T], max),
		inUse:     make(map[int]*PooledObject[T]),
	}, nil
}

// Start pre-creates min objects and places them in the pool.
func (oc *objectCache[T]) Start() error {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()

	if oc.closed {
		return ErrPoolClosed
	}

	for i := 0; i < oc.min; i++ {
		obj, err := oc.creator()
		if err != nil {
			return err
		}
		oc.nextID++
		po := &PooledObject[T]{id: oc.nextID, obj: obj}
		oc.pool <- po
		oc.current++
	}
	oc.highWaterMark = oc.current
	oc.lowWaterMark = oc.current
	return nil
}

// Close drains all idle objects, destroys them and any in-use objects,
// and marks the pool as closed.
func (oc *objectCache[T]) Close() error {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()

	oc.closed = true

	// Drain idle objects from the channel
	for {
		select {
		case po := <-oc.pool:
			if oc.destroyer != nil {
				_ = oc.destroyer(po.obj)
			}
			oc.current--
		default:
			goto drained
		}
	}
drained:

	// Destroy in-use objects
	for id, po := range oc.inUse {
		if oc.destroyer != nil {
			_ = oc.destroyer(po.obj)
		}
		delete(oc.inUse, id)
		oc.current--
	}

	return nil
}

func (oc *objectCache[T]) Creator() ObjectCreator[T] {
	return oc.creator
}

func (oc *objectCache[T]) Destroyer() ObjectHandler[T] {
	return oc.destroyer
}

func (oc *objectCache[T]) Min() int {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()
	return oc.min
}

func (oc *objectCache[T]) SetMin(min int) {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()
	if min > 0 && min <= oc.max {
		oc.min = min
	}
}

func (oc *objectCache[T]) Max() int {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()
	return oc.max
}

func (oc *objectCache[T]) SetMax(max int) {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()
	if max >= oc.min && max > 0 {
		oc.max = max
	}
}

func (oc *objectCache[T]) Current() int {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()
	return oc.current
}

func (oc *objectCache[T]) HighWaterMark() int {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()
	return oc.highWaterMark
}

func (oc *objectCache[T]) LowWaterMark() int {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()
	return oc.lowWaterMark
}

func (oc *objectCache[T]) IdleTimeout() int {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()
	return oc.idleTimeout
}

func (oc *objectCache[T]) SetIdleTimeout(timeout int) {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()
	oc.idleTimeout = timeout
}

func (oc *objectCache[T]) MaxWait() int {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()
	return oc.maxWait
}

func (oc *objectCache[T]) SetMaxWait(maxWait int) {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()
	oc.maxWait = maxWait
}

// Checkout retrieves an object from the pool. If the pool is empty and the
// current count is below max, a new object is created. Otherwise it blocks
// up to MaxWait seconds for an object to become available.
func (oc *objectCache[T]) Checkout() (v T, err error) {
	oc.mutex.Lock()
	if oc.closed {
		oc.mutex.Unlock()
		err = ErrPoolClosed
		return
	}
	oc.mutex.Unlock()

	var pooledObj *PooledObject[T]

	// Fast path: try to grab an idle object without blocking.
	select {
	case po := <-oc.pool:
		pooledObj = po
	default:
		// No idle object available — try to create one.
		oc.mutex.Lock()
		if oc.current < oc.max {
			oc.nextID++
			id := oc.nextID
			oc.current++
			if oc.current > oc.highWaterMark {
				oc.highWaterMark = oc.current
			}
			oc.mutex.Unlock()

			var newObj T
			newObj, err = oc.creator()
			if err != nil {
				// Roll back the current count on creation failure.
				oc.mutex.Lock()
				oc.current--
				oc.mutex.Unlock()
				return
			}
			pooledObj = &PooledObject[T]{id: id, obj: newObj}
		} else {
			// At max capacity — release the lock and wait.
			maxWait := oc.maxWait
			oc.mutex.Unlock()

			select {
			case po := <-oc.pool:
				pooledObj = po
			case <-time.After(time.Duration(maxWait) * time.Second):
				err = ErrCacheFull
				return
			}
		}
	}

	if pooledObj != nil {
		oc.mutex.Lock()
		oc.inUse[pooledObj.id] = pooledObj
		oc.mutex.Unlock()
		v = pooledObj.obj
	}
	return
}

// Checkin returns an object to the pool, making it available for future Checkout calls.
func (oc *objectCache[T]) Checkin(obj T) {
	oc.mutex.Lock()
	var found *PooledObject[T]
	var foundID int
	for id, po := range oc.inUse {
		if assertion.Equal(po.obj, obj) {
			found = po
			foundID = id
			break
		}
	}
	if found == nil {
		oc.mutex.Unlock()
		return
	}
	delete(oc.inUse, foundID)
	oc.mutex.Unlock()

	// Return to the idle channel (non-blocking — channel is sized to max).
	select {
	case oc.pool <- found:
	default:
		// Pool channel is full (shouldn't happen), destroy the object.
		if oc.destroyer != nil {
			_ = oc.destroyer(found.obj)
		}
		oc.mutex.Lock()
		oc.current--
		if oc.current < oc.lowWaterMark {
			oc.lowWaterMark = oc.current
		}
		oc.mutex.Unlock()
	}
}

// Delete removes a specific in-use object from the pool and destroys it.
func (oc *objectCache[T]) Delete(obj T) {
	oc.mutex.Lock()
	for id, po := range oc.inUse {
		if assertion.Equal(po.obj, obj) {
			delete(oc.inUse, id)
			oc.current--
			if oc.current < oc.lowWaterMark {
				oc.lowWaterMark = oc.current
			}
			oc.mutex.Unlock()
			if oc.destroyer != nil {
				_ = oc.destroyer(po.obj)
			}
			return
		}
	}
	oc.mutex.Unlock()
}

// Clear removes all idle objects from the pool and destroys them.
// In-use objects are not affected.
func (oc *objectCache[T]) Clear() {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()

	for {
		select {
		case po := <-oc.pool:
			if oc.destroyer != nil {
				_ = oc.destroyer(po.obj)
			}
			oc.current--
		default:
			if oc.current < oc.lowWaterMark {
				oc.lowWaterMark = oc.current
			}
			return
		}
	}
}

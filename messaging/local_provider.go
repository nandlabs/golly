package messaging

import (
	"errors"
	"math/rand"
	"net/url"
	"sync"

	"oss.nandlabs.io/golly/ioutils"
)

const (
	LocalMsgScheme        = "chan"
	unnamedListeners      = "__unnamed_listeners__"
	defaultChannelBufSize = 256
)

var (
	localProviderSchemes = []string{LocalMsgScheme}
	// ErrChannelFull is returned when the message channel buffer is full.
	ErrChannelFull = errors.New("message channel is full")
	// ErrProviderClosed is returned when the provider has been closed.
	ErrProviderClosed = errors.New("provider is closed")
)

// LocalProvider is an implementation of the Provider interface that uses
// buffered Go channels for in-memory message passing.
type LocalProvider struct {
	mutex        sync.RWMutex
	destinations map[string]chan Message
	listeners    map[string]map[string][]func(msg Message)
	closed       bool
}

func (lp *LocalProvider) Id() string {
	return "local-channel"
}

func (lp *LocalProvider) NewMessage(scheme string, options ...Option) (msg Message, err error) {
	msg, err = NewLocalMessage()
	return
}

// getChan returns the channel for the given URL host, creating a buffered channel if needed.
// Uses double-check locking: fast-path with read lock, slow-path with write lock.
func (lp *LocalProvider) getChan(url *url.URL) (chan Message, error) {
	// Fast path: read lock
	lp.mutex.RLock()
	if lp.closed {
		lp.mutex.RUnlock()
		return nil, ErrProviderClosed
	}
	result, ok := lp.destinations[url.Host]
	lp.mutex.RUnlock()
	if ok {
		return result, nil
	}

	// Slow path: write lock
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	if lp.closed {
		return nil, ErrProviderClosed
	}
	// Re-check after acquiring write lock (another goroutine may have created it)
	result, ok = lp.destinations[url.Host]
	if !ok {
		result = make(chan Message, defaultChannelBufSize)
		lp.destinations[url.Host] = result
	}
	return result, nil
}

// Send sends a message to the channel identified by the URL host.
// Returns ErrChannelFull if the buffer is full, ErrProviderClosed if closed.
func (lp *LocalProvider) Send(url *url.URL, msg Message, options ...Option) (err error) {
	destination, err := lp.getChan(url)
	if err != nil {
		return err
	}
	// Recover from send-on-closed-channel panic (race between Send and Close)
	defer func() {
		if r := recover(); r != nil {
			err = ErrProviderClosed
		}
	}()
	select {
	case destination <- msg:
		logger.TraceF("sent message to channel %s", url.Host)
		return nil
	default:
		return ErrChannelFull
	}
}

// SendBatch sends a batch of messages. Stops and returns on first error.
func (lp *LocalProvider) SendBatch(url *url.URL, msgs []Message, options ...Option) error {
	for _, message := range msgs {
		if err := lp.Send(url, message); err != nil {
			return err
		}
	}
	return nil
}

// Receive blocks until a single message is available and returns it.
// Returns ErrProviderClosed if the channel is closed before a message arrives.
func (lp *LocalProvider) Receive(url *url.URL, options ...Option) (Message, error) {
	receiver, err := lp.getChan(url)
	if err != nil {
		return nil, err
	}
	msg, ok := <-receiver
	if !ok {
		return nil, ErrProviderClosed
	}
	return msg, nil
}

// ReceiveBatch blocks until at least one message is available, then drains
// all immediately available messages and returns them.
// Returns ErrProviderClosed if the channel is closed before any message arrives.
func (lp *LocalProvider) ReceiveBatch(url *url.URL, options ...Option) ([]Message, error) {
	receiver, err := lp.getChan(url)
	if err != nil {
		return nil, err
	}
	// Block until at least one message arrives
	msg, ok := <-receiver
	if !ok {
		return nil, ErrProviderClosed
	}
	msgs := []Message{msg}
	// Drain any additional immediately available messages
	for {
		select {
		case m, ok := <-receiver:
			if !ok {
				return msgs, nil
			}
			msgs = append(msgs, m)
		default:
			return msgs, nil
		}
	}
}

// AddListener registers a listener function for messages on the given URL.
// For unnamed listeners, all registered listeners are invoked for each message.
// For named listeners, one listener is randomly selected per message.
// Only the first call for a given host starts the dispatch goroutine.
func (lp *LocalProvider) AddListener(url *url.URL, listener func(msg Message), options ...Option) error {
	channel, err := lp.getChan(url)
	if err != nil {
		return err
	}
	lp.mutex.Lock()
	defer lp.mutex.Unlock()

	createListener := false
	if _, ok := lp.listeners[url.Host]; !ok {
		lp.listeners[url.Host] = make(map[string][]func(msg Message))
		createListener = true
	}

	optionsResolver := NewOptionsResolver(options...)
	namedListener, hasNamed := ResolveOptValue[string]("NamedListener", optionsResolver)
	if hasNamed {
		lp.listeners[url.Host][namedListener] = append(lp.listeners[url.Host][namedListener], listener)
	} else {
		lp.listeners[url.Host][unnamedListeners] = append(lp.listeners[url.Host][unnamedListeners], listener)
	}

	if createListener {
		go lp.dispatchMessages(url.Host, channel)
	}
	return nil
}

// dispatchMessages reads messages from the channel and dispatches them to
// registered listeners. Takes a snapshot of listeners under read lock to
// avoid data races with concurrent AddListener calls.
func (lp *LocalProvider) dispatchMessages(host string, channel chan Message) {
	for m := range channel {
		lp.mutex.RLock()
		hostListeners, ok := lp.listeners[host]
		if !ok {
			lp.mutex.RUnlock()
			continue
		}
		// Snapshot listeners to avoid holding lock during dispatch
		snapshot := make(map[string][]func(msg Message), len(hostListeners))
		for name, fns := range hostListeners {
			fnsCopy := make([]func(msg Message), len(fns))
			copy(fnsCopy, fns)
			snapshot[name] = fnsCopy
		}
		lp.mutex.RUnlock()

		for name, fns := range snapshot {
			if name == unnamedListeners {
				for _, l := range fns {
					go l(m)
				}
			} else if len(fns) > 0 {
				idx := rand.Intn(len(fns))
				go fns[idx](m)
			}
		}
	}
}

// Setup initializes the provider's internal state. Must be called before use.
func (lp *LocalProvider) Setup() error {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	lp.destinations = make(map[string]chan Message)
	lp.listeners = make(map[string]map[string][]func(msg Message))
	lp.closed = false
	return nil
}

// Close closes all destination channels and marks the provider as closed.
// Dispatch goroutines will exit when their channels are closed.
func (lp *LocalProvider) Close() error {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	lp.closed = true
	for dest, ch := range lp.destinations {
		logger.TraceF("closing channel for destination %s", dest)
		ioutils.CloseChannel[Message](ch)
	}
	return nil
}

// Schemes returns the URL schemes supported by this provider.
func (lp *LocalProvider) Schemes() []string {
	return localProviderSchemes
}

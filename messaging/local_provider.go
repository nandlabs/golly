package messaging

import (
	"context"
	"errors"
	"hash/fnv"
	"math/rand"
	"net/url"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"oss.nandlabs.io/golly/ioutils"
)

const (
	LocalMsgScheme        = "chan"
	unnamedListeners      = "__unnamed_listeners__"
	defaultChannelBufSize = 256
	// defaultAckTimeout is used when RequireAckOpt is set but the caller
	// does not supply an AckTimeoutOpt. It bounds how long a dispatcher
	// waits before assuming a listener will not ack.
	defaultAckTimeout = 30 * time.Second
	// defaultMaxDeliveryAttempts is used when the URL is configured with
	// RequireAckOpt but no MaxDeliveryAttemptsOpt. A value of 3 matches
	// the broker-side default documented in broker_options.go.
	defaultMaxDeliveryAttempts = 3
)

var (
	localProviderSchemes = []string{LocalMsgScheme}
	// ErrChannelFull is returned when the message channel buffer is full
	// and BlockOnFullOpt is not set.
	ErrChannelFull = errors.New("message channel is full")
	// ErrProviderClosed is returned when the provider has been closed.
	ErrProviderClosed = errors.New("provider is closed")
)

// dispatchConfig captures the per-URL dispatch semantics gathered from
// the options passed to the FIRST AddListener call for a URL. Subsequent
// AddListener calls to the same URL preserve the initial config —
// changing dispatch semantics mid-flight would silently reinterpret
// previously registered listeners, which we consider a footgun.
type dispatchConfig struct {
	requireAck          bool
	ackTimeout          time.Duration
	maxDeliveryAttempts int
	deadLetter          *url.URL
	concurrency         int // 0 == unbounded
}

// LocalProvider is an implementation of the Provider interface that uses
// buffered Go channels for in-memory message passing.
//
// Delivery semantics (see doc.go for a fuller treatment):
//
//   - Without RequireAckOpt on the listener: at-most-once, fire-and-forget.
//     Send returns as soon as the message is queued; each listener runs in
//     its own goroutine (bounded by ConsumerConcurrencyOpt when set).
//     Handler failures are silently dropped — Rsvp is a no-op.
//
//   - With RequireAckOpt: at-least-once with configurable redelivery.
//     The dispatcher waits (up to AckTimeoutOpt or the built-in default)
//     for the listener to call Rsvp. A Nack + RequeueOpt requeues the
//     message; other Nacks or ack-timeouts route to the DeadLetter
//     destination when configured, or drop when not. MaxDeliveryAttemptsOpt
//     caps redelivery.
type LocalProvider struct {
	mutex        sync.RWMutex
	destinations map[string]chan Message
	// bufSizes records the per-host buffer size chosen at destination
	// creation. It is used to explain observed backpressure and does not
	// affect runtime behavior after creation.
	bufSizes  map[string]int
	listeners map[string]map[string][]func(msg Message)
	// dispatchConfigs is keyed by URL host and captured from the first
	// AddListener call. It drives ack semantics, concurrency limits, and
	// dead-letter routing on the dispatch goroutine.
	dispatchConfigs map[string]*dispatchConfig
	// dispatcherRunning tracks whether the dispatch goroutine for a host
	// has been started, so subsequent AddListener calls do not spawn a
	// second goroutine per host.
	dispatcherRunning map[string]bool
	// concurrencyLimiters is a per-host semaphore for bounded dispatch.
	concurrencyLimiters map[string]chan struct{}
	// inFlight counts in-progress dispatches (per host and total). It
	// drives Drain().
	inFlight *sync.WaitGroup
	// defaultBufSize is the buffer size used when LocalBufferSizeOpt is
	// not passed at first-Send. Set via SetDefaultBufferSize.
	defaultBufSize int
	// observer is fired around Send / dispatch / Rsvp. Wrapped in a
	// pointer atomic so SetObserver can swap concrete types (nil,
	// recording, no-op sentinel) without a lock.
	observer atomic.Pointer[observerBox]
	closed   bool
}

// observerBox is a stable concrete type stored in the atomic pointer.
// It lets SetObserver swap in a fresh box holding any Observer without
// tripping atomic.Value's type-invariant.
type observerBox struct {
	obs Observer
}

// SetDefaultBufferSize overrides the buffer size the provider uses when
// no per-URL LocalBufferSizeOpt is supplied. It only affects destinations
// created after the call — existing destinations keep their original
// buffer.
func (lp *LocalProvider) SetDefaultBufferSize(n int) {
	if n <= 0 {
		return
	}
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	lp.defaultBufSize = n
}

// SetObserver installs an Observer that will be invoked around Send /
// dispatch / Rsvp. Passing nil clears any previously registered observer.
// This satisfies the ObservableProvider capability interface.
func (lp *LocalProvider) SetObserver(obs Observer) {
	if obs == nil {
		lp.observer.Store(nil)
		return
	}
	lp.observer.Store(&observerBox{obs: obs})
}

// currentObserver returns the currently registered observer or the
// no-op sentinel so callers can invoke hooks without a nil check.
func (lp *LocalProvider) currentObserver() Observer {
	if box := lp.observer.Load(); box != nil && box.obs != nil {
		return box.obs
	}
	return noopObserver{}
}

func (lp *LocalProvider) Id() string {
	return "local-channel"
}

func (lp *LocalProvider) NewMessage(scheme string, options ...Option) (msg Message, err error) {
	msg, err = NewLocalMessage()
	return
}

// getChan returns the channel for the given URL host, creating a buffered
// channel if needed. Buffer size defaults to lp.defaultBufSize (or the
// package default when unset) and may be overridden per-URL by the first
// caller passing LocalBufferSizeOpt.
func (lp *LocalProvider) getChan(url *url.URL, options ...Option) (chan Message, error) {
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
		bufSize := lp.defaultBufSize
		if bufSize <= 0 {
			bufSize = defaultChannelBufSize
		}
		if v, has := GetOptValue[int](LocalBufferSizeOpt, options...); has && v > 0 {
			bufSize = v
		}
		result = make(chan Message, bufSize)
		lp.destinations[url.Host] = result
		lp.bufSizes[url.Host] = bufSize
	}
	return result, nil
}

// Send sends a message to the channel identified by the URL host.
// Behavior depends on BlockOnFullOpt: without it, a full buffer returns
// ErrChannelFull; with it, Send blocks until room is available or the
// provider is closed.
func (lp *LocalProvider) Send(url *url.URL, msg Message, options ...Option) (err error) {
	return lp.sendCtx(context.Background(), url, msg, options...)
}

// sendCtx is the internal Send implementation shared by Send and SendCtx.
// It applies the RoutingKeyOpt to the message (when the message is Keyed
// and no key is set already), honors BlockOnFullOpt and the ack-plumbing
// hooks, and fires the OnSend observer callback around the operation.
func (lp *LocalProvider) sendCtx(ctx context.Context, u *url.URL, msg Message, options ...Option) (err error) {
	start := time.Now()
	obs := lp.currentObserver()
	defer func() {
		obs.OnSend(u, msg, err, time.Since(start))
	}()

	// If a routing key was supplied via option and the message is Keyed
	// with no key yet, propagate it. Explicit keys already on the message
	// are respected.
	if v, has := GetOptValue[string](RoutingKeyOpt, options...); has && v != "" {
		if k, ok := msg.(Keyed); ok && k.RoutingKey() == "" {
			k.SetRoutingKey(v)
		}
	}

	destination, err := lp.getChan(u, options...)
	if err != nil {
		return err
	}

	blockOnFull, _ := GetOptValue[bool](BlockOnFullOpt, options...)

	// Recover from send-on-closed-channel panic (race between Send and Close).
	defer func() {
		if r := recover(); r != nil {
			err = ErrProviderClosed
		}
	}()

	if blockOnFull {
		select {
		case destination <- msg:
			logger.TraceF("sent message to channel %s", u.Host)
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	select {
	case destination <- msg:
		logger.TraceF("sent message to channel %s", u.Host)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrChannelFull
	}
}

// SendBatch sends a batch of messages. Stops and returns on first error.
func (lp *LocalProvider) SendBatch(url *url.URL, msgs []Message, options ...Option) error {
	for _, message := range msgs {
		if err := lp.Send(url, message, options...); err != nil {
			return err
		}
	}
	return nil
}

// Receive blocks until a single message is available and returns it.
// Returns ErrProviderClosed if the channel is closed before a message arrives.
func (lp *LocalProvider) Receive(u *url.URL, options ...Option) (Message, error) {
	obs := lp.currentObserver()
	receiver, err := lp.getChan(u, options...)
	if err != nil {
		obs.OnReceive(u, nil, err)
		return nil, err
	}
	msg, ok := <-receiver
	if !ok {
		obs.OnReceive(u, nil, ErrProviderClosed)
		return nil, ErrProviderClosed
	}
	obs.OnReceive(u, msg, nil)
	return msg, nil
}

// ReceiveBatch blocks until at least one message is available, then drains
// all immediately available messages and returns them.
func (lp *LocalProvider) ReceiveBatch(u *url.URL, options ...Option) ([]Message, error) {
	obs := lp.currentObserver()
	receiver, err := lp.getChan(u, options...)
	if err != nil {
		obs.OnReceive(u, nil, err)
		return nil, err
	}
	// Block until at least one message arrives.
	msg, ok := <-receiver
	if !ok {
		obs.OnReceive(u, nil, ErrProviderClosed)
		return nil, ErrProviderClosed
	}
	msgs := []Message{msg}
	obs.OnReceive(u, msg, nil)
	// Drain any additional immediately available messages.
	for {
		select {
		case m, ok := <-receiver:
			if !ok {
				return msgs, nil
			}
			msgs = append(msgs, m)
			obs.OnReceive(u, m, nil)
		default:
			return msgs, nil
		}
	}
}

// AddListener registers a listener function for messages on the given URL.
//
// Dispatch semantics captured on the FIRST AddListener call for a URL
// come from RequireAckOpt / AckTimeoutOpt / MaxDeliveryAttemptsOpt /
// DeadLetterOpt / ConsumerConcurrencyOpt. Subsequent AddListener calls
// may still pass options — they are ignored and a warning is logged when
// they conflict with the initial config.
//
// For unnamed listeners, every registered listener runs per message. For
// named listeners, one listener within the named group runs per message
// (chosen randomly, or deterministically by routing key when the message
// implements Keyed).
func (lp *LocalProvider) AddListener(url *url.URL, listener func(msg Message), options ...Option) error {
	channel, err := lp.getChan(url, options...)
	if err != nil {
		return err
	}
	lp.mutex.Lock()
	defer lp.mutex.Unlock()

	if _, ok := lp.listeners[url.Host]; !ok {
		lp.listeners[url.Host] = make(map[string][]func(msg Message))
	}

	optionsResolver := NewOptionsResolver(options...)
	namedListener, hasNamed := ResolveOptValue[string](NamedListener, optionsResolver)
	if hasNamed {
		lp.listeners[url.Host][namedListener] = append(lp.listeners[url.Host][namedListener], listener)
	} else {
		lp.listeners[url.Host][unnamedListeners] = append(lp.listeners[url.Host][unnamedListeners], listener)
	}

	// Capture dispatch configuration on first registration for this host.
	if _, exists := lp.dispatchConfigs[url.Host]; !exists {
		cfg := newDispatchConfig(options...)
		lp.dispatchConfigs[url.Host] = cfg
		if cfg.concurrency > 0 {
			lp.concurrencyLimiters[url.Host] = make(chan struct{}, cfg.concurrency)
		}
	}

	if !lp.dispatcherRunning[url.Host] {
		lp.dispatcherRunning[url.Host] = true
		go lp.dispatchMessages(url.Host, channel)
	}
	return nil
}

// newDispatchConfig extracts the ack / concurrency / DLQ options from the
// caller-supplied options into a struct the dispatcher consults per
// message. Defaults match the LocalProvider's documented behavior.
func newDispatchConfig(options ...Option) *dispatchConfig {
	cfg := &dispatchConfig{
		ackTimeout:          defaultAckTimeout,
		maxDeliveryAttempts: defaultMaxDeliveryAttempts,
	}
	if v, has := GetOptValue[bool](RequireAckOpt, options...); has {
		cfg.requireAck = v
	}
	if v, has := GetOptValue[time.Duration](AckTimeoutOpt, options...); has && v > 0 {
		cfg.ackTimeout = v
	}
	if v, has := GetOptValue[int](MaxDeliveryAttemptsOpt, options...); has && v > 0 {
		cfg.maxDeliveryAttempts = v
	}
	if v, has := GetOptValue[string](DeadLetterOpt, options...); has && v != "" {
		if parsed, err := url.Parse(v); err == nil {
			cfg.deadLetter = parsed
		}
	}
	if v, has := GetOptValue[int](ConsumerConcurrencyOpt, options...); has && v > 0 {
		cfg.concurrency = v
	}
	return cfg
}

// RemoveListeners drops every listener registered for the URL. Idempotent.
func (lp *LocalProvider) RemoveListeners(url *url.URL) error {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	if lp.closed {
		return ErrProviderClosed
	}
	delete(lp.listeners, url.Host)
	return nil
}

// RemoveNamedListener drops the listeners registered under the given named
// group for the URL. Idempotent.
func (lp *LocalProvider) RemoveNamedListener(url *url.URL, name string) error {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	if lp.closed {
		return ErrProviderClosed
	}
	if groups, ok := lp.listeners[url.Host]; ok {
		delete(groups, name)
	}
	return nil
}

// dispatchMessages reads messages from the channel and dispatches them to
// registered listeners. Behavior varies by dispatch config (ack-aware vs
// fire-and-forget) and by the presence of a routing key on the message.
func (lp *LocalProvider) dispatchMessages(host string, channel chan Message) {
	for m := range channel {
		lp.mutex.RLock()
		hostListeners, ok := lp.listeners[host]
		if !ok {
			lp.mutex.RUnlock()
			continue
		}
		cfg := lp.dispatchConfigs[host]
		limiter := lp.concurrencyLimiters[host]
		// Snapshot listeners to avoid holding lock during dispatch.
		snapshot := make(map[string][]func(msg Message), len(hostListeners))
		for name, fns := range hostListeners {
			fnsCopy := make([]func(msg Message), len(fns))
			copy(fnsCopy, fns)
			snapshot[name] = fnsCopy
		}
		lp.mutex.RUnlock()

		lp.dispatchOne(host, m, snapshot, cfg, limiter, channel)
	}
}

// dispatchOne fires the listener set for a single message. When cfg is
// ack-aware, dispatch is synchronous and follows the RequireAck /
// AckTimeout / Requeue / DLQ contract. When cfg is nil or requireAck is
// false, dispatch remains fire-and-forget (goroutine-per-listener,
// bounded by the concurrency semaphore when configured).
func (lp *LocalProvider) dispatchOne(host string, m Message, snapshot map[string][]func(msg Message), cfg *dispatchConfig, limiter chan struct{}, channel chan Message) {
	// Select the fan-out set: with a routing key + multiple named groups
	// we route to exactly one group. Unnamed listeners always broadcast.
	unnamed := snapshot[unnamedListeners]
	namedGroups := lp.selectNamedGroups(snapshot, m)

	if cfg != nil && cfg.requireAck {
		lp.dispatchAckAware(host, m, unnamed, namedGroups, cfg, channel)
		return
	}

	// Fire-and-forget path.
	fireListener := func(l func(msg Message)) {
		lp.inFlight.Add(1)
		if limiter != nil {
			limiter <- struct{}{}
		}
		go func() {
			defer lp.inFlight.Done()
			defer func() {
				if limiter != nil {
					<-limiter
				}
			}()
			l(m)
		}()
	}

	for _, l := range unnamed {
		fireListener(l)
	}
	for _, fns := range namedGroups {
		if len(fns) == 0 {
			continue
		}
		fireListener(pickListener(fns, m))
	}
}

// selectNamedGroups returns the named groups that should receive the
// message. With a routing key AND >1 named groups, exactly one group is
// chosen via hash(key) % len(sortedGroupNames). Without a key, all named
// groups are returned (each caller picks one listener within).
func (lp *LocalProvider) selectNamedGroups(snapshot map[string][]func(msg Message), m Message) map[string][]func(msg Message) {
	// Collect the named group names (excluding unnamed).
	names := make([]string, 0, len(snapshot))
	for name := range snapshot {
		if name == unnamedListeners {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return nil
	}
	key := RoutingKeyOf(m)
	if key == "" || len(names) == 1 {
		result := make(map[string][]func(msg Message), len(names))
		for _, name := range names {
			result[name] = snapshot[name]
		}
		return result
	}
	sort.Strings(names)
	idx := int(hashKey(key) % uint32(len(names)))
	selected := names[idx]
	return map[string][]func(msg Message){selected: snapshot[selected]}
}

// pickListener chooses one listener from the group. With a routing key
// it is deterministic (hash(key) % len). Without a key, it is uniformly
// random — matching the pre-v1.7 named-group behavior.
func pickListener(fns []func(msg Message), m Message) func(msg Message) {
	if len(fns) == 1 {
		return fns[0]
	}
	if key := RoutingKeyOf(m); key != "" {
		return fns[int(hashKey(key)%uint32(len(fns)))]
	}
	return fns[rand.Intn(len(fns))]
}

// hashKey computes a stable 32-bit FNV-1a hash of the routing key. FNV
// is a good fit here — cheap, no external state, well-distributed for
// short strings, and stable across process restarts.
func hashKey(key string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return h.Sum32()
}

// dispatchAckAware runs the ack-tracking dispatch loop for a single message.
// It fires each selected listener SEQUENTIALLY (per group) so ack signals
// are unambiguous; concurrency limits still apply across messages via
// the outer read loop. On Nack + Requeue, it re-enqueues the message
// (up to MaxDeliveryAttempts). On terminal failure it routes to the
// dead-letter destination when configured or drops otherwise.
func (lp *LocalProvider) dispatchAckAware(host string, m Message, unnamed []func(msg Message), namedGroups map[string][]func(msg Message), cfg *dispatchConfig, channel chan Message) {
	// Build the ordered list of listeners to invoke. For ack-aware
	// dispatch we flatten broadcast + named-group picks into a sequence.
	targets := make([]func(msg Message), 0, len(unnamed)+len(namedGroups))
	targets = append(targets, unnamed...)
	for _, fns := range namedGroups {
		if len(fns) == 0 {
			continue
		}
		targets = append(targets, pickListener(fns, m))
	}
	if len(targets) == 0 {
		return
	}

	// Track this delivery attempt.
	if lm, ok := m.(*LocalMessage); ok {
		lm.incrementAttempts()
	}

	// Prepare an ack channel and bind it to the message so Rsvp can signal us.
	ackCh := make(chan ackSignal, 1)
	if lm, ok := m.(*LocalMessage); ok {
		lm.bindAck(ackCh)
	}

	obs := lp.currentObserver()
	u := &url.URL{Scheme: LocalMsgScheme, Host: host}

	// Fire targets. Each target gets its own goroutine so a slow listener
	// doesn't starve the others; the first Rsvp wins.
	for _, l := range targets {
		l := l
		lp.inFlight.Add(1)
		go func() {
			defer lp.inFlight.Done()
			defer func() {
				if r := recover(); r != nil {
					logger.ErrorF("messaging: listener panic on %s: %v", host, r)
				}
			}()
			l(m)
		}()
	}

	// Wait for ack, timeout, or provider close.
	var sig ackSignal
	timer := time.NewTimer(cfg.ackTimeout)
	select {
	case sig = <-ackCh:
		timer.Stop()
	case <-timer.C:
		logger.WarnF("messaging: ack timeout after %v on %s (attempt %d)",
			cfg.ackTimeout, host, currentAttempts(m))
		sig = ackSignal{ok: false, requeue: true} // timeout treated as retryable failure
	}

	if sig.ok {
		obs.OnAck(u, m)
		return
	}
	obs.OnNack(u, m, sig.requeue)

	// Nack path — decide between requeue, DLQ, or drop.
	attempts := int(currentAttempts(m))
	if sig.requeue && attempts < cfg.maxDeliveryAttempts {
		// Redeliver: clear the ack channel binding so the next dispatch
		// installs a fresh one, then re-enqueue.
		if lm, ok := m.(*LocalMessage); ok {
			lm.bindAck(nil)
		}
		defer func() { _ = recover() }() // provider may have closed
		select {
		case channel <- m:
			logger.TraceF("messaging: requeued message on %s (attempt %d/%d)",
				host, attempts, cfg.maxDeliveryAttempts)
		default:
			logger.WarnF("messaging: requeue on %s failed (buffer full); routing to DLQ", host)
			lp.routeToDLQ(host, m, cfg)
		}
		return
	}
	// Terminal failure: DLQ or drop.
	lp.routeToDLQ(host, m, cfg)
}

// currentAttempts returns the delivery-attempt count for a message when
// it is a *LocalMessage, and 0 otherwise. The typed helper keeps
// dispatchAckAware's control flow readable.
func currentAttempts(m Message) int32 {
	if lm, ok := m.(*LocalMessage); ok {
		return lm.deliveryAttempts()
	}
	return 0
}

// routeToDLQ sends the message to the dispatch config's dead-letter
// destination when one is set, or drops it otherwise. Send failures are
// logged; the dispatcher does not surface them to the original send site.
func (lp *LocalProvider) routeToDLQ(host string, m Message, cfg *dispatchConfig) {
	if cfg == nil || cfg.deadLetter == nil {
		logger.WarnF("messaging: dropping message on %s after %d attempts (no DLQ configured)",
			host, currentAttempts(m))
		return
	}
	// Reset ack binding before dispatching to the DLQ; the DLQ destination
	// should not receive our ack channel by accident.
	if lm, ok := m.(*LocalMessage); ok {
		lm.bindAck(nil)
	}
	if err := GetManager().Send(cfg.deadLetter, m); err != nil {
		logger.ErrorF("messaging: dead-letter send to %s failed: %v", cfg.deadLetter, err)
		return
	}
	logger.InfoF("messaging: routed message from %s to DLQ %s after %d attempts",
		host, cfg.deadLetter, currentAttempts(m))
}

// Setup initializes the provider's internal state. Must be called before use.
// Setup is idempotent — repeated calls reset the internal maps and clear
// the closed flag, matching the LifecycleProvider Start contract.
func (lp *LocalProvider) Setup() error {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	lp.destinations = make(map[string]chan Message)
	lp.bufSizes = make(map[string]int)
	lp.listeners = make(map[string]map[string][]func(msg Message))
	lp.dispatchConfigs = make(map[string]*dispatchConfig)
	lp.dispatcherRunning = make(map[string]bool)
	lp.concurrencyLimiters = make(map[string]chan struct{})
	lp.inFlight = &sync.WaitGroup{}
	lp.closed = false
	if lp.defaultBufSize <= 0 {
		lp.defaultBufSize = defaultChannelBufSize
	}
	return nil
}

// Close closes all destination channels and marks the provider as closed.
// Dispatch goroutines exit when their channels are closed.
func (lp *LocalProvider) Close() error {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	if lp.closed {
		return nil
	}
	lp.closed = true
	for dest, ch := range lp.destinations {
		logger.TraceF("closing channel for destination %s", dest)
		ioutils.CloseChannel(ch)
	}
	return nil
}

// Drain waits up to timeout for all in-flight dispatches to complete.
// This is the graceful-shutdown hook used by AsLifecycleComponent's Stop
// func. Returns nil once the wait group drains, or ErrDrainTimeout when
// the timeout elapses first.
func (lp *LocalProvider) Drain(timeout time.Duration) error {
	lp.mutex.RLock()
	wg := lp.inFlight
	lp.mutex.RUnlock()
	if wg == nil {
		return nil
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	if timeout <= 0 {
		<-done
		return nil
	}
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return ErrDrainTimeout
	}
}

// ErrDrainTimeout is returned by LocalProvider.Drain when the graceful
// wait exceeds the caller-supplied budget. Callers typically log it and
// proceed to Close — the fixed release of channels is preferable to
// leaking the provider.
var ErrDrainTimeout = errors.New("messaging: drain timeout")

// Schemes returns the URL schemes supported by this provider.
func (lp *LocalProvider) Schemes() []string {
	return localProviderSchemes
}

// --- ProducerCtx / ReceiverCtx capability implementation ---

// SendCtx sends msg with context-aware backpressure. When BlockOnFullOpt
// is set the send blocks until the buffer has room or ctx cancels;
// otherwise it returns ErrChannelFull immediately on a full buffer.
func (lp *LocalProvider) SendCtx(ctx context.Context, u *url.URL, msg Message, opts ...Option) error {
	return lp.sendCtx(ctx, u, msg, opts...)
}

// SendBatchCtx sends a batch, honoring ctx between messages.
func (lp *LocalProvider) SendBatchCtx(ctx context.Context, u *url.URL, msgs []Message, opts ...Option) error {
	for _, m := range msgs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := lp.sendCtx(ctx, u, m, opts...); err != nil {
			return err
		}
	}
	return nil
}

// ReceiveCtx blocks until a message arrives, ctx cancels, or the provider
// closes.
func (lp *LocalProvider) ReceiveCtx(ctx context.Context, u *url.URL, opts ...Option) (Message, error) {
	obs := lp.currentObserver()
	receiver, err := lp.getChan(u, opts...)
	if err != nil {
		obs.OnReceive(u, nil, err)
		return nil, err
	}
	select {
	case msg, ok := <-receiver:
		if !ok {
			obs.OnReceive(u, nil, ErrProviderClosed)
			return nil, ErrProviderClosed
		}
		obs.OnReceive(u, msg, nil)
		return msg, nil
	case <-ctx.Done():
		obs.OnReceive(u, nil, ctx.Err())
		return nil, ctx.Err()
	}
}

// ReceiveBatchCtx blocks for at least one message subject to ctx, then
// drains any additional messages that are immediately available.
func (lp *LocalProvider) ReceiveBatchCtx(ctx context.Context, u *url.URL, opts ...Option) ([]Message, error) {
	obs := lp.currentObserver()
	receiver, err := lp.getChan(u, opts...)
	if err != nil {
		obs.OnReceive(u, nil, err)
		return nil, err
	}
	select {
	case msg, ok := <-receiver:
		if !ok {
			obs.OnReceive(u, nil, ErrProviderClosed)
			return nil, ErrProviderClosed
		}
		msgs := []Message{msg}
		obs.OnReceive(u, msg, nil)
		for {
			select {
			case m, ok := <-receiver:
				if !ok {
					return msgs, nil
				}
				msgs = append(msgs, m)
				obs.OnReceive(u, m, nil)
			default:
				return msgs, nil
			}
		}
	case <-ctx.Done():
		obs.OnReceive(u, nil, ctx.Err())
		return nil, ctx.Err()
	}
}

// AddListenerCtx registers a listener and installs a ctx watcher that
// removes the listener (via RemoveListeners) when ctx is canceled. Note
// this removes ALL unnamed listeners for the URL when the anonymous
// listener path is used, matching the granularity of ListenerRemover.
// Callers that need finer control should use AddListener with a
// NamedListener option and RemoveNamedListener directly.
func (lp *LocalProvider) AddListenerCtx(ctx context.Context, u *url.URL, l func(Message), opts ...Option) error {
	if err := lp.AddListener(u, l, opts...); err != nil {
		return err
	}
	if ctx == nil || ctx.Done() == nil {
		return nil
	}
	go func() {
		<-ctx.Done()
		name, hasNamed := GetOptValue[string](NamedListener, opts...)
		if hasNamed {
			_ = lp.RemoveNamedListener(u, name)
			return
		}
		_ = lp.RemoveListeners(u)
	}()
	return nil
}

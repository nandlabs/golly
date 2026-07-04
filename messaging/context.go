package messaging

import (
	"context"
	"net/url"
)

// ProducerCtx is an optional capability a Producer may implement to
// support context-aware send operations. Providers signal support by
// satisfying this interface in addition to Producer.
//
// Callers should prefer the package-level helpers (SendCtx,
// SendBatchCtx) rather than type-asserting directly — the helpers fall
// back to the non-ctx methods with best-effort cancellation when the
// backend does not implement ProducerCtx. Network-backed implementations
// (SQS, Pub/Sub, Kafka, …) should implement ProducerCtx so the context
// propagates into the underlying SDK for genuine cancellation and
// deadline support.
type ProducerCtx interface {
	SendCtx(ctx context.Context, u *url.URL, msg Message, opts ...Option) error
	SendBatchCtx(ctx context.Context, u *url.URL, msgs []Message, opts ...Option) error
}

// ReceiverCtx mirrors Receiver with context support. Providers that can
// propagate the context into their underlying SDK (long-poll cancellation,
// per-listener shutdown, …) should implement this interface.
type ReceiverCtx interface {
	ReceiveCtx(ctx context.Context, u *url.URL, opts ...Option) (Message, error)
	ReceiveBatchCtx(ctx context.Context, u *url.URL, opts ...Option) ([]Message, error)
	AddListenerCtx(ctx context.Context, u *url.URL, l func(Message), opts ...Option) error
}

// runCtx runs op in a goroutine and races it against ctx cancellation.
// On ctx cancellation, returns ctx.Err() — op may still complete in the
// background. This is best-effort cancellation: callers regain control
// immediately, but the underlying send / receive cannot always be
// aborted mid-flight.
func runCtx[T any](ctx context.Context, op func() (T, error)) (T, error) {
	type result struct {
		v   T
		err error
	}
	done := make(chan result, 1)
	go func() {
		v, err := op()
		done <- result{v, err}
	}()
	select {
	case r := <-done:
		return r.v, r.err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

// runCtxErr is the void-returning variant of runCtx for ops that only
// return error.
func runCtxErr(ctx context.Context, op func() error) error {
	_, err := runCtx(ctx, func() (struct{}, error) {
		return struct{}{}, op()
	})
	return err
}

// SendCtx sends msg via p with context support. If p implements
// ProducerCtx the call is delegated; otherwise Send runs in a goroutine
// raced against ctx.
func SendCtx(ctx context.Context, p Producer, u *url.URL, msg Message, opts ...Option) error {
	if pc, ok := p.(ProducerCtx); ok {
		return pc.SendCtx(ctx, u, msg, opts...)
	}
	return runCtxErr(ctx, func() error { return p.Send(u, msg, opts...) })
}

// SendBatchCtx mirrors Producer.SendBatch with context support.
func SendBatchCtx(ctx context.Context, p Producer, u *url.URL, msgs []Message, opts ...Option) error {
	if pc, ok := p.(ProducerCtx); ok {
		return pc.SendBatchCtx(ctx, u, msgs, opts...)
	}
	return runCtxErr(ctx, func() error { return p.SendBatch(u, msgs, opts...) })
}

// ReceiveCtx mirrors Receiver.Receive with context support.
func ReceiveCtx(ctx context.Context, r Receiver, u *url.URL, opts ...Option) (Message, error) {
	if rc, ok := r.(ReceiverCtx); ok {
		return rc.ReceiveCtx(ctx, u, opts...)
	}
	return runCtx(ctx, func() (Message, error) { return r.Receive(u, opts...) })
}

// ReceiveBatchCtx mirrors Receiver.ReceiveBatch with context support.
func ReceiveBatchCtx(ctx context.Context, r Receiver, u *url.URL, opts ...Option) ([]Message, error) {
	if rc, ok := r.(ReceiverCtx); ok {
		return rc.ReceiveBatchCtx(ctx, u, opts...)
	}
	return runCtx(ctx, func() ([]Message, error) { return r.ReceiveBatch(u, opts...) })
}

// AddListenerCtx mirrors Receiver.AddListener with context support.
// Providers that implement ReceiverCtx typically tear the listener down
// when ctx is canceled; the fallback path returns immediately after the
// underlying AddListener call and leaves ctx cancellation to the caller.
func AddListenerCtx(ctx context.Context, r Receiver, u *url.URL, l func(Message), opts ...Option) error {
	if rc, ok := r.(ReceiverCtx); ok {
		return rc.AddListenerCtx(ctx, u, l, opts...)
	}
	return runCtxErr(ctx, func() error { return r.AddListener(u, l, opts...) })
}

package messaging

import "oss.nandlabs.io/golly/clients"

const (
	CircuitBreakerOpts = "CircuitBreakerOption"
	RetryOpts          = "RetryOption"
	NamedListener      = "NamedListener"
)

type Option struct {
	Key   string
	Value interface{}
}

type OptionsBuilder struct {
	options []Option
}

type OptionsResolver struct {
	opts map[string]interface{}
}

func NewOptionsResolver(options ...Option) (optsResolver *OptionsResolver) {
	optsResolver = &OptionsResolver{opts: make(map[string]interface{})}

	if len(options) > 0 {
		for _, option := range options {
			optsResolver.opts[option.Key] = option.Value
		}
	}
	return
}

// TODO check if we can pool this for performance
func NewOptionsBuilder() *OptionsBuilder {
	return &OptionsBuilder{}
}

// TODO check if you need to pool this for performance
func (ob *OptionsBuilder) Add(key string, value interface{}) *OptionsBuilder {
	ob.options = append(ob.options, Option{
		Key:   key,
		Value: value,
	})
	return ob
}

func (ob *OptionsBuilder) Build() []Option {
	return ob.options
}

func (ob *OptionsBuilder) AddCircuitBreaker(failureThreshold, successThreshold uint64, maxHalfOpen,
	timeout uint32) *OptionsBuilder {
	breakerInfo := &clients.BreakerInfo{
		FailureThreshold: failureThreshold,
		SuccessThreshold: successThreshold,
		MaxHalfOpen:      maxHalfOpen,
		Timeout:          timeout,
	}
	return ob.Add(CircuitBreakerOpts, breakerInfo)
}

func (ob *OptionsBuilder) AddRetryHandler(maxRetries, wait int) *OptionsBuilder {
	retryInfo := &clients.RetryInfo{
		MaxRetries: maxRetries,
		Wait:       wait,
	}
	return ob.Add(RetryOpts, retryInfo)
}

// AddRetryHandlerWithBackoff configures retries with exponential backoff.
// Parameters:
//   - maxRetries: maximum number of retry attempts.
//   - waitMs: base wait time in milliseconds.
//   - multiplier: backoff multiplier (e.g. 2.0 for doubling). Values <= 0 default to 2.
//   - maxWaitMs: upper bound for backoff in milliseconds. 0 means no cap.
//   - jitter: when true, adds random jitter to prevent thundering-herd.
func (ob *OptionsBuilder) AddRetryHandlerWithBackoff(maxRetries, waitMs int, multiplier float64, maxWaitMs int, jitter bool) *OptionsBuilder {
	retryInfo := &clients.RetryInfo{
		MaxRetries:  maxRetries,
		Wait:        waitMs,
		Exponential: true,
		Multiplier:  multiplier,
		MaxWait:     maxWaitMs,
		Jitter:      jitter,
	}
	return ob.Add(RetryOpts, retryInfo)
}

func (ob *OptionsBuilder) AddNamedListener(name string) *OptionsBuilder {
	return ob.Add(NamedListener, name)
}

func GetOptValue[T any](key string, opts ...Option) (value T, has bool) {
	defer func() {
		if r := recover(); r != nil {
			has = false
		}
	}()
	for _, opt := range opts {
		if opt.Key == key {
			value = opt.Value.(T)
			has = true
			return
		}
	}
	return
}

func ResolveOptValue[T any](key string, optionsResolver *OptionsResolver) (value T, has bool) {
	defer func() {
		if r := recover(); r != nil {
			has = false
		}
	}()

	var val any
	val, has = optionsResolver.opts[key]
	if has {
		value = val.(T)
	}

	return
}

func (or *OptionsResolver) Get(key string) (value interface{}, has bool) {
	value, has = or.opts[key]
	return
}

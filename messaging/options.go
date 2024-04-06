package messaging

import "go.nandlabs.io/golly/clients"

const (
	CircuitBreakerOpts = "CircuitBreakerOption"
	RetryOpts          = "CircuitBreakerOption"
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

	if options != nil && len(options) > 0 {
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

func (or *OptionsResolver) GetCircuitBreaker() (breakerInfo *clients.BreakerInfo, has bool) {
	var v interface{}
	if v, has = or.opts[CircuitBreakerOpts]; has {
		breakerInfo = v.(*clients.BreakerInfo) // TODO check if this is of the type breaker info
	}
	return
}

func (or *OptionsResolver) GetRetryInfo() (retryInfo *clients.RetryInfo, has bool) {
	var v interface{}
	if v, has = or.opts[RetryOpts]; has {
		retryInfo = v.(*clients.RetryInfo) // TODO check if this is of the type breaker info
	}
	return
}

func (or *OptionsResolver) Get(key string) (value interface{}, has bool) {
	value, has = or.opts[RetryOpts]
	return
}

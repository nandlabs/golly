package messaging

//var (
//	knownProviders = make(map[string]Provider)
//)
//
//// Messaging is a wrapper on the Provider interface
//// TODO :: should Messaging implement the Provider interface?
//type Messaging struct {
//	// TODO :: this should contain the circuit_breaker and retry info
//	retryInfo      *clients.RetryInfo
//	circuitBreaker *clients.CircuitBreaker
//}
//
//func NewMessaging() *Messaging {
//	return &Messaging{}
//}
//
//// Retry sets the maximum number of retries and wait interval in seconds between retries.
//// The Messaging client does not retry by default. If retry configuration is set along with UseCircuitBreaker then the retry config
//// is ignored
//func (m *Messaging) Retry(maxRetries, wait int) *Messaging {
//	m.retryInfo = &clients.RetryInfo{
//		MaxRetries: maxRetries,
//		Wait:       wait,
//	}
//	return m
//}
//
//// UseCircuitBreaker sets the circuit breaker configuration for this messaging client.
//// The circuit breaker pattern has higher precedence than retry pattern. If both are set then the retry configuration is
//// ignored.
//func (m *Messaging) UseCircuitBreaker(failureThreshold, successThreshold uint64, maxHalfOpen, timeout uint32) *Messaging {
//	breakerInfo := &clients.BreakerInfo{
//		FailureThreshold: failureThreshold,
//		SuccessThreshold: successThreshold,
//		MaxHalfOpen:      maxHalfOpen,
//		Timeout:          timeout,
//	}
//	m.circuitBreaker = clients.NewCB(breakerInfo)
//	return m
//}
//
//func Register(url *url.URL, provider Provider) {
//	if knownProviders == nil {
//		knownProviders = make(map[string]Provider)
//	}
//	knownProviders[url.String()] = provider
//}
//
//func (m *Messaging) AddListener(url *url.URL, listener func(msg Message)) (err error) {
//	var provider Provider
//	provider, err = m.getProvider(url)
//	if err == nil {
//		err = provider.AddListener(url, listener)
//	}
//	return
//}
//
//func (m *Messaging) getProvider(url *url.URL) (provider Provider, err error) {
//	supports := false
//	provider, supports = knownProviders[url.String()]
//	if !supports {
//		err = errors.New("unsupported provider with url " + url.String())
//	}
//	return
//}
//
//func (m *Messaging) Send(url *url.URL, msg Message) (err error) {
//	var provider Provider
//	provider, err = m.getProvider(url)
//	if err == nil {
//		if m.circuitBreaker != nil {
//			err = m.circuitBreaker.CanExecute()
//			if err == nil {
//				err = provider.Send(url, msg)
//				m.circuitBreaker.OnExecution(err != nil)
//			}
//		} else if m.retryInfo != nil {
//			err = provider.Send(url, msg)
//			for i := 0; err != nil && 1 < m.retryInfo.MaxRetries; i++ {
//				err = fnutils.ExecuteAfterSecs(func() {
//					err = provider.Send(url, msg)
//				}, m.retryInfo.Wait)
//				if err != nil {
//					return
//				}
//			}
//		} else {
//			err = provider.Send(url, msg)
//		}
//	}
//	return
//}
//
//func (m *Messaging) SendBatch(url *url.URL, msg ...Message) (err error) {
//	var provider Provider
//	provider, err = m.getProvider(url)
//	if err == nil {
//		if m.circuitBreaker != nil {
//			err = m.circuitBreaker.CanExecute()
//			if err == nil {
//				err = provider.SendBatch(url, msg...)
//				m.circuitBreaker.OnExecution(err != nil)
//			}
//		} else if m.retryInfo != nil {
//			err = provider.SendBatch(url, msg...)
//			for i := 0; err != nil && 1 < m.retryInfo.MaxRetries; i++ {
//				err = fnutils.ExecuteAfterSecs(func() {
//					err = provider.SendBatch(url, msg...)
//				}, m.retryInfo.Wait)
//				if err != nil {
//					return
//				}
//			}
//		} else {
//			err = provider.SendBatch(url, msg...)
//		}
//	}
//	return
//}
//
//func (m *Messaging) Receive(url *url.URL) (msg Message, err error) {
//	var provider Provider
//	provider, err = m.getProvider(url)
//	if err == nil {
//		msg, err = provider.Receive(url)
//	}
//	return
//}
//
//func (m *Messaging) ReceiveBatch(url *url.URL) (msgs []Message, err error) {
//	var provider Provider
//	provider, err = m.getProvider(url)
//	if err == nil {
//		msgs, err = provider.ReceiveBatch(url)
//	}
//	return
//}

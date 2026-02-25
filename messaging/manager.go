package messaging

import (
	"fmt"
	"net/url"
	"sync"

	"oss.nandlabs.io/golly/errutils"
)

var defaultManager Manager
var mutex sync.Mutex

// Manager interface defines an abstraction for messaging providers that can be registered
type Manager interface {
	Provider
	Wait()
	Register(Provider)
}

// managerImpl struct is used to manage the known Messaging providers.
// It includes a mutex to handle concurrent access to the known providers
type managerImpl struct {
	knownProviders map[string]Provider
	mutex          sync.Mutex
	waitgroup      sync.WaitGroup
}

// Id returns the id of the manager
func (m *managerImpl) Id() string {
	return "default-manager"
}

// getFor returns the provider for the given scheme
func (m *managerImpl) getFor(scheme string) (provider Provider, err error) {
	var ok bool
	provider, ok = m.knownProviders[scheme]
	if !ok {
		err = fmt.Errorf("unsupported scheme %s", scheme)
	}
	return
}

// Send is a helper function that sends a message using the appropriate provider
func (m *managerImpl) Send(u *url.URL, msg Message, options ...Option) (err error) {
	var provider Provider
	provider, err = m.getFor(u.Scheme)
	if err == nil {
		err = provider.Send(u, msg, options...)
	}
	return
}

// Receive receives a single message using the appropriate provider
func (m *managerImpl) Receive(u *url.URL, options ...Option) (msg Message, err error) {
	var provider Provider
	provider, err = m.getFor(u.Scheme)
	if err == nil {
		msg, err = provider.Receive(u, options...)
	}
	return
}

// AddListener registers a listener for the message using the appropriate provider
func (m *managerImpl) AddListener(u *url.URL, listener func(msg Message), options ...Option) (err error) {
	var provider Provider
	provider, err = m.getFor(u.Scheme)
	if err == nil {
		err = provider.AddListener(u, listener, options...)
	}
	return
}

// ReceiveBatch receives a batch of messages using the appropriate provider
func (m *managerImpl) ReceiveBatch(u *url.URL, options ...Option) (msgs []Message, err error) {
	var provider Provider
	provider, err = m.getFor(u.Scheme)
	if err == nil {
		msgs, err = provider.ReceiveBatch(u, options...)
	}
	return
}

// SendBatch sends a batch of messages using the appropriate provider
func (m *managerImpl) SendBatch(u *url.URL, msgs []Message, options ...Option) (err error) {
	var provider Provider
	provider, err = m.getFor(u.Scheme)
	if err == nil {
		err = provider.SendBatch(u, msgs, options...)
	}
	return
}

// Schemes returns the supported URL schemes by the known providers
func (m *managerImpl) Schemes() (schemes []string) {
	for k := range m.knownProviders {
		if k == "" {
			continue
		}
		schemes = append(schemes, k)
	}
	return
}

// NewMessage creates a new message using the appropriate provider
func (m *managerImpl) NewMessage(scheme string, options ...Option) (msg Message, err error) {
	var provider Provider
	provider, err = m.getFor(scheme)
	if err == nil {
		msg, err = provider.NewMessage(scheme, options...)
	}
	return
}

// Setup performs the initial setup of the messaging manager
func (m *managerImpl) Setup() (err error) {
	m.waitgroup.Add(1)
	return
}

// Register registers a messaging provider with the manager
func (m *managerImpl) Register(provider Provider) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, s := range provider.Schemes() {
		if m.knownProviders == nil {
			m.knownProviders = make(map[string]Provider)
		}
		if _, ok := m.knownProviders[s]; !ok {
			m.knownProviders[s] = provider
		}
	}
}

// Close function calls closing of all providers
func (m *managerImpl) Close() (err error) {
	var multiError *errutils.MultiError

	for _, provider := range m.knownProviders {
		providerErr := provider.Close()
		if providerErr != nil {
			if multiError == nil {
				multiError = errutils.NewMultiErr(providerErr)
			} else {
				multiError.Add(providerErr)
			}
		}
	}
	if multiError != nil {
		// TODO check bug why multi error is retuned to calling function as not nil always
		err = multiError

	}
	m.waitgroup.Done()
	return
}

func (m *managerImpl) Wait() {
	m.waitgroup.Wait()
}

// Setup function initializes the default manager

// GetManager returns the facade messaging instance
func GetManager() Manager {

	if defaultManager == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if defaultManager == nil {
			defaultManager = &managerImpl{
				knownProviders: make(map[string]Provider),
				mutex:          sync.Mutex{},
			}
			_ = defaultManager.Setup()
			localProvider := &LocalProvider{}
			_ = localProvider.Setup()
			defaultManager.Register(localProvider)
		}
	}

	return defaultManager
}

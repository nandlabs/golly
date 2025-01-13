package messaging

import (
	"math/rand"
	"net/url"
	"sync"

	"oss.nandlabs.io/golly/ioutils"
)

const (
	LocalMsgScheme   = "chan"
	unnamedListeners = "__unnamed_listeners__"
)

var localProviderSchemes = []string{LocalMsgScheme}

// LocalProvider is an implementation of the Provider interface
type LocalProvider struct {
	mutex        sync.Mutex
	destinations map[string]chan Message
	listeners    map[string]map[string][]func(msg Message)
}

func (lp *LocalProvider) Id() string {
	return "local-channel"
}

func (lp *LocalProvider) NewMessage(scheme string, options ...Option) (msg Message, err error) {
	msg, err = NewLocalMessage()
	return
}

func (lp *LocalProvider) getChan(url *url.URL) (result chan Message) {
	var ok bool
	result, ok = lp.destinations[url.Host]
	if !ok {
		lp.mutex.Lock()
		defer lp.mutex.Unlock()
		localMsgChannel := make(chan Message)
		lp.destinations[url.Host] = localMsgChannel
		result = localMsgChannel
	}
	return
}

func (lp *LocalProvider) Send(url *url.URL, msg Message, options ...Option) (err error) {
	destination := lp.getChan(url)
	go func() {
		logger.TraceF("sending message to channel %s", url.Host)
		destination <- msg
	}()
	return
}

func (lp *LocalProvider) SendBatch(url *url.URL, msgs []Message, options ...Option) (err error) {
	for _, message := range msgs {
		err = lp.Send(url, message)
		if err != nil {
			return
		}
	}
	return
}

func (lp *LocalProvider) Receive(url *url.URL, options ...Option) (msg Message, err error) {
	receiver := lp.getChan(url)
	for m := range receiver {
		msg = m
	}
	return
}

func (lp *LocalProvider) ReceiveBatch(url *url.URL, options ...Option) (msgs []Message, err error) {
	receiver := lp.getChan(url)
	for m := range receiver {
		msgs = append(msgs, m)
	}
	return
}

func (lp *LocalProvider) AddListener(url *url.URL, listener func(msg Message), options ...Option) (err error) {
	// Get channel first before locking to avoid dead locl
	channel := lp.getChan(url)
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
		go func() {

			for m := range channel {
				for name, listener := range lp.listeners[url.Host] {
					if name == unnamedListeners {
						for _, l := range listener {
							go l(m)
						}
					} else {
						lIDx := rand.Intn(len(listener))
						go listener[lIDx](m)
					}
				}
			}
		}()
	}
	return
}

func (lp *LocalProvider) Setup() (err error) {
	lp.mutex = sync.Mutex{}
	lp.destinations = make(map[string]chan Message)
	lp.listeners = make(map[string]map[string][]func(msg Message))
	return nil
}

func (lp *LocalProvider) Close() (err error) {
	for dest, ch := range lp.destinations {
		logger.TraceF("closing channel for desination %s", dest)
		ioutils.CloseChannel[Message](ch)
	}
	return
}

func (lp *LocalProvider) Schemes() (schemes []string) {
	schemes = localProviderSchemes
	return
}

package messaging

import (
	"errors"
	"net/url"
	"sync"
)

const (
	chanScheme = "chan"
)

var localProviderSchemes = []string{chanScheme}

// LocalProvider is an implementation of the Provider interface
type LocalProvider struct {
	mutex sync.Mutex
	local map[*url.URL]chan Message
}

func (lp *LocalProvider) NewMessage(scheme string, options ...Option) (msg Message, err error) {
	msg = NewLocalMessage()
	return
}

func (lp *LocalProvider) getChan(url *url.URL) (result chan Message) {
	var ok bool
	result, ok = lp.local[url]
	if !ok {
		lp.mutex.Lock()
		defer lp.mutex.Unlock()
		localMsgChannel := make(chan Message)
		lp.local[url] = localMsgChannel
		result = localMsgChannel
	}
	return
}

func (lp *LocalProvider) Send(url *url.URL, msg Message, options ...Option) (err error) {
	destination := lp.getChan(url)
	go func() {
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
	localListener := lp.getChan(url)
	for {
		val, ok := <-localListener
		if !ok {
			err = errors.New("channel is closed")
			return
		}
		listener(val)
	}
}

func (lp *LocalProvider) Setup() {
	lp.mutex = sync.Mutex{}
	lp.local = make(map[*url.URL]chan Message)
}

func (lp *LocalProvider) Schemes() (schemes []string) {
	schemes = localProviderSchemes
	return
}

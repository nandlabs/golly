package messaging

import (
	"bytes"
	"reflect"
)

type LocalMessage struct {
	*BaseMessage
}

func NewLocalMessage() *LocalMessage {
	return &LocalMessage{
		&BaseMessage{
			headers:     make(map[string]interface{}),
			headerTypes: make(map[string]reflect.Kind),
			body:        &bytes.Buffer{},
		},
	}
}

func (lm *LocalMessage) Rsvp(yes bool, options ...Option) (err error) {
	// Local message does not support RSVP
	return
}

// multiple endpoint support support

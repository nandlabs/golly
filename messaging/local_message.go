package messaging

import (
	"bytes"
	"reflect"

	"oss.nandlabs.io/golly/uuid"
)

type LocalMessage struct {
	*BaseMessage
}

func NewLocalMessage() (msg Message, err error) {
	var uid *uuid.UUID
	uid, err = uuid.V4()
	if err != nil {
		return
	}
	msg = &LocalMessage{
		BaseMessage: &BaseMessage{
			id:          uid.String(),
			headers:     make(map[string]interface{}),
			headerTypes: make(map[string]reflect.Kind),
			body:        &bytes.Buffer{},
		},
	}
	return
}

func (lm *LocalMessage) Rsvp(yes bool, options ...Option) (err error) {
	// Local message does not support RSVP
	return
}

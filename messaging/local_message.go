package messaging

type LocalMessage struct {
	*BaseMessage
}

func NewLocalMessage() (msg Message, err error) {
	var baseMsg *BaseMessage
	baseMsg, err = NewBaseMessage()
	if err == nil {
		msg = &LocalMessage{
			BaseMessage: baseMsg,
		}
	}
	return
}

func (lm *LocalMessage) Rsvp(yes bool, options ...Option) (err error) {
	// Local message does not support RSVP
	return
}

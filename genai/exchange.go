package genai

import (
	"bytes"
	"fmt"
	"net/url"

	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/ioutils"
)

// Actor is the type of the actor that sent the message
type Actor string

const (
	// UserActor represents a user
	UserActor Actor = "USER"
	// SystemActor represents the system
	SystemActor Actor = "SYSTEM"
	//AIActor represents an AI
	AIActor Actor = "AI"
	// FunctionActor represents a function
	FunctionActor Actor = "FUNCTION"
	//ToolActor represents a tool
	ToolActor Actor = "TOOL"
	//AgentActor represents an agent
	AgentActor Actor = "AGENT"
	//ContextActor represents a context
	PromptActor Actor = "PROMPT"
)

// Exchange is the interface that represents an exchange between users and the ai
type Exchange interface {
	// Id of the Exchange
	Id() string
	// Attributes of the Exchange
	Attributes() map[string]any
	// MsgsFrmActors returns true if the exchange has messages from the given actor
	HasMsgsFrmActors(actor ...Actor) bool
	//Messages of the exchange
	Messages() []*Message
	// MsgsByMime returns the messages of the exchange that match the given MIME type
	MsgsByMime(mime string) []*Message
	// MsgsFrmActors returns the messages of the exchange that match the given actor
	MsgsByActors(actor ...Actor) []*Message
	// AddAll adds all the messages to the exchange
	Add(message ...*Message)
	// AddTxtMsg adds a new text message to the exchnage
	AddTxtMsg(text string, actor Actor) (*Message, error)
	// AddJsonMsg adds a new JSON message to the exchnage
	AddJsonMsg(data interface{}, actor Actor) (*Message, error)
	// AddFileMsg adds a new file message to the exchnage.
	AddFileMsg(u string, mimeType string, actor Actor) (*Message, error)
	// AddBinMsg adds a new binary message to the exchnage
	AddBinMsg(data []byte, mimeType string, actor Actor) (*Message, error)
}

type exchangeImpl struct {
	//id of the message
	id string
	//includes the Attributes interface from the utils package
	attributes map[string]any
	//messages of the message
	messages []*Message
}

// NewExchange creates a new message
func NewExchange(id string) Exchange {

	return &exchangeImpl{
		id:         id,
		attributes: make(map[string]any),
	}
}

// Id returns the id of the message
func (e *exchangeImpl) Id() string {
	return e.id
}

// Attributes returns the attributes of the message
func (e *exchangeImpl) Attributes() map[string]any {
	return e.attributes
}

// HasMsgsFrmActors returns true if the message has messages from the given actor
func (e *exchangeImpl) HasMsgsFrmActors(actor ...Actor) bool {
	for _, message := range e.messages {
		for _, a := range actor {
			if message.Actor() != a {
				return false
			}
		}
	}
	return true
}

// Messages returns the messages of the message
func (e *exchangeImpl) Messages() []*Message {
	return e.messages
}

// CurrentMessage returns the last message of the message
func (e *exchangeImpl) CurrentMessage() *Message {
	if len(e.messages) > 0 {
		return e.messages[len(e.messages)-1]
	}
	return nil
}

// MsgsByMime returns the messages of the message that match the given MIME type
func (e *exchangeImpl) MsgsByMime(mime string) (messages []*Message) {
	for _, msg := range e.messages {
		if msg.Mime() == mime {
			messages = append(messages, msg)
		}
	}
	return
}

// MsgsByActor returns the messages of the message that match the given actor
func (e *exchangeImpl) MsgsByActors(actor ...Actor) (messages []*Message) {
	for _, message := range e.messages {
		for _, a := range actor {
			if message.Actor() == a {
				messages = append(messages, message)
			}
		}
	}
	return
}

// AddPart adds a new message to the beginning of the message
func (e *exchangeImpl) Prepend(message ...*Message) {
	e.messages = append(message, e.messages...)
}

// Addd adds a new message to the exchnage
func (e *exchangeImpl) Add(messages ...*Message) {
	e.messages = append(e.messages, messages...)
}

// AddTxtMsg adds a new text message to the exchnage
func (e *exchangeImpl) AddTxtMsg(text string, actor Actor) (message *Message, err error) {
	buf := new(bytes.Buffer)
	message = &Message{
		rwer:     buf,
		mimeType: ioutils.MimeTextPlain,
		msgActor: actor,
	}
	_, err = buf.WriteString(text)
	e.messages = append(e.messages, message)
	return
}

// AddMsgFrmTemplate adds a new text message to the exchnage
func (e *exchangeImpl) AddMsgFrmTemplate(templateId string, parameters map[string]any, actor Actor) (message *Message, err error) {
	template := GetPromptTemplate(templateId)
	if template == nil {
		err = fmt.Errorf("template %s not found", templateId)
		return
	}
	buf := new(bytes.Buffer)
	err = template.WriteTo(buf, parameters)
	if err != nil {
		return
	}
	message = &Message{
		rwer:     buf,
		mimeType: ioutils.MimeTextPlain,
		msgActor: actor,
	}
	e.messages = append(e.messages, message)
	return
}

// AddJsonMsg adds a new JSON message to the exchnage
func (e *exchangeImpl) AddJsonMsg(data interface{}, actor Actor) (message *Message, err error) {
	var c codec.Codec
	c, err = codec.GetDefault(ioutils.MimeApplicationJSON)
	if err != nil {
		return
	}
	message = &Message{
		rwer:     new(bytes.Buffer),
		mimeType: ioutils.MimeApplicationJSON,
		msgActor: actor,
	}
	err = c.Write(data, message.rwer)
	if err != nil {
		return
	}
	e.messages = append(e.messages, message)
	return
}

// AddFileMsg adds a new file message to the exchnage
func (e *exchangeImpl) AddFileMsg(u string, mimeType string, actor Actor) (message *Message, err error) {
	fileUrl, err := url.Parse(u)

	if err == nil {
		message = &Message{
			u:        fileUrl,
			mimeType: mimeType,
			msgActor: actor,
		}
		e.messages = append(e.messages, message)
	}
	return
}

// AddBinMsg adds a new binary message to the exchnage
func (e *exchangeImpl) AddBinMsg(data []byte, mimeType string, actor Actor) (message *Message, err error) {
	message = &Message{
		rwer:     new(bytes.Buffer),
		mimeType: mimeType,
		msgActor: actor,
	}
	_, err = message.Write(data)
	e.messages = append(e.messages, message)
	return
}

// LastMessage returns the last message of the exchange
func (e *exchangeImpl) LastMessage() *Message {
	if len(e.messages) > 0 {
		return e.messages[len(e.messages)-1]
	}
	return nil
}

package messaging

import (
	"io"
)

// Header defines all the header interfaces required by the messaging clients
type Header interface {
	// SetHeader sets the byte header value for the Message header
	SetHeader(key string, value []byte)
	// SetStrHeader sets the string header value for the Message header
	SetStrHeader(key string, value string)
	// SetBoolHeader sets the boolean header value for the Message header
	SetBoolHeader(key string, value bool)
	// SetIntHeader sets the int header value for the Message header
	SetIntHeader(key string, value int)
	// SetInt8Header sets the int8 header value for the Message header
	SetInt8Header(key string, value int8)
	// SetInt16Header sets the int16 header value for the Message header
	SetInt16Header(key string, value int16)
	// SetInt32Header sets the int32 header value for the Message header
	SetInt32Header(key string, value int32)
	// SetInt64Header sets the int64 header value for the Message header
	SetInt64Header(key string, value int64)
	// SetFloatHeader sets the float32 header value for the Message header
	SetFloatHeader(key string, value float32)
	// SetFloat64Header sets the float64 header value for the Message header
	SetFloat64Header(key string, value float64)

	// GetHeader returns the value of the key set in the headers if exists in the byte[] value
	GetHeader(key string) (value []byte, exists bool)
	// GetStrHeader returns the value of the key set in the headers if exists in the string value
	GetStrHeader(key string) (value string, exists bool)
	// GetBoolHeader returns the value of the key set in the headers if exists in the bool value
	GetBoolHeader(key string) (value bool, exists bool)
	// GetIntHeader returns the value of the key set in the headers if exists in the int value
	GetIntHeader(key string) (value int, exists bool)
	// GetInt8Header returns the value of the key set in the headers if exists in the int8 value
	GetInt8Header(key string) (value int8, exists bool)
	// GetInt16Header returns the value of the key set in the headers if exists in the int16 value
	GetInt16Header(key string) (value int16, exists bool)
	// GetInt32Header returns the value of the key set in the headers if exists in the int32 value
	GetInt32Header(key string) (value int32, exists bool)
	// GetInt64Header returns the value of the key set in the headers if exists in the int64 value
	GetInt64Header(key string) (value int64, exists bool)
	// GetFloatHeader returns the value of the key set in the headers if exists in the float32 value
	GetFloatHeader(key string) (value float32, exists bool)
	// GetFloat64Header returns the value of the key set in the headers if exists in the float64 value
	GetFloat64Header(key string) (value float64, exists bool)
}

// Body defines all the body interfaces required by the body of the messaging client
type Body interface {
	// SetBodyStr sets the string body to the Message structure
	SetBodyStr(in string) (int, error)
	// SetBodyBytes sets the byte[] body to the Message structure
	SetBodyBytes(int []byte) (int, error)
	// SetFrom sets the Reader body to the Message structure
	SetFrom(content io.Reader) (int64, error)
	// WriteJSON sets the JSON body to the Message structure
	WriteJSON(int interface{}) error
	// WriteXML sets the XML body to the Message structure
	WriteXML(in interface{}) error
	// WriteContent sets the custom body type based on the contentType to the Message structure
	WriteContent(in interface{}, contentType string) error

	// ReadBody reads the Reader body from the Message structure
	ReadBody() io.Reader
	// ReadBytes reads the []byte body from the Message structure
	ReadBytes() []byte
	// ReadAsStr reads the string body from the Message structure
	ReadAsStr() string
	// ReadJSON reads the JSON body from the Message structure
	ReadJSON(out interface{}) error
	// ReadXML reads the XML body from the Message structure
	ReadXML(out interface{}) error
	// ReadContent reads the content body based on the contentType from the Message structure
	ReadContent(out interface{}, contentType string) error
}

// Message interface wil be implemented by all third party implementation such as
//aws - sns, sqs,
//gcp -> pub/sub, gcm,
//messaging -> amqp, kafka
type Message interface {
	Header
	Body
	// Rsvp function provides a facade to acknowledge the message to the provider indicating the acceptance or rejection
	//as mentioned by the first bool parameter.
	//Additional options can be set for indicating further actions.
	//This functionality is purely dependent on the capability of the provider to accept an acknowledgement.
	Rsvp(bool, ...Option) error
}

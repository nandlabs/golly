package messaging

import (
	"bytes"
	"io"
	"reflect"

	"oss.nandlabs.io/golly/codec"
	"oss.nandlabs.io/golly/ioutils"
	"oss.nandlabs.io/golly/uuid"
)

type BaseMessage struct {
	id          string
	headers     map[string]interface{}
	headerTypes map[string]reflect.Kind
	body        *bytes.Buffer
}

func NewBaseMessage() (baseMsg *BaseMessage, err error) {
	var uid *uuid.UUID
	uid, err = uuid.V4()
	if err == nil {
		baseMsg = &BaseMessage{
			id:          uid.String(),
			headers:     make(map[string]interface{}),
			headerTypes: make(map[string]reflect.Kind),
			body:        &bytes.Buffer{},
		}
	}
	return
}

func (bm *BaseMessage) Id() string {
	return bm.id
}

func (bm *BaseMessage) SetBodyStr(input string) (n int, err error) {
	n, err = bm.body.WriteString(input)
	return
}

func (bm *BaseMessage) SetBodyBytes(input []byte) (n int, err error) {
	n, err = bm.body.Write(input)
	return
}

func (bm *BaseMessage) SetFrom(content io.Reader) (n int64, err error) {
	n, err = io.Copy(bm.body, content)
	return
}

func (bm *BaseMessage) WriteJSON(input interface{}) (err error) {
	err = bm.WriteContent(input, ioutils.MimeApplicationJSON)
	return
}

func (bm *BaseMessage) WriteXML(input interface{}) (err error) {
	err = bm.WriteContent(input, ioutils.MimeTextXML)
	return
}

func (bm *BaseMessage) WriteContent(input interface{}, contentType string) (err error) {
	var cdc codec.Codec
	// TODO : provide options to customize codec options
	cdc, err = codec.GetDefault(contentType)
	if err == nil {
		err = cdc.Write(input, bm.body)
	}
	return
}

func (bm *BaseMessage) ReadBody() io.Reader {
	return bm.body
}

func (bm *BaseMessage) ReadBytes() []byte {
	return bm.body.Bytes()
}

func (bm *BaseMessage) ReadAsStr() string {
	return bm.body.String()
}

func (bm *BaseMessage) ReadJSON(out interface{}) (err error) {
	err = bm.ReadContent(out, ioutils.MimeApplicationJSON)
	return
}

func (bm *BaseMessage) ReadXML(out interface{}) (err error) {
	err = bm.ReadContent(out, ioutils.MimeTextXML)
	return
}

func (bm *BaseMessage) ReadContent(out interface{}, contentType string) (err error) {
	var cdc codec.Codec
	// TODO: provide options to customize codec options
	cdc, err = codec.GetDefault(contentType)
	if err == nil {
		err = cdc.Read(bm.body, out)
	}
	return
}

func (bm *BaseMessage) SetHeader(key string, value []byte) {
	bm.headers[key] = value
	bm.headerTypes[key] = reflect.Array
}

func (bm *BaseMessage) SetStrHeader(key string, value string) {
	bm.headers[key] = value
	bm.headerTypes[key] = reflect.String
}

func (bm *BaseMessage) SetBoolHeader(key string, value bool) {
	bm.headers[key] = value
	bm.headerTypes[key] = reflect.Bool
}

func (bm *BaseMessage) SetIntHeader(key string, value int) {
	bm.headers[key] = value
	bm.headerTypes[key] = reflect.Int
}

func (bm *BaseMessage) SetInt8Header(key string, value int8) {
	bm.headers[key] = value
	bm.headerTypes[key] = reflect.Int8
}

func (bm *BaseMessage) SetInt16Header(key string, value int16) {
	bm.headers[key] = value
	bm.headerTypes[key] = reflect.Int16
}

func (bm *BaseMessage) SetInt32Header(key string, value int32) {
	bm.headers[key] = value
	bm.headerTypes[key] = reflect.Int32
}

func (bm *BaseMessage) SetInt64Header(key string, value int64) {
	bm.headers[key] = value
	bm.headerTypes[key] = reflect.Int64
}

func (bm *BaseMessage) SetFloatHeader(key string, value float32) {
	bm.headers[key] = value
	bm.headerTypes[key] = reflect.Float32
}

func (bm *BaseMessage) SetFloat64Header(key string, value float64) {
	bm.headers[key] = value
	bm.headerTypes[key] = reflect.Float64
}

func (bm *BaseMessage) GetHeader(key string) (value []byte, exists bool) {
	var v interface{}
	v, exists = bm.headers[key]
	if exists {
		value = v.([]byte)
	}
	return
}

func (bm *BaseMessage) GetStrHeader(key string) (value string, exists bool) {
	var v interface{}
	v, exists = bm.headers[key]
	if exists {
		value = v.(string)
	}
	return
}

func (bm *BaseMessage) GetBoolHeader(key string) (value bool, exists bool) {
	var v interface{}
	v, exists = bm.headers[key]
	if exists {
		value = v.(bool)
	}
	return
}

func (bm *BaseMessage) GetIntHeader(key string) (value int, exists bool) {
	var v interface{}
	v, exists = bm.headers[key]
	if exists {
		value = v.(int)
	}
	return
}

func (bm *BaseMessage) GetInt8Header(key string) (value int8, exists bool) {
	var v interface{}
	v, exists = bm.headers[key]
	if exists {
		value = v.(int8)
	}
	return
}

func (bm *BaseMessage) GetInt16Header(key string) (value int16, exists bool) {
	var v interface{}
	v, exists = bm.headers[key]
	if exists {
		value = v.(int16)
	}
	return
}

func (bm *BaseMessage) GetInt32Header(key string) (value int32, exists bool) {
	var v interface{}
	v, exists = bm.headers[key]
	if exists {
		value = v.(int32)
	}
	return
}

func (bm *BaseMessage) GetInt64Header(key string) (value int64, exists bool) {
	var v interface{}
	v, exists = bm.headers[key]
	if exists {
		value = v.(int64)
	}
	return
}

func (bm *BaseMessage) GetFloatHeader(key string) (value float32, exists bool) {
	var v interface{}
	v, exists = bm.headers[key]
	if exists {
		value = v.(float32)
	}
	return
}

func (bm *BaseMessage) GetFloat64Header(key string) (value float64, exists bool) {
	var v interface{}
	v, exists = bm.headers[key]
	if exists {
		value = v.(float64)
	}
	return
}

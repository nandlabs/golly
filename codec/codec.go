package codec

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"

	"oss.nandlabs.io/golly/textutils"
)

const (
	JSON                      = "application/json"
	XML                       = "text/xml"
	XmlApplicationContentType = "application/xml"
	YAML                      = "text/yaml"
	defaultValidateOnRead     = false
	defaultValidateBefWrite   = false
	ValidateOnRead            = "ValidateOnRead"
	ValidateBefWrite          = "ValidateBefWrite"
	Charset                   = "charset"
	JsonEscapeHTML            = "JsonEscapeHTML"
	PrettyPrint               = "PrettyPrint"
)

// StringEncoder Interface
type StringEncoder interface {
	//EncodeToString will encode a type to string
	EncodeToString(v interface{}) (string, error)
}

// BytesEncoder Interface
type BytesEncoder interface {
	// EncodeToBytes will encode the provided type to []byte
	EncodeToBytes(v interface{}) ([]byte, error)
}

// StringDecoder Interface
type StringDecoder interface {
	//DecodeString will decode  a type from string
	DecodeString(s string, v interface{}) error
}

// BytesDecoder Interface
type BytesDecoder interface {
	//DecodeBytes will decode a type from an array of bytes
	DecodeBytes(b []byte, v interface{}) error
}

// Encoder Interface
type Encoder interface {
	StringEncoder
	BytesEncoder
}

// Decoder Interface
type Decoder interface {
	StringDecoder
	BytesDecoder
}

type ReaderWriter interface {
	//Write a type to writer
	Write(v interface{}, w io.Writer) error
	//Read a type from a reader
	Read(r io.Reader, v interface{}) error
}

type Validator interface {
	Validate() (bool, []error)
}

// Codec Interface
type Codec interface {
	Decoder
	Encoder
	ReaderWriter
	//SetOption sets an option to the reader and writer
	SetOption(key string, value interface{})
}

type BaseCodec struct {
	readerWriter ReaderWriter
	options      map[string]interface{}
	once         sync.Once
}

// getDefaultCodecOption returns the default codec option
func getDefaultCodecOption() (defaultCodecOption map[string]interface{}) {
	defaultCodecOption = make(map[string]interface{})
	defaultCodecOption[ValidateOnRead] = defaultValidateOnRead
	defaultCodecOption[ValidateBefWrite] = defaultValidateBefWrite
	return

}

func (bc *BaseCodec) SetOption(key string, value interface{}) {
	bc.once.Do(func() {
		if bc.options == nil {
			bc.options = make(map[string]interface{})
		}
	})

	bc.options[key] = value
}

// GetDefault function creates an instance of codec based on the contentType and defaultOptions
func GetDefault(contentType string) (Codec, error) {
	return Get(contentType, getDefaultCodecOption())
}

// Get function creates an instance of codec based on the contentType and Options
func Get(contentType string, options map[string]interface{}) (c Codec, err error) {

	bc := &BaseCodec{
		options: options,
	}
	typ := contentType
	if strings.Contains(contentType, textutils.SemiColonStr) {
		values := strings.Split(contentType, textutils.SemiColonStr)
		l := len(values)
		for i := 0; i < l; i++ {
			val := strings.TrimSpace(values[i])
			if i == 0 {
				typ = val
			} else {
				if strings.HasPrefix(val, Charset) {
					charset := strings.Split(val, textutils.EqualStr)
					if len(charset) == 2 {
						//Charset is added to the options but not used by the known json,xml and yaml Read Writers
						bc.SetOption(Charset, strings.TrimSpace(charset[1]))
					}
				}
			}
		}
	}

	switch typ {
	case JSON:
		{
			bc.readerWriter = &jsonRW{options: options}

		}
	case XML, XmlApplicationContentType:
		{
			bc.readerWriter = &xmlRW{options: options}
		}
	case YAML:
		{
			bc.readerWriter = &yamlRW{options: options}
		}
	default:
		err = fmt.Errorf("Unsupported contentType %s", contentType)
	}

	if err == nil {
		c = bc
	}

	return
}

func (bc *BaseCodec) DecodeString(s string, v interface{}) error {
	r := strings.NewReader(s)
	return bc.Read(r, v)
}

func (bc *BaseCodec) DecodeBytes(b []byte, v interface{}) error {
	r := bytes.NewReader(b)
	return bc.Read(r, v)
}

// EncodeToBytes :
func (bc *BaseCodec) EncodeToBytes(v interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	e := bc.Write(v, buf)
	if e == nil {
		return buf.Bytes(), e
	} else {
		return nil, e
	}
}

func (bc *BaseCodec) EncodeToString(v interface{}) (string, error) {
	buf := &bytes.Buffer{}
	e := bc.Write(v, buf)
	if e == nil {
		return buf.String(), e
	} else {
		return textutils.EmptyStr, e
	}
}

func (bc *BaseCodec) Read(r io.Reader, v interface{}) (err error) {

	err = bc.readerWriter.Read(r, v)
	//Check if validation is  required after read
	if err == nil && bc.options != nil {
		if v, ok := bc.options[ValidateOnRead]; ok && v.(bool) {
			err = structValidator.Validate(v)
		}
	}
	return
}

func (bc *BaseCodec) Write(v interface{}, w io.Writer) (err error) {

	//Check if validation is  required before write
	if bc.options != nil {
		if v, ok := bc.options[ValidateBefWrite]; ok && v.(bool) {
			err = structValidator.Validate(v)
		}
	}
	if err == nil {
		err = bc.readerWriter.Write(v, w)
	}
	return
}

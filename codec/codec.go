package codec

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"

	"oss.nandlabs.io/golly/ioutils"
	"oss.nandlabs.io/golly/managers"
	"oss.nandlabs.io/golly/textutils"
)

const (
	defaultValidateOnRead   = false
	defaultValidateBefWrite = false
	ValidateOnRead          = "ValidateOnRead"
	ValidateBefWrite        = "ValidateBefWrite"
	Charset                 = "charset"
	JsonEscapeHTML          = "JsonEscapeHTML"
	PrettyPrint             = "PrettyPrint"
)

var codecManager = managers.NewItemManager[ReaderWriter]()

// StringEncoder Interface
type StringEncoder interface {
	// EncodeToString will encode a type to string
	EncodeToString(v interface{}) (string, error)
}

// BytesEncoder Interface
type BytesEncoder interface {
	// EncodeToBytes will encode the provided type to []byte
	EncodeToBytes(v interface{}) ([]byte, error)
}

// StringDecoder Interface
type StringDecoder interface {
	// DecodeString will decode  a type from string
	DecodeString(s string, v interface{}) error
}

// BytesDecoder Interface
type BytesDecoder interface {
	// DecodeBytes will decode a type from an array of bytes
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

// ReaderWriter is an interface that defines methods for writing and reading
// data to and from an io.Writer and io.Reader, respectively.
//
// Write writes the given value to the provided writer.
// It takes an interface{} value and an io.Writer, and returns an error if the write operation fails.
//
// Read reads data from the provided reader into the given value.
// It takes an io.Reader and an interface{} value, and returns an error if the read operation fails.
type ReaderWriter interface {
	// Write a type to writer
	Write(v interface{}, w io.Writer) error
	// Read a type from a reader
	Read(r io.Reader, v interface{}) error
	// MimeTypes returns a slice of strings representing the MIME types
	MimeTypes() []string
}

// Validator is an interface that defines a method for validating an object.
// The Validate method returns a boolean indicating whether the validation was
// successful, and a slice of errors detailing any validation issues.
type Validator interface {
	Validate() (bool, []error)
}

// Codec Interface
type Codec interface {
	Decoder
	Encoder
	ReaderWriter
	// SetOption sets an option to the reader and writer
	SetOption(key string, value interface{})
}

// BaseCodec is a struct that encapsulates a ReaderWriter interface, a set of options,
// and a sync.Once instance to ensure that certain operations are only performed once.
// It is designed to handle encoding and decoding operations with customizable options.
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

// SetOption sets an option for the BaseCodec instance. It initializes the options map
// if it hasn't been initialized yet. This method is thread-safe and can be called
// concurrently.
//
// Parameters:
//   - key: The option key as a string.
//   - value: The option value as an interface{}.
func (bc *BaseCodec) SetOption(key string, value interface{}) {
	bc.once.Do(func() {
		if bc.options == nil {
			bc.options = make(map[string]interface{})
		}
	})

	bc.options[key] = value
}

// MimeTypes
func (bc *BaseCodec) MimeTypes() []string {
	return bc.readerWriter.MimeTypes()
}

// GetDefault function creates an instance of codec based on the contentType and defaultOptions
func GetDefault(contentType string) (Codec, error) {
	return Get(contentType, getDefaultCodecOption())
}

// JsonCodec Provides a JSONCodec
// JsonCodec returns a Codec for handling JSON data.
// It retrieves the default Codec for the MIME type "application/json".
// If there is an error during retrieval, it is ignored and the default Codec is returned.
func JsonCodec() Codec {
	c, _ := GetDefault(ioutils.MimeApplicationJSON)
	return c
}

// XmlCodec returns a Codec for handling XML data.
// It retrieves the default Codec associated with the MIME type for XML text.
// The function ignores any error that might occur during the retrieval process.
func XmlCodec() Codec {
	c, _ := GetDefault(ioutils.MimeTextXML)
	return c
}

// YamlCodec Provides a YamlCodec
func YamlCodec() Codec {
	c, _ := GetDefault(ioutils.MimeTextYAML)
	return c
}

// Get returns a Codec based on the provided content type and options.
// It supports JSON, XML, and YAML content types. If the content type
// contains a charset, it is added to the options but not used by the
// known JSON, XML, and YAML Read Writers.
//
// Parameters:
//   - contentType: A string representing the MIME type of the content.
//   - options: A map of options to configure the Codec.
//
// Returns:
//   - c: A Codec configured for the specified content type.
//   - err: An error if the content type is unsupported.
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
			} else if strings.HasPrefix(val, Charset) {
				charset := strings.Split(val, textutils.EqualStr)
				if len(charset) == 2 {
					// Charset is added to the options but not used by the known json,xml and yaml Read Writers
					bc.SetOption(Charset, strings.TrimSpace(charset[1]))
				}
			}
		}
	}

	switch typ {
	case ioutils.MimeApplicationJSON:
		{
			bc.readerWriter = &jsonRW{options: options}

		}
	case ioutils.MimeTextXML, ioutils.MimeApplicationXML:
		{
			bc.readerWriter = &xmlRW{options: options}
		}
	case ioutils.MimeTextYAML:
		{
			bc.readerWriter = &yamlRW{options: options}
		}
	default:

		readerWriter := codecManager.Get(contentType)
		if readerWriter != nil {
			bc.readerWriter = readerWriter
		} else {
			err = fmt.Errorf("unsupported contentType %s", contentType)
		}
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
		return buf.Bytes(), nil
	}
	return nil, e
}

func (bc *BaseCodec) EncodeToString(v interface{}) (string, error) {
	buf := &bytes.Buffer{}
	e := bc.Write(v, buf)
	if e == nil {
		return buf.String(), nil
	}
	return textutils.EmptyStr, e
}

func (bc *BaseCodec) Read(r io.Reader, v interface{}) (err error) {

	err = bc.readerWriter.Read(r, v)
	// Check if validation is  required after read
	if err == nil && bc.options != nil {
		if v, ok := bc.options[ValidateOnRead]; ok && v.(bool) {
			err = structValidator.Validate(v)
		}
	}
	return
}

func (bc *BaseCodec) Write(v interface{}, w io.Writer) (err error) {

	// Check if validation is  required before write
	if bc.options != nil {
		if opt, ok := bc.options[ValidateBefWrite]; ok && opt.(bool) {
			err = structValidator.Validate(v)
		}
	}
	if err == nil {
		err = bc.readerWriter.Write(v, w)
	}
	return
}

func Register(contentType string, readerWriter ReaderWriter) {
	codecManager.Register(contentType, readerWriter)
}

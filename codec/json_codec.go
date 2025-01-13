package codec

import (
	"encoding/json"
	"io"

	"oss.nandlabs.io/golly/codec/validator"
	"oss.nandlabs.io/golly/ioutils"
)

const (
	jsonPrettyPrintPrefix = ""
	jsonPrettyPrintIndent = "  "
)

var structValidator = validator.NewStructValidator()
var jsonmimeTypes = []string{ioutils.MimeApplicationJSON}

type jsonRW struct {
	options map[string]interface{}
}

// Write encodes the given value v into JSON and writes it to the provided io.Writer w.
// It supports options for escaping HTML and pretty-printing the JSON output.
// The options are specified in the jsonRW struct's options map with the keys JsonEscapeHTML and PrettyPrint.
//
// Parameters:
//   - v: The value to be encoded into JSON.
//   - w: The io.Writer to write the JSON output to.
//
// Returns:
//   - error: An error if the encoding or writing fails, otherwise nil.
func (j *jsonRW) Write(v interface{}, w io.Writer) error {
	//only utf-8 charset is supported
	var escapeHtml = false
	var prettyPrint = false
	if j.options != nil {
		if v, ok := j.options[JsonEscapeHTML]; ok {
			escapeHtml = v.(bool)
		}

		if v, ok := j.options[PrettyPrint]; ok {
			prettyPrint = v.(bool)
		}

	}
	encoder := json.NewEncoder(w)
	if prettyPrint {
		encoder.SetIndent(jsonPrettyPrintPrefix, jsonPrettyPrintIndent)
	}
	encoder.SetEscapeHTML(escapeHtml)
	return encoder.Encode(v)

}

// Read reads JSON-encoded data from the provided io.Reader and decodes it into the specified interface{}.
// It returns an error if the decoding process fails.
//
// Parameters:
//
//	r - the io.Reader to read the JSON data from
//	v - the interface{} to decode the JSON data into
//
// Returns:
//
//	error - an error if the decoding process fails, or nil if successful
func (j *jsonRW) Read(r io.Reader, v interface{}) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(v)
}

// MimeTypes returns a slice of strings representing the MIME types
// that are supported by the jsonRW codec.
func (j *jsonRW) MimeTypes() []string {
	return jsonmimeTypes
}

package codec

import (
	"io"

	"github.com/goccy/go-yaml"
	"oss.nandlabs.io/golly/ioutils"
)

var yamlmimeTypes = []string{ioutils.MimeTextYAML}

type yamlRW struct {
	options map[string]interface{}
}

// Write encodes the given value v into YAML format and writes it to the provided io.Writer w.
// It returns an error if the encoding process fails.
//
// Parameters:
//
//	v - The value to be encoded into YAML format.
//	w - The io.Writer where the encoded YAML data will be written.
//
// Returns:
//
//	error - An error if the encoding process fails, otherwise nil.
func (y *yamlRW) Write(v interface{}, w io.Writer) error {
	encoder := yaml.NewEncoder(w)
	return encoder.Encode(v)
}

// Read reads YAML-encoded data from the provided io.Reader and decodes it into the provided interface{}.
// It returns an error if the decoding process fails.
//
// Parameters:
//   - r: An io.Reader from which the YAML data will be read.
//   - v: A pointer to the value where the decoded data will be stored.
//
// Returns:
//   - error: An error if the decoding process fails, otherwise nil.
func (y *yamlRW) Read(r io.Reader, v interface{}) error {
	decoder := yaml.NewDecoder(r)
	return decoder.Decode(v)
}

// MimeTypes returns a slice of strings representing the MIME types
// that are supported by the yamlRW codec.
func (y *yamlRW) MimeTypes() []string {
	return yamlmimeTypes
}

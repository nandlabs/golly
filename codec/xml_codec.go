package codec

import (
	"encoding/xml"
	"io"

	"oss.nandlabs.io/golly/ioutils"
)

const (
	xmlPrettyPrintPrefix = ""
	xmlPrettyPrintIndent = "    "
)

var xmlmimeTypes = []string{ioutils.MimeApplicationXML, ioutils.MimeTextXML}

type xmlRW struct {
	options map[string]interface{}
}

// Write encodes the given value v into XML format and writes it to the provided io.Writer w.
// If the PrettyPrint option is set to true in x.options, the output will be indented for readability.
//
// Parameters:
//   - v: The value to be encoded into XML.
//   - w: The io.Writer to which the encoded XML will be written.
//
// Returns:
//   - error: An error if the encoding or writing process fails, otherwise nil.
func (x *xmlRW) Write(v interface{}, w io.Writer) error {
	encoder := xml.NewEncoder(w)
	var prettyPrint = false
	if x.options != nil {
		if opt, ok := x.options[PrettyPrint]; ok {
			prettyPrint = opt.(bool)
		}
	}
	if prettyPrint {
		encoder.Indent(xmlPrettyPrintPrefix, xmlPrettyPrintIndent)
	}
	return encoder.Encode(v)

}

// Read reads XML data from the provided io.Reader and decodes it into the provided interface{}.
// It uses the xml.NewDecoder to decode the XML data.
// Parameters:
//   - r: An io.Reader from which the XML data will be read.
//   - v: A pointer to the value where the decoded XML data will be stored.
//
// Returns:
//   - An error if the decoding fails, otherwise nil.
func (x *xmlRW) Read(r io.Reader, v interface{}) error {
	decoder := xml.NewDecoder(r)
	return decoder.Decode(v)
}

// MimeTypes returns a slice of strings representing the MIME types
// that are supported by the xmlRW codec.
func (x *xmlRW) MimeTypes() []string {
	return xmlmimeTypes
}

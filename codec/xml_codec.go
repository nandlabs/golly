package codec

import (
	"encoding/xml"
	"io"
)

const (
	xmlPrettyPrintPrefix = ""
	xmlPrettyPrintIndent = "    "
)

type xmlRW struct {
	options map[string]interface{}
}

func (x *xmlRW) Write(v interface{}, w io.Writer) error {
	encoder := xml.NewEncoder(w)
	var prettyPrint = false
	if x.options != nil {
		if v, ok := x.options[PrettyPrint]; ok {
			prettyPrint = v.(bool)
		}
	}
	if prettyPrint {
		encoder.Indent(xmlPrettyPrintPrefix, xmlPrettyPrintIndent)
	}
	return encoder.Encode(v)

}

func (x *xmlRW) Read(r io.Reader, v interface{}) error {
	decoder := xml.NewDecoder(r)
	return decoder.Decode(v)
}

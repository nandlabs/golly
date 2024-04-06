package codec

import (
	"encoding/json"
	"io"

	"oss.nandlabs.io/golly/codec/validator"
)

const (
	jsonPrettyPrintPrefix = ""
	jsonPrettyPrintIndent = "  "
)

var structValidator = validator.NewStructValidator()

type jsonRW struct {
	options map[string]interface{}
}

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

func (j *jsonRW) Read(r io.Reader, v interface{}) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(v)
}

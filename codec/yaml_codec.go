package codec

import (
	"io"

	"gopkg.in/yaml.v3"
)

type yamlRW struct {
	options map[string]interface{}
}

func (y *yamlRW) Write(v interface{}, w io.Writer) error {
	encoder := yaml.NewEncoder(w)
	return encoder.Encode(v)
}

func (y *yamlRW) Read(r io.Reader, v interface{}) error {
	decoder := yaml.NewDecoder(r)
	return decoder.Decode(v)
}

package config

import (
	"encoding/json"

	"gopkg.in/yaml.v2"
)

type Codec interface {
	Name() string
	Unmarshal([]byte, interface{}) error
}

var (
	codecs map[string]Codec
)

func init() {
	codecs = make(map[string]Codec)
	RegisterCodec(&YamlCodec{})
	RegisterCodec(&JSONCodec{})
}

func RegisterCodec(c Codec) {
	codecs[c.Name()] = c
}

func GetCodec(name string) Codec {
	return codecs[name]
}

type YamlCodec struct{}

func (*YamlCodec) Name() string {
	return "yaml"
}

func (c *YamlCodec) Unmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}

type JSONCodec struct{}

func (*JSONCodec) Name() string {
	return "json"
}

func (c *JSONCodec) Unmarshal(in []byte, out interface{}) error {
	return json.Unmarshal(in, out)
}

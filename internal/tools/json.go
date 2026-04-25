package tools

import (
	"avito-shop/internal/config"
	"encoding/json"

	"github.com/bytedance/sonic"
)

type JSONCodec interface {
	Marshal(data any) ([]byte, error)
	MarshalIndent(v interface{}, prefix, indent string) ([]byte, error)
	Unmarshal(buf []byte, val interface{}) error
}

func NewJSONCodec() JSONCodec {
	switch config.App.Tools.JSON {
	case "sonic":
		return sonic.ConfigDefault
	default:
		return jsonGo{}
	}
}

type jsonGo struct{}

func (j jsonGo) Marshal(data any) ([]byte, error) {
	return json.Marshal(data)
}

func (j jsonGo) MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

func (j jsonGo) Unmarshal(buf []byte, val interface{}) error {
	return json.Unmarshal(buf, val)
}

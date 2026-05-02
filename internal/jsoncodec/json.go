package jsoncodec

import (
	"encoding/json"

	"github.com/bytedance/sonic"
)

type JSONCodec interface {
	Marshal(data any) ([]byte, error)
	MarshalIndent(v any, prefix, indent string) ([]byte, error)
	Unmarshal(buf []byte, val any) error
}

func NewJSONCodec(codecType string) JSONCodec {
	switch codecType {
	case "sonic":
		return sonic.ConfigDefault
	default:
		return jsonGo{}
	}
}

type jsonGo struct{}

func (jsonGo) Marshal(data any) ([]byte, error) {
	return json.Marshal(data)
}

func (jsonGo) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

func (jsonGo) Unmarshal(buf []byte, val any) error {
	return json.Unmarshal(buf, val)
}

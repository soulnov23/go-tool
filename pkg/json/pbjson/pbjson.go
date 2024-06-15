// Package pbjson
package pbjson

import (
	"github.com/soulnov23/go-tool/pkg/json/jsoniter"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	marshaler = protojson.MarshalOptions{
		Multiline:       false,
		Indent:          "",
		AllowPartial:    true,
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: true,
	}
	unmarshaler = protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: true,
		RecursionLimit: 0,
	}
)

// Marshal
func Marshal(value any) ([]byte, error) {
	input, ok := value.(proto.Message)
	if !ok {
		return jsoniter.Marshal(value)
	}
	return marshaler.Marshal(input)
}

// Unmarshal
func Unmarshal(data []byte, value any) error {
	input, ok := value.(proto.Message)
	if !ok {
		return jsoniter.Unmarshal(data, value)
	}
	return unmarshaler.Unmarshal(data, input)
}

package serialization

import (
	jsoniter "github.com/json-iterator/go"
)

func init() {
	RegisterSerializer(SerializationTypeJSON, &JSONSerialization{
		API: jsoniter.Config{
			IndentionStep:           0,
			MarshalFloatWith6Digits: false,
			EscapeHTML:              true,
			SortMapKeys:             true,
			// https://github.com/json-iterator/go/blob/master/adapter.go:100
			// 当用interface{}来Unmarshal接收值的时候jsoniter会解析成float64，有精度丢失，UseNumber=true使用Number类型接收，后续通过接口转换成需要的类型
			UseNumber: true,
			// 允许定义的Struct中有未知的字段
			DisallowUnknownFields:         false,
			TagKey:                        "json",
			OnlyTaggedField:               true,
			ValidateJsonRawMessage:        true,
			ObjectFieldMustBeSimpleString: false,
			CaseSensitive:                 true,
		}.Froze(),
	})
}

type JSONSerialization struct {
	API jsoniter.API
}

func (s *JSONSerialization) Unmarshal(in []byte, body interface{}) error {
	return s.API.Unmarshal(in, body)
}

func (s *JSONSerialization) Marshal(body interface{}) ([]byte, error) {
	return s.API.Marshal(body)
}

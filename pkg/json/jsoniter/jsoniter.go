// Package jsoniter
package jsoniter

import (
	jsoniter "github.com/json-iterator/go"
)

var api jsoniter.API

func init() {
	api = jsoniter.Config{
		IndentionStep:           0,
		MarshalFloatWith6Digits: false,
		EscapeHTML:              false,
		SortMapKeys:             false,
		// https://github.com/json-iterator/go/blob/master/adapter.go:100
		// 当用any来Unmarshal接收值的时候jsoniter会解析成float64，有精度丢失，UseNumber=true使用Number类型接收，后续通过接口转换成需要的类型
		UseNumber: true,
		// 允许定义的Struct中有未知的字段
		DisallowUnknownFields:         false,
		TagKey:                        "json",
		OnlyTaggedField:               true,
		ValidateJsonRawMessage:        true,
		ObjectFieldMustBeSimpleString: false,
		CaseSensitive:                 true,
	}.Froze()
}

// Marshal
func Marshal(value any) ([]byte, error) {
	return api.Marshal(value)
}

// Unmarshal
func Unmarshal(data []byte, value any) error {
	return api.Unmarshal(data, value)
}

// MarshalToString
func MarshalToString(value any) (string, error) {
	return api.MarshalToString(value)
}

// UnmarshalFromString
func UnmarshalFromString(data string, value any) error {
	return api.UnmarshalFromString(data, value)
}

// Stringify 使json struct字符串化
func Stringify(value any) string {
	data, err := MarshalToString(value)
	if err != nil {
		return ""
	}
	return data
}

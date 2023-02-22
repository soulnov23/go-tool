package utils

import (
	"reflect"
	"strings"
	"unsafe"
)

func Byte2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func String2Byte(s string) (b []byte) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}

func String2Map(data string, fieldSep string, valueSep string) map[string]string {
	recordMap := map[string]string{}
	fieldSlice := strings.Split(data, fieldSep)
	for _, kv := range fieldSlice {
		valueSlice := strings.Split(kv, valueSep)
		if len(valueSlice) == 2 {
			recordMap[valueSlice[0]] = valueSlice[1]
		}
	}
	return recordMap
}

func Map2String(recordMap map[string]string) string {
	var builder strings.Builder
	for key, value := range recordMap {
		builder.WriteString(key + "=" + value + "&")
	}
	builder.Len()
	return builder.String()[0 : builder.Len()-1]
}

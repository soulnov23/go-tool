package utils

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/soulnov23/go-tool/pkg/json"
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

func Intf2String(row interface{}) string {
	switch v := row.(type) {
	case nil:
		return ""
	case *string:
		return fmt.Sprintf("%v", *v)
	case *bool:
		return fmt.Sprintf("%v", *v)
	case *uint8:
		return fmt.Sprintf("%v", *v)
	case *uint16:
		return fmt.Sprintf("%v", *v)
	case *uint32:
		return fmt.Sprintf("%v", *v)
	case *uint64:
		return fmt.Sprintf("%v", *v)
	case *int8:
		return fmt.Sprintf("%v", *v)
	case *int16:
		return fmt.Sprintf("%v", *v)
	case *int32:
		return fmt.Sprintf("%v", *v)
	case *int64:
		return fmt.Sprintf("%v", *v)
	case *float32:
		return fmt.Sprintf("%v", *v)
	case *float64:
		return fmt.Sprintf("%v", *v)
	case *int:
		return fmt.Sprintf("%v", *v)
	case *uint:
		return fmt.Sprintf("%v", *v)
	case *[]byte:
		return fmt.Sprintf("%v", *v)
	case string, bool, uint8, uint16, uint32, uint64, int8, int16, int32, int64, float32, float64, int, uint:
		return fmt.Sprintf("%v", v)
	case []byte:
		return string(v)
	case *struct{}:
		result, err := json.Marshal(*v)
		if err != nil {
			return ""
		}
		return string(result)
	case *interface{}:
		return Intf2String(*v)
	case interface{}:
		switch vv := v.(type) {
		case string, bool, uint8, uint16, uint32, uint64, int8, int16, int32, int64, float32, float64, int, uint:
			return fmt.Sprintf("%v", vv)
		case []byte:
			return string(vv)
		default:
			return fmt.Sprintf("%v", vv)
		}
	default:
		return fmt.Sprintf("%v", v)
	}
}

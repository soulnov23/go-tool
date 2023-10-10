package strconv

import (
	"fmt"
	"sort"
	"strings"

	"github.com/soulnov23/go-tool/pkg/json"
)

func StringToMap(data string, fieldSep string, valueSep string) map[string]string {
	recordMap := map[string]string{}
	fieldSlice := strings.Split(data, fieldSep)
	for _, kv := range fieldSlice {
		valueSlice := strings.Split(kv, valueSep)
		if len(valueSlice) == 2 {
			recordMap[valueSlice[0]] = valueSlice[1]
		} else if len(valueSlice) == 1 && strings.Count(kv, valueSep) == 1 {
			recordMap[valueSlice[0]] = ""
		}
	}
	return recordMap
}

func MapToString(recordMap map[string]string, sorted bool) string {
	size := len(recordMap)
	if size == 0 {
		return ""
	}

	var builder strings.Builder
	if sorted {
		keys := make([]string, 0, size)
		for key := range recordMap {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			builder.WriteString(key + "=" + recordMap[key] + "&")
		}
	} else {
		for key, value := range recordMap {
			builder.WriteString(key + "=" + value + "&")
		}
	}
	return builder.String()[0 : builder.Len()-1]
}

func AnyToString(row any) string {
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
	case *any:
		return AnyToString(*v)
	case any:
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

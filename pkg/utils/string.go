package utils

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/soulnov23/go-tool/pkg/json/jsoniter"
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
	case []byte:
		return BytesToString(v)
	case *[]byte:
		return BytesToString(*v)
	case string:
		return v
	case *string:
		return *v
	case bool:
		return strconv.FormatBool(v)
	case *bool:
		return strconv.FormatBool(*v)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case *uint:
		return strconv.FormatUint(uint64(*v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case *uint8:
		return strconv.FormatUint(uint64(*v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case *uint16:
		return strconv.FormatUint(uint64(*v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case *uint32:
		return strconv.FormatUint(uint64(*v), 10)
	case uint64:
		return strconv.FormatUint(uint64(v), 10)
	case *uint64:
		return strconv.FormatUint(uint64(*v), 10)
	case int:
		return strconv.FormatInt(int64(v), 10)
	case *int:
		return strconv.FormatInt(int64(*v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case *int8:
		return strconv.FormatInt(int64(*v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case *int16:
		return strconv.FormatInt(int64(*v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case *int32:
		return strconv.FormatInt(int64(*v), 10)
	case int64:
		return strconv.FormatInt(int64(v), 10)
	case *int64:
		return strconv.FormatInt(int64(*v), 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case *float32:
		return strconv.FormatFloat(float64(*v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case *float64:
		return strconv.FormatFloat(float64(*v), 'f', -1, 64)
	case time.Time:
		return strconv.FormatInt(v.Unix(), 10)
	case *time.Time:
		return strconv.FormatInt(v.Unix(), 10)
	case struct{}, *struct{}:
		return jsoniter.Stringify(v)
	case json.Number:
		return v.String()
	case *json.Number:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

type item struct {
	prefixKey string
	value     map[string]any
}

func FlattenMap(recordMap map[string]any) map[string]any {
	result := map[string]any{}
	var stack Stack
	stack.Push(&item{"", recordMap})

	for i := stack.Pop(); i != nil; i = stack.Pop() {
		current := i.(*item)
		for key, value := range current.value {
			flattenKey := key
			if current.prefixKey != "" {
				flattenKey = current.prefixKey + "_" + key
			}
			switch v := value.(type) {
			case map[string]any:
				stack.Push(&item{flattenKey, v})
			default:
				result[flattenKey] = value
			}
		}
	}

	return result
}

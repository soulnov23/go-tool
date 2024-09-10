package utils

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/soulnov23/go-tool/pkg/json/pbjson"
)

func Stringify(value any) string {
	result, err := pbjson.Marshal(value)
	if err != nil {
		return ""
	}
	return BytesToString(result)
}

func StringToMap(data string, fieldSep string, valueSep string) map[string]string {
	result := map[string]string{}
	fieldSlice := strings.Split(data, fieldSep)
	for _, kv := range fieldSlice {
		valueSlice := strings.Split(kv, valueSep)
		if len(valueSlice) == 2 {
			result[valueSlice[0]] = valueSlice[1]
		} else if len(valueSlice) == 1 && strings.Count(kv, valueSep) == 1 {
			result[valueSlice[0]] = ""
		}
	}
	return result
}

func MapToString(record map[string]string, sorted bool) string {
	size := len(record)
	if size == 0 {
		return ""
	}

	var builder strings.Builder
	if sorted {
		keys := make([]string, 0, size)
		for key := range record {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			builder.WriteString(key + "=" + record[key] + "&")
		}
	} else {
		for key, value := range record {
			builder.WriteString(key + "=" + value + "&")
		}
	}
	return builder.String()[0 : builder.Len()-1]
}

func AnyToString(value any) string {
	switch v := value.(type) {
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
		return strconv.FormatInt(v.UnixMilli(), 10)
	case *time.Time:
		return strconv.FormatInt(v.UnixMilli(), 10)
	case struct{}, *struct{}:
		return Stringify(v)
	case json.Number:
		return v.String()
	case *json.Number:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

package utils

import (
	"sort"
	"strings"
)

func StringToMap(data string, fieldSep string, valueSep string) map[string]string {
	if data == "" {
		return map[string]string{}
	}
	fieldSlice := strings.Split(data, fieldSep)
	result := make(map[string]string, len(fieldSlice))
	for _, kv := range fieldSlice {
		if kv == "" {
			continue
		}
		key, value, found := strings.Cut(kv, valueSep)
		if found {
			result[key] = value
		}
	}
	return result
}

func MapToString(data map[string]string, fieldSep string, valueSep string, sorted bool) string {
	size := len(data)
	if size == 0 {
		return ""
	}
	fields := make([]string, 0, size)
	if sorted {
		keys := make([]string, 0, size)
		for key := range data {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fields = append(fields, key+valueSep+data[key])
		}
	} else {
		for key, value := range data {
			fields = append(fields, key+valueSep+value)
		}
	}
	return strings.Join(fields, fieldSep)
}

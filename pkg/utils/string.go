package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/soulnov23/go-tool/pkg/json/pbjson"
)

var (
	LowerCaseLettersCharset = []rune("abcdefghijklmnopqrstuvwxyz")
	UpperCaseLettersCharset = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	LettersCharset          = append(LowerCaseLettersCharset, UpperCaseLettersCharset...)
	NumbersCharset          = []rune("0123456789")
	AlphanumericCharset     = append(LettersCharset, NumbersCharset...)
	SpecialCharset          = []rune("!@#$%^&*()_+-=[]{}|;':\",./<>?")
	AllCharset              = append(AlphanumericCharset, SpecialCharset...)

	// bearer:disable go_lang_permissive_regex_validation
	splitWordReg = regexp.MustCompile(`([a-z])([A-Z0-9])|([a-zA-Z])([0-9])|([0-9])([a-zA-Z])|([A-Z])([A-Z])([a-z])`)
	// bearer:disable go_lang_permissive_regex_validation
	splitNumberLetterReg = regexp.MustCompile(`([0-9])([a-zA-Z])`)
	maximumCapacity      = math.MaxInt>>1 + 1
)

// RandomString return a random string.
// Play: https://go.dev/play/p/rRseOQVVum4
func RandomString(size int, charset []rune) string {
	if size <= 0 {
		size = 16
	}
	if len(charset) <= 0 {
		charset = AllCharset
	}

	// see https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
	sb := strings.Builder{}
	sb.Grow(size)
	// Calculate the number of bits required to represent the charset,
	// e.g., for 62 characters, it would need 6 bits (since 62 -> 64 = 2^6)
	letterIdBits := int(math.Log2(float64(nearestPowerOfTwo(len(charset)))))
	// Determine the corresponding bitmask,
	// e.g., for 62 characters, the bitmask would be 111111.
	var letterIdMask int64 = 1<<letterIdBits - 1
	// Available count, since rand.Int64() returns a non-negative number, the first bit is fixed, so there are 63 random bits
	// e.g., for 62 characters, this value is 10 (63 / 6).
	letterIdMax := 63 / letterIdBits
	// Generate the random string in a loop.
	for i, cache, remain := size-1, rand.Int64(), letterIdMax; i >= 0; {
		// Regenerate the random number if all available bits have been used
		if remain == 0 {
			cache, remain = rand.Int64(), letterIdMax
		}
		// Select a character from the charset
		if idx := int(cache & letterIdMask); idx < len(charset) {
			sb.WriteRune(charset[idx])
			i--
		}
		// Shift the bits to the right to prepare for the next character selection,
		// e.g., for 62 characters, shift by 6 bits.
		cache >>= letterIdBits
		// Decrease the remaining number of uses for the current random number.
		remain--
	}
	return sb.String()
}

// nearestPowerOfTwo returns the nearest power of two.
func nearestPowerOfTwo(cap int) int {
	n := cap - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	if n < 0 {
		return 1
	}
	if n >= maximumCapacity {
		return maximumCapacity
	}
	return n + 1
}

// Words splits string into an array of its words.
func Words(str string) []string {
	str = splitWordReg.ReplaceAllString(str, `$1$3$5$7 $2$4$6$8$9`)
	// example: Int8Value => Int 8Value => Int 8 Value
	str = splitNumberLetterReg.ReplaceAllString(str, "$1 $2")
	var result strings.Builder
	for _, r := range str {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else {
			result.WriteRune(' ')
		}
	}
	return strings.Fields(result.String())
}

func Bytesify(value any) []byte {
	result, _ := pbjson.Marshal(value)
	return result
}

func Stringify(value any) string {
	result, _ := pbjson.Marshal(value)
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
	case *any:
		return AnyToString(*v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

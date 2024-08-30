//go:build go1.20

package utils

import (
	"unsafe"
)

// https://github.com/golang/go/issues/53003
func BytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

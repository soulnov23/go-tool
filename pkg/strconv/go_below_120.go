//go:build !go1.20

package strconv

import (
	"reflect"
	"unsafe"
)

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func StringToBytes(s string) (b []byte) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}

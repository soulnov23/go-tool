package ring

import (
	"testing"
	"unsafe"
)

func TestAnySize(t *testing.T) {
	var value any
	// any相当于两个指针大小，64为操作系统，指针大小是8，这里就是16
	t.Log(unsafe.Sizeof(value))
}

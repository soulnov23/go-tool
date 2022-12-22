package buffer

import (
	"testing"

	"github.com/SoulNov23/go-tool/pkg/unsafe"
)

func TestBuffer(t *testing.T) {
	buf := unsafe.String2Byte("hello world")
	lkBuffer := NewBuffer()
	lkBuffer.Write(buf)
	t.Logf("len: %d", lkBuffer.Len())
	lkBuffer.Write(buf)
	t.Logf("len: %d", lkBuffer.Len())
	lkBuffer.Write(buf)
	t.Logf("len: %d", lkBuffer.Len())

	res, err := lkBuffer.Peek(10)
	if err != nil {
		t.Logf("linkedBuffer.Peek: %v", err)
	} else {
		t.Logf("buf: %s", unsafe.Byte2String(res))
	}
	res, err = lkBuffer.Peek(20)
	if err != nil {
		t.Logf("linkedBuffer.Peek: %v", err)
	} else {
		t.Logf("buf: %s", unsafe.Byte2String(res))
	}
	res, err = lkBuffer.Peek(30)
	if err != nil {
		t.Logf("linkedBuffer.Peek: %v", err)
	} else {
		t.Logf("buf: %s", unsafe.Byte2String(res))
	}
	res, err = lkBuffer.Peek(40)
	if err != nil {
		t.Logf("linkedBuffer.Peek: %v", err)
	} else {
		t.Logf("buf: %s", unsafe.Byte2String(res))
	}

	t.Logf("len: %d", lkBuffer.Len())
	res, err = lkBuffer.Peek(10)
	if err != nil {
		t.Logf("linkedBuffer.Peek: %v", err)
	} else {
		lkBuffer.Skip(10)
		t.Logf("buf: %s", unsafe.Byte2String(res))
	}
	t.Logf("len: %d", lkBuffer.Len())
	res, err = lkBuffer.Peek(20)
	if err != nil {
		t.Logf("linkedBuffer.Peek: %v", err)
	} else {
		lkBuffer.Skip(20)
		t.Logf("buf: %s", unsafe.Byte2String(res))
	}
	t.Logf("len: %d", lkBuffer.Len())
	res, err = lkBuffer.Peek(30)
	if err != nil {
		t.Logf("linkedBuffer.Peek: %v", err)
	} else {
		lkBuffer.Skip(30)
		t.Logf("buf: %s", unsafe.Byte2String(res))
	}
	t.Logf("len: %d", lkBuffer.Len())
	res, err = lkBuffer.Peek(40)
	if err != nil {
		t.Logf("linkedBuffer.Peek: %v", err)
	} else {
		lkBuffer.Skip(40)
		t.Logf("buf: %s", unsafe.Byte2String(res))
	}
	lkBuffer.OptimizeMemory()

	DeleteBuffer(lkBuffer)
}

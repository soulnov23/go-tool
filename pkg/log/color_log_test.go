package log

import "testing"

func TestTinyLog(t *testing.T) {
	ColorDebugf("hello world")
	ColorInfof("hello world")
	ColorWarnf("hello world")
	ColorErrorf("hello world")
	ColorFatalf("hello world")

	ColorDebugf("hello %s", "world")
	ColorInfof("hello %s", "world")
	ColorWarnf("hello %s", "world")
	ColorErrorf("hello %s", "world")
	ColorFatalf("hello %s", "world")
}

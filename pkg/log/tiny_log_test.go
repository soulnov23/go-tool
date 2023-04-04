package log

import "testing"

func TestTinyLog(t *testing.T) {
	Debug("hello %s", "world")
	Info("hello %s", "world")
	Warn("hello %s", "world")
	Error("hello %s", "world")
	Fatal("hello %s", "world")
}

package log

import "testing"

func Test(t *testing.T) {
	Debug("hello %s", "world")
	Info("hello %s", "world")
	Warn("hello %s", "world")
	Error("hello %s", "world")
	Fatal("hello %s", "world")
}

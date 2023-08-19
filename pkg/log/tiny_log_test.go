package log

import "testing"

func TestTinyLog(t *testing.T) {
	Debugf("hello world")
	Infof("hello world")
	Warnf("hello world")
	Errorf("hello world")
	Fatalf("hello world")

	Debugf("hello %s", "world")
	Infof("hello %s", "world")
	Warnf("hello %s", "world")
	Errorf("hello %s", "world")
	Fatalf("hello %s", "world")
}

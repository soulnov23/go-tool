package utils

import (
	"testing"
)

func TestStack(t *testing.T) {
	var s Stack
	s.Push("1")
	s.Push("2")
	s.Push("3")
	t.Log(s.Pop())
	t.Log(s.Pop())
	t.Log(s.Pop())
}

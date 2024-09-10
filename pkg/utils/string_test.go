package utils

import (
	"testing"
)

func TestString(t *testing.T) {
	BytesToString([]byte{})
	StringToBytes("")
}

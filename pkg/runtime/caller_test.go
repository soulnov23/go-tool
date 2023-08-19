package runtime

import (
	"fmt"
	"testing"
)

func TestGetCaller(t *testing.T) {
	print(0, "hello world")
	print(1, "hello world")
	print(2, "hello world")

	print(0, "hello %s", "world")
	print(1, "hello %s", "world")
	print(2, "hello %s", "world")
}

func print(skip int, formatter string, args ...any) {
	fmt.Printf("%s %s\n", Caller(skip), fmt.Sprintf(formatter, args...))
}

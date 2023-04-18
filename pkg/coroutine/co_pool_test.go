package coroutine

import (
	"runtime"
	"testing"
)

func TestCoPool(t *testing.T) {
	pool := New(t.Logf, 1)
	for i := 0; i < 60; i++ {
		fn := func(args ...any) {
			t.Logf("key: %s, value: %d, runtime.NumGoroutine: %d", args[0], args[1], runtime.NumGoroutine())
		}
		pool.Run(fn, "index", i)
	}
	pool.Close()
}

package coroutine

import (
	"runtime"
	"testing"
)

func TestCoPool(t *testing.T) {
	pool := NewPool(t.Logf, 10)
	fn := func() {
		t.Logf("runtime.NumGoroutine: %d", runtime.NumGoroutine())
	}
	for {
		pool.Run(fn)
	}
}

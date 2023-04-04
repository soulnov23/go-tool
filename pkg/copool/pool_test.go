package copool

import (
	"runtime"
	"testing"
)

func TestCoPool(t *testing.T) {
	pool := NewPool(t.Logf, 10)
	handler := func() {
		t.Logf("runtime.NumGoroutine: %d", runtime.NumGoroutine())
	}
	for {
		pool.Run(handler)
	}
}

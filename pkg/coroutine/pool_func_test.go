package coroutine

import (
	"runtime"
	"testing"
)

func TestPoolFunc(t *testing.T) {
	pool, _ := NewPoolFunc(10, 10000, t.Errorf, func(arg any) {
		t.Logf("index: %d", arg)
		t.Logf("runtime.NumGoroutine: %d", runtime.NumGoroutine())
	})
	for i := 0; i < 60; i++ {
		pool.Invoke(i)
	}
	pool.Wait()
}

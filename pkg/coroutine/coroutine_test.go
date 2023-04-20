package coroutine

import (
	"runtime"
	"runtime/debug"
	"testing"

	"github.com/soulnov23/go-tool/pkg/utils"
)

func TestStack(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			buffer := make([]byte, 10*1024)
			runtime.Stack(buffer, false)
			t.Logf("[PANIC] %v\n%s", err, utils.Byte2String(buffer))
			t.Logf("[PANIC] %v\n%s", err, utils.Byte2String(debug.Stack()))
		}
	}()
	panic("hello world")
}

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

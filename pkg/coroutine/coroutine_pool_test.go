package coroutine

import (
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"github.com/soulnov23/go-tool/pkg/utils"
)

func TestStack(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			buffer := make([]byte, 10*1024)
			runtime.Stack(buffer, false)
			t.Logf("[PANIC] %v\n%s", err, utils.BytesToString(buffer))
			t.Logf("[PANIC] %v\n%s", err, utils.BytesToString(debug.Stack()))
		}
	}()
	panic("hello world")
}

func TestPool(t *testing.T) {
	pool := NewPool(32, t.Errorf)
	for i := range 10000 {
		fn := func(args ...any) {
			t.Logf("%s: %d", time.Now().Format(time.DateTime+".000"), runtime.NumGoroutine())
		}
		pool.Go(fn, i)
	}
	pool.Wait()
	t.Logf("wait 1")
	for i := range 10000 {
		fn := func(args ...any) {
			t.Logf("%s: %d", time.Now().Format(time.DateTime+".000"), runtime.NumGoroutine())
		}
		pool.Go(fn, i)
	}
	pool.Wait()
	t.Logf("wait 2")
	pool.Close()
}

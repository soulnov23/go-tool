package copool

import (
	"runtime"
	"testing"

	"github.com/SoulNov23/go-tool/pkg/log"
)

func TestCoPool(t *testing.T) {
	config := &log.LogConfig{
		Writer: "console",
		Level:  "debug",
	}
	zapLog, err := log.NewZapLog(config)
	if err != nil {
		t.Logf("log.NewZapLog: %s", err.Error())
	}
	pool := NewPool(zapLog, 10)
	handler := func() {
		t.Logf("runtime.NumGoroutine: %d", runtime.NumGoroutine())
	}
	for {
		pool.Run(handler)
	}
}

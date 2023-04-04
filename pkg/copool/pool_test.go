package copool

import (
	"runtime"
	"testing"

	"github.com/soulnov23/go-tool/pkg/log"
)

func TestCoPool(t *testing.T) {
	zapLog, err := log.NewZapLog(log.DefaultConsoleLogConfig)
	if err != nil {
		t.Logf("log.NewZapLog: %s", err.Error())
	}
	pool := NewPool(zapLog.Errorf, 10)
	handler := func() {
		t.Logf("runtime.NumGoroutine: %d", runtime.NumGoroutine())
	}
	for {
		pool.Run(handler)
	}
}

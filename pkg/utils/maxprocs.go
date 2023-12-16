package utils

import (
	"time"

	"go.uber.org/automaxprocs/maxprocs"
)

func UpdateGOMAXPROCS(printf func(formatter string, args ...any), interval time.Duration) func() {
	_, _ = maxprocs.Set(maxprocs.Logger(printf))
	done := make(chan struct{})
	go func() {
		tick := time.NewTicker(interval)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				_, _ = maxprocs.Set(maxprocs.Logger(printf))
			case <-done:
				return
			}
		}
	}()
	return func() { close(done) }
}

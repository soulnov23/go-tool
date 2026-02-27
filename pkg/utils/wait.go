package utils

import (
	"context"
	"time"
)

func Wait(ctx context.Context, timeout time.Duration, interval time.Duration, fn func() bool) {
	ctx, timeoutCancel := context.WithTimeout(ctx, timeout)
	defer timeoutCancel()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if fn() {
				return
			}
		}
	}
}

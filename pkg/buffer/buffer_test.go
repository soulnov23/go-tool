package buffer

import (
	"context"
	"testing"
	"time"

	"github.com/soulnov23/go-tool/pkg/utils"
)

func TestBuffer(t *testing.T) {
	lkBuffer := New()

	write := make(chan struct{})
	read := make(chan struct{})

	timeout := 3 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	buf := utils.String2Byte("hello world")

	go func(ctx context.Context, lkBuffer *Buffer) {
		for {
			select {
			case <-ctx.Done():
				write <- struct{}{}
				return
			default:
				lkBuffer.Write(buf)
			}
		}
	}(ctx, lkBuffer)

	go func(ctx context.Context, lkBuffer *Buffer) {
		for {
			select {
			case <-ctx.Done():
				read <- struct{}{}
				return
			default:
				res, err := lkBuffer.Peek(40)
				if err != nil {
					t.Logf("Buffer.Peek: %v", err)
				} else {
					lkBuffer.Skip(cap(res))
					t.Logf("buf: %s", utils.Byte2String(res))
				}
				lkBuffer.GC()
			}
		}
	}(ctx, lkBuffer)

	go func(ctx context.Context, lkBuffer *Buffer) {
		for {
			select {
			case <-ctx.Done():
				read <- struct{}{}
				return
			default:
				res, err := lkBuffer.Read(40)
				if err != nil {
					t.Logf("Buffer.Read: %v", err)
				} else {
					t.Logf("buf: %s", utils.Byte2String(res))
				}
				lkBuffer.GC()
			}
		}
	}(ctx, lkBuffer)

	time.Sleep(timeout)

	lkBuffer.Close()

	cancel()

	<-write
	<-read
}

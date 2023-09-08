package ringbuffer

import (
	"context"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/soulnov23/go-tool/pkg/log"
	"go.uber.org/zap"
)

func TestAnySize(t *testing.T) {
	var value any
	// any相当于两个指针大小，64为操作系统，指针大小是8，这里就是16
	t.Log(unsafe.Sizeof(value))
}

func TestRoundUpToPower2(t *testing.T) {
	t.Log(roundUpToPower2(0))
	t.Log(roundUpToPower2(1))
	t.Log(roundUpToPower2(2))
	t.Log(roundUpToPower2(3))
	t.Log(roundUpToPower2(4))
	t.Log(roundUpToPower2(7))
	t.Log(roundUpToPower2(15))
	t.Log(roundUpToPower2(21))
	t.Log(roundUpToPower2(33))
}

func TestRingBuffer(t *testing.T) {
	queue := New(4)

	timeout := 10 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	enWait := &sync.WaitGroup{}
	for i := 0; i < 8; i++ {
		enWait.Add(1)
		go func(ctx context.Context, queue *RingBuffer) {
			defer enWait.Done()
			for {
				select {
				case <-ctx.Done():
					log.DebugFields("ctx done")
					return
				default:
					if queue.Enqueue("ringbuffer") != nil {
						log.DebugFields("full", zap.Uint64("size", queue.Size()))
					}
					//log.DebugFields("Enqueue", zap.Uint64("size", queue.Size()))
				}
			}
		}(ctx, queue)
	}

	deWait := &sync.WaitGroup{}
	for i := 0; i < 8; i++ {
		deWait.Add(1)
		go func(ctx context.Context, queue *RingBuffer) {
			defer deWait.Done()
			for {
				select {
				case <-ctx.Done():
					log.DebugFields("ctx done")
					return
				default:
					if queue.Dequeue() == nil {
						log.DebugFields("empty", zap.Uint64("size", queue.Size()))
					}
					//log.DebugFields("Dequeue", zap.Uint64("size", queue.Size()))
				}
			}
		}(ctx, queue)
	}

	time.Sleep(timeout)

	cancel()
	enWait.Wait()
	deWait.Wait()
}

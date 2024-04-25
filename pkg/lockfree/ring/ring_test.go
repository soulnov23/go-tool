package ringbuffer

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"go.uber.org/zap"

	"github.com/soulnov23/go-tool/pkg/log"
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
	queue := New[string](512)

	timeout := 10 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	enWait := &sync.WaitGroup{}
	var enCount uint64
	for i := 0; i < 8; i++ {
		enWait.Add(1)
		go func(ctx context.Context, queue *Ring[string]) {
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
					log.DebugFields("Enqueue", zap.Uint64("size", queue.Size()))
					atomic.AddUint64(&enCount, uint64(1))
				}
			}
		}(ctx, queue)
	}

	deWait := &sync.WaitGroup{}
	var deCount uint64
	for i := 0; i < 8; i++ {
		deWait.Add(1)
		go func(ctx context.Context, queue *Ring[string]) {
			defer deWait.Done()
			for {
				select {
				case <-ctx.Done():
					log.DebugFields("ctx done")
					return
				default:
					if queue.Dequeue() == "" {
						log.DebugFields("empty", zap.Uint64("size", queue.Size()))
					}
					log.DebugFields("Dequeue", zap.Uint64("size", queue.Size()))
					atomic.AddUint64(&deCount, uint64(1))
				}
			}
		}(ctx, queue)
	}

	time.Sleep(timeout)

	cancel()
	enWait.Wait()
	deWait.Wait()

	log.DebugFields("", zap.Uint64("enCount", enCount), zap.Uint64("deCount", deCount))
}

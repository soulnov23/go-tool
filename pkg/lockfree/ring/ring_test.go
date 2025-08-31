package ring

import (
	"context"
	"sync"
	"sync/atomic"
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
	glog, err := log.GetDefaultLogger()
	if err != nil {
		t.Logf("log.GetDefaultLogger: %v", err)
		return
	}

	queue := New(512)

	timeout := 10 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	var enWait sync.WaitGroup
	var enCount atomic.Uint64
	for i := 0; i < 8; i++ {
		enWait.Add(1)
		go func(ctx context.Context, queue *Queue) {
			defer enWait.Done()
			for {
				select {
				case <-ctx.Done():
					glog.DebugFields("ctx done")
					return
				default:
					if queue.Enqueue("ringbuffer") == ErrQueueFull {
						glog.DebugFields("full", zap.Uint64("size", queue.Size()))
					}
					glog.DebugFields("Enqueue", zap.Uint64("size", queue.Size()))
					enCount.Add(1)
				}
			}
		}(ctx, queue)
	}

	var deWait sync.WaitGroup
	var deCount atomic.Uint64
	for i := 0; i < 8; i++ {
		deWait.Add(1)
		go func(ctx context.Context, queue *Queue) {
			defer deWait.Done()
			for {
				select {
				case <-ctx.Done():
					glog.DebugFields("ctx done")
					return
				default:
					_, err := queue.Dequeue()
					if err == ErrQueueEmpty {
						glog.DebugFields("empty", zap.Uint64("size", queue.Size()))
					}
					glog.DebugFields("Dequeue", zap.Uint64("size", queue.Size()))
					deCount.Add(1)
				}
			}
		}(ctx, queue)
	}

	time.Sleep(timeout)

	cancel()
	enWait.Wait()
	deWait.Wait()

	glog.DebugFields("", zap.Uint64("enCount", enCount.Load()), zap.Uint64("deCount", deCount.Load()))
}

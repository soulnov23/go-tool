package ringbuffer

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/soulnov23/go-tool/pkg/log"
	convert "github.com/soulnov23/go-tool/pkg/strconv"
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
	jsonLog, err := log.New(log.JsonConfig)
	if err != nil {
		t.Logf("new log: %s", err.Error())
		return
	}

	queue := New(4)

	timeout := 20 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	enWait := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		enWait.Add(1)
		go func(ctx context.Context, queue *RingBuffer) {
			defer enWait.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					queue.Enqueue("123")
					queue.Enqueue("456")
					queue.Enqueue("789")
				}
			}
		}(ctx, queue)
	}

	deWait := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		deWait.Add(1)
		go func(ctx context.Context, queue *RingBuffer) {
			defer deWait.Done()
			for {
				select {
				case <-ctx.Done():
					break
				default:
					value := queue.Dequeue()
					t.Log(value)
					var data []string
					for _, node := range queue.nodes {
						data = append(data, convert.AnyToString(node.value))
					}
					jsonLog.DebugFields("debug", zap.Uint64("capacity", queue.capacity), zap.Uint64("size", queue.Size()),
						zap.Uint64("head", atomic.LoadUint64(&queue.head)), zap.Uint64("tail", atomic.LoadUint64(&queue.tail)),
						zap.Reflect("nodes", data))
				}
			}
		}(ctx, queue)
	}

	time.Sleep(timeout)

	cancel()
	enWait.Wait()
	deWait.Wait()
}

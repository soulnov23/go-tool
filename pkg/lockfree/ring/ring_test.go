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

	queue := New(2) // 最小容量，提高并发碰撞概率

	timeout := 30 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	var enWait sync.WaitGroup
	var enSuccess atomic.Uint64 // 入队成功次数
	var enSeq atomic.Uint64     // 全局递增序列号，确保每个入队值唯一
	for i := 0; i < 32; i++ {
		enWait.Add(1)
		go func(ctx context.Context, queue *Queue) {
			defer enWait.Done()
			for {
				select {
				case <-ctx.Done():
					glog.DebugFields("ctx done")
					return
				default:
					seq := enSeq.Add(1)
					if queue.Enqueue(seq) == ErrQueueFull {
						glog.DebugFields("full", zap.Uint64("size", queue.Size()))
					} else {
						enSuccess.Add(1)
					}
				}
			}
		}(ctx, queue)
	}

	var deWait sync.WaitGroup
	var deSuccess atomic.Uint64 // 出队成功次数
	var seen sync.Map           // 检测重复值
	for i := 0; i < 32; i++ {
		deWait.Add(1)
		go func(ctx context.Context, queue *Queue) {
			defer deWait.Done()
			for {
				select {
				case <-ctx.Done():
					glog.DebugFields("ctx done")
					return
				default:
					value, err := queue.Dequeue()
					if err == ErrQueueEmpty {
						glog.DebugFields("empty", zap.Uint64("size", queue.Size()))
					} else {
						if _, loaded := seen.LoadOrStore(value, true); loaded {
							t.Errorf("duplicate value: %v", value)
						}
						deSuccess.Add(1)
					}
				}
			}
		}(ctx, queue)
	}

	time.Sleep(timeout)

	cancel()
	enWait.Wait()
	deWait.Wait()

	// 正确性校验：入队成功数 == 出队成功数 + 队列剩余
	en := enSuccess.Load()
	de := deSuccess.Load()
	remain := queue.Size()
	glog.DebugFields("", zap.Uint64("enSuccess", en), zap.Uint64("deSuccess", de), zap.Uint64("remain", remain))
	if en != de+remain {
		t.Fatalf("data inconsistency: enSuccess(%d) != deSuccess(%d) + remain(%d)", en, de, remain)
	}
}

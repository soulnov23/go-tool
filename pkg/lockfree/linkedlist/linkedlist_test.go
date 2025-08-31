package linkedlist

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/soulnov23/go-tool/pkg/log"
	"go.uber.org/zap"
)

func TestQueue(t *testing.T) {
	glog, err := log.GetDefaultLogger()
	if err != nil {
		t.Logf("log.GetDefaultLogger: %v", err)
		return
	}

	queue := New()

	timeout := 10 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	enWait := &sync.WaitGroup{}
	var enCount uint64
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
					queue.Enqueue("linkedlist")
					glog.DebugFields("Enqueue", zap.Uint64("size", queue.Size()))
					atomic.AddUint64(&enCount, uint64(1))
				}
			}
		}(ctx, queue)
	}

	deWait := &sync.WaitGroup{}
	var deCount uint64
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
					if queue.Dequeue() == nil {
						glog.DebugFields("empty", zap.Uint64("size", queue.Size()))
					}
					glog.DebugFields("Dequeue", zap.Uint64("size", queue.Size()))
					atomic.AddUint64(&deCount, uint64(1))
				}
			}
		}(ctx, queue)
	}

	time.Sleep(timeout)

	cancel()
	enWait.Wait()
	deWait.Wait()

	glog.DebugFields("", zap.Uint64("enCount", enCount), zap.Uint64("deCount", deCount))
}

func TestAddUint64(t *testing.T) {
	glog, err := log.GetDefaultLogger()
	if err != nil {
		t.Logf("log.GetDefaultLogger: %v", err)
		return
	}

	var value uint64
	glog.DebugFields("value", zap.Uint64("value", value))
	atomic.AddUint64(&value, ^uint64(0))
	glog.DebugFields("value", zap.Uint64("value", value))
	atomic.AddUint64(&value, uint64(1))
	glog.DebugFields("value", zap.Uint64("value", value))

	atomicValue := &atomic.Uint64{}
	glog.DebugFields("atomicValue", zap.Uint64("atomicValue", atomicValue.Load()))
	atomicValue.Add(^uint64(0))
	glog.DebugFields("atomicValue", zap.Uint64("atomicValue", atomicValue.Load()))
	atomicValue.Add(uint64(1))
	glog.DebugFields("atomicValue", zap.Uint64("atomicValue", atomicValue.Load()))
}

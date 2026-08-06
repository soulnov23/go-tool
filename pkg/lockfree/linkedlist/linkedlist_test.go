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
					log.DefaultLogger.DebugFields("ctx done")
					return
				default:
					queue.Enqueue("linkedlist")
					log.DefaultLogger.DebugFields("Enqueue", zap.Uint64("size", queue.Size()))
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
					log.DefaultLogger.DebugFields("ctx done")
					return
				default:
					if queue.Dequeue() == nil {
						log.DefaultLogger.DebugFields("empty", zap.Uint64("size", queue.Size()))
					}
					log.DefaultLogger.DebugFields("Dequeue", zap.Uint64("size", queue.Size()))
					atomic.AddUint64(&deCount, uint64(1))
				}
			}
		}(ctx, queue)
	}

	time.Sleep(timeout)

	cancel()
	enWait.Wait()
	deWait.Wait()

	log.DefaultLogger.DebugFields("", zap.Uint64("enCount", enCount), zap.Uint64("deCount", deCount))
}

func TestAddUint64(t *testing.T) {
	var value uint64
	log.DefaultLogger.DebugFields("value", zap.Uint64("value", value))
	atomic.AddUint64(&value, ^uint64(0))
	log.DefaultLogger.DebugFields("value", zap.Uint64("value", value))
	atomic.AddUint64(&value, uint64(1))
	log.DefaultLogger.DebugFields("value", zap.Uint64("value", value))

	atomicValue := &atomic.Uint64{}
	log.DefaultLogger.DebugFields("atomicValue", zap.Uint64("atomicValue", atomicValue.Load()))
	atomicValue.Add(^uint64(0))
	log.DefaultLogger.DebugFields("atomicValue", zap.Uint64("atomicValue", atomicValue.Load()))
	atomicValue.Add(uint64(1))
	log.DefaultLogger.DebugFields("atomicValue", zap.Uint64("atomicValue", atomicValue.Load()))
}

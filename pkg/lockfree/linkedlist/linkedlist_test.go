package linkedlist

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/soulnov23/go-tool/pkg/log"
)

func TestQueue(t *testing.T) {
	queue := New()

	timeout := 10 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	enWait := &sync.WaitGroup{}
	var enCount uint64
	for i := 0; i < 8; i++ {
		enWait.Add(1)
		go func(ctx context.Context, queue *LinkedList) {
			defer enWait.Done()
			for {
				select {
				case <-ctx.Done():
					log.DebugFields("ctx done")
					return
				default:
					queue.Enqueue("linkedlist")
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
		go func(ctx context.Context, queue *LinkedList) {
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

func TestAddUint64(t *testing.T) {
	var value uint64
	log.Debug(value)
	atomic.AddUint64(&value, ^uint64(0))
	log.Debug(value)
	atomic.AddUint64(&value, uint64(1))
	log.Debug(value)

	atomicValue := &atomic.Uint64{}
	log.Debug(atomicValue.Load())
	atomicValue.Add(^uint64(0))
	log.Debug(atomicValue.Load())
	atomicValue.Add(uint64(1))
	log.Debug(atomicValue.Load())
}

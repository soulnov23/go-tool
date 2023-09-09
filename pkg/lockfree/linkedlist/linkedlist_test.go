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
				}
			}
		}(ctx, queue)
	}

	deWait := &sync.WaitGroup{}
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
				}
			}
		}(ctx, queue)
	}

	time.Sleep(timeout)

	cancel()
	enWait.Wait()
	deWait.Wait()
}

func TestAddUint64(t *testing.T) {
	var value uint64
	log.Debug(value)
	atomic.AddUint64(&value, ^uint64(0))
	log.Debug(value)
	atomic.AddUint64(&value, uint64(1))
	log.Debug(value)
}

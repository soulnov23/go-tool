package lockfree

import (
	"context"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	queue := NewQueue()

	write := make(chan struct{})
	read := make(chan struct{})

	timeout := 3 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context, queue *Queue) {
		for {
			select {
			case <-ctx.Done():
				write <- struct{}{}
				return
			default:
				queue.EnQueue("hello world")
			}
		}
	}(ctx, queue)

	go func(ctx context.Context, queue *Queue) {
		for {
			select {
			case <-ctx.Done():
				read <- struct{}{}
				return
			default:
				temp := queue.DeQueue()
				if temp == nil {
					t.Log("empty")
				}
			}
		}
	}(ctx, queue)

	go func(ctx context.Context, queue *Queue) {
		for {
			select {
			case <-ctx.Done():
				read <- struct{}{}
				return
			default:
				temp := queue.DeQueue()
				if temp == nil {
					t.Log("empty")
				} else {
					t.Log(temp)
				}
			}
		}
	}(ctx, queue)

	go func(ctx context.Context, queue *Queue) {
		for {
			select {
			case <-ctx.Done():
				read <- struct{}{}
				return
			default:
				temp := queue.DeQueue()
				if temp == nil {
					t.Log("empty")
				} else {
					t.Log(temp)
				}
			}
		}
	}(ctx, queue)

	time.Sleep(timeout)

	cancel()

	<-write
	<-read
	<-read
}

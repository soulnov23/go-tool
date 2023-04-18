package lockfree

import (
	"context"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	queue := New()

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
				queue.PushBack("hello world")
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
				temp := queue.PopFront()
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
				temp := queue.PopFront()
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
				temp := queue.PopFront()
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

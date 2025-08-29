package coroutine

import (
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/soulnov23/go-tool/pkg/utils"
)

var tasks sync.Pool

func init() {
	tasks.New = func() any {
		return &task{}
	}
}

type task struct {
	fn   func(...any)
	args []any
}

type Pool struct {
	capacity uint64
	length   atomic.Uint64
	taskChan chan *task
	printf   func(formatter string, args ...any)
	wg       sync.WaitGroup
}

func NewPool(poolCapacity int, printf func(formatter string, args ...any)) *Pool {
	return &Pool{
		capacity: uint64(poolCapacity),
		taskChan: make(chan *task),
		printf:   printf,
	}
}

func (pool *Pool) Run(fn func(...any), args ...any) {
	for {
		length := pool.length.Load()
		if length >= pool.capacity {
			break
		}
		if pool.length.CompareAndSwap(length, length+1) {
			go func() {
				defer pool.length.Add(^uint64(0))
				for task := range pool.taskChan {
					func() {
						defer func() {
							if err := recover(); err != nil {
								pool.printf("[PANIC] %v\n%s\n", err, utils.BytesToString(debug.Stack()))
							}
							pool.wg.Done()
							task.fn = nil
							task.args = nil
							tasks.Put(task)
						}()
						task.fn(task.args...)
					}()
				}
			}()
			break
		}
	}
	task := tasks.Get().(*task)
	task.fn = fn
	task.args = args
	pool.wg.Add(1)
	pool.taskChan <- task
}

func (pool *Pool) Length() uint64 {
	return pool.length.Load()
}

func (pool *Pool) Capacity() uint64 {
	return pool.capacity
}

func (pool *Pool) Wait() {
	pool.wg.Wait()
}

func (pool *Pool) Close() {
	close(pool.taskChan)
}

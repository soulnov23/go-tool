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
	size     atomic.Uint64
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

func (pool *Pool) Go(fn func(...any), args ...any) {
	pool.spawnWorker()
	task := tasks.Get().(*task)
	task.fn = fn
	task.args = args
	pool.wg.Add(1)
	pool.taskChan <- task
}

func (pool *Pool) spawnWorker() {
	for {
		size := pool.size.Load()
		if size >= pool.capacity {
			break
		}
		if pool.size.CompareAndSwap(size, size+1) {
			go pool.worker()
			break
		}
	}
}

func (pool *Pool) worker() {
	defer pool.size.Add(^uint64(0))
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
}

func (pool *Pool) Size() uint64 {
	return pool.size.Load()
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

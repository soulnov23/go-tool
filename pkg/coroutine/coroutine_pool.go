package coroutine

import (
	"runtime/debug"
	"sync"

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
	taskChan chan *task
	printf   func(formatter string, args ...any)
	wg       sync.WaitGroup
}

func NewPool(poolCapacity int, printf func(formatter string, args ...any)) *Pool {
	pool := &Pool{
		taskChan: make(chan *task),
		printf:   printf,
	}
	for range poolCapacity {
		go pool.worker()
	}
	return pool
}

func (pool *Pool) Go(fn func(...any), args ...any) {
	task := tasks.Get().(*task)
	task.fn = fn
	task.args = args
	pool.wg.Add(1)
	pool.taskChan <- task
}

func (pool *Pool) worker() {
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

func (pool *Pool) Wait() {
	pool.wg.Wait()
}

func (pool *Pool) Close() {
	close(pool.taskChan)
}

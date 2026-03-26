package coroutine

import (
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/soulnov23/go-tool/pkg/lockfree/ring"
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
	queue  *ring.Queue
	printf func(formatter string, args ...any)
	wg     sync.WaitGroup
	closed atomic.Bool
}

func NewPool(poolCapacity int, printf func(formatter string, args ...any)) *Pool {
	// 队列容量设为 worker 数量的 4 倍，提供足够缓冲
	queueCapacity := max(uint64(poolCapacity)*4, 64)
	pool := &Pool{
		queue:  ring.New(queueCapacity),
		printf: printf,
	}
	for range poolCapacity {
		go pool.worker()
	}
	return pool
}

func (pool *Pool) Go(fn func(...any), args ...any) {
	if pool.closed.Load() {
		panic("send task to closed pool")
	}
	task := tasks.Get().(*task)
	task.fn = fn
	task.args = args
	pool.wg.Add(1)
	for pool.queue.Enqueue(task) != nil {
		if pool.closed.Load() {
			pool.wg.Done()
			task.fn = nil
			task.args = nil
			tasks.Put(task)
			panic("send task to closed pool")
		}
		runtime.Gosched()
	}
}

func (pool *Pool) worker() {
	for {
		value, err := pool.queue.Dequeue()
		if err != nil {
			// 队列为空，检查是否已关闭且所有任务已完成
			if pool.closed.Load() && pool.queue.IsEmpty() {
				return
			}
			runtime.Gosched()
			continue
		}
		task := value.(*task)
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
	pool.closed.Store(true)
}

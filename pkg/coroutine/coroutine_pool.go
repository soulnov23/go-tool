package coroutine

import (
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/soulnov23/go-tool/pkg/lockfree/ring"
	"github.com/soulnov23/go-tool/pkg/utils"
)

var (
	tasks   sync.Pool
	wgTasks sync.WaitGroup
	works   sync.Pool
	wgWorks sync.WaitGroup
)

func init() {
	tasks.New = func() any {
		return &task{
			referCount: 1,
		}
	}
	works.New = func() any {
		return &worker{
			referCount: 1,
		}
	}
}

type task struct {
	fn         func(...any)
	args       []any
	referCount int32
}

func newTask(fn func(...any), args ...any) *task {
	wgTasks.Add(1)
	task := tasks.Get().(*task)
	task.fn = fn
	task.args = append(task.args, args...)
	return task
}

func (task *task) delete() {
	if atomic.AddInt32(&task.referCount, -1) == 0 {
		wgTasks.Done()
		task.fn = nil
		task.args = nil
		tasks.Put(task)
	}
}

type worker struct {
	pool       *Pool
	referCount int32
}

func newWork(pool *Pool) *worker {
	worker := works.Get().(*worker)
	worker.pool = pool
	return worker
}

func (worker *worker) run() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				worker.pool.printf("[PANIC] %v\n%s\n", err, utils.BytesToString(debug.Stack()))
			}
			worker.pool.decWorker()
			worker.delete()
		}()
		for {
			value := worker.pool.taskQueue.Dequeue()
			if value == nil {
				return
			}
			task := value.(*task)
			defer task.delete()
			task.fn(task.args...)
		}
	}()
}

func (worker *worker) delete() {
	if atomic.AddInt32(&worker.referCount, -1) == 0 {
		worker.pool = nil
		works.Put(worker)
	}
}

type Pool struct {
	capacity   uint64
	workerSize *atomic.Uint64
	taskQueue  *ring.Queue
	printf     func(formatter string, args ...any)
}

func NewPool(poolCapacity int, taskCapacity int, printf func(formatter string, args ...any)) *Pool {
	return &Pool{
		capacity:   uint64(poolCapacity),
		workerSize: &atomic.Uint64{},
		taskQueue:  ring.New(uint64(taskCapacity)),
		printf:     printf,
	}
}

func (pool *Pool) Run(fn func(...any), args ...any) {
	task := newTask(fn, args...)
	for {
		if pool.taskQueue.Enqueue(task) == nil {
			break
		}
		// queue is full
		runtime.Gosched()
	}
	if pool.worker() == 0 || pool.worker() < pool.capacity {
		pool.incWorker()
		worker := newWork(pool)
		worker.run()
	}
}

func (pool *Pool) worker() uint64 {
	return pool.workerSize.Load()
}

func (pool *Pool) incWorker() {
	pool.workerSize.Add(1)
	wgWorks.Add(1)
}

func (pool *Pool) decWorker() {
	pool.workerSize.Add(^uint64(0))
	wgWorks.Done()
}

func (pool *Pool) Wait() {
	wgTasks.Wait()
	wgWorks.Wait()
}

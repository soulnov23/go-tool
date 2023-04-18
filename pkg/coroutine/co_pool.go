package coroutine

import (
	"sync/atomic"

	"github.com/soulnov23/go-tool/pkg/lockfree"
)

type Pool struct {
	printf    func(formatter string, args ...any)
	taskQueue *lockfree.Queue

	poolSize    uint32 // 协程池额定大小
	workerCount uint32 // 协程池实际大小
}

func New(printf func(formatter string, args ...any), poolSize int) *Pool {
	return &Pool{
		printf:      printf,
		taskQueue:   lockfree.New(),
		poolSize:    uint32(poolSize),
		workerCount: 0,
	}
}

func (pool *Pool) Run(fn func(...any), args ...any) {
	task := newTask(fn, args...)
	pool.taskQueue.PushBack(task)
	if pool.worker() == 0 || pool.worker() < atomic.LoadUint32(&pool.poolSize) {
		pool.incWorker()
		work := newWork(pool)
		work.run()
	}
}

func (pool *Pool) worker() uint32 {
	return atomic.LoadUint32(&pool.workerCount)
}

func (pool *Pool) incWorker() {
	atomic.AddUint32(&pool.workerCount, 1)
}

func (pool *Pool) decWorker() {
	atomic.AddUint32(&pool.workerCount, ^uint32(0))
}

func (pool *Pool) Close() {
	for {
		if pool.taskQueue.Len() == 0 && pool.worker() == 0 {
			break
		}
	}
}

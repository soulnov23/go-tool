package coroutine

import (
	"sync"
	"sync/atomic"
)

var works sync.Pool

func init() {
	works.New = func() any {
		return &Work{
			referCount: 1,
		}
	}
}

type Work struct {
	pool       *Pool
	referCount int32
}

func NewWork(pool *Pool) *Work {
	work := works.Get().(*Work)
	work.pool = pool
	return work
}

func (work *Work) Run() {
	Go(work.pool.printf, func() {
		for {
			value := work.pool.taskQueue.DeQueue()
			if value == nil {
				work.pool.decWorker()
				work.Close()
				return
			}
			task := value.(*Task)
			task.fn()
			task.Close()
		}
	})
}

func (work *Work) Close() {
	if atomic.AddInt32(&work.referCount, -1) == 0 {
		work.pool = nil
		works.Put(work)
	}
}

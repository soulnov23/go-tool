package coroutine

import (
	"sync"
	"sync/atomic"
)

var works sync.Pool

func init() {
	works.New = func() any {
		return &work{
			referCount: 1,
		}
	}
}

type work struct {
	pool       *Pool
	referCount int32
}

func newWork(pool *Pool) *work {
	work := works.Get().(*work)
	work.pool = pool
	return work
}

func (work *work) run() {
	Go(work.pool.printf, func() {
		for {
			value := work.pool.taskQueue.PopFront()
			if value == nil {
				work.pool.decWorker()
				work.close()
				return
			}
			task := value.(*task)
			task.fn(task.args...)
			task.close()
		}
	})
}

func (work *work) close() {
	if atomic.AddInt32(&work.referCount, -1) == 0 {
		work.pool = nil
		works.Put(work)
	}
}

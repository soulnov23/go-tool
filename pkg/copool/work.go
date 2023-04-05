package copool

import (
	"sync"
	"sync/atomic"

	"github.com/soulnov23/go-tool/pkg/co"
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

func DeleteWork(work *Work) {
	if work == nil {
		return
	}
	if atomic.AddInt32(&work.referCount, -1) == 0 {
		work.pool = nil
		works.Put(work)
	}
}

func (work *Work) Run() {
	co.Go(work.pool.printf, func() {
		for {
			value := work.pool.taskQueue.DeQueue()
			if value == nil {
				work.pool.decWorker()
				DeleteWork(work)
				return
			}
			task := value.(*Task)
			task.handler()
			DeleteTask(task)
		}
	})
}

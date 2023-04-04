package copool

import (
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/soulnov23/go-tool/pkg/utils"
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
	go func() {
		for {
			value := work.pool.taskQueue.DeQueue()
			if value == nil {
				work.pool.decWorker()
				DeleteWork(work)
				return
			}
			task := value.(*Task)
			func() {
				defer func() {
					if e := recover(); e != nil {
						buffer := make([]byte, 10*1024)
						runtime.Stack(buffer, false)
						work.pool.printf("[PANIC] %v\n%s", e, utils.Byte2String(buffer))
					}
				}()
				task.handler()
			}()
			DeleteTask(task)
		}
	}()
}

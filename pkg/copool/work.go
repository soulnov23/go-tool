package copool

import (
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/SoulNov23/go-tool/pkg/unsafe"
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
			var task *Task
			work.pool.lock.Lock()
			if work.pool.head != nil {
				task = work.pool.head
				work.pool.head = task.next
			}
			work.pool.lock.Unlock()
			if task == nil {
				work.pool.decWorker()
				DeleteWork(work)
				return
			}
			func() {
				defer func() {
					if e := recover(); e != nil {
						buffer := make([]byte, 10*1024)
						runtime.Stack(buffer, false)
						work.pool.log.Errorf("[PANIC]%v\n%s\n", e, unsafe.Byte2String(buffer))
					}
				}()
				task.handler()
			}()
			DeleteTask(task)
		}
	}()
}

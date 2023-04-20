package coroutine

import (
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/soulnov23/go-tool/pkg/utils"
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
	go func() {
		for {
			value := work.pool.taskQueue.PopFront()
			if value == nil {
				work.pool.decWorker()
				work.close()
				return
			}
			task := value.(*task)
			// 增加func()包住是为了防止task的执行panic导致协程池有协程退出，协程应该由协程池管理，退出条件是没有可执行任务了
			// 其次协程退出没有调用decWorker，会导致pool.Close()锁死
			func() {
				defer func() {
					if err := recover(); err != nil {
						work.pool.printf("[PANIC] %v\n%s", err, utils.Byte2String(debug.Stack()))
					}
				}()
				task.fn(task.args...)
			}()
			task.close()
		}
	}()
}

func (work *work) close() {
	if atomic.AddInt32(&work.referCount, -1) == 0 {
		work.pool = nil
		works.Put(work)
	}
}

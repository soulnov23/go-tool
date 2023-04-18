package coroutine

import (
	"sync"
	"sync/atomic"
)

var tasks sync.Pool

func init() {
	tasks.New = func() any {
		return &task{
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
	task := tasks.Get().(*task)
	task.fn = fn
	task.args = append(task.args, args...)
	return task
}

func (task *task) close() {
	if atomic.AddInt32(&task.referCount, -1) == 0 {
		task.fn = nil
		task.args = nil
		tasks.Put(task)
	}
}

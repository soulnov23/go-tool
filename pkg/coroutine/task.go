package coroutine

import (
	"sync"
	"sync/atomic"
)

var tasks sync.Pool

func init() {
	tasks.New = func() any {
		return &Task{
			referCount: 1,
		}
	}
}

type Task struct {
	fn         func()
	referCount int32
}

func NewTask(fn func()) *Task {
	task := tasks.Get().(*Task)
	task.fn = fn
	return task
}

func (task *Task) Close() {
	if atomic.AddInt32(&task.referCount, -1) == 0 {
		task.fn = nil
		tasks.Put(task)
	}
}

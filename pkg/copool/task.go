package copool

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
	handler    func()
	next       *Task
	referCount int32
}

func NewTask(handler func()) *Task {
	task := tasks.Get().(*Task)
	task.handler = handler
	return task
}

func DeleteTask(task *Task) {
	if task == nil {
		return
	}
	if atomic.AddInt32(&task.referCount, -1) == 0 {
		task.handler, task.next = nil, nil
		tasks.Put(task)
	}
}

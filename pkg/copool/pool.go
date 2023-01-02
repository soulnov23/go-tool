package copool

import (
	"sync"
	"sync/atomic"

	"github.com/SoulNov23/go-tool/pkg/log"
)

type Pool struct {
	log  log.Logger
	head *Task
	tail *Task
	lock sync.Mutex

	poolSize int32

	workerCount int32
}

func NewPool(log log.Logger, poolSize int) *Pool {
	return &Pool{
		log:         log,
		head:        nil,
		tail:        nil,
		poolSize:    int32(poolSize),
		workerCount: 0,
	}
}

func DeletePool(pool *Pool) {
	if pool == nil {
		return
	}
	atomic.StoreInt32(&pool.workerCount, 0)
	for task := pool.head; task != nil; {
		next := task.next
		DeleteTask(task)
		task = next
	}
	pool.head, pool.tail = nil, nil
}

func (pool *Pool) Run(handler func()) {
	task := NewTask(handler)
	pool.lock.Lock()
	if pool.head == nil {
		pool.head = task
		pool.tail = task
	} else {
		pool.tail.next = task
		pool.tail = task
	}
	pool.lock.Unlock()
	if pool.worker() == 0 || pool.worker() < atomic.LoadInt32(&pool.poolSize) {
		pool.incWorker()
		work := NewWork(pool)
		work.Run()
	}
}

func (pool *Pool) worker() int32 {
	return atomic.LoadInt32(&pool.workerCount)
}

func (pool *Pool) incWorker() {
	atomic.AddInt32(&pool.workerCount, 1)
}

func (pool *Pool) decWorker() {
	atomic.AddInt32(&pool.workerCount, -1)
}

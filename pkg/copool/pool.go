package copool

import (
	"sync/atomic"

	"github.com/soulnov23/go-tool/pkg/lockfree"
	"github.com/soulnov23/go-tool/pkg/log"
)

type Pool struct {
	log       log.Logger
	taskQueue *lockfree.Queue

	poolSize    uint32 // 协程池额定大小
	workerCount uint32 // 协程池实际大小
}

func NewPool(log log.Logger, poolSize int) *Pool {
	return &Pool{
		log:         log,
		taskQueue:   lockfree.NewQueue(),
		poolSize:    uint32(poolSize),
		workerCount: 0,
	}
}

func DeletePool(pool *Pool) {
	if pool == nil {
		return
	}
	for {
		value := pool.taskQueue.DeQueue()
		if value == nil {
			break
		}
		task := value.(*Task)
		DeleteTask(task)
	}
	atomic.StoreUint32(&pool.workerCount, 0)
}

func (pool *Pool) Run(handler func()) {
	task := NewTask(handler)
	pool.taskQueue.EnQueue(task)
	if pool.worker() == 0 || pool.worker() < atomic.LoadUint32(&pool.poolSize) {
		pool.incWorker()
		work := NewWork(pool)
		work.Run()
	}
}

func (pool *Pool) worker() uint32 {
	return atomic.LoadUint32(&pool.workerCount)
}

func (pool *Pool) incWorker() {
	atomic.AddUint32(&pool.workerCount, 1)
}

func (pool *Pool) decWorker() {
	atomic.AddUint32(&pool.workerCount, ^uint32(0))
}

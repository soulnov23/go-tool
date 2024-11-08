package coroutine

import (
	"runtime/debug"
	"sync"

	"github.com/panjf2000/ants/v2"
	"github.com/soulnov23/go-tool/pkg/utils"
)

type Pool struct {
	*ants.Pool
	printf func(formatter string, args ...any)
	wg     *sync.WaitGroup
}

func NewPool(poolCapacity int, taskCapacity int, printf func(formatter string, args ...any)) *Pool {
	pool, _ := ants.NewPool(poolCapacity, ants.WithMaxBlockingTasks(taskCapacity), ants.WithPanicHandler(func(err any) {
		printf("[PANIC] %v\n%s", err, utils.BytesToString(debug.Stack()))
	}))
	return &Pool{
		Pool:   pool,
		printf: printf,
		wg:     &sync.WaitGroup{},
	}
}

func (pool *Pool) Run(task func()) {
	for {
		err := pool.Submit(func() {
			pool.wg.Add(1)
			defer pool.wg.Done()
			task()
		})
		if err == nil {
			break
		}
		if err == ants.ErrPoolOverload {
			continue
		}
		pool.printf("ants.Pool.Submit failed: %v", err)
	}
}

func (pool *Pool) Wait() {
	pool.wg.Wait()
	pool.Release()
}

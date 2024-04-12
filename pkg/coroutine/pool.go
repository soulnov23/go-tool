package coroutine

import (
	"runtime/debug"
	"sync"

	"github.com/panjf2000/ants/v2"

	"github.com/soulnov23/go-tool/pkg/utils"
)

type Pool struct {
	*ants.Pool
	wg *sync.WaitGroup
}

func NewPool(poolSize int, taskSize int, printf func(formatter string, args ...any)) (*Pool, error) {
	pool, err := ants.NewPool(poolSize, ants.WithMaxBlockingTasks(taskSize), ants.WithPanicHandler(func(err any) {
		printf("[PANIC] %v\n%s", err, utils.BytesToString(debug.Stack()))
	}))
	if err != nil {
		return nil, err
	}
	return &Pool{
		Pool: pool,
		wg:   &sync.WaitGroup{},
	}, nil
}

func (pool *Pool) Submit(task func()) error {
	pool.wg.Add(1)
	return pool.Pool.Submit(func() {
		defer pool.wg.Done()
		task()
	})
}

func (pool *Pool) Wait() {
	pool.wg.Wait()
	pool.Pool.Release()
}

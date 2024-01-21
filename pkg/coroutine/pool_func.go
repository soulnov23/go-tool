package coroutine

import (
	"runtime/debug"
	"sync"

	"github.com/panjf2000/ants/v2"
	"github.com/soulnov23/go-tool/pkg/utils"
)

type PoolFunc struct {
	*ants.PoolWithFunc
	wg *sync.WaitGroup
}

func NewPoolFunc(poolSize int, taskSize int, printf func(formatter string, args ...any), task func(any)) (*PoolFunc, error) {
	pf := &PoolFunc{
		wg: &sync.WaitGroup{},
	}
	pool, err := ants.NewPoolWithFunc(poolSize, func(arg any) {
		defer pf.wg.Done()
		task(arg)
	}, ants.WithMaxBlockingTasks(taskSize), ants.WithPanicHandler(func(err any) {
		printf("[PANIC] %v\n%s", err, utils.BytesToString(debug.Stack()))
	}))
	if err != nil {
		return nil, err
	}
	pf.PoolWithFunc = pool
	return pf, nil
}

func (pool *PoolFunc) Invoke(arg any) error {
	pool.wg.Add(1)
	return pool.PoolWithFunc.Invoke(arg)
}

func (pool *PoolFunc) Wait() {
	pool.wg.Wait()
	pool.PoolWithFunc.Release()
}

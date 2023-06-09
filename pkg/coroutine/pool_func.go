package coroutine

import (
	"sync"

	"github.com/panjf2000/ants/v2"
)

type PoolFunc struct {
	*ants.PoolWithFunc
	wg *sync.WaitGroup
}

func NewPoolFunc(size int, task func(any)) (*PoolFunc, error) {
	pf := &PoolFunc{
		wg: &sync.WaitGroup{},
	}
	pool, err := ants.NewPoolWithFunc(size, func(arg any) {
		defer pf.wg.Done()
		task(arg)
	})
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

func (pool *PoolFunc) Release() {
	pool.wg.Wait()
	pool.PoolWithFunc.Release()
}

package coroutine

import (
	"sync"

	"github.com/panjf2000/ants/v2"
)

type Pool struct {
	*ants.Pool
	wg *sync.WaitGroup
}

func NewPool(size int) (*Pool, error) {
	pool, err := ants.NewPool(size)
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

func (pool *Pool) Release() {
	pool.wg.Wait()
	pool.Pool.Release()
}

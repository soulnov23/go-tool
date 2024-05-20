package coroutine

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/soulnov23/go-tool/pkg/utils"
)

/*
1. 多协程并发执行时，一个协程抛出panic会导致所有协程全部退出，一个协程的panic只能该协程调用recover捕获，所以每个协程都要执行defer func() { recover() }()
2. 如果我们有批量的任务需要执行，肯定通过并发调用来提高性能，同时我们不希望其中一个调用失败就导致所有的任务都退出，而是要继续执行完其它的任务，封装co.GoAndWait接口
*/
func GoAndWait(fns ...func() error) error {
	var (
		wg     sync.WaitGroup
		once   sync.Once
		fnsErr error
	)
	for _, fn := range fns {
		wg.Add(1)
		go func(fn func() error) {
			defer func() {
				if err := recover(); err != nil {
					strErr := fmt.Sprintf("[PANIC] %v\n%s\n", err, utils.BytesToString(debug.Stack()))
					once.Do(func() {
						fnsErr = errors.New(strErr)
					})
				}
				wg.Done()
			}()
			if err := fn(); err != nil {
				once.Do(func() {
					fnsErr = err
				})
			}
		}(fn)
	}
	wg.Wait()
	return fnsErr
}

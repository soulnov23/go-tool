package coroutine

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/soulnov23/go-tool/pkg/utils"
)

/*
1. 多协程并发执行时，一个协程抛出panic会导致所有协程全部退出，一个协程的panic只能该协程调用recover捕获，所以每个协程都要执行defer func() { recover() }()
2. 如果我们有批量的任务需要执行，肯定通过并发调用来提高性能，同时我们不希望其中一个调用失败就导致所有的任务都退出，而是要继续执行完其它的任务，封装co.GoAndWait接口
*/
func GoAndWait(fns ...func() error) error {
	var (
		wg   sync.WaitGroup
		once sync.Once
		err  error
	)
	for _, fn := range fns {
		wg.Add(1)
		go func(fn func() error) {
			defer func() {
				if e := recover(); e != nil {
					buffer := make([]byte, 10*1024)
					runtime.Stack(buffer, false)
					strErr := fmt.Sprintf("[PANIC] %v\n%s", e, utils.Byte2String(buffer))
					once.Do(func() {
						err = errors.New(strErr)
					})
				}
				wg.Done()
			}()
			if e := fn(); e != nil {
				once.Do(func() {
					err = e
				})
			}
		}(fn)
	}
	wg.Wait()
	return err
}

func Go(printf func(formatter string, args ...any), fn func()) {
	go func() {
		defer func() {
			if e := recover(); e != nil {
				buffer := make([]byte, 10*1024)
				runtime.Stack(buffer, false)
				printf("[PANIC] %v\n%s", e, utils.Byte2String(buffer))
			}
		}()
		fn()
	}()
}

// 如果我们有定时任务需要执行，同时我们不希望失败就退出，而是要继续执行，封装co.GoAndRetry接口
func GoAndRetry(printf func(formatter string, args ...any), fn func(), retryDelay int) {
	go func() {
		defer func() {
			if e := recover(); e != nil {
				buffer := make([]byte, 10*1024)
				runtime.Stack(buffer, false)
				printf("[PANIC] %v\n%s", e, utils.Byte2String(buffer))
			}
			time.Sleep(time.Duration(retryDelay) * time.Millisecond)
			GoAndRetry(printf, fn, retryDelay)
		}()
		fn()
	}()
}

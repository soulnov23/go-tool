package coroutine

import (
	"errors"
	"fmt"
	"runtime/debug"
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
		wg     sync.WaitGroup
		once   sync.Once
		fnsErr error
	)
	for _, fn := range fns {
		wg.Add(1)
		go func(fn func() error) {
			defer func() {
				if err := recover(); err != nil {
					strErr := fmt.Sprintf("[PANIC] %v\n%s", err, utils.Byte2String(debug.Stack()))
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

// 谨慎使用，仅适用旁路分支不需要对panic做处理的场景下
// 考虑fn是事件循环模型，由于panic协程退出了，主协程没有感知还在运行，实际上svr已经无法提供服务了
func Go(printf func(formatter string, args ...any), fn func(args ...any), args ...any) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				printf("[PANIC] %v\n%s", err, utils.Byte2String(debug.Stack()))
			}
		}()
		fn(args...)
	}()
}

// 如果我们有定时任务需要执行，同时我们不希望失败就退出，而是要继续执行，封装co.GoAndRetry接口
func GoAndRetry(printf func(formatter string, args ...any), retryDelay int, fn func(args ...any) error, args ...any) {
	go func() {
		defer func() {
			// panic打印后退出协程
			if err := recover(); err != nil {
				printf("[PANIC] %v\n%s", err, utils.Byte2String(debug.Stack()))
			}
		}()
		// 任务报错sleep后继续执行
		if err := fn(args...); err != nil {
			time.Sleep(time.Duration(retryDelay) * time.Millisecond)
			GoAndRetry(printf, retryDelay, fn, args...)
		}
	}()
}

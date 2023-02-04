package co

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/SoulNov23/go-tool/pkg/log"

	"github.com/SoulNov23/go-tool/pkg/unsafe"
)

/*
1. 多协程并发执行时，一个协程抛出panic会导致所有协程全部退出，一个协程的panic只能该协程调用recover捕获，所以每个协程都要执行defer func() { recover() }()
2. 如果我们有批量的任务需要执行，肯定通过并发调用来提高性能，同时我们不希望其中一个调用失败就导致所有的任务都退出，而是要继续执行完其它的任务，封装co.GoAndWait接口
*/
func GoAndWait(handlers ...func() error) error {
	var (
		wg   sync.WaitGroup
		once sync.Once
		err  error
	)
	for _, handler := range handlers {
		wg.Add(1)
		go func(handler func() error) {
			defer func() {
				if e := recover(); e != nil {
					buffer := make([]byte, 10*1024)
					runtime.Stack(buffer, false)
					strErr := fmt.Sprintf("[PANIC] %v\n%s\n", e, unsafe.Byte2String(buffer))
					once.Do(func() {
						err = errors.New(strErr)
					})
				}
				wg.Done()
			}()
			if e := handler(); e != nil {
				once.Do(func() {
					err = e
				})
			}
		}(handler)
	}
	wg.Wait()
	return err
}

func Go(log log.Logger, handler func()) {
	go func() {
		defer func() {
			if e := recover(); e != nil {
				buffer := make([]byte, 10*1024)
				runtime.Stack(buffer, false)
				log.Errorf("[PANIC] %v\n%s\n", e, unsafe.Byte2String(buffer))
			}
		}()
		handler()
	}()
}

// 如果我们有定时任务需要执行，同时我们不希望失败就退出，而是要继续执行，封装co.GoAndRetry接口
func GoAndRetry(log log.Logger, handler func(), retryDuration time.Duration) {
	go func() {
		defer func() {
			if e := recover(); e != nil {
				buffer := make([]byte, 10*1024)
				runtime.Stack(buffer, false)
				log.Errorf("[PANIC] %v\n%s\n", e, unsafe.Byte2String(buffer))
			}
			time.Sleep(retryDuration)
			GoAndRetry(log, handler, retryDuration)
		}()
		handler()
	}()
}

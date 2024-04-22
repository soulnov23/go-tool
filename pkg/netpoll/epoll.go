package netpoll

import (
	"errors"
	"fmt"
	"runtime"
	"sync/atomic"
	"unsafe"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

type Epoll struct {
	fd            int
	wakeOperator  *FDOperator // eventfd, wake epoll_wait
	events        []EpollEvent
	operatorCache *operatorCache
	triggerBuf    []byte
	trigger       atomic.Uint32
	close         chan struct{}
	info          func(msg string, fields ...zap.Field)
}

func NewEpoll(info func(msg string, fields ...zap.Field)) (*Epoll, error) {
	fd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, fmt.Errorf("unix.EpollCreate1: %v", err)
	}
	epoll := &Epoll{
		fd:            fd,
		events:        make([]EpollEvent, 128), // https://github.com/golang/go/blob/master/src/runtime/netpoll_epoll.go#L114
		operatorCache: newOperatorCache(),
		triggerBuf:    make([]byte, 8),
		close:         make(chan struct{}, 1),
		info:          info,
	}
	eventFD, err := unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("unix.Eventfd: %v", err)
	}
	operator := epoll.Alloc()
	operator.FD = eventFD
	operator.Epoll = epoll
	if err := epoll.Control(operator, Readable); err != nil {
		unix.Close(eventFD)
		unix.Close(fd)
		return nil, fmt.Errorf("epoll_fd[%d] epoll.Control event_fd[%d]: %v", fd, eventFD, err)
	}
	epoll.wakeOperator = operator
	epoll.info("new epoll", zap.Int("epoll_fd", fd), zap.Int("event_fd", eventFD))
	return epoll, nil
}

func (epoll *Epoll) FD() int {
	return epoll.fd
}

func (epoll *Epoll) Control(operator *FDOperator, event int) error {
	if operator == nil {
		return errors.New("operator is nil")
	}
	epollEvent := &EpollEvent{}
	*(**FDOperator)(unsafe.Pointer(&epollEvent.Data)) = operator
	switch event {
	case Readable:
		epollEvent.Events = ReadFlags
		return EpollCtl(epoll.fd, unix.EPOLL_CTL_ADD, operator.FD, epollEvent)
	case ModReadable:
		epollEvent.Events = ReadFlags
		return EpollCtl(epoll.fd, unix.EPOLL_CTL_MOD, operator.FD, epollEvent)
	case Writable:
		epollEvent.Events = WriteFlags
		return EpollCtl(epoll.fd, unix.EPOLL_CTL_ADD, operator.FD, epollEvent)
	case ModWritable:
		epollEvent.Events = WriteFlags
		return EpollCtl(epoll.fd, unix.EPOLL_CTL_MOD, operator.FD, epollEvent)
	case ReadWritable:
		epollEvent.Events = ReadFlags | WriteFlags
		return EpollCtl(epoll.fd, unix.EPOLL_CTL_ADD, operator.FD, epollEvent)
	case ModReadWritable:
		epollEvent.Events = ReadFlags | WriteFlags
		return EpollCtl(epoll.fd, unix.EPOLL_CTL_MOD, operator.FD, epollEvent)
	case Detach:
		return EpollCtl(epoll.fd, unix.EPOLL_CTL_DEL, operator.FD, nil)
	default:
		return fmt.Errorf("event[%d] not support", event)
	}
}

func (epoll *Epoll) Wait() error {
	epoll.info("wait epoll", zap.Int("epoll_fd", epoll.fd), zap.Int("event_fd", epoll.wakeOperator.FD))
	// 先epoll_wait阻塞等待
	msec := -1
	for {
		n, err := EpollWait(epoll.fd, epoll.events, msec)
		if err != nil && err != unix.EINTR {
			return fmt.Errorf("unix.EpollWait: %v", err)
		}
		// 轮询没有事件发生，直接阻塞协程，然后主动切换协程
		if n <= 0 {
			msec = -1
			runtime.Gosched()
			continue
		}
		msec = 0
		if epoll.handle(n) {
			if err := epoll.Control(epoll.wakeOperator, Detach); err != nil {
				epoll.info("epoll.Control event_fd failed", zap.Int("epoll_fd", epoll.fd), zap.Int("event_fd", epoll.wakeOperator.FD), zap.Error(err))
			}
			unix.Close(epoll.wakeOperator.FD)
			unix.Close(epoll.fd)
			epoll.close <- struct{}{}
			epoll.trigger.Store(0)
			epoll.info("exit gracefully", zap.Int("epoll_fd", epoll.fd), zap.Int("event_fd", epoll.wakeOperator.FD))
			return nil
		}
	}
}

func (epoll *Epoll) handle(eventSize int) bool {
	exit := false
	var hups []*FDOperator
	for i := 0; i < eventSize; i++ {
		event := epoll.events[i]
		operator := *(**FDOperator)(unsafe.Pointer(&event.Data))
		epoll.info("wake epoll", zap.Int("epoll_fd", epoll.fd), zap.Int("client_fd", operator.FD), zap.String("event", EventString(event.Events)))

		// 通过write event fd主动触发循环优雅退出
		if operator.FD == epoll.wakeOperator.FD {
			_, _ = unix.Read(epoll.wakeOperator.FD, epoll.triggerBuf)
			if epoll.triggerBuf[0] > 0 {
				exit = true
			}
			continue
		}

		if event.Events&(unix.EPOLLRDHUP|unix.EPOLLHUP|unix.EPOLLERR) != 0 {
			if operator != nil && operator.OnHup != nil {
				hups = append(hups, operator)
			}
		}

		if event.Events&(unix.EPOLLIN) != 0 {
			if operator != nil && operator.OnRead != nil {
				operator.OnRead(epoll, operator)
			}
		}

		if event.Events&unix.EPOLLOUT != 0 {
			if operator != nil && operator.OnWrite != nil {
				operator.OnWrite(epoll, operator)
			}
		}
	}
	for _, operator := range hups {
		if err := epoll.Control(operator, Detach); err != nil {
			epoll.info("epoll.Control event_fd failed", zap.Int("epoll_fd", epoll.fd), zap.Int("event_fd", epoll.wakeOperator.FD), zap.Error(err))
		}
		operator.OnHup(epoll, operator)
		epoll.Free(operator)
	}
	epoll.operatorCache.free()
	// 是否退出循环：否
	return exit
}

func (epoll *Epoll) Close() error {
	// 防止重复主动触发
	if epoll.trigger.Add(1) > 1 {
		return nil
	}
	if _, err := unix.Write(epoll.wakeOperator.FD, []byte{1, 0, 0, 0, 0, 0, 0, 1}); err != nil {
		return fmt.Errorf("epoll_fd[%d] write event_fd[%d]: %v", epoll.fd, epoll.wakeOperator.FD, err)
	}
	<-epoll.close
	return nil
}

func (epoll *Epoll) Alloc() *FDOperator {
	return epoll.operatorCache.alloc()
}

func (epoll *Epoll) Free(operator *FDOperator) {
	epoll.operatorCache.freeable(operator)
}

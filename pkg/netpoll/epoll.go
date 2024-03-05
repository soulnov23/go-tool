package netpoll

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"unsafe"

	"github.com/soulnov23/go-tool/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

type FDOperator struct {
	// FD is file descriptor, poll will bind when register.
	FD int

	// Desc provides three callbacks for fd's reading, writing or hanging events.
	OnRead  func()
	OnWrite func()
	OnHup   func()

	// Epoll is the registered location of the file descriptor.
	Epoll *Epoll
}

type Epoll struct {
	fd         int
	operator   *FDOperator
	events     []EpollEvent
	triggerBuf []byte
	trigger    atomic.Uint32
	close      chan struct{}
	log.Logger
}

func NewEpoll(eventSize int, log log.Logger) (*Epoll, error) {
	fd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, fmt.Errorf("unix.EpollCreate1: %v", err)
	}
	epoll := &Epoll{
		fd:         fd,
		events:     make([]EpollEvent, eventSize),
		triggerBuf: make([]byte, 8),
		close:      make(chan struct{}, 1),
		Logger:     log,
	}
	eventFD, err := unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("unix.Eventfd: %v", err)
	}
	operator := &FDOperator{
		FD:    eventFD,
		Epoll: epoll,
	}
	if err := epoll.Control(operator, Readable); err != nil {
		unix.Close(eventFD)
		unix.Close(fd)
		return nil, fmt.Errorf("epoll_fd[%d] epoll.Control event_fd[%d]: %v", fd, eventFD, err)
	}
	epoll.operator = operator
	return epoll, nil
}

func (epoll *Epoll) Control(operator *FDOperator, event int) error {
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
	epoll.InfoFields("wait epoll", zap.Int("epoll_fd", epoll.fd), zap.Int("event_fd", epoll.operator.FD))
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
			epoll.Control(epoll.operator, Detach)
			unix.Close(epoll.operator.FD)
			unix.Close(epoll.fd)
			epoll.close <- struct{}{}
			epoll.trigger.Store(0)
			epoll.InfoFields("exit gracefully", zap.Int("epoll_fd", epoll.fd))
			return nil
		}
	}
}

func (epoll *Epoll) handle(eventSize int) bool {
	exit := false
	for i := 0; i < eventSize; i++ {
		event := epoll.events[i]
		operator := *(**FDOperator)(unsafe.Pointer(&event.Data))
		epoll.InfoFields("wake epoll", zap.Int("epoll_fd", epoll.fd), zap.Int("client_fd", operator.FD), zap.String("event", EventString(event.Events)))

		// 通过write event fd主动触发循环优雅退出
		if operator.FD == epoll.operator.FD {
			unix.Read(epoll.operator.FD, epoll.triggerBuf)
			if epoll.triggerBuf[0] > 0 {
				exit = true
			}
			continue
		}

		if event.Events&(unix.EPOLLRDHUP|unix.EPOLLHUP|unix.EPOLLERR) != 0 {
			operator.OnHup()
			epoll.Control(operator, Detach)
			continue
		}

		if event.Events&(unix.EPOLLIN|unix.EPOLLPRI) != 0 {
			operator.OnRead()
		}

		if event.Events&unix.EPOLLOUT != 0 {
			operator.OnWrite()
		}
	}
	// 是否退出循环：否
	return exit
}

/*
func (epoll *Epoll) handleAccept(fd int) {
	for {
		connFD, addr, err := unix.Accept4(fd, unix.SOCK_CLOEXEC)
		if err != nil {
			if err == unix.EAGAIN {
				break
			} else if err == unix.EINTR {
				continue
			} else {
				epoll.Logger.ErrorFields("accept client", zap.Error(err), zap.Int("epoll_fd", epoll.epollFD), zap.Int("client_fd", fd))
				continue
			}
		}
		ip, err := utils.ResolveSockaddrIP(addr)
		if err != nil {
			epoll.Logger.ErrorFields("get client remote ip", zap.Error(err), zap.Int("epoll_fd", epoll.epollFD), zap.Int("client_fd", fd))
			continue
		}
		local := epoll.listens[fd]
		epoll.Logger.DebugFields("accept client", zap.Int("epoll_fd", epoll.epollFD), zap.Int("client_fd", fd), zap.String("remote_address", ip), zap.String("local_address", local))
		utils.SetSocketCloseExec(connFD)
		if err := utils.SetSocketNonBlock(connFD); err != nil {
			unix.Close(connFD)
			epoll.Logger.ErrorFields("set client fd non-blocking", zap.Error(err), zap.Int("client_fd", fd))
			continue
		}
		if err := utils.SetSocketTCPNodelay(connFD); err != nil {
			unix.Close(connFD)
			epoll.Logger.ErrorFields("set client fd tcp no delay", zap.Error(err), zap.Int("client_fd", fd))
			continue
		}
		if err := Control(epoll.epollFD, connFD, Readable); err != nil {
			unix.Close(connFD)
			epoll.Logger.ErrorFields("epoll control client fd", zap.Error(err), zap.Int("epoll_fd", epoll.epollFD), zap.Int("client_fd", fd), zap.String("epoll_event", EventString(Readable)))
			continue
		}
		operator := epoll.operators[fd]
		tcpConn := NewTCPConn(epoll.Logger, epoll.epollFD, connFD, local, ip, operator)
		epoll.tcpConns[connFD] = tcpConn
	}
}
*/

func (epoll *Epoll) Close() error {
	// 防止重复主动触发
	if epoll.trigger.Add(1) > 1 {
		return nil
	}
	if _, err := unix.Write(epoll.operator.FD, []byte{1, 0, 0, 0, 0, 0, 0, 1}); err != nil {
		return fmt.Errorf("epoll_fd[%d] write event_fd[%d]: %v", epoll.fd, epoll.operator.FD, err)
	}
	<-epoll.close
	return nil
}

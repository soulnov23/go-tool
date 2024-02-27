package net

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"syscall"

	"github.com/soulnov23/go-tool/pkg/utils"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

type Epoll struct {
	fd         int
	eventFD    int
	listens    map[int]string
	events     []syscall.EpollEvent
	triggerBuf []byte
	trigger    uint32
	close      chan struct{}

	Operator

	infof  func(formatter string, args ...any)
	errorf func(formatter string, args ...any)
}

type Operator interface {
	OnRequest()
	OnTask()
	OnExit()
}

func NewEpoll(eventSize int, infof, errorf func(formatter string, args ...any)) (*Epoll, error) {
	fd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		return nil, fmt.Errorf("syscall.EpollCreate1: %v", err)
	}
	epoll := &Epoll{
		fd:         fd,
		listens:    make(map[int]string),
		events:     make([]syscall.EpollEvent, eventSize),
		triggerBuf: make([]byte, 8),
		close:      make(chan struct{}, 1),
		infof:      infof,
		errorf:     errorf,
	}
	eventFD, err := unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
	if err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("unix.Eventfd: %v", err)
	}
	if err := epoll.Control(eventFD, Readable); err != nil {
		syscall.Close(eventFD)
		syscall.Close(fd)
		return nil, fmt.Errorf("epoll_fd[%d] epoll.Control event_fd[%d]: %v", fd, eventFD, err)
	}
	epoll.eventFD = eventFD
	return epoll, nil
}

func (epoll *Epoll) Control(fd int, event int) error {
	epollEvent := &syscall.EpollEvent{
		Fd: int32(fd),
	}
	switch event {
	case Readable:
		epollEvent.Events = ReadFlags
		return syscall.EpollCtl(epoll.fd, syscall.EPOLL_CTL_ADD, fd, epollEvent)
	case ModReadable:
		epollEvent.Events = ReadFlags
		return syscall.EpollCtl(epoll.fd, syscall.EPOLL_CTL_MOD, fd, epollEvent)
	case Writable:
		epollEvent.Events = WriteFlags
		return syscall.EpollCtl(epoll.fd, syscall.EPOLL_CTL_ADD, fd, epollEvent)
	case ModWritable:
		epollEvent.Events = WriteFlags
		return syscall.EpollCtl(epoll.fd, syscall.EPOLL_CTL_MOD, fd, epollEvent)
	case ReadWritable:
		epollEvent.Events = ReadFlags | WriteFlags
		return syscall.EpollCtl(epoll.fd, syscall.EPOLL_CTL_ADD, fd, epollEvent)
	case ModReadWritable:
		epollEvent.Events = ReadFlags | WriteFlags
		return syscall.EpollCtl(epoll.fd, syscall.EPOLL_CTL_MOD, fd, epollEvent)
	case Detach:
		return syscall.EpollCtl(epoll.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	default:
		return fmt.Errorf("event[%d] not support", event)
	}
}

func (epoll *Epoll) Wait() error {
	// 先epoll_wait阻塞等待
	msec := -1
	for {
		n, err := syscall.EpollWait(epoll.fd, epoll.events, msec)
		if err != nil && err != syscall.EINTR {
			return fmt.Errorf("syscall.EpollWait: %v", err)
		}
		// 轮询没有事件发生，直接阻塞协程，然后主动切换协程
		if n <= 0 {
			msec = -1
			runtime.Gosched()
			continue
		}
		msec = 0
		if epoll.handle(n) {
			epoll.close <- struct{}{}
			return nil
		}
	}
}

func (epoll *Epoll) handle(eventSize int) bool {
	exit := false
	for i := 0; i < eventSize; i++ {
		fd := int(epoll.events[i].Fd)
		event := epoll.events[i].Events
		epoll.infof("client_fd[%d] event[%s] wake epoll_fd[%d]", fd, EventString(event), epoll.fd)

		// 通过write event fd触发
		if fd == epoll.eventFD {
			exit = epoll.handleEventFD()
			continue
		}

		if event&(syscall.EPOLLRDHUP|syscall.EPOLLHUP|syscall.EPOLLERR) != 0 {
			if !ok {
				continue
			}
			epoll.Logger.DebugFields("close client", zap.Int("epoll_fd", epoll.epollFD), zap.Int("client_fd", fd), zap.String("remote_address", conn.remoteAddr), zap.String("local_address", conn.localAddr))
			DeleteTCPConn(conn)
			delete(epoll.tcpConns, fd)
			continue
		}

		if event&(syscall.EPOLLIN|syscall.EPOLLPRI) != 0 {
			if _, ok := epoll.listens[fd]; ok {
				epoll.handleAccept(fd)
				continue
			}
			if !ok {
				continue
			}
			conn.handleRead()
			conn.Operator.OnRequest(conn)
		}

		if event&syscall.EPOLLOUT != 0 {
			if !ok {
				continue
			}
			conn.handleWrite()
		}
	}
	// 是否退出循环：否
	return exit
}

func (epoll *Epoll) handleEventFD() bool {
	offset := 0
	for {
		n, err := syscall.Read(epoll.eventFD, epoll.triggerBuf)
		if err != nil {
			if err == syscall.EAGAIN {
				break
			} else if err == syscall.EINTR {
				continue
			} else {
				epoll.errorf("epoll_fd[%d] syscall.Read fd[%d]: %v", epoll.fd, epoll.eventFD, err)
				break
			}
		}
		offset += n
		if n == 0 || offset == 8 {
			break
		}
	}
	atomic.StoreUint32(&epoll.trigger, 0)
	// 主动触发循环优雅退出
	if epoll.triggerBuf[0] > 0 {
		epoll.infof("exit gracefully")
		epoll.OnExit()
		syscall.Close(epoll.eventFD)
		syscall.Close(epoll.fd)
		return true
	}
	// 主动触发循环执行异步任务
	epoll.infof("task trigger")
	epoll.OnTask()
	return false
}

func (epoll *Epoll) handleAccept(fd int) {
	for {
		connFD, addr, err := syscall.Accept4(fd, syscall.SOCK_CLOEXEC)
		if err != nil {
			if err == syscall.EAGAIN {
				break
			} else if err == syscall.EINTR {
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
			syscall.Close(connFD)
			epoll.Logger.ErrorFields("set client fd non-blocking", zap.Error(err), zap.Int("client_fd", fd))
			continue
		}
		if err := utils.SetSocketTCPNodelay(connFD); err != nil {
			syscall.Close(connFD)
			epoll.Logger.ErrorFields("set client fd tcp no delay", zap.Error(err), zap.Int("client_fd", fd))
			continue
		}
		if err := Control(epoll.epollFD, connFD, Readable); err != nil {
			syscall.Close(connFD)
			epoll.Logger.ErrorFields("epoll control client fd", zap.Error(err), zap.Int("epoll_fd", epoll.epollFD), zap.Int("client_fd", fd), zap.String("epoll_event", EventString(Readable)))
			continue
		}
		operator := epoll.operators[fd]
		tcpConn := NewTCPConn(epoll.Logger, epoll.epollFD, connFD, local, ip, operator)
		epoll.tcpConns[connFD] = tcpConn
	}
}

func (epoll *Epoll) Trigger() error {
	// 防止重复主动触发
	if atomic.AddUint32(&epoll.trigger, 1) > 1 {
		return nil
	}
	if _, err := syscall.Write(epoll.eventFD, []byte{0, 0, 0, 0, 0, 0, 0, 1}); err != nil {
		epoll.Logger.ErrorFields("write event fd trigger", zap.Error(err), zap.Int("epoll_fd", epoll.epollFD), zap.Int("event_fd", epoll.eventFD))
		return fmt.Errorf("write event fd trigger: " + err.Error())
	}
	return nil
}

func (epoll *Epoll) Close() error {
	if _, err := syscall.Write(epoll.eventFD, []byte{1, 0, 0, 0, 0, 0, 0, 1}); err != nil {
		epoll.Logger.ErrorFields("write event fd close", zap.Error(err), zap.Int("epoll_fd", epoll.epollFD), zap.Int("event_fd", epoll.eventFD))
		return fmt.Errorf("write event fd close: " + err.Error())
	}
	<-epoll.close
	return nil
}

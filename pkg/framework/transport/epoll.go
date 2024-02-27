package transport

import (
	"errors"
	"runtime"
	"sync/atomic"
	"syscall"

	"github.com/soulnov23/go-tool/pkg/log"
	"github.com/soulnov23/go-tool/pkg/utils"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

const (
	ReadFlags  = syscall.EPOLLIN | syscall.EPOLLRDHUP | syscall.EPOLLHUP | syscall.EPOLLERR | syscall.EPOLLPRI
	WriteFlags = unix.EPOLLET | syscall.EPOLLOUT | syscall.EPOLLHUP | syscall.EPOLLERR
)

const (
	Readable = iota
	ModReadable
	Writable
	ModWritable
	ReadWritable
	ModReadWritable
	Detach
)

func EventString(event uint32) string {
	eventString := "EPOLLET"
	if event&unix.EPOLLET != 0 {
		eventString += "EPOLLET"
	}
	if event&syscall.EPOLLIN != 0 {
		eventString += "|EPOLLIN"
	}
	if event&syscall.EPOLLPRI != 0 {
		eventString += "|EPOLLPRI"
	}
	if event&syscall.EPOLLOUT != 0 {
		eventString += "|EPOLLOUT"
	}
	if event&syscall.EPOLLHUP != 0 {
		eventString += "|EPOLLHUP"
	}
	if event&syscall.EPOLLRDHUP != 0 {
		eventString += "|EPOLLRDHUP"
	}
	if event&syscall.EPOLLERR != 0 {
		eventString += "|EPOLLERR"
	}
	return eventString
}

func Control(epollFD int, fd int, event int) error {
	evt := &syscall.EpollEvent{
		Fd: int32(fd),
	}
	switch event {
	case Readable:
		evt.Events = ReadFlags
		return syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, evt)
	case ModReadable:
		evt.Events = ReadFlags
		return syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_MOD, fd, evt)
	case Writable:
		evt.Events = WriteFlags
		return syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, evt)
	case ModWritable:
		evt.Events = WriteFlags
		return syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_MOD, fd, evt)
	case ReadWritable:
		evt.Events = ReadFlags | WriteFlags
		return syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, evt)
	case ModReadWritable:
		evt.Events = ReadFlags | WriteFlags
		return syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_MOD, fd, evt)
	case Detach:
		return syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_DEL, fd, nil)
	default:
		return errors.New("event not support")
	}
}

type Epoll struct {
	log.Logger
	epollFD    int
	eventFD    int
	listens    map[int]string
	operators  map[int]Operator
	events     []syscall.EpollEvent
	tcpConns   map[int]*TCPConnection
	triggerBuf []byte
	trigger    uint32
	close      chan struct{}
}

func NewEpoll(log log.Logger, eventSize int) (*Epoll, error) {
	epollFD, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		log.ErrorFields("epoll create", zap.Error(err))
		return nil, errors.New("epoll create: " + err.Error())
	}
	eventFD, err := unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
	if err != nil {
		syscall.Close(epollFD)
		log.ErrorFields("new event fd", zap.Error(err))
		return nil, errors.New("new event fd: " + err.Error())
	}
	if err := Control(epollFD, eventFD, Readable); err != nil {
		syscall.Close(eventFD)
		syscall.Close(epollFD)
		log.ErrorFields("epoll control event fd", zap.Error(err), zap.Int("epoll_fd", epollFD), zap.Int("event_fd", eventFD))
		return nil, errors.New("epoll control event fd: " + err.Error())
	}
	log.DebugFields("new epoll success", zap.Int("epoll_fd", epollFD), zap.Int("event_fd", eventFD))
	return &Epoll{
		Logger:     log,
		epollFD:    epollFD,
		eventFD:    eventFD,
		listens:    make(map[int]string),
		operators:  make(map[int]Operator),
		events:     make([]syscall.EpollEvent, eventSize),
		tcpConns:   make(map[int]*TCPConnection),
		triggerBuf: make([]byte, 8),
		close:      make(chan struct{}, 1),
	}, nil
}

func (epoll *Epoll) Listen(network string, address string, backlog int, operator Operator) error {
	listenFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		epoll.Logger.ErrorFields("new listen fd", zap.Error(err))
		return errors.New("new listen fd: " + err.Error())
	}
	utils.SetSocketCloseExec(listenFD)
	if err := utils.SetSocketNonBlock(listenFD); err != nil {
		syscall.Close(listenFD)
		epoll.Logger.ErrorFields("set listen fd non-blocking", zap.Error(err), zap.Int("listen_fd", listenFD))
		return errors.New("set listen fd non-blocking: " + err.Error())
	}
	if err := utils.SetSocketReuseaddr(listenFD); err != nil {
		syscall.Close(listenFD)
		epoll.Logger.ErrorFields("set listen fd reuse address", zap.Error(err), zap.Int("listen_fd", listenFD))
		return errors.New("set listen fd reuse address: " + err.Error())
	}
	if err := utils.SetSocketReUsePort(listenFD); err != nil {
		syscall.Close(listenFD)
		epoll.Logger.ErrorFields("set listen fd reuse port", zap.Error(err), zap.Int("listen_fd", listenFD))
		return errors.New("set listen fd reuse port: " + err.Error())
	}
	sa, err := utils.ResolveSockaddr(network, address)
	if err != nil {
		syscall.Close(listenFD)
		epoll.Logger.ErrorFields("get listen fd bind address", zap.Error(err), zap.Int("listen_fd", listenFD), zap.String("network", network), zap.String("address", address))
		return errors.New("get listen fd bind address: " + err.Error())
	}
	if err := syscall.Bind(listenFD, sa); err != nil {
		syscall.Close(listenFD)
		epoll.Logger.ErrorFields("set listen fd bind address", zap.Error(err), zap.Int("listen_fd", listenFD), zap.String("network", network), zap.String("address", address), zap.Reflect("sockaddr", sa))
		return errors.New("set listen fd bind address: " + err.Error())
	}
	if err := syscall.Listen(listenFD, backlog); err != nil {
		syscall.Close(listenFD)
		epoll.Logger.ErrorFields("set listen fd backlog", zap.Error(err), zap.Int("listen_fd", listenFD), zap.Int("backlog", backlog))
		return errors.New("set listen fd backlog: " + err.Error())
	}
	if err := Control(epoll.epollFD, listenFD, Readable); err != nil {
		syscall.Close(listenFD)
		epoll.Logger.ErrorFields("epoll control listen fd", zap.Error(err), zap.Int("epoll_fd", epoll.epollFD), zap.Int("listen_fd", listenFD))
		return errors.New("epoll control listen fd: " + err.Error())
	}
	epoll.listens[listenFD] = address
	epoll.operators[listenFD] = operator
	epoll.Logger.DebugFields("listen success", zap.Int("listen_fd", listenFD), zap.String("network", network), zap.String("address", address))
	return nil
}

func (epoll *Epoll) Wait() error {
	// 先epoll_wait阻塞等待
	msec := -1
	for {
		n, err := syscall.EpollWait(epoll.epollFD, epoll.events, msec)
		if err != nil && err != syscall.EINTR {
			epoll.Logger.ErrorFields("epoll wait", zap.Error(err), zap.Int("epoll_fd", epoll.epollFD))
			return errors.New("epoll wait: " + err.Error())
		}
		// 轮询没有事件发生，直接阻塞协程，然后主动切换协程
		if n <= 0 {
			msec = -1
			runtime.Gosched()
			continue
		}
		msec = 0
		if epoll.handler(n) {
			epoll.close <- struct{}{}
			return nil
		}
	}
}

func (epoll *Epoll) handler(eventSize int) bool {
	for i := 0; i < eventSize; i++ {
		fd := int(epoll.events[i].Fd)
		evt := epoll.events[i].Events
		epoll.Logger.DebugFields("wake epoll wait", zap.Int("epoll_fd", epoll.epollFD), zap.String("epoll_event", EventString(evt)), zap.Int("client_fd", fd))

		// 通过write evfd触发
		if fd == epoll.eventFD {
			if epoll.handlerEventFD() {
				return true
			}
			continue
		}

		conn, ok := epoll.tcpConns[fd]
		if evt&(syscall.EPOLLRDHUP|syscall.EPOLLHUP|syscall.EPOLLERR) != 0 {
			if !ok {
				continue
			}
			epoll.Logger.DebugFields("close client", zap.Int("epoll_fd", epoll.epollFD), zap.Int("client_fd", fd), zap.String("remote_address", conn.remoteAddr), zap.String("local_address", conn.localAddr))
			DeleteTCPConn(conn)
			delete(epoll.tcpConns, fd)
			continue
		}

		if evt&(syscall.EPOLLIN|syscall.EPOLLPRI) != 0 {
			if _, ok := epoll.listens[fd]; ok {
				epoll.handlerAccept(fd)
				continue
			}
			if !ok {
				continue
			}
			conn.handlerRead()
			conn.Operator.OnRequest(conn)
		}

		if evt&syscall.EPOLLOUT != 0 {
			if !ok {
				continue
			}
			conn.handlerWrite()
		}
	}
	// 是否退出循环：否
	return false
}

func (epoll *Epoll) handlerEventFD() bool {
	offset := 0
	for {
		n, err := syscall.Read(epoll.eventFD, epoll.triggerBuf)
		if err != nil {
			if err == syscall.EAGAIN {
				break
			} else if err == syscall.EINTR {
				continue
			} else {
				epoll.Logger.ErrorFields("read event fd", zap.Error(err), zap.Int("epoll_fd", epoll.epollFD), zap.Int("event_fd", epoll.eventFD))
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
		epoll.Logger.DebugFields("exit gracefully")
		for fd, conn := range epoll.tcpConns {
			epoll.Logger.DebugFields("close client", zap.Int("epoll_fd", epoll.epollFD), zap.Int("client_fd", fd), zap.String("remote_address", conn.remoteAddr), zap.String("local_address", conn.localAddr))
			DeleteTCPConn(conn)
			delete(epoll.tcpConns, fd)
		}
		for fd := range epoll.listens {
			Control(epoll.epollFD, fd, Detach)
			syscall.Close(fd)
			delete(epoll.listens, fd)
		}
		syscall.Close(epoll.eventFD)
		syscall.Close(epoll.epollFD)
		return true
	}
	// 主动触发循环执行异步任务
	epoll.Logger.DebugFields("trigger")
	return false
}

func (epoll *Epoll) handlerAccept(fd int) {
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
		return errors.New("write event fd trigger: " + err.Error())
	}
	return nil
}

func (epoll *Epoll) Close() error {
	if _, err := syscall.Write(epoll.eventFD, []byte{1, 0, 0, 0, 0, 0, 0, 1}); err != nil {
		epoll.Logger.ErrorFields("write event fd close", zap.Error(err), zap.Int("epoll_fd", epoll.epollFD), zap.Int("event_fd", epoll.eventFD))
		return errors.New("write event fd close: " + err.Error())
	}
	<-epoll.close
	return nil
}

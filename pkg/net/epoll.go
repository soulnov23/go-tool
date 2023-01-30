package net

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/SoulNov23/go-tool/pkg/log"
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

type Server interface {
	OnAccept(conn *TcpConn)
	OnClose(conn *TcpConn)
	OnRead(conn *TcpConn)
}

type Epoll struct {
	log        log.Logger
	epollFD    int
	eventFD    int
	listens    map[int]string
	events     []syscall.EpollEvent
	tcpConns   map[int]*TcpConn
	connLock   sync.Mutex
	triggerBuf []byte
	trigger    uint32
	close      chan struct{}

	server Server
}

func NewEpoll(log log.Logger, eventSize int, server Server) (*Epoll, error) {
	epollFD, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		wrapErr := errors.New("syscall.EpollCreate1: " + err.Error())
		log.Error(wrapErr)
		return nil, wrapErr
	}
	eventFD, err := unix.Eventfd(0, unix.EFD_NONBLOCK|unix.EFD_CLOEXEC)
	if err != nil {
		syscall.Close(epollFD)
		wrapErr := errors.New("net.Eventfd: " + err.Error())
		log.Error(wrapErr)
		return nil, wrapErr
	}
	if err := Control(epollFD, eventFD, Readable); err != nil {
		syscall.Close(eventFD)
		syscall.Close(epollFD)
		wrapErr := errors.New("net.Control: " + err.Error())
		log.Error(wrapErr)
		return nil, wrapErr
	}
	log.Debugf("create epoll fd: %d, event fd: %d", epollFD, eventFD)
	return &Epoll{
		log:        log,
		epollFD:    epollFD,
		eventFD:    eventFD,
		listens:    make(map[int]string),
		events:     make([]syscall.EpollEvent, eventSize),
		tcpConns:   make(map[int]*TcpConn),
		triggerBuf: make([]byte, 8),
		close:      make(chan struct{}, 1),
		server:     server,
	}, nil
}

func (ep *Epoll) Listen(network string, address string, backlog int) error {
	listenFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		wrapErr := errors.New("syscall.Socket: " + err.Error())
		ep.log.Error(wrapErr)
		return wrapErr
	}
	SetSocketCloseExec(listenFD)
	if err := SetSocketNonBlock(listenFD); err != nil {
		syscall.Close(listenFD)
		wrapErr := errors.New("SetSocketNonBlock: " + err.Error())
		ep.log.Error(wrapErr)
		return wrapErr
	}
	if err := SetSocketReuseaddr(listenFD); err != nil {
		syscall.Close(listenFD)
		wrapErr := errors.New("SetSocketReuseaddr: " + err.Error())
		ep.log.Error(wrapErr)
		return wrapErr
	}
	if err := SetSocketReUsePort(listenFD); err != nil {
		syscall.Close(listenFD)
		wrapErr := errors.New("SetSocketReUsePort: " + err.Error())
		ep.log.Error(wrapErr)
		return wrapErr
	}
	sa, err := GetSocketAddr(network, address)
	if err != nil {
		syscall.Close(listenFD)
		wrapErr := errors.New("net.GetSocketAddr: " + err.Error())
		ep.log.Error(wrapErr)
		return wrapErr
	}
	if err := syscall.Bind(listenFD, sa); err != nil {
		syscall.Close(listenFD)
		wrapErr := errors.New("syscall.Bind: " + err.Error())
		ep.log.Error(wrapErr)
		return wrapErr
	}
	if err := syscall.Listen(listenFD, backlog); err != nil {
		syscall.Close(listenFD)
		wrapErr := errors.New("syscall.Listen: " + err.Error())
		ep.log.Error(wrapErr)
		return wrapErr
	}
	if err := Control(ep.epollFD, listenFD, Readable); err != nil {
		syscall.Close(listenFD)
		wrapErr := errors.New("net.Control: " + err.Error())
		ep.log.Error(wrapErr)
		return wrapErr
	}
	ep.listens[listenFD] = address
	ep.log.Debugf("listen %s success, fd: %d", address, listenFD)
	return nil
}

func (ep *Epoll) Wait() error {
	// 先epoll_wait阻塞等待
	msec := -1
	for {
		n, err := syscall.EpollWait(ep.epollFD, ep.events, msec)
		if err != nil && err != syscall.EINTR {
			wrapErr := errors.New("syscall.EpollWait: " + err.Error())
			ep.log.Error(wrapErr)
			return wrapErr
		}
		// 轮询没有事件发生，直接阻塞协程，然后主动切换协程
		if n <= 0 {
			msec = -1
			runtime.Gosched()
			continue
		}
		msec = 0
		if ep.handler() {
			return nil
		}
	}
}

func (ep *Epoll) handler() bool {
	ep.log.Debugf("wake epoll fd: %d", ep.epollFD)
	for i := range ep.events {
		fd := int(ep.events[i].Fd)
		// 通过write evfd触发
		if fd == ep.eventFD {
			// 主动触发循环
			ep.log.Debug("trigger")
			// 将eventfd清零
			syscall.Read(fd, ep.triggerBuf)
			atomic.StoreUint32(&ep.trigger, 0)
			// TODO执行异步任务
			continue
		}
		evt := ep.events[i].Events
		switch {
		case evt&(syscall.EPOLLRDHUP|syscall.EPOLLHUP|syscall.EPOLLERR) != 0:
			ep.connLock.Lock()
			conn, ok := ep.tcpConns[fd]
			if !ok {
				ep.connLock.Unlock()
				continue
			}
			ep.log.Debugf("close %s->%s, fd: %d", conn.remoteAddr, conn.localAddr, fd)
			ep.server.OnClose(conn)
			DeleteTcpConn(conn)
			delete(ep.tcpConns, fd)
			ep.connLock.Unlock()
		case evt&(syscall.EPOLLIN|syscall.EPOLLPRI) != 0:
			if _, ok := ep.listens[fd]; ok {
				ep.handlerAccept(fd)
				continue
			}
			ep.tcpConns[fd].handlerRead()
			ep.server.OnRead(ep.tcpConns[fd])
		case evt&syscall.EPOLLOUT != 0:
			ep.tcpConns[fd].handlerWrite()
		}
	}
	// 是否退出循环：否
	return false
}

func (ep *Epoll) handlerAccept(fd int) {
	for {
		connFD, addr, err := syscall.Accept4(fd, syscall.SOCK_CLOEXEC)
		if err != nil {
			if err == syscall.EAGAIN {
				break
			} else if err == syscall.EINTR {
				continue
			} else {
				ep.log.Errorf("syscall.Accept4: %v", err)
				continue
			}
		}
		ip, err := GetSocketIP(addr)
		if err != nil {
			ep.log.Errorf("GetSocketIP: %s", err.Error())
			continue
		}
		ep.log.Debugf("accept %s->%s fd: %d", ip, ep.listens[fd], connFD)
		SetSocketCloseExec(connFD)
		if err := SetSocketNonBlock(connFD); err != nil {
			syscall.Close(connFD)
			ep.log.Errorf("SetSocketNonBlock: %v", err)
			continue
		}
		if err := SetSocketTcpNodelay(connFD); err != nil {
			syscall.Close(connFD)
			ep.log.Errorf("SetSocketTcpNodelay: %v", err)
			continue
		}
		if err := Control(ep.epollFD, connFD, Readable); err != nil {
			syscall.Close(connFD)
			ep.log.Errorf("net.Control: %v", err)
			continue
		}
		tcpConn := NewTcpConn(ep.log, ep.epollFD, connFD, ep.listens[fd], ip)
		ep.connLock.Lock()
		ep.tcpConns[connFD] = tcpConn
		ep.server.OnAccept(tcpConn)
		ep.connLock.Unlock()
	}
}

func (ep *Epoll) Trigger() error {
	// 防止重复主动触发
	if atomic.AddUint32(&ep.trigger, 1) > 1 {
		return nil
	}
	if _, err := syscall.Write(ep.eventFD, []byte{0, 0, 0, 0, 0, 0, 0, 1}); err != nil {
		wrapErr := errors.New("notify eventFD trigger: " + err.Error())
		ep.log.Error(wrapErr)
		return wrapErr
	}
	return nil
}

func (ep *Epoll) Close() {
	syscall.Close(ep.eventFD)
	syscall.Close(ep.epollFD)
}

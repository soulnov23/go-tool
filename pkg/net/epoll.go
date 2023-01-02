package net

import (
	"errors"
	"net"
	"net/netip"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/SoulNov23/go-tool/pkg/buffer"
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
}

func NewEpoll(log log.Logger, eventSize int) (*Epoll, error) {
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
	log.Debugf("Epoll create success, Epoll fd: %d, event fd: %d", epollFD, eventFD)
	return &Epoll{
		log:        log,
		epollFD:    epollFD,
		eventFD:    eventFD,
		listens:    make(map[int]string),
		events:     make([]syscall.EpollEvent, eventSize),
		tcpConns:   make(map[int]*TcpConn),
		triggerBuf: make([]byte, 8),
		close:      make(chan struct{}, 1),
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
	addr, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		syscall.Close(listenFD)
		wrapErr := errors.New("net.ResolveTCPAddr: " + err.Error())
		ep.log.Error(wrapErr)
		return wrapErr
	}
	sa := &syscall.SockaddrInet4{
		Port: addr.Port,
	}
	for i := 0; i < net.IPv4len; i++ {
		sa.Addr[i] = addr.IP[i]
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
			if _, ok := ep.tcpConns[fd]; !ok {
				ep.connLock.Unlock()
				continue
			}
			ep.log.Debugf("close ip: %s, fd: %d", ep.tcpConns[fd].remoteAddr, fd)
			Control(ep.epollFD, fd, Detach)
			delete(ep.tcpConns, fd)
			ep.connLock.Unlock()
		case evt&(syscall.EPOLLIN|syscall.EPOLLPRI) != 0:
			if _, ok := ep.listens[fd]; ok {
				ep.handlerAccept(fd)
			} else {
				ep.tcpConns[fd].handlerRead()
			}
		case evt&syscall.EPOLLOUT != 0:
			ep.tcpConns[fd].handlerWrite()
		default:
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
		sa, ok := addr.(*syscall.SockaddrInet4)
		if !ok {
			ep.log.Errorf("convert addr to syscall.SockaddrInet4 failed")
			continue
		}
		ip := netip.AddrFrom4(sa.Addr).String() + ":" + strconv.Itoa(sa.Port)
		ep.log.Debugf("accept ip: %v, fd: %d", ip, connFD)
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
		tcpConn := &TcpConn{
			log:         ep.log,
			epollFD:     ep.epollFD,
			fd:          connFD,
			localAddr:   ep.listens[fd],
			remoteAddr:  ip,
			readBuffer:  buffer.NewBuffer(),
			writeBuffer: buffer.NewBuffer(),
		}
		ep.connLock.Lock()
		ep.tcpConns[connFD] = tcpConn
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

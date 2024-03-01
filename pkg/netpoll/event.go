package netpoll

import (
	"syscall"
	"unsafe"

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

func EpollCtl(epfd int, op int, fd int, event *EpollEvent) error {
	var err error
	_, _, err = syscall.RawSyscall6(syscall.SYS_EPOLL_CTL, uintptr(epfd), uintptr(op), uintptr(fd), uintptr(unsafe.Pointer(event)), 0, 0)
	if err == syscall.Errno(0) {
		err = nil
	}
	return err
}

func EpollWait(epfd int, events []EpollEvent, msec int) (int, error) {
	p := unsafe.Pointer(&events[0])
	var r uintptr
	var err error
	if msec == 0 {
		r, _, err = syscall.RawSyscall6(syscall.SYS_EPOLL_PWAIT, uintptr(epfd), uintptr(p), uintptr(len(events)), 0, 0, 0)
	} else {
		r, _, err = syscall.Syscall6(syscall.SYS_EPOLL_PWAIT, uintptr(epfd), uintptr(p), uintptr(len(events)), uintptr(msec), 0, 0)
	}
	if err == syscall.Errno(0) {
		err = nil
	}
	return int(r), err
}

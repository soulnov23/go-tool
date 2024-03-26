package netpoll

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	ReadFlags  = unix.EPOLLIN | unix.EPOLLRDHUP | unix.EPOLLHUP | unix.EPOLLERR
	WriteFlags = unix.EPOLLOUT | unix.EPOLLHUP | unix.EPOLLERR
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
	var eventString string
	switch {
	case event&unix.EPOLLIN != 0:
		eventString += "|EPOLLIN"
	case event&unix.EPOLLOUT != 0:
		eventString += "|EPOLLOUT"
	case event&unix.EPOLLHUP != 0:
		eventString += "|EPOLLHUP"
	case event&unix.EPOLLRDHUP != 0:
		eventString += "|EPOLLRDHUP"
	case event&unix.EPOLLERR != 0:
		eventString += "|EPOLLERR"
	}
	return eventString[1:]
}

func EpollCtl(epfd int, op int, fd int, event *EpollEvent) error {
	var err error
	_, _, err = unix.RawSyscall6(unix.SYS_EPOLL_CTL, uintptr(epfd), uintptr(op), uintptr(fd), uintptr(unsafe.Pointer(event)), 0, 0)
	if err == unix.Errno(0) {
		err = nil
	}
	return err
}

func EpollWait(epfd int, events []EpollEvent, msec int) (int, error) {
	p := unsafe.Pointer(&events[0])
	var r uintptr
	var err error
	if msec == 0 {
		r, _, err = unix.RawSyscall6(unix.SYS_EPOLL_PWAIT, uintptr(epfd), uintptr(p), uintptr(len(events)), 0, 0, 0)
	} else {
		r, _, err = unix.Syscall6(unix.SYS_EPOLL_PWAIT, uintptr(epfd), uintptr(p), uintptr(len(events)), uintptr(msec), 0, 0)
	}
	if err == unix.Errno(0) {
		err = nil
	}
	return int(r), err
}

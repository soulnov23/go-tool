package netpoll

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	ReadFlags  = unix.EPOLLIN | unix.EPOLLRDHUP | unix.EPOLLHUP | unix.EPOLLERR | unix.EPOLLPRI
	WriteFlags = unix.EPOLLET | unix.EPOLLOUT | unix.EPOLLHUP | unix.EPOLLERR
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
	if event&unix.EPOLLIN != 0 {
		eventString += "|EPOLLIN"
	}
	if event&unix.EPOLLPRI != 0 {
		eventString += "|EPOLLPRI"
	}
	if event&unix.EPOLLOUT != 0 {
		eventString += "|EPOLLOUT"
	}
	if event&unix.EPOLLHUP != 0 {
		eventString += "|EPOLLHUP"
	}
	if event&unix.EPOLLRDHUP != 0 {
		eventString += "|EPOLLRDHUP"
	}
	if event&unix.EPOLLERR != 0 {
		eventString += "|EPOLLERR"
	}
	return eventString
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

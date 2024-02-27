package net

import (
	"syscall"

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

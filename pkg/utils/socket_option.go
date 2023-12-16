package utils

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func SetSocketBlock(fd int) error {
	return syscall.SetNonblock(fd, false)
}

func SetSocketNonBlock(fd int) error {
	return syscall.SetNonblock(fd, true)
}

func SetSocketReuseaddr(fd int) error {
	var flags int = 1
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, flags)
}

func SetSocketReUsePort(fd int) error {
	var flags int = 1
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, unix.SO_REUSEPORT, flags)
}

func SetSocketTCPNodelay(fd int) error {
	var flags int = 1
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.TCP_NODELAY, flags)
}

func SetSocketKeepAlive(fd int) error {
	var flags int = 1
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, flags)
}

func SetSocketCloseExec(fd int) {
	syscall.CloseOnExec(fd)
}

func SetSocketRecvBufSize(fd int, bufSize int) error {
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_RCVBUF, bufSize)
}

func SetSocketSendBufSize(fd int, bufSize int) error {
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_SNDBUF, bufSize)
}

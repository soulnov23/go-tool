package netpoll

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
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
}

func SetSocketReUsePort(fd int) error {
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1)
}

func SetSocketTCPNodelay(fd int) error {
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.TCP_NODELAY, 1)
}

// cnt保活探测次数
// intvl保活探测间隔时间
// idle连接空闲时间阈值
func SetSocketKeepAlive(fd int, cnt int, intvl int, idle int) error {
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); err != nil {
		return err
	}
	if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPCNT, cnt); err != nil {
		return err
	}
	if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, intvl); err != nil {
		return err
	}
	if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, idle); err != nil {
		return err
	}
	return nil
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

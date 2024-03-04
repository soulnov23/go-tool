package netpoll

import (
	"golang.org/x/sys/unix"
)

func SetSocketBlock(fd int) error {
	return unix.SetNonblock(fd, false)
}

func SetSocketNonBlock(fd int) error {
	return unix.SetNonblock(fd, true)
}

func SetSocketReuseaddr(fd int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
}

func SetSocketReUsePort(fd int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
}

func SetSocketTCPNodelay(fd int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.TCP_NODELAY, 1)
}

// cnt保活探测次数
// intvl保活探测间隔时间
// idle连接空闲时间阈值
func SetSocketKeepAlive(fd int, cnt int, intvl int, idle int) error {
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1); err != nil {
		return err
	}
	if err := unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPCNT, cnt); err != nil {
		return err
	}
	if err := unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, intvl); err != nil {
		return err
	}
	if err := unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPIDLE, idle); err != nil {
		return err
	}
	return nil
}

func SetSocketCloseExec(fd int) {
	unix.CloseOnExec(fd)
}

func SetSocketRecvBufSize(fd int, bufSize int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, bufSize)
}

func SetSocketSendBufSize(fd int, bufSize int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_SNDBUF, bufSize)
}

//go:build !386 && !amd64 && linux

package netpoll

// 除x86外内核的epoll_event都不是packed，Data按8字节对齐，都是16字节
type EpollEvent struct {
	Events uint32
	_pad   uint32
	Data   [8]byte // to match amd64
}

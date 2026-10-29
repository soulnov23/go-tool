//go:build (386 || amd64) && linux

package netpoll

// x86_64内核把epoll_event声明成packed，i386上u64只按4字节对齐，都是12字节，Data偏移4
type EpollEvent struct {
	Events uint32
	Data   [8]byte // unaligned uintptr
}

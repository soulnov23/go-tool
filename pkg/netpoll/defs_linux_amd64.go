//go:build amd64 && linux

package netpoll

type EpollEvent struct {
	Events uint32
	Data   [8]byte // unaligned uintptr
}

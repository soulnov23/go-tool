//go:build 386 && linux

package netpoll

type EpollEvent struct {
	Events uint32
	Data   [8]byte // to match amd64
}

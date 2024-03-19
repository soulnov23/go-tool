package netpoll

import _ "unsafe"

//go:linkname MaxListenerBacklog net.maxListenerBacklog
func MaxListenerBacklog() int

package netpoll

import (
	"fmt"
	"net"

	"golang.org/x/sys/unix"
)

func ResolveAddr(network, address string) (net.Addr, error) {
	var addr net.Addr
	var err error
	switch network {
	case "tcp", "tcp4":
		if addr, err = net.ResolveTCPAddr(network, address); err != nil {
			return nil, fmt.Errorf("net.ResolveTCPAddr network[%s] address[%s]: %v", network, address, err)
		}
	case "udp", "udp4":
		if addr, err = net.ResolveUDPAddr(network, address); err != nil {
			return nil, fmt.Errorf("net.ResolveUDPAddr network[%s] address[%s]: %v", network, address, err)
		}
	case "unix", "unixgram", "unixpacket":
		if addr, err = net.ResolveUnixAddr(network, address); err != nil {
			return nil, fmt.Errorf("net.ResolveUnixAddr network[%s] address[%s]: %v", network, address, err)
		}
	default:
		return nil, fmt.Errorf("network[%s] not support", network)
	}
	return addr, nil
}

func ResolveSockaddr(network, address string) (unix.Sockaddr, error) {
	switch network {
	case "tcp", "tcp4":
		addr, err := net.ResolveTCPAddr(network, address)
		if err != nil {
			return nil, fmt.Errorf("net.ResolveTCPAddr network[%s] address[%s]: %v", network, address, err)
		}
		return AddrToSockaddrInet4(addr)
	case "udp", "udp4":
		addr, err := net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, fmt.Errorf("net.ResolveUDPAddr network[%s] address[%s]: %v", network, address, err)
		}
		return AddrToSockaddrInet4(addr)
	case "unix", "unixgram", "unixpacket":
		addr, err := net.ResolveUnixAddr(network, address)
		if err != nil {
			return nil, fmt.Errorf("net.ResolveUnixAddr network[%s] address[%s]: %v", network, address, err)
		}
		sau := &unix.SockaddrUnix{
			Name: addr.Name,
		}
		return sau, nil
	default:
		return nil, fmt.Errorf("network[%s] not support", network)
	}
}

func AddrToSockaddrInet4(addr net.Addr) (unix.Sockaddr, error) {
	switch a := addr.(type) {
	case *net.TCPAddr:
		sa4 := &unix.SockaddrInet4{
			Port: a.Port,
		}
		copy(sa4.Addr[:], a.IP.To4())
		return sa4, nil
	case *net.UDPAddr:
		sa4 := &unix.SockaddrInet4{
			Port: a.Port,
		}
		copy(sa4.Addr[:], a.IP.To4())
		return sa4, nil
	default:
		return nil, fmt.Errorf("addr[%T] not support", addr)
	}
}

func SockaddrToAddr(network string, sockaddr unix.Sockaddr) (net.Addr, error) {
	switch sa := sockaddr.(type) {
	case *unix.SockaddrInet4:
		return SockaddrInet4ToAddr(network, sa)
	case *unix.SockaddrUnix:
		return &net.UnixAddr{Name: sa.Name, Net: network}, nil
	default:
		return nil, fmt.Errorf("sockaddr[%T] not support", sockaddr)
	}
}

func SockaddrInet4ToAddr(network string, sa4 *unix.SockaddrInet4) (net.Addr, error) {
	switch network {
	case "tcp", "tcp4":
		return &net.TCPAddr{IP: net.IP(sa4.Addr[:]), Port: sa4.Port}, nil
	case "udp", "udp4":
		return &net.UDPAddr{IP: net.IP(sa4.Addr[:]), Port: sa4.Port}, nil
	default:
		return nil, fmt.Errorf("network[%s] not support", network)
	}
}

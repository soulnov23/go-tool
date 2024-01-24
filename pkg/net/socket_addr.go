package net

import (
	"errors"
	"fmt"
	"net"
	"syscall"
)

func ResolveAddr(network string, address string) (net.Addr, error) {
	var addr net.Addr
	var err error
	switch network {
	case "tcp", "tcp4", "tcp6":
		if addr, err = net.ResolveTCPAddr(network, address); err != nil {
			return nil, fmt.Errorf("net.ResolveTCPAddr: ", err)
		}
	case "udp", "udp4", "udp6":
		if addr, err = net.ResolveUDPAddr(network, address); err != nil {
			return nil, fmt.Errorf("net.ResolveUDPAddr: ", err)
		}
	case "unix", "unixgram", "unixpacket":
		if addr, err = net.ResolveUnixAddr(network, address); err != nil {
			return nil, fmt.Errorf("net.ResolveUnixAddr: ", err)
		}
	case "ip", "ip4", "ip6":
		if addr, err = net.ResolveIPAddr(network, address); err != nil {
			return nil, fmt.Errorf("net.ResolveIPAddr: ", err)
		}
	default:
		return nil, errors.New("network not support")
	}
	return addr, nil
}

func ResolveSockaddr(network string, address string) (syscall.Sockaddr, error) {
	switch network {
	case "tcp", "tcp4":
		addr, err := net.ResolveTCPAddr(network, address)
		if err != nil {
			return nil, fmt.Errorf("net.ResolveTCPAddr: %v", err)
		}
		sa4 := &syscall.SockaddrInet4{
			Port: addr.Port,
		}
		copy(sa4.Addr[:], addr.IP.To4())
		return sa4, nil
	case "tcp6":
		addr, err := net.ResolveTCPAddr(network, address)
		if err != nil {
			return nil, fmt.Errorf("net.ResolveTCPAddr: %v", err)
		}
		sa6 := &syscall.SockaddrInet6{
			Port: addr.Port,
		}
		copy(sa6.Addr[:], addr.IP.To16())
		if addr.Zone != "" {
			intf, err := net.InterfaceByName(addr.Zone)
			if err != nil {
				return nil, fmt.Errorf("net.InterfaceByName: %v", err)
			}
			sa6.ZoneId = uint32(intf.Index)
		}
		return sa6, nil
	case "udp", "udp4":
		addr, err := net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, fmt.Errorf("net.ResolveUDPAddr: %v", err)
		}
		sa4 := &syscall.SockaddrInet4{
			Port: addr.Port,
		}
		copy(sa4.Addr[:], addr.IP.To4())
		return sa4, nil
	case "udp6":
		addr, err := net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, fmt.Errorf("net.ResolveUDPAddr: %v", err)
		}
		sa6 := &syscall.SockaddrInet6{
			Port: addr.Port,
		}
		copy(sa6.Addr[:], addr.IP.To16())
		if addr.Zone != "" {
			intf, err := net.InterfaceByName(addr.Zone)
			if err != nil {
				return nil, fmt.Errorf("net.InterfaceByName: %v", err)
			}
			sa6.ZoneId = uint32(intf.Index)
		}
		return sa6, nil
	case "unix", "unixgram", "unixpacket":
		addr, err := net.ResolveUnixAddr(network, address)
		if err != nil {
			return nil, fmt.Errorf("net.ResolveUnixAddr: %v", err)
		}
		sau := &syscall.SockaddrUnix{
			Name: addr.Name,
		}
		return sau, nil
	case "ip", "ip4":
		addr, err := net.ResolveIPAddr(network, address)
		if err != nil {
			return nil, fmt.Errorf("net.ResolveIPAddr: %v", err)
		}
		sa4 := &syscall.SockaddrInet4{}
		copy(sa4.Addr[:], addr.IP.To4())
		return sa4, nil
	case "ip6":
		addr, err := net.ResolveIPAddr(network, address)
		if err != nil {
			return nil, fmt.Errorf("net.ResolveIPAddr: %v", err)
		}
		sa6 := &syscall.SockaddrInet6{}
		copy(sa6.Addr[:], addr.IP.To16())
		if addr.Zone != "" {
			intf, err := net.InterfaceByName(addr.Zone)
			if err != nil {
				return nil, fmt.Errorf("net.InterfaceByName: %v", err)
			}
			sa6.ZoneId = uint32(intf.Index)
		}
		return sa6, nil
	default:
		return nil, errors.New("network not support")
	}
}

func SockaddrToAddr(sockaddr syscall.Sockaddr) (net.Addr, error) {
	switch sa := sockaddr.(type) {
	case *syscall.SockaddrInet4:
		return &net.TCPAddr{IP: net.IPv4(sa.Addr[0], sa.Addr[1], sa.Addr[2], sa.Addr[3]), Port: sa.Port}, nil
	case *syscall.SockaddrInet6:
		intf, err := net.InterfaceByIndex(int(sa.ZoneId))
		if err != nil {
			return nil, fmt.Errorf("net.InterfaceByIndex: %v", err)
		}
		return &net.TCPAddr{IP: net.IP(sa.Addr[:]), Port: sa.Port, Zone: intf.Name}, nil
	case *syscall.SockaddrUnix:
		return &net.UnixAddr{Name: sa.Name}, nil
	default:
		return nil, errors.New("sockaddr not support")
	}
}

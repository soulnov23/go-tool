package netpoll

import (
	"errors"
	"net"

	"golang.org/x/sys/unix"
)

func ResolveAddr(network string, address string) (net.Addr, error) {
	var addr net.Addr
	var err error
	switch network {
	case "tcp", "tcp4", "tcp6":
		if addr, err = net.ResolveTCPAddr(network, address); err != nil {
			return nil, err
		}
	case "udp", "udp4", "udp6":
		if addr, err = net.ResolveUDPAddr(network, address); err != nil {
			return nil, err
		}
	case "unix", "unixgram", "unixpacket":
		if addr, err = net.ResolveUnixAddr(network, address); err != nil {
			return nil, err
		}
	case "ip", "ip4", "ip6":
		if addr, err = net.ResolveIPAddr(network, address); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("network not support")
	}
	return addr, nil
}

func ResolveSockaddr(network string, address string) (unix.Sockaddr, error) {
	switch network {
	case "tcp", "tcp4":
		addr, err := net.ResolveTCPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return AddrToSockaddrInet4(addr)
	case "tcp6":
		addr, err := net.ResolveTCPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return AddrToSockaddrInet6(addr)
	case "udp", "udp4":
		addr, err := net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return AddrToSockaddrInet4(addr)
	case "udp6":
		addr, err := net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return AddrToSockaddrInet6(addr)
	case "unix", "unixgram", "unixpacket":
		addr, err := net.ResolveUnixAddr(network, address)
		if err != nil {
			return nil, err
		}
		sau := &unix.SockaddrUnix{
			Name: addr.Name,
		}
		return sau, nil
	case "ip", "ip4":
		addr, err := net.ResolveIPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return AddrToSockaddrInet4(addr)
	case "ip6":
		addr, err := net.ResolveIPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return AddrToSockaddrInet6(addr)
	default:
		return nil, errors.New("network not support")
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
	case *net.IPAddr:
		sa4 := &unix.SockaddrInet4{}
		copy(sa4.Addr[:], a.IP.To4())
		return sa4, nil
	default:
		return nil, errors.New("addr not support")
	}
}

func AddrToSockaddrInet6(addr net.Addr) (unix.Sockaddr, error) {
	switch a := addr.(type) {
	case *net.TCPAddr:
		sa6 := &unix.SockaddrInet6{
			Port: a.Port,
		}
		copy(sa6.Addr[:], a.IP.To16())
		if a.Zone != "" {
			intf, err := net.InterfaceByName(a.Zone)
			if err != nil {
				return nil, err
			}
			sa6.ZoneId = uint32(intf.Index)
		}
		return sa6, nil
	case *net.UDPAddr:
		sa6 := &unix.SockaddrInet6{
			Port: a.Port,
		}
		copy(sa6.Addr[:], a.IP.To16())
		if a.Zone != "" {
			intf, err := net.InterfaceByName(a.Zone)
			if err != nil {
				return nil, err
			}
			sa6.ZoneId = uint32(intf.Index)
		}
		return sa6, nil
	case *net.IPAddr:
		sa6 := &unix.SockaddrInet6{}
		copy(sa6.Addr[:], a.IP.To16())
		if a.Zone != "" {
			intf, err := net.InterfaceByName(a.Zone)
			if err != nil {
				return nil, err
			}
			sa6.ZoneId = uint32(intf.Index)
		}
		return sa6, nil
	default:
		return nil, errors.New("addr not support")
	}
}

func SockaddrToAddr(network string, sockaddr unix.Sockaddr) (net.Addr, error) {
	switch sa := sockaddr.(type) {
	case *unix.SockaddrInet4:
		return SockaddrInet4ToAddr(network, sa)
	case *unix.SockaddrInet6:
		return SockaddrInet6ToAddr(network, sa)
	case *unix.SockaddrUnix:
		return &net.UnixAddr{Name: sa.Name}, nil
	default:
		return nil, errors.New("sockaddr not support")
	}
}

func SockaddrInet4ToAddr(network string, sa4 *unix.SockaddrInet4) (net.Addr, error) {
	switch network {
	case "tcp", "tcp4":
		return &net.TCPAddr{IP: net.IP(sa4.Addr[:]), Port: sa4.Port}, nil
	case "udp", "udp4":
		return &net.UDPAddr{IP: net.IP(sa4.Addr[:]), Port: sa4.Port}, nil
	case "ip", "ip4":
		return &net.IPAddr{IP: net.IP(sa4.Addr[:])}, nil
	default:
		return nil, errors.New("network not support")
	}
}

func SockaddrInet6ToAddr(network string, sa6 *unix.SockaddrInet6) (net.Addr, error) {
	switch network {
	case "tcp6":
		intf, err := net.InterfaceByIndex(int(sa6.ZoneId))
		if err != nil {
			return nil, err
		}
		return &net.TCPAddr{IP: net.IP(sa6.Addr[:]), Port: sa6.Port, Zone: intf.Name}, nil
	case "udp6":
		intf, err := net.InterfaceByIndex(int(sa6.ZoneId))
		if err != nil {
			return nil, err
		}
		return &net.UDPAddr{IP: net.IP(sa6.Addr[:]), Port: sa6.Port, Zone: intf.Name}, nil
	case "ip6":
		intf, err := net.InterfaceByIndex(int(sa6.ZoneId))
		if err != nil {
			return nil, err
		}
		return &net.IPAddr{IP: net.IP(sa6.Addr[:]), Zone: intf.Name}, nil
	default:
		return nil, errors.New("network not support")
	}
}

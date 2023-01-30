package net

import (
	"errors"
	"net"
	"net/netip"
	"strconv"
	"syscall"
)

func GetSocketAddr(network string, address string) (syscall.Sockaddr, error) {
	tcpAddr, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		return nil, errors.New("net.ResolveTCPAddr: " + err.Error())
	}

	if network == "tcp" || network == "tcp4" {
		sa4 := &syscall.SockaddrInet4{
			Port: tcpAddr.Port,
		}
		copy(sa4.Addr[:], tcpAddr.IP.To4())
		return sa4, nil
	} else if network == "tcp6" {
		sa6 := &syscall.SockaddrInet6{
			Port: tcpAddr.Port,
		}
		copy(sa6.Addr[:], tcpAddr.IP.To16())
		if tcpAddr.Zone != "" {
			intf, err := net.InterfaceByName(tcpAddr.Zone)
			if err != nil {
				return nil, errors.New("net.InterfaceByName: " + err.Error())
			}
			sa6.ZoneId = uint32(intf.Index)
		}
		return sa6, nil
	} else {
		return nil, errors.New("network " + network + " not support")
	}
}

func GetSocketIP(addr syscall.Sockaddr) (string, error) {
	if sa4, ok := addr.(*syscall.SockaddrInet4); ok {
		return netip.AddrFrom4(sa4.Addr).String() + ":" + strconv.Itoa(sa4.Port), nil
	} else if sa6, ok := addr.(*syscall.SockaddrInet6); ok {
		if sa6.ZoneId != 0 {
			return netip.AddrFrom16(sa6.Addr).String() + "%" + strconv.Itoa(sa6.Port) + ":" + strconv.Itoa(sa6.Port), nil
		}
		return netip.AddrFrom16(sa6.Addr).String() + "" + ":" + strconv.Itoa(sa6.Port), nil
	} else {
		return "", errors.New("addrInet not support")
	}
}

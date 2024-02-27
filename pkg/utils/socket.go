package utils

import (
	"fmt"
	"syscall"
)

func Socket(network string) (int, error) {
	switch network {
	case "tcp", "tcp4":
		fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
		if err != nil {
			return 0, fmt.Errorf("syscall.Socket network[%s]: %v", network, err)
		}
		return fd, nil
	case "tcp6":
		fd, err := syscall.Socket(syscall.AF_INET6, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
		if err != nil {
			return 0, fmt.Errorf("syscall.Socket network[%s]: %v", network, err)
		}
		return fd, nil
	case "udp", "udp4":
		fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
		if err != nil {
			return 0, fmt.Errorf("syscall.Socket network[%s]: %v", network, err)
		}
		return fd, nil
	case "udp6":
		fd, err := syscall.Socket(syscall.AF_INET6, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
		if err != nil {
			return 0, fmt.Errorf("syscall.Socket network[%s]: %v", network, err)
		}
		return fd, nil
	case "unix":
		fd, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
		if err != nil {
			return 0, fmt.Errorf("syscall.Socket network[%s]: %v", network, err)
		}
		return fd, nil
	case "unixgram":
		fd, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_DGRAM, 0)
		if err != nil {
			return 0, fmt.Errorf("syscall.Socket network[%s]: %v", network, err)
		}
		return fd, nil
	case "unixpacket":
		fd, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_SEQPACKET, 0)
		if err != nil {
			return 0, fmt.Errorf("syscall.Socket network[%s]: %v", network, err)
		}
		return fd, nil
	default:
		return 0, fmt.Errorf("network[%s] not support", network)
	}
}

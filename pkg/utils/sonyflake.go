package utils

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/sony/sonyflake"
)

var sf *sonyflake.Sonyflake

func init() {
	var err error
	sf, err = sonyflake.New(sonyflake.Settings{
		StartTime: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		MachineID: func() (uint16, error) {
			as, err := net.InterfaceAddrs()
			if err != nil {
				return 0, err
			}
			for _, a := range as {
				ipnet, ok := a.(*net.IPNet)
				if !ok || ipnet.IP.IsLoopback() {
					continue
				}
				ip := ipnet.IP.To4()
				if ip == nil {
					continue
				}
				return uint16(ip[2])<<8 + uint16(ip[3]), nil
			}
			return 0, errors.New("init sonyflake: not found ip address")
		},
	})
	if err != nil {
		panic(err)
	}
}

func GenerateID() (string, error) {
	nextID, err := sf.NextID()
	if err != nil {
		return "", fmt.Errorf("generate snowflake id: %w", err)
	}
	return time.Now().Format("20060102150405") + strconv.FormatUint(nextID, 10), nil
}

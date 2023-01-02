// Package main 客户端测试程序
package main

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/SoulNov23/go-tool/pkg/unsafe"
)

func TestAccept(t *testing.T) {
	close := make(chan struct{})
	timeout := 30 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				close <- struct{}{}
				return
			default:
				conn, err := net.DialTimeout("tcp", "0.0.0.0:8000", timeout)
				if err != nil {
					t.Logf("net.DialTimeout: %s", err.Error())
				}
				conn.Close()

				conn, err = net.DialTimeout("tcp", "0.0.0.0:8080", timeout)
				if err != nil {
					t.Logf("net.DialTimeout: %s", err.Error())
				}
				conn.Close()
			}
		}
	}(ctx)

	time.Sleep(timeout)
	cancel()
	<-close
}

func TestRead(t *testing.T) {
	close := make(chan struct{})
	timeout := 30 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	tcpConn, err := net.DialTimeout("tcp", "0.0.0.0:8000", timeout)
	if err != nil {
		t.Errorf("net.DialTimeout: %s", err.Error())
	}

	httpConn, err := net.DialTimeout("tcp", "0.0.0.0:8080", timeout)
	if err != nil {
		t.Errorf("net.DialTimeout: %s", err.Error())
	}

	buf := unsafe.String2Byte("hello world")

	go func(ctx context.Context, tcpConn net.Conn, httpConn net.Conn) {
		for {
			select {
			case <-ctx.Done():
				close <- struct{}{}
				return
			default:
				tcpConn.Write(buf)
				httpConn.Write(buf)
			}
		}
	}(ctx, tcpConn, httpConn)

	time.Sleep(timeout)
	cancel()
	<-close
}

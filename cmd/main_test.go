// Package main 客户端测试程序
package main

import (
	"context"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/SoulNov23/go-tool/pkg/unsafe"
)

func TestEchoConnect(t *testing.T) {
	close := make(chan struct{})
	timeout := 30 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < runtime.NumCPU(); i++ {
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					close <- struct{}{}
					return
				default:
					conn, err := net.DialTimeout("tcp", "0.0.0.0:8000", 10*time.Millisecond)
					if err != nil {
						t.Logf("net.DialTimeout: %s", err.Error())
					} else {
						conn.Close()
					}

					conn, err = net.DialTimeout("tcp", "0.0.0.0:8080", 10*time.Millisecond)
					if err != nil {
						t.Logf("net.DialTimeout: %s", err.Error())
					} else {
						conn.Close()
					}
				}
			}
		}(ctx)
	}

	time.Sleep(timeout)
	cancel()
	<-close
}

func TestKeepConnect(t *testing.T) {
	timeout := 30 * time.Second

	for i := 0; i < 100000; i++ {
		_, err := net.DialTimeout("tcp", "0.0.0.0:8000", 1000*time.Millisecond)
		if err != nil {
			t.Logf("net.DialTimeout: %s", err.Error())
		}

		_, err = net.DialTimeout("tcp", "0.0.0.0:8080", 1000*time.Millisecond)
		if err != nil {
			t.Logf("net.DialTimeout: %s", err.Error())
		}
	}

	time.Sleep(timeout)
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

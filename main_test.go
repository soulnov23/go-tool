// Package main 客户端测试程序
package main

import (
	"context"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/soulnov23/go-tool/pkg/utils"
)

func TestConnect(t *testing.T) {
	timeout := 3 * time.Second

	a, err := net.DialTimeout("tcp", "0.0.0.0:6666", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("net.DialTimeout: %s", err.Error())
	}

	b, err := net.DialTimeout("tcp", "0.0.0.0:8888", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("net.DialTimeout: %s", err.Error())
	}

	time.Sleep(timeout)

	a.Close()
	b.Close()
}

func TestConcurrentConnect(t *testing.T) {
	var conns []net.Conn
	timeout := 10 * time.Second

	for i := 0; i < 100000; i++ {
		conn, err := net.DialTimeout("tcp", "0.0.0.0:6666", 100*time.Millisecond)
		if err != nil {
			t.Errorf("net.DialTimeout: %s", err.Error())
		} else {
			conns = append(conns, conn)
		}

		conn, err = net.DialTimeout("tcp", "0.0.0.0:8888", 100*time.Millisecond)
		if err != nil {
			t.Errorf("net.DialTimeout: %s", err.Error())
		} else {
			conns = append(conns, conn)
		}
	}

	time.Sleep(timeout)

	for _, conn := range conns {
		if err := conn.Close(); err != nil {
			t.Errorf("net.Conn.Close: %s", err.Error())
		}
	}
}

func TestConcurrentConnectAndClose(t *testing.T) {
	close := make(chan struct{})
	timeout := 10 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < runtime.NumCPU(); i++ {
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					close <- struct{}{}
					return
				default:
					conn, err := net.DialTimeout("tcp", "0.0.0.0:6666", 100*time.Millisecond)
					if err != nil {
						t.Logf("net.DialTimeout: %s", err.Error())
					} else {
						conn.Close()
					}

					conn, err = net.DialTimeout("tcp", "0.0.0.0:8888", 100*time.Millisecond)
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

func TestRead(t *testing.T) {
	tcpConn, err := net.DialTimeout("tcp", "0.0.0.0:6666", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("net.DialTimeout: %s", err.Error())
	}

	httpConn, err := net.DialTimeout("tcp", "0.0.0.0:8888", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("net.DialTimeout: %s", err.Error())
	}

	buf := utils.StringToBytes("hello world")
	tcpConn.Write(buf)
	httpConn.Write(buf)
	tcpConn.Close()
	httpConn.Close()
}

func TestConcurrentRead(t *testing.T) {
	close := make(chan struct{})
	timeout := 10 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	tcpConn, err := net.DialTimeout("tcp", "0.0.0.0:6666", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("net.DialTimeout: %s", err.Error())
	}

	httpConn, err := net.DialTimeout("tcp", "0.0.0.0:8888", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("net.DialTimeout: %s", err.Error())
	}

	buf := utils.StringToBytes("hello world")

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

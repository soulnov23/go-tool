package net

import (
	"syscall"

	"github.com/SoulNov23/go-tool/pkg/buffer"
	"github.com/SoulNov23/go-tool/pkg/cache"
	"github.com/SoulNov23/go-tool/pkg/log"
	"github.com/SoulNov23/go-tool/pkg/unsafe"
)

type TcpConn struct {
	log         log.Logger
	epollFD     int
	fd          int
	localAddr   string
	remoteAddr  string
	readBuffer  *buffer.LinkedBuffer
	writeBuffer *buffer.LinkedBuffer
}

func (conn *TcpConn) Peek(size int) ([]byte, error) {
	return conn.readBuffer.Peek(size)
}

func (conn *TcpConn) Skip(size int) error {
	return conn.readBuffer.Skip(size)
}

func (conn *TcpConn) Next(size int) ([]byte, error) {
	return conn.readBuffer.Next(size)
}

func (conn *TcpConn) Read() {
	buf := cache.New(buffer.Block8k)
	offset := 0
	for {
		n, err := syscall.Read(conn.fd, buf[offset:])
		if err != nil {
			if err == syscall.EAGAIN {
				break
			} else if err == syscall.EINTR {
				continue
			} else {
				conn.log.Errorf("syscall.Read: %v", err)
				continue
			}
		}
		offset += n
		if n == 0 || offset == 1024 {
			break
		}
	}
	conn.log.Debugf("read: %s", unsafe.Byte2String(buf[:]))
	conn.readBuffer.Append(buf)
}

func (conn *TcpConn) Write(buf []byte) {
	offset := 0
	for {
		n, err := syscall.Write(conn.fd, buf[offset:])
		if err != nil {
			if err == syscall.EAGAIN {
				if err := Control(conn.epollFD, conn.fd, ModReadWritable); err != nil {
					conn.log.Errorf("net.Control: " + err.Error())
				}
				conn.writeBuffer.Append(buf)
				break
			} else if err == syscall.EINTR {
				continue
			} else {
				conn.log.Errorf("syscall.Write: %v", err)
				continue
			}
		}
		offset += n
		if n == 0 || offset == len(buf) {
			break
		}
	}
	conn.log.Debugf("write: %s", unsafe.Byte2String(buf)[:offset])
}

func (conn *TcpConn) ReWrite() {
	if conn.writeBuffer.Len() == 0 {
		return
	}
	buf, err := conn.Peek(conn.writeBuffer.Len())
	if err != nil {
		conn.log.Errorf("TcpConn.Peek: " + err.Error())
	}
	conn.Write(buf)
}

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

func NewTcpConn(log log.Logger, epollFD int, fd int, localAddr string, remoteAddr string) *TcpConn {
	return &TcpConn{
		log:         log,
		epollFD:     epollFD,
		fd:          fd,
		localAddr:   localAddr,
		remoteAddr:  remoteAddr,
		readBuffer:  buffer.NewBuffer(),
		writeBuffer: buffer.NewBuffer(),
	}
}

func DeleteTcpConn(conn *TcpConn) {
	conn.log, conn.epollFD, conn.localAddr, conn.remoteAddr = nil, -1, "", ""
	Control(conn.epollFD, conn.fd, Detach)
	syscall.Close(conn.fd)
	conn.fd = -1
	buffer.DeleteBuffer(conn.readBuffer)
	buffer.DeleteBuffer(conn.writeBuffer)
}

func (conn *TcpConn) Peek(size int) ([]byte, error) {
	return conn.readBuffer.Peek(size)
}

func (conn *TcpConn) Skip(size int) error {
	return conn.readBuffer.Skip(size)
}

func (conn *TcpConn) Read(size int) ([]byte, error) {
	return conn.readBuffer.Read(size)
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
				conn.writeBuffer.Write(buf)
				break
			} else if err == syscall.EINTR {
				continue
			} else {
				conn.log.Errorf("syscall.Write: %v", err)
				break
			}
		}
		offset += n
		if n == 0 || offset == len(buf) {
			break
		}
	}
	conn.log.Debugf("write: %s", unsafe.Byte2String(buf)[:offset])
}

func (conn *TcpConn) handlerRead() {
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
				break
			}
		}
		offset += n
		if n == 0 || offset == 1024 {
			break
		}
	}
	conn.log.Debugf("read: %s", unsafe.Byte2String(buf[:]))
	conn.readBuffer.Write(buf)
	cache.Delete(buf)
}

func (conn *TcpConn) handlerWrite() {
	if conn.writeBuffer.Len() == 0 {
		return
	}
	buf, err := conn.writeBuffer.Peek(conn.writeBuffer.Len())
	if err != nil {
		conn.log.Errorf("TcpConn.writeBuffer.Peek: " + err.Error())
		// 发送失败，再触发一次EPOLLOUT
		if err := Control(conn.epollFD, conn.fd, ModReadWritable); err != nil {
			conn.log.Errorf("net.Control: " + err.Error())
		}
		return
	}
	conn.writeBuffer.Skip(cap(buf))
	conn.writeBuffer.GC()
	conn.Write(buf)
}

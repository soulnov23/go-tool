package net

import (
	"syscall"

	"github.com/SoulNov23/go-tool/pkg/buffer"
	"github.com/SoulNov23/go-tool/pkg/cache"
	"github.com/SoulNov23/go-tool/pkg/log"
	"github.com/SoulNov23/go-tool/pkg/unsafe"
)

type Operator interface {
	OnAccept(conn *TcpConn)
	OnClose(conn *TcpConn)
	OnRead(conn *TcpConn)
}

type TcpConn struct {
	log         log.Logger
	epollFD     int
	fd          int
	localAddr   string
	remoteAddr  string
	readBuffer  *buffer.LinkedBuffer
	writeBuffer *buffer.LinkedBuffer
	operator    Operator
}

func NewTcpConn(log log.Logger, epollFD int, fd int, localAddr string, remoteAddr string, operator Operator) *TcpConn {
	return &TcpConn{
		log:         log,
		epollFD:     epollFD,
		fd:          fd,
		localAddr:   localAddr,
		remoteAddr:  remoteAddr,
		readBuffer:  buffer.NewBuffer(),
		writeBuffer: buffer.NewBuffer(),
		operator:    operator,
	}
}

func DeleteTcpConn(conn *TcpConn) {
	conn.log, conn.epollFD, conn.localAddr, conn.remoteAddr = nil, -1, "", ""
	Control(conn.epollFD, conn.fd, Detach)
	syscall.Close(conn.fd)
	buffer.DeleteBuffer(conn.readBuffer)
	buffer.DeleteBuffer(conn.writeBuffer)
}

func (conn *TcpConn) LocalAddr() string {
	return conn.localAddr
}

func (conn *TcpConn) RemoteAddr() string {
	return conn.remoteAddr
}

func (conn *TcpConn) ReadBufferLen() uint64 {
	return conn.readBuffer.Len()
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
		if n == 0 || offset == buffer.Block8k {
			break
		}
	}
	conn.log.Debugf("read: %s", unsafe.Byte2String(buf[:]))
	conn.readBuffer.Write(buf)
	cache.Delete(buf)
}

func (conn *TcpConn) handlerWrite() {
	buf, err := conn.writeBuffer.Peek(int(conn.writeBuffer.Len()))
	if err != nil {
		// 数据发送完了返回err
		conn.log.Errorf("TcpConn.writeBuffer.Peek: " + err.Error())
		return
	}
	conn.writeBuffer.Skip(cap(buf))
	conn.writeBuffer.GC()
	conn.Write(buf)
}

package net

import (
	"syscall"

	"github.com/soulnov23/go-tool/pkg/buffer"
	"github.com/soulnov23/go-tool/pkg/cache"
	"github.com/soulnov23/go-tool/pkg/log"
	"github.com/soulnov23/go-tool/pkg/utils"
	"go.uber.org/zap"
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
					conn.log.ErrorFields("epoll control client fd", zap.Error(err), zap.Int("epoll_fd", conn.epollFD), zap.Int("client_fd", conn.fd), zap.String("epoll_event", EventString(ModReadWritable)))
					break
				}
				conn.writeBuffer.Write(buf)
				break
			} else if err == syscall.EINTR {
				continue
			} else {
				conn.log.ErrorFields("write client fd", zap.Error(err), zap.Int("epoll_fd", conn.epollFD), zap.Int("client_fd", conn.fd))
				break
			}
		}
		offset += n
		if n == 0 || offset == len(buf) {
			break
		}
	}
	conn.log.DebugFields("write success", zap.String("msg", utils.Byte2String(buf[:offset])), zap.Int("epoll_fd", conn.epollFD), zap.Int("client_fd", conn.fd))
}

func (conn *TcpConn) handlerRead() {
	conn.readBuffer.GC()
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
				conn.log.ErrorFields("read client fd", zap.Error(err), zap.Int("epoll_fd", conn.epollFD), zap.Int("client_fd", conn.fd))
				break
			}
		}
		offset += n
		if n == 0 || offset == buffer.Block8k {
			break
		}
	}
	conn.log.DebugFields("read success", zap.String("msg", utils.Byte2String(buf[:offset])), zap.Int("epoll_fd", conn.epollFD), zap.Int("client_fd", conn.fd))
	conn.readBuffer.Write(buf[:offset])
}

func (conn *TcpConn) handlerWrite() {
	buf, err := conn.writeBuffer.Peek(int(conn.writeBuffer.Len()))
	if err != nil {
		// 数据发送完了返回err
		conn.log.ErrorFields("peek write buffer", zap.Error(err), zap.Int("epoll_fd", conn.epollFD), zap.Int("client_fd", conn.fd))
		return
	}
	conn.writeBuffer.Skip(cap(buf))
	conn.writeBuffer.GC()
	conn.Write(buf)
}

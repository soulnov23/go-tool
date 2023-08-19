package net

import (
	"syscall"

	"github.com/soulnov23/go-tool/pkg/buffer"
	"github.com/soulnov23/go-tool/pkg/cache"
	"github.com/soulnov23/go-tool/pkg/log"
	"go.uber.org/zap"
)

type Operator interface {
	OnRead(conn *TcpConn)
}

type TcpConn struct {
	log         log.Logger
	epollFD     int
	fd          int
	localAddr   string
	remoteAddr  string
	readBuffer  *buffer.Buffer
	writeBuffer *buffer.Buffer
	operator    Operator
}

func NewTcpConn(log log.Logger, epollFD int, fd int, localAddr string, remoteAddr string, operator Operator) *TcpConn {
	return &TcpConn{
		log:         log,
		epollFD:     epollFD,
		fd:          fd,
		localAddr:   localAddr,
		remoteAddr:  remoteAddr,
		readBuffer:  buffer.New(),
		writeBuffer: buffer.New(),
		operator:    operator,
	}
}

func DeleteTcpConn(conn *TcpConn) {
	conn.log, conn.epollFD, conn.localAddr, conn.remoteAddr = nil, -1, "", ""
	Control(conn.epollFD, conn.fd, Detach)
	syscall.Close(conn.fd)
	conn.readBuffer.Close()
	conn.writeBuffer.Close()
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
			if err == syscall.EINTR /*中断信号触发系统调用中断直接忽略继续读取*/ {
				continue
			} else if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK /*非阻塞IO没有数据可读时直接返回等待OUT事件再次触发，不打印了不然日志太多*/ {
				if err := Control(conn.epollFD, conn.fd, ModReadWritable); err != nil {
					conn.log.ErrorFields("epoll control client fd", zap.Error(err), zap.Int("epoll_fd", conn.epollFD), zap.Int("client_fd", conn.fd), zap.String("epoll_event", EventString(ModReadWritable)))
					break
				}
				conn.writeBuffer.Write(buf)
				break
			} else if err == syscall.EBADF || err == syscall.EINVAL /*fd被关闭已经是无效的文件描述符，在epoll事件模型中把HUP放最前面了，这里不会发生*/ {
				goto ERROR
			} else if err == syscall.EPIPE /*broken pipe在write进行中对端意外关闭连接，TCP发起RST报文，触发SIGPIPE信号*/ {
				goto ERROR
			} else {
				goto ERROR
			}
		ERROR:
			conn.log.ErrorFields("write client fd", zap.Error(err), zap.Int("epoll_fd", conn.epollFD), zap.Int("client_fd", conn.fd))
			break
		}
		offset += n
		if offset == len(buf) /*buf全部写进去了*/ {
			break
		}
	}
	conn.log.DebugFields("write success", zap.ByteString("buffer", buf[:offset]), zap.Int("epoll_fd", conn.epollFD), zap.Int("client_fd", conn.fd))
}

func (conn *TcpConn) handlerRead() {
	conn.readBuffer.GC()
	buf := cache.New(buffer.Block8k)
	offset := 0
	for {
		n, err := syscall.Read(conn.fd, buf[offset:])
		if err != nil {
			if err == syscall.EINTR /*中断信号触发系统调用中断直接忽略继续读取*/ {
				continue
			} else if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK /*非阻塞IO没有数据可读时直接返回等待OUT事件再次触发，不打印了不然日志太多*/ {
				break
			} else if err == syscall.EBADF || err == syscall.EINVAL /*fd被关闭已经是无效的文件描述符，在epoll事件模型中把HUP放最前面了，这里不会发生*/ {
				goto ERROR
			} else if err == syscall.ECONNRESET /*connection reset by peer在read进行中对端意外关闭连接，TCP发起RST报文*/ {
				goto ERROR
			} else {
				goto ERROR
			}
		ERROR:
			conn.log.ErrorFields("read client fd", zap.Error(err), zap.Int("epoll_fd", conn.epollFD), zap.Int("client_fd", conn.fd))
			break
		}
		offset += n
		if n == 0 /*在read进行中对端主动关闭连接调用了Close，TCP发起FIN报文*/ || offset == buffer.Block8k /*读取8k就走避免饥饿连接*/ {
			break
		}
	}
	conn.log.DebugFields("read success", zap.ByteString("buffer", buf[:offset]), zap.Int("epoll_fd", conn.epollFD), zap.Int("client_fd", conn.fd))
	conn.readBuffer.Write(buf[:offset])
}

func (conn *TcpConn) handlerWrite() {
	buf, err := conn.writeBuffer.Peek(int(conn.writeBuffer.Len()))
	if err != nil {
		// 数据发送完了返回err
		conn.log.ErrorFields("peek write buffer", zap.Error(err), zap.Int("epoll_fd", conn.epollFD), zap.Int("client_fd", conn.fd))
		return
	}
	size := len(buf)
	conn.writeBuffer.Skip(size)
	conn.writeBuffer.GC()
	conn.Write(buf[:size])
}

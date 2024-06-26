package transport

import (
	"fmt"
	"net"
	"runtime"

	"github.com/soulnov23/go-tool/pkg/buffer"
	"github.com/soulnov23/go-tool/pkg/cache"
	"github.com/soulnov23/go-tool/pkg/framework/log"
	"github.com/soulnov23/go-tool/pkg/netpoll"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

func init() {
	RegisterServerTransportFunc("tcp", newServerTransportTCP)
	RegisterServerTransportFunc("tcp4", newServerTransportTCP)
	RegisterServerTransportFunc("tcp6", newServerTransportTCP)
}

type serverTransportTCP struct {
	address       string
	network       string
	protocol      string
	epolls        []*netpoll.Epoll
	localAddr     net.Addr
	localSockAddr unix.Sockaddr
	opts          *ServerTransportOptions
}

func newServerTransportTCP(address, network, protocol string, opts ...ServerTransportOption) ServerTransport {
	transport := &serverTransportTCP{
		address:  address,
		network:  network,
		protocol: protocol,
		opts: &ServerTransportOptions{
			coreSize: runtime.GOMAXPROCS(0),
		},
	}
	for _, opt := range opts {
		opt(transport.opts)
	}
	return transport
}

func (t *serverTransportTCP) ListenAndServe() error {
	for i := 0; i < t.opts.coreSize; i++ {
		epoll, err := netpoll.NewEpoll(log.DefaultLogger.InfoFields)
		if err != nil {
			return fmt.Errorf("netpoll.NewEpoll: %v", err)
		}

		addr, err := netpoll.ResolveAddr(t.network, t.address)
		if err != nil {
			return fmt.Errorf("netpoll.ResolveAddr: %v", err)
		}
		t.localAddr = addr

		sockaddr, err := netpoll.ResolveSockaddr(t.network, t.address)
		if err != nil {
			return fmt.Errorf("netpoll.ResolveSockaddr: %v", err)
		}
		t.localSockAddr = sockaddr

		listenFD, err := netpoll.Socket(t.network)
		if err != nil {
			return fmt.Errorf("netpoll.Socket[%s]: %v", t.network, err)
		}
		if err := netpoll.SetSocketReuseaddr(listenFD); err != nil {
			unix.Close(listenFD)
			return fmt.Errorf("netpoll.SetSocketReuseaddr[%d]: %v", listenFD, err)
		}
		if err := netpoll.SetSocketReUsePort(listenFD); err != nil {
			unix.Close(listenFD)
			return fmt.Errorf("netpoll.SetSocketReUsePort[%d]: %v", listenFD, err)
		}
		if err := unix.Bind(listenFD, t.localSockAddr); err != nil {
			unix.Close(listenFD)
			return fmt.Errorf("unix.Bind[%d] address[%s] network[%s]: %v", listenFD, t.network, t.address, err)
		}
		backlog := netpoll.MaxListenerBacklog()
		if err := unix.Listen(listenFD, backlog); err != nil {
			unix.Close(listenFD)
			return fmt.Errorf("unix.Listen[%d] backlog[%d]: %v", listenFD, backlog, err)
		}
		log.DefaultLogger.InfoFields("listen success", zap.Int("epoll_fd", epoll.FD()), zap.Int("listen_fd", listenFD), zap.String("network", t.network), zap.String("address", t.address))

		operator := epoll.Alloc()
		operator.FD = listenFD
		operator.Epoll = epoll
		operator.OnRead = t.accept
		if err := epoll.Control(operator, netpoll.ReadWritable); err != nil {
			unix.Close(listenFD)
			return fmt.Errorf("epoll_fd[%d] epoll.Control listen_fd[%d]: %v", epoll.FD(), listenFD, err)
		}
		go func() {
			if err := epoll.Wait(); err != nil {
				log.DefaultLogger.FatalFields("epoll.Wait", zap.Error(err), zap.Reflect("service_transport", t))
				panic(fmt.Sprintf("epoll.Wait: %v", err))
			}
		}()
	}
	return nil
}

func (t *serverTransportTCP) accept(epoll *netpoll.Epoll, operator *netpoll.FDOperator) {
	for {
		clientFD, addr, err := unix.Accept4(operator.FD, unix.SOCK_CLOEXEC)
		if err != nil {
			if err == unix.EAGAIN {
				break
			} else if err == unix.EINTR {
				continue
			} else {
				log.DefaultLogger.ErrorFields("accept client", zap.Error(err), zap.Int("listen_fd", operator.FD))
				continue
			}
		}
		if err := netpoll.SetSocketNonBlock(clientFD); err != nil {
			unix.Close(clientFD)
			log.DefaultLogger.ErrorFields("netpoll.SetSocketNonBlock", zap.Error(err), zap.Int("client_fd", clientFD))
			continue
		}
		netpoll.SetSocketCloseExec(clientFD)
		if err := netpoll.SetSocketTCPNodelay(clientFD); err != nil {
			unix.Close(clientFD)
			log.DefaultLogger.ErrorFields("netpoll.SetSocketTCPNodelay", zap.Error(err), zap.Int("client_fd", clientFD))
			continue
		}
		remoteAddr, err := netpoll.SockaddrToAddr(t.network, addr)
		if err != nil {
			log.DefaultLogger.ErrorFields("netpoll.SockaddrToAddr", zap.Error(err), zap.Reflect("sockaddr", addr))
			continue
		}
		clientOperator := epoll.Alloc()
		clientOperator.FD = clientFD
		clientOperator.Epoll = epoll
		clientOperator.OnRead = t.read
		clientOperator.OnWrite = t.write
		clientOperator.OnHup = t.hup
		clientOperator.Data = &tcpConnection{
			fd:          clientFD,
			localAddr:   t.localAddr,
			remoteAddr:  remoteAddr,
			readBuffer:  buffer.New(),
			writeBuffer: buffer.New(),
		}
		if err := operator.Epoll.Control(clientOperator, netpoll.Readable); err != nil {
			unix.Close(clientFD)
			log.DefaultLogger.ErrorFields("epoll.Control", zap.Error(err), zap.Int("epoll_fd", epoll.FD()), zap.Int("client_fd", clientFD), zap.String("epoll_event", netpoll.EventString(netpoll.Readable)))
			continue
		}
		log.DefaultLogger.InfoFields("accept success", zap.Int("epoll_fd", epoll.FD()), zap.Int("listen_fd", operator.FD), zap.Int("client_fd", clientOperator.FD), zap.String("remote_address", remoteAddr.String()), zap.String("local_address", t.localAddr.String()))
	}
}

func (t *serverTransportTCP) read(epoll *netpoll.Epoll, operator *netpoll.FDOperator) {
	tcpConn, ok := operator.Data.(*tcpConnection)
	if !ok || tcpConn == nil {
		log.DefaultLogger.ErrorFields("data is not tcpConnection", zap.Reflect("operator", operator))
		return
	}
	tcpConn.readBuffer.GC()
	buf := cache.New(buffer.Block8k)
	offset := 0
	for {
		n, err := unix.Read(tcpConn.fd, buf[offset:])
		if err != nil {
			if err == unix.EINTR /*中断信号触发系统调用中断直接忽略继续读取*/ {
				continue
			} else if err == unix.EAGAIN || err == unix.EWOULDBLOCK /*非阻塞IO没有数据可读时直接返回等待OUT事件再次触发，不打印了不然日志太多*/ {
				break
			} else if err == unix.EBADF || err == unix.EINVAL /*fd被关闭已经是无效的文件描述符，在epoll事件模型中把HUP放最前面了，这里不会发生*/ {
				goto ERROR
			} else if err == unix.ECONNRESET /*connection reset by peer在read进行中对端意外关闭连接，TCP发起RST报文*/ {
				goto ERROR
			} else {
				goto ERROR
			}
		ERROR:
			log.DefaultLogger.ErrorFields("unix.Read", zap.Error(err), zap.Int("epoll_fd", operator.Epoll.FD()), zap.Int("client_fd", operator.FD))
			break
		}
		offset += n
		if n == 0 /*在read进行中对端主动关闭连接调用了Close，TCP发起FIN报文*/ || offset == buffer.Block8k /*读取8k就走避免饥饿连接*/ {
			break
		}
	}
	tcpConn.readBuffer.Write(buf[:offset])
	log.DefaultLogger.InfoFields("read success", zap.Int("epoll_fd", operator.Epoll.FD()), zap.Int("client_fd", operator.FD), zap.ByteString("buffer", buf[:offset]))
}

func (t *serverTransportTCP) write(epoll *netpoll.Epoll, operator *netpoll.FDOperator) {
	tcpConn, ok := operator.Data.(*tcpConnection)
	if !ok || tcpConn == nil {
		log.DefaultLogger.ErrorFields("data is not tcpConnection", zap.Reflect("operator", operator))
		return
	}
	buf, err := tcpConn.writeBuffer.Peek(int(tcpConn.writeBuffer.Size()))
	if err != nil {
		// 数据发送完了返回
		return
	}

	offset := 0
	for {
		n, err := unix.Write(tcpConn.fd, buf[offset:])
		if err != nil {
			if err == unix.EINTR /*中断信号触发系统调用中断直接忽略继续读取*/ {
				continue
			} else if err == unix.EAGAIN || err == unix.EWOULDBLOCK /*非阻塞IO没有数据可读时直接返回等待OUT事件再次触发，不打印了不然日志太多*/ {
				break
			} else if err == unix.EBADF || err == unix.EINVAL /*fd被关闭已经是无效的文件描述符，在epoll事件模型中把HUP放最前面了，这里不会发生*/ {
				goto ERROR
			} else if err == unix.EPIPE /*broken pipe在write进行中对端意外关闭连接，TCP发起RST报文，触发SIGPIPE信号*/ {
				goto ERROR
			} else {
				goto ERROR
			}
		ERROR:
			log.DefaultLogger.ErrorFields("unix.Write", zap.Error(err), zap.Int("epoll_fd", operator.Epoll.FD()), zap.Int("client_fd", operator.FD))
			break
		}
		offset += n
		if offset == len(buf) /*buf全部写进去了*/ {
			break
		}
	}
	_ = tcpConn.writeBuffer.Skip(offset)
	tcpConn.writeBuffer.GC()
	log.DefaultLogger.InfoFields("write success", zap.Int("epoll_fd", operator.Epoll.FD()), zap.Int("client_fd", operator.FD), zap.ByteString("buffer", buf[:offset]))
}

func (t *serverTransportTCP) hup(epoll *netpoll.Epoll, operator *netpoll.FDOperator) {
	unix.Close(operator.FD)

	tcpConn, ok := operator.Data.(*tcpConnection)
	if !ok || tcpConn == nil {
		log.DefaultLogger.ErrorFields("data is not tcpConnection", zap.Reflect("operator", operator))
		return
	}
	tcpConn.readBuffer.Delete()
	tcpConn.writeBuffer.Delete()

	log.DefaultLogger.InfoFields("close success", zap.Int("epoll_fd", epoll.FD()), zap.Int("client_fd", operator.FD), zap.String("remote_address", tcpConn.remoteAddr.String()), zap.String("local_address", tcpConn.localAddr.String()))
}

func (t *serverTransportTCP) Close() {
	for _, epoll := range t.epolls {
		epoll.Close()
	}
}

type Operator interface {
	OnRequest(conn *tcpConnection)
}

type tcpConnection struct {
	fd          int
	localAddr   net.Addr
	remoteAddr  net.Addr
	readBuffer  *buffer.Buffer
	writeBuffer *buffer.Buffer
	Operator
}

func (conn *tcpConnection) LocalAddr() net.Addr {
	return conn.localAddr
}

func (conn *tcpConnection) RemoteAddr() net.Addr {
	return conn.remoteAddr
}

func (conn *tcpConnection) ReadBufferSize() uint64 {
	return conn.readBuffer.Size()
}

func (conn *tcpConnection) Peek(size int) ([]byte, error) {
	return conn.readBuffer.Peek(size)
}

func (conn *tcpConnection) Skip(size int) error {
	return conn.readBuffer.Skip(size)
}

func (conn *tcpConnection) Read(size int) ([]byte, error) {
	return conn.readBuffer.Read(size)
}

func (conn *tcpConnection) Write(buf []byte) {
	conn.writeBuffer.Write(buf)
}

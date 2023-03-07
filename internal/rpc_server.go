package internal

import (
	"github.com/soulnov23/go-tool/pkg/net"
	"go.uber.org/zap"
)

type RPCServer struct {
	CallLog *zap.Logger
	RunLog  *zap.SugaredLogger
}

func (svr *RPCServer) OnAccept(conn *net.TcpConn) {
	// TODO
}

func (svr *RPCServer) OnClose(conn *net.TcpConn) {
	// TODO
}

func (svr *RPCServer) OnRead(conn *net.TcpConn) {
}

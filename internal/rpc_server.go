package internal

import (
	"github.com/SoulNov23/go-tool/pkg/log"
	"github.com/SoulNov23/go-tool/pkg/net"
)

type RPCServer struct {
	CallLog log.Logger
	RunLog  log.Logger
}

func (svr *RPCServer) OnAccept(conn *net.TcpConn) {
	// TODO
}

func (svr *RPCServer) OnClose(conn *net.TcpConn) {
	// TODO
}

func (svr *RPCServer) OnRead(conn *net.TcpConn) {
	svr.RunLog.Debug("TODO")
}

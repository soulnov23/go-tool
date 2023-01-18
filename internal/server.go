package internal

import (
	"github.com/SoulNov23/go-tool/pkg/log"
	"github.com/SoulNov23/go-tool/pkg/net"
)

type Server struct {
	CallLog log.Logger
	RunLog  log.Logger
}

func (svr *Server) OnAccept(conn *net.TcpConn) {
	// TODO
}

func (svr *Server) OnClose(conn *net.TcpConn) {
	// TODO
}

func (svr *Server) OnRead(conn *net.TcpConn) {
	svr.RunLog.Debug("TODO")
}

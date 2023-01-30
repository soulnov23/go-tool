package internal

import (
	"github.com/SoulNov23/go-tool/pkg/log"
	"github.com/SoulNov23/go-tool/pkg/net"
)

type HTTPServer struct {
	CallLog log.Logger
	RunLog  log.Logger
}

func (svr *HTTPServer) OnAccept(conn *net.TcpConn) {
	// TODO
}

func (svr *HTTPServer) OnClose(conn *net.TcpConn) {
	// TODO
}

func (svr *HTTPServer) OnRead(conn *net.TcpConn) {
	svr.RunLog.Debug("TODO")
}

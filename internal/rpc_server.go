package internal

import (
	"github.com/google/uuid"
	"github.com/soulnov23/go-tool/pkg/net"
	"go.uber.org/zap"
)

type RPCServer struct {
	CallLog    *zap.Logger
	RunLog     *zap.SugaredLogger
	oldCallLog *zap.Logger
	oldRunLog  *zap.SugaredLogger
}

func (svr *RPCServer) OnAccept(conn *net.TcpConn) {
	// TODO
}

func (svr *RPCServer) OnClose(conn *net.TcpConn) {
	// TODO
}

func (svr *RPCServer) OnRead(conn *net.TcpConn) {
	svr.setLog()
	defer svr.resetLog()
	svr.RunLog.Debug("TODO")
}

func (svr *RPCServer) setLog() {
	svr.oldCallLog = svr.CallLog
	svr.oldRunLog = svr.RunLog
	uuid := uuid.New().String()
	svr.CallLog = svr.CallLog.With(zap.String("uuid", uuid))
	svr.RunLog = svr.RunLog.With(zap.String("uuid", uuid))
}

func (svr *RPCServer) resetLog() {
	svr.CallLog = svr.oldCallLog
	svr.RunLog = svr.oldRunLog
}

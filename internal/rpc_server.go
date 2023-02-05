package internal

import (
	"github.com/SoulNov23/go-tool/pkg/net"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RPCServer struct {
	CallLog    *zap.SugaredLogger
	RunLog     *zap.SugaredLogger
	oldCallLog *zap.SugaredLogger
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

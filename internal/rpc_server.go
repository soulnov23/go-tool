package internal

import (
	"time"

	"github.com/google/uuid"
	"github.com/soulnov23/go-tool/pkg/log"
	"github.com/soulnov23/go-tool/pkg/net"
	"github.com/soulnov23/go-tool/pkg/utils"
	"go.uber.org/zap"
)

type RPCServer struct {
	FrameLog log.Logger
	RunLog   log.Logger
}

func (svr *RPCServer) OnRead(conn *net.TcpConn) {
	bufferLen := conn.ReadBufferLen()
	buf, err := conn.Read(int(bufferLen))
	// read buffer没数据了
	if err != nil {
		return
	}
	svr.FrameLog.DebugFields("codec success",
		zap.String("remote_address", conn.RemoteAddr()),
		zap.String("local_address", conn.LocalAddr()),
		zap.ByteString("body", buf))

	traceId := uuid.New().String()
	log := svr.RunLog.With(zap.String("trace_id", traceId))
	defer log.Sync()

	response := "{\"ret_code\":0,\"msg\":\"ok\"}"

	begin := time.Now()
	time.Sleep(666 * time.Millisecond)
	timeUsed := time.Since(begin).Milliseconds()
	log.InfoFields("call",
		zap.String("remote_address", conn.RemoteAddr()),
		zap.String("local_address", conn.LocalAddr()),
		zap.ByteString("request", buf),
		zap.String("response", response),
		zap.Int64("time_used", timeUsed))

	log.DebugFields("handle begin", zap.ByteString("request", buf))
	log.DebugFields("handle end", zap.String("response", response))
	conn.Write(utils.String2Byte(response))
}

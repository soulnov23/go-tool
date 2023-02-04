package internal

import (
	"strconv"
	"strings"

	"github.com/SoulNov23/go-tool/pkg/log"
	"github.com/SoulNov23/go-tool/pkg/net"
	"github.com/SoulNov23/go-tool/pkg/unsafe"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type HTTPServer struct {
	CallLog    log.Logger
	RunLog     log.Logger
	oldCallLog log.Logger
	oldRunLog  log.Logger
}

func (svr *HTTPServer) OnAccept(conn *net.TcpConn) {
	// TODO
}

func (svr *HTTPServer) OnClose(conn *net.TcpConn) {
	// TODO
}

func (svr *HTTPServer) OnRead(conn *net.TcpConn) {
	svr.setLog()
	defer svr.resetLog()
	bufferLen := conn.ReadBufferLen()
	buf, err := conn.Read(int(bufferLen))
	// read buffer没数据了
	if err != nil {
		return
	}
	index := strings.Index(unsafe.Byte2String(buf), "\r\n\r\n")
	if index == -1 {
		svr.setBad(conn)
		return
	}
	sliceTemp := strings.Split(unsafe.Byte2String(buf[:index]), "\r\n")
	if len(sliceTemp) < 1 {
		svr.RunLog.Error("HTTP protocol format invalid")
		svr.setBad(conn)
		return
	}
	var method, uri, version string
	header := map[string]string{}
	for index, line := range sliceTemp {
		// request line
		if index == 0 {
			requestLine := strings.Split(line, " ")
			if len(requestLine) != 3 {
				svr.RunLog.Error("HTTP protocol format invalid")
				svr.setBad(conn)
				return
			}
			version = requestLine[2]
			if version != "HTTP/1.0" && version != "HTTP/1.1" {
				svr.RunLog.Error("HTTP version not support")
				svr.setBad(conn)
				return
			}
			method = requestLine[0]
			if method != "GET" && method != "POST" {
				svr.RunLog.Error("HTTP method not support")
				svr.setBad(conn)
				return
			}
			uri = requestLine[1]
			continue
		}
		// header
		sliceKV := strings.Split(line, ": ")
		if len(sliceKV) != 2 {
			svr.RunLog.Error("HTTP header not support")
			svr.setBad(conn)
			return
		}
		header[sliceKV[0]] = sliceKV[1]
	}
	svr.RunLog.Debugf("Version: %s", version)
	svr.RunLog.Debugf("Method: %s", method)
	svr.RunLog.Debugf("%s->%s%s", conn.RemoteAddr(), conn.LocalAddr(), uri)
	for key, value := range header {
		svr.RunLog.Debugf("%s: %s", key, value)
	}
	var body string
	if method == "POST" {
		strLength, ok := header["Content-Length"]
		if !ok {
			svr.RunLog.Error("HTTP body is empty")
			svr.setBad(conn)
			return
		}
		length, err := strconv.Atoi(strLength)
		if err != nil {
			svr.RunLog.Error("HTTP header Content-Length format invalid")
			svr.setBad(conn)
			return
		}
		if index+length > int(bufferLen) {
			svr.RunLog.Error("HTTP body is larger than 8k")
			svr.setBad(conn)
			return
		}
		body = unsafe.Byte2String(buf[index+8 : index+8+length])
		svr.RunLog.Debugf("Body: %s", body)
	}
	svr.setOK(conn)
}

func (svr *HTTPServer) setLog() {
	svr.oldCallLog = svr.CallLog
	svr.oldRunLog = svr.RunLog
	uuid := uuid.New().String()
	svr.CallLog = svr.CallLog.With(zap.String("uuid", uuid))
	svr.RunLog = svr.RunLog.With(zap.String("uuid", uuid))
}

func (svr *HTTPServer) resetLog() {
	svr.CallLog = svr.oldCallLog
	svr.RunLog = svr.oldRunLog
}

func (svr *HTTPServer) setOK(conn *net.TcpConn) {
	conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 42\r\n\r\n{\"msg\":\"ok\",\"need_resend\":\"false\",\"ret\":0}"))
}

func (svr *HTTPServer) setBad(conn *net.TcpConn) {
	conn.Write([]byte("HTTP/1.1 400 Bad Request\r\nContent-Length: 0\r\n\r\n"))
}

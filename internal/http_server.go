package internal

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/soulnov23/go-tool/pkg/json"
	"github.com/soulnov23/go-tool/pkg/net"
	"github.com/soulnov23/go-tool/pkg/utils"
	"go.uber.org/zap"
)

type HTTPServer struct {
	CallLog    *zap.Logger
	RunLog     *zap.SugaredLogger
	oldCallLog *zap.Logger
	oldRunLog  *zap.SugaredLogger
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
	buf, err := conn.Peek(int(bufferLen))
	// read buffer没数据了
	if err != nil {
		return
	}
	index := strings.Index(utils.Byte2String(buf), "\r\n\r\n")
	if index == -1 {
		svr.setBad(conn)
		return
	}
	sliceTemp := strings.Split(utils.Byte2String(buf[:index]), "\r\n")
	if len(sliceTemp) < 1 {
		svr.RunLog.Error("HTTP protocol format invalid")
		svr.setBad(conn)
		return
	}
	var method, url, version string
	header := map[string]string{}
	cookie := map[string]string{}
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
			url = requestLine[1]
			continue
		}
		// header
		headerSlice := strings.Split(line, ": ")
		if len(headerSlice) != 2 {
			svr.RunLog.Error("HTTP header not support")
			svr.setBad(conn)
			return
		}
		// cookie
		if headerSlice[0] == "Cookie" {
			cookie = utils.String2Map(headerSlice[1], "; ", "=")
		} else {
			header[headerSlice[0]] = headerSlice[1]
		}
	}
	var query, body string
	if method == "GET" {
		querySlice := strings.Split(url, "?")
		if len(querySlice) == 2 {
			url = querySlice[0]
			query = querySlice[1]
		}
	} else if method == "POST" {
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
			svr.RunLog.Debug("HTTP body not complete, continue")
			return
		}
		body = utils.Byte2String(buf[index+4 : index+4+length])
	}
	svr.RunLog.Debugf("Version: %s, Method: %s, %s->%s%s", version, method, conn.RemoteAddr(), conn.LocalAddr(), url)
	svr.RunLog.Debugf("Header: %s", json.Stringify(header))
	svr.RunLog.Debugf("Cookie: %s", json.Stringify(cookie))
	svr.RunLog.Debugf("Query: %s", query)
	svr.RunLog.Debugf("Body: %s", body)
	response, err := svr.Handler(conn, version, method, url, query, body, header, cookie)
	if err != nil {
		svr.RunLog.Errorf("svr.Handler: %s", err.Error())
		svr.setBad(conn)
		return
	}
	svr.setOK(conn, response)
}

func (svr *HTTPServer) Handler(conn *net.TcpConn, version, method, url, query, body string, header, cookie map[string]string) (string, error) {
	begin := time.Now()
	// TODO
	time.Sleep(666 * time.Millisecond)
	timeUsed := time.Since(begin).Milliseconds()
	response := "{\"msg\":\"ok\",\"need_resend\":\"false\",\"ret\":0}"
	if method == "GET" {
		svr.RunLog.Debugf("Request: %s", query)
		svr.RunLog.Debugf("Response: %s", response)
		svr.CallLog.Info("call",
			zap.String("RemoteAddr", conn.RemoteAddr()),
			zap.String("LocalAddr", conn.LocalAddr()),
			zap.String("HttpVersion", version),
			zap.String("HttpMethod", method),
			zap.String("HttpURL", url),
			zap.String("HttpHeaders", json.Stringify(header)),
			zap.String("HttpCookies", json.Stringify(cookie)),
			zap.String("HttpQuery", query),
			zap.String("Request", query),
			zap.String("Response", response),
			zap.Int64("TimeUsed", timeUsed))
	} else if method == "POST" {
		svr.RunLog.Debugf("Request: %s", body)
		svr.RunLog.Debugf("Response: %s", response)
		svr.CallLog.Info("call",
			zap.String("RemoteAddr", conn.RemoteAddr()),
			zap.String("LocalAddr", conn.LocalAddr()),
			zap.String("HttpVersion", version),
			zap.String("HttpMethod", method),
			zap.String("HttpURL", url),
			zap.String("HttpHeaders", json.Stringify(header)),
			zap.String("HttpCookies", json.Stringify(cookie)),
			zap.String("HttpQuery", query),
			zap.String("Request", body),
			zap.String("Response", response),
			zap.Int64("TimeUsed", timeUsed))
	}
	return response, nil
}

func (svr *HTTPServer) setLog() {
	svr.oldCallLog = svr.CallLog
	svr.oldRunLog = svr.RunLog
	uuid := uuid.New().String()
	svr.CallLog = svr.CallLog.With(zap.String("UUID", uuid))
	svr.RunLog = svr.RunLog.With(zap.String("UUID", uuid))
}

func (svr *HTTPServer) resetLog() {
	svr.CallLog = svr.oldCallLog
	svr.RunLog = svr.oldRunLog
}

func (svr *HTTPServer) setOK(conn *net.TcpConn, response string) {
	httpRsp := "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n"
	httpRsp += "Content-Length: " + strconv.Itoa(len(response)) + "\r\n\r\n" + response
	conn.Write(utils.String2Byte(httpRsp))
}

func (svr *HTTPServer) setBad(conn *net.TcpConn) {
	conn.Write(utils.String2Byte("HTTP/1.1 400 Bad Request\r\nContent-Length: 0\r\n\r\n"))
}

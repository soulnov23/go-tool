package internal

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/soulnov23/go-tool/pkg/log"
	"github.com/soulnov23/go-tool/pkg/net"
	"github.com/soulnov23/go-tool/pkg/utils"
	"go.uber.org/zap"
)

type HTTPServer struct {
	FrameLog log.Logger
	RunLog   log.Logger
}

func (svr *HTTPServer) OnRead(conn *net.TcpConn) {
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
		svr.FrameLog.ErrorFields("http protocol format invalid")
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
				svr.FrameLog.ErrorFields("http protocol format invalid")
				svr.setBad(conn)
				return
			}
			version = requestLine[2]
			if version != "HTTP/1.0" && version != "HTTP/1.1" {
				svr.FrameLog.ErrorFields("http version not support")
				svr.setBad(conn)
				return
			}
			method = requestLine[0]
			if method != "GET" && method != "POST" {
				svr.FrameLog.ErrorFields("http method not support")
				svr.setBad(conn)
				return
			}
			url = requestLine[1]
			continue
		}
		// header
		headerSlice := strings.Split(line, ": ")
		if len(headerSlice) != 2 {
			svr.FrameLog.ErrorFields("http header not support")
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
			svr.FrameLog.ErrorFields("http body is empty")
			svr.setBad(conn)
			return
		}
		length, err := strconv.Atoi(strLength)
		if err != nil {
			svr.FrameLog.ErrorFields("http header Content-Length format invalid")
			svr.setBad(conn)
			return
		}
		if index+length > int(bufferLen) {
			svr.FrameLog.DebugFields("http body not complete, continue")
			return
		}
		body = utils.Byte2String(buf[index+4 : index+4+length])
	}
	svr.FrameLog.DebugFields("codec success",
		zap.String("remote_address", conn.RemoteAddr()),
		zap.String("local_address", conn.LocalAddr()),
		zap.String("http_version", version),
		zap.String("http_method", method),
		zap.String("http_url", url),
		zap.Any("http_header", header),
		zap.Any("http_cookie", cookie),
		zap.String("http_query", query),
		zap.String("http_body", body))

	traceId := uuid.New().String()
	log := svr.RunLog.With(zap.String("trace_id", traceId))
	defer log.Sync()
	response, err := svr.handle(conn, version, method, url, query, body, header, cookie, log)
	if err != nil {
		svr.FrameLog.ErrorFields("http server handle", zap.Error(err))
		svr.setBad(conn)
		return
	}
	svr.setOK(conn, response)
}

func (svr *HTTPServer) handle(conn *net.TcpConn, version, method, url, query, body string, header, cookie map[string]string, log log.Logger) (string, error) {
	var request string
	if method == "GET" {
		request = query
	} else if method == "POST" {
		request = body
	}
	response := "{\"msg\":\"ok\",\"need_resend\":\"false\",\"ret\":0}"

	begin := time.Now()
	time.Sleep(666 * time.Millisecond)
	timeUsed := time.Since(begin).Milliseconds()
	log.InfoFields("call",
		zap.String("remote_address", conn.RemoteAddr()),
		zap.String("local_address", conn.LocalAddr()),
		zap.String("http_version", version),
		zap.String("http_method", method),
		zap.String("http_url", url),
		zap.Any("http_header", header),
		zap.Any("http_cookie", cookie),
		zap.String("http_query", query),
		zap.String("http_body", body),
		zap.String("request", request),
		zap.String("response", response),
		zap.Int64("time_used", timeUsed))

	log.DebugFields("handle begin", zap.String("request", request))
	log.DebugFields("handle end", zap.String("response", response))
	return response, nil
}

func (svr *HTTPServer) setOK(conn *net.TcpConn, response string) {
	httpRsp := "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n"
	httpRsp += "Content-Length: " + strconv.Itoa(len(response)) + "\r\n\r\n" + response
	conn.Write(utils.String2Byte(httpRsp))
}

func (svr *HTTPServer) setBad(conn *net.TcpConn) {
	conn.Write(utils.String2Byte("HTTP/1.1 400 Bad Request\r\nContent-Length: 0\r\n\r\n"))
}

package status

import (
	"net/http"

	"github.com/soulnov23/go-tool/pkg/json"
	gcode "google.golang.org/genproto/googleapis/rpc/code"
)

var (
	HTTPStatusCode = map[string]int{
		"OK":                  http.StatusOK,
		"INVALID_ARGUMENT":    http.StatusBadRequest,
		"FAILED_PRECONDITION": http.StatusBadRequest,
		"OUT_OF_RANGE":        http.StatusBadRequest,
		"UNAUTHENTICATED":     http.StatusUnauthorized,
		"PERMISSION_DENIED":   http.StatusForbidden,
		"NOT_FOUND":           http.StatusNotFound,
		"ABORTED":             http.StatusConflict,
		"ALREADY_EXISTS":      http.StatusConflict,
		"RESOURCE_EXHAUSTED":  http.StatusTooManyRequests,
		"CANCELLED":           499,
		"DATA_LOSS":           http.StatusInternalServerError,
		"UNKNOWN":             http.StatusInternalServerError,
		"INTERNAL":            http.StatusInternalServerError,
		"UNIMPLEMENTED":       http.StatusNotImplemented,
		"UNAVAILABLE":         http.StatusServiceUnavailable,
		"DEADLINE_EXCEEDED":   http.StatusGatewayTimeout,
	}
)

type Status struct {
	Name string `json:"name"`
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

// 200 OK
var New = func() *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_OK)],
		Code: 0,
		Msg:  "ok",
	}
}

// 400 Bad Request
func NewInvalidArgument(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_INVALID_ARGUMENT)],
		Code: code,
		Msg:  msg,
	}
}

// 400 Bad Request
func NewFailedPrecondition(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_FAILED_PRECONDITION)],
		Code: code,
		Msg:  msg,
	}
}

// 400 Bad Request
func NewOutOfRange(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_OUT_OF_RANGE)],
		Code: code,
		Msg:  msg,
	}
}

// 401 Unauthorized
func NewUnauthenticated(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_UNAUTHENTICATED)],
		Code: code,
		Msg:  msg,
	}
}

// 403 Forbidden
func NewPermissionDenied(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_PERMISSION_DENIED)],
		Code: code,
		Msg:  msg,
	}
}

// 404 Not Found
func NewNotFound(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_NOT_FOUND)],
		Code: code,
		Msg:  msg,
	}
}

// 409 Conflict
func NewAborted(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_ABORTED)],
		Code: code,
		Msg:  msg,
	}
}

// 409 Conflict
func NewAlreadyExists(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_ALREADY_EXISTS)],
		Code: code,
		Msg:  msg,
	}
}

// 429 Too Many Requests
func NewResourceExhausted(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_RESOURCE_EXHAUSTED)],
		Code: code,
		Msg:  msg,
	}
}

// 499 Client Closed Request
func NewCancelled(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_CANCELLED)],
		Code: code,
		Msg:  msg,
	}
}

// 500 Internal Server Error
func NewDataLoss(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_DATA_LOSS)],
		Code: code,
		Msg:  msg,
	}
}

// 500 Internal Server Error
func NewUnknown(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_UNKNOWN)],
		Code: code,
		Msg:  msg,
	}
}

// 500 Internal Server Error
func NewInternal(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_INTERNAL)],
		Code: code,
		Msg:  msg,
	}
}

// 501 Not Implemented
func NewUnimplemented(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_UNIMPLEMENTED)],
		Code: code,
		Msg:  msg,
	}
}

// 503 Service Unavailable
func NewUnavailable(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_UNAVAILABLE)],
		Code: code,
		Msg:  msg,
	}
}

// 504 Gateway Timeout
func NewDeadlineExceeded(code int32, msg string) *Status {
	return &Status{
		Name: gcode.Code_name[int32(gcode.Code_DEADLINE_EXCEEDED)],
		Code: code,
		Msg:  msg,
	}
}

func (s *Status) OK() bool {
	if s == nil {
		return true
	}
	return s.Code == 0
}

func (s *Status) Error() string {
	return json.Stringify(s)
}

func (s *Status) HTTPCode() int {
	if s == nil {
		return http.StatusOK
	}
	return HTTPStatusCode[s.Name]
}

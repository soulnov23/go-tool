package errors

import (
	"fmt"
	"net/http"

	"github.com/soulnov23/go-tool/pkg/json/jsoniter"
)

//go:generate protoc --proto_path=. --go_out=paths=source_relative:. wrap_error.proto

/*
1xx: Informational - Request received, continuing process
2xx: Success - The action was successfully received, understood, and accepted
3xx: Redirection - Further action must be taken in order to complete the request
4xx: Client Error - The request contains bad syntax or cannot be fulfilled
5xx: Server Error - The server failed to fulfill an apparently valid request
*/

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return jsoniter.Stringify(e)
}

func (e *Error) OK() bool {
	if e == nil {
		return true
	}
	return e.Code < 300
}

// nil
var New = func() *Error {
	return &Error{}
}

func Parse(err string) *Error {
	e := New()
	if errr := jsoniter.UnmarshalFromString(err, e); errr != nil {
		return nil
	}
	return e
}

func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if innerErr, ok := err.(*Error); ok && innerErr != nil {
		return innerErr
	}
	return Parse(err.Error())
}

func Equal(err1 error, err2 error) bool {
	verr1, ok1 := err1.(*Error)
	verr2, ok2 := err2.(*Error)

	if ok1 != ok2 {
		return false
	}

	if !ok1 {
		return err1 == err2
	}

	if verr1.Code != verr2.Code {
		return false
	}

	if verr1.Name != verr2.Name {
		return false
	}

	return true
}

// 100 Continue
func NewContinue() *Error {
	return &Error{
		Code:   http.StatusContinue,
		Status: http.StatusText(http.StatusContinue),
	}
}

// 200 OK
func NewOK() *Error {
	return &Error{
		Code:   http.StatusOK,
		Status: http.StatusText(http.StatusOK),
	}
}

// 201 Created
func NewCreated() *Error {
	return &Error{
		Code:   http.StatusCreated,
		Status: http.StatusText(http.StatusCreated),
	}
}

// 204 No Content
func NewNoContent() *Error {
	return &Error{
		Code:   http.StatusNoContent,
		Status: http.StatusText(http.StatusNoContent),
	}
}

// 300 Multiple Choices
func NewMultipleChoices(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusMultipleChoices,
		Status:  http.StatusText(http.StatusMultipleChoices),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

// 301 Moved Permanently
func NewMovedPermanently(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusMovedPermanently,
		Status:  http.StatusText(http.StatusMovedPermanently),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

// 302 Found
func NewFound(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusFound,
		Status:  http.StatusText(http.StatusFound),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

// 400 Bad Request
func NewBadRequest(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusBadRequest,
		Status:  http.StatusText(http.StatusBadRequest),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

// 401 Unauthorized
func NewUnauthorized(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusUnauthorized,
		Status:  http.StatusText(http.StatusUnauthorized),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

// 403 Forbidden
func NewForbidden(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusForbidden,
		Status:  http.StatusText(http.StatusForbidden),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

// 404 Not Found
func NewNotFound(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusNotFound,
		Status:  http.StatusText(http.StatusNotFound),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

// 405 Method Not Allowed
func NewMethodNotAllowed(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusMethodNotAllowed,
		Status:  http.StatusText(http.StatusMethodNotAllowed),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

// 408 Request Timeout
func NewRequestTimeout(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusRequestTimeout,
		Status:  http.StatusText(http.StatusRequestTimeout),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

// 409 Conflict
func NewConflict(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusConflict,
		Status:  http.StatusText(http.StatusConflict),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

// 500 Internal Server Error
func NewInternalServerError(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusInternalServerError,
		Status:  http.StatusText(http.StatusInternalServerError),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

// 501 Not Implemented
func NewNotImplemented(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusNotImplemented,
		Status:  http.StatusText(http.StatusNotImplemented),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

// 502 Bad Gateway
func NewBadGateway(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusBadGateway,
		Status:  http.StatusText(http.StatusBadGateway),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

// 503 Service Unavailable
func NewServiceUnavailable(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusServiceUnavailable,
		Status:  http.StatusText(http.StatusServiceUnavailable),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

// 504 Gateway Timeout
func NewGatewayTimeout(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusGatewayTimeout,
		Status:  http.StatusText(http.StatusGatewayTimeout),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

// 504 HTTP Version Not Supported
func NewHTTPVersionNotSupported(name string, formatter string, args ...any) *Error {
	return &Error{
		Code:    http.StatusHTTPVersionNotSupported,
		Status:  http.StatusText(http.StatusHTTPVersionNotSupported),
		Name:    name,
		Message: fmt.Sprintf(formatter, args...),
	}
}

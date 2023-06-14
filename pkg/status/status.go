package status

import (
	"net/http"

	"github.com/soulnov23/go-tool/pkg/json"
	gcode "google.golang.org/genproto/googleapis/rpc/code"
	gstatus "google.golang.org/genproto/googleapis/rpc/status"
)

const (
	HTTP_CODE_RANGE = 1_000 * 1_000
	RPC_CODE_RANGE  = 1_000
	MSG_RANGE       = ": "
)

// Status.Code ${http_code}+${rcp_code}+${code}
type Status struct {
	Status *gstatus.Status `json:"status"`
}

func NewOk(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusOK*HTTP_CODE_RANGE + int32(gcode.Code_OK)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusOK) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_OK)] + MSG_RANGE + message,
		},
	}
}

func NewCancelled(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    499*HTTP_CODE_RANGE + int32(gcode.Code_CANCELLED)*RPC_CODE_RANGE + code,
			Message: "Client Closed Request" + MSG_RANGE + gcode.Code_name[int32(gcode.Code_CANCELLED)] + MSG_RANGE + message,
		},
	}
}

func NewUnknown(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusInternalServerError*HTTP_CODE_RANGE + int32(gcode.Code_UNKNOWN)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusInternalServerError) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_UNKNOWN)] + MSG_RANGE + message,
		},
	}
}

func NewInvalidArgument(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusBadRequest*HTTP_CODE_RANGE + int32(gcode.Code_INVALID_ARGUMENT)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusBadRequest) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_INVALID_ARGUMENT)] + MSG_RANGE + message,
		},
	}
}

func NewDeadlineExceeded(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusGatewayTimeout*HTTP_CODE_RANGE + int32(gcode.Code_DEADLINE_EXCEEDED)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusGatewayTimeout) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_DEADLINE_EXCEEDED)] + MSG_RANGE + message,
		},
	}
}

func NewNotFound(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusNotFound*HTTP_CODE_RANGE + int32(gcode.Code_NOT_FOUND)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusNotFound) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_NOT_FOUND)] + MSG_RANGE + message,
		},
	}
}

func NewAlreadyExists(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusConflict*HTTP_CODE_RANGE + int32(gcode.Code_ALREADY_EXISTS)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusConflict) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_ALREADY_EXISTS)] + MSG_RANGE + message,
		},
	}
}

func NewPermissionDenied(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusForbidden*HTTP_CODE_RANGE + int32(gcode.Code_PERMISSION_DENIED)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusForbidden) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_PERMISSION_DENIED)] + MSG_RANGE + message,
		},
	}
}

func NewUnauthenticated(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusUnauthorized*HTTP_CODE_RANGE + int32(gcode.Code_UNAUTHENTICATED)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusUnauthorized) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_UNAUTHENTICATED)] + MSG_RANGE + message,
		},
	}
}

func NewResourceExhausted(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusTooManyRequests*HTTP_CODE_RANGE + int32(gcode.Code_RESOURCE_EXHAUSTED)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusTooManyRequests) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_RESOURCE_EXHAUSTED)] + MSG_RANGE + message,
		},
	}
}

func NewFailedPrecondition(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusBadRequest*HTTP_CODE_RANGE + int32(gcode.Code_FAILED_PRECONDITION)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusBadRequest) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_FAILED_PRECONDITION)] + MSG_RANGE + message,
		},
	}
}

func NewAborted(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusConflict*HTTP_CODE_RANGE + int32(gcode.Code_ABORTED)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusConflict) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_ABORTED)] + MSG_RANGE + message,
		},
	}
}

func NewOutOfRange(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusBadRequest*HTTP_CODE_RANGE + int32(gcode.Code_OUT_OF_RANGE)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusBadRequest) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_OUT_OF_RANGE)] + MSG_RANGE + message,
		},
	}
}

func NewUnimplemented(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusNotImplemented*HTTP_CODE_RANGE + int32(gcode.Code_UNIMPLEMENTED)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusNotImplemented) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_UNIMPLEMENTED)] + MSG_RANGE + message,
		},
	}
}

func NewInternal(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusInternalServerError*HTTP_CODE_RANGE + int32(gcode.Code_INTERNAL)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusInternalServerError) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_INTERNAL)] + MSG_RANGE + message,
		},
	}
}

func NewUnavailable(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusServiceUnavailable*HTTP_CODE_RANGE + int32(gcode.Code_UNAVAILABLE)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusServiceUnavailable) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_UNAVAILABLE)] + MSG_RANGE + message,
		},
	}
}

func NewDataLoss(code int32, message string) *Status {
	return &Status{
		Status: &gstatus.Status{
			Code:    http.StatusInternalServerError*HTTP_CODE_RANGE + int32(gcode.Code_DATA_LOSS)*RPC_CODE_RANGE + code,
			Message: http.StatusText(http.StatusInternalServerError) + MSG_RANGE + gcode.Code_name[int32(gcode.Code_DATA_LOSS)] + MSG_RANGE + message,
		},
	}
}

func (s *Status) Error() string {
	return json.Stringify(s.Status)
}

func (s *Status) HttpCode() int32 {
	if s == nil || s.Status == nil {
		return http.StatusOK
	}
	return s.Status.Code / HTTP_CODE_RANGE
}

func (s *Status) RpcCode() int32 {
	if s == nil || s.Status == nil {
		return int32(gcode.Code_OK)
	}
	return (s.Status.Code / RPC_CODE_RANGE) % RPC_CODE_RANGE
}

func (s *Status) Code() int32 {
	if s == nil || s.Status == nil {
		return int32(gcode.Code_OK)
	}
	return s.Status.Code % RPC_CODE_RANGE
}

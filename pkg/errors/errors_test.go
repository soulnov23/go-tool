package errors

import (
	"errors"
	"testing"

	"github.com/soulnov23/go-tool/pkg/json"
)

func Test(t *testing.T) {
	err := NewInternalServerError("NOT_FOUND_USER", "not found user")
	t.Log(err.Code, err.Status, err.Name, err.Msg, err.Error())
	if err.OK() {
		t.Log("OK")
	}

	tempErr := errors.New("not found user")
	err = FromError(tempErr)
	if err.OK() {
		t.Log("OK") // err解析失败，不可用
	}

	tempErr = errors.New(`{"code":500,"status":"Internal Server Error","name":"NOT_FOUND_USER","msg":"not found user"}`)
	err = FromError(tempErr)
	if err.OK() {
		t.Log("OK")
	}

	temp := json.Stringify(New())
	t.Log(temp)

	temp = json.Stringify(nil)
	t.Log(temp)

	t.Log(123456789 / 1_000_000)
	t.Log((123456789 / 1_000) % 1_000)
	t.Log(123456789 % 1_000)
}
